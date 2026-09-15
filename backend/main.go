package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	serviceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
	kubernetesService  = "https://kubernetes.default.svc"
)

type Checker struct {
	client           *http.Client
	kubernetesClient *http.Client
	frontendURL      string
	namespace        string
	token            string
}

type StatusResponse struct {
	API     string `json:"api"`
	Network string `json:"network"`
	Message string `json:"message"`
}

type NetworkPolicyList struct {
	Items []NetworkPolicy `json:"items"`
}

type NetworkPolicy struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec NetworkPolicySpec `json:"spec"`
}

type NetworkPolicySpec struct {
	PodSelector struct {
		MatchLabels map[string]string `json:"matchLabels"`
	} `json:"podSelector"`
	PolicyTypes []string               `json:"policyTypes"`
	Ingress     []NetworkPolicyIngress `json:"ingress"`
}

type NetworkPolicyIngress struct {
	From  []NetworkPolicyPeer `json:"from"`
	Ports []NetworkPolicyPort `json:"ports"`
}

type NetworkPolicyPeer struct {
	PodSelector struct {
		MatchLabels map[string]string `json:"matchLabels"`
	} `json:"podSelector"`
}

type NetworkPolicyPort struct {
	Protocol string          `json:"protocol"`
	Port     json.RawMessage `json:"port"`
}

func (c *Checker) checkFrontend() error {
	response, err := c.client.Get(c.frontendURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("frontend returned HTTP %d", response.StatusCode)
	}

	return nil
}

func (c *Checker) missingNetworkPolicies(ctx context.Context) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/apis/networking.k8s.io/v1/namespaces/%s/networkpolicies", kubernetesService, c.namespace), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)

	response, err := c.kubernetesClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Kubernetes API returned HTTP %d", response.StatusCode)
	}

	var policies NetworkPolicyList
	if err := json.NewDecoder(response.Body).Decode(&policies); err != nil {
		return nil, err
	}

	return evaluateNetworkPolicies(policies.Items), nil
}

func evaluateNetworkPolicies(policies []NetworkPolicy) []string {
	required := map[string]bool{
		"default-deny-ingress":      false,
		"allow-client-to-frontend":  false,
		"allow-frontend-to-backend": false,
		"allow-backend-to-frontend": false,
	}
	for _, policy := range policies {
		switch policy.Metadata.Name {
		case "default-deny-ingress":
			required[policy.Metadata.Name] = isDefaultDenyIngress(policy)
		case "allow-client-to-frontend":
			required[policy.Metadata.Name] = allowsIngressFromAnySource(policy, "frontend", 80)
		case "allow-frontend-to-backend":
			required[policy.Metadata.Name] = allowsIngress(policy, "backend", "frontend", 8000)
		case "allow-backend-to-frontend":
			required[policy.Metadata.Name] = allowsIngress(policy, "frontend", "backend", 80)
		}
	}

	missing := make([]string, 0, len(required))
	for name, present := range required {
		if !present {
			missing = append(missing, name)
		}
	}
	return missing
}

func isDefaultDenyIngress(policy NetworkPolicy) bool {
	return len(policy.Spec.PodSelector.MatchLabels) == 0 && len(policy.Spec.Ingress) == 0 && contains(policy.Spec.PolicyTypes, "Ingress")
}

func allowsIngress(policy NetworkPolicy, destination, source string, port int) bool {
	if policy.Spec.PodSelector.MatchLabels["app"] != destination || !contains(policy.Spec.PolicyTypes, "Ingress") {
		return false
	}
	for _, rule := range policy.Spec.Ingress {
		for _, peer := range rule.From {
			if peer.PodSelector.MatchLabels["app"] != source {
				continue
			}
			for _, allowedPort := range rule.Ports {
				if allowedPort.Protocol == "TCP" && string(allowedPort.Port) == fmt.Sprintf("%d", port) {
					return true
				}
			}
		}
	}
	return false
}

func allowsIngressFromAnySource(policy NetworkPolicy, destination string, port int) bool {
	if policy.Spec.PodSelector.MatchLabels["app"] != destination || !contains(policy.Spec.PolicyTypes, "Ingress") {
		return false
	}
	for _, rule := range policy.Spec.Ingress {
		if len(rule.From) != 0 {
			continue
		}
		for _, allowedPort := range rule.Ports {
			if allowedPort.Protocol == "TCP" && string(allowedPort.Port) == fmt.Sprintf("%d", port) {
				return true
			}
		}
	}
	return false
}

func contains(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

func (c *Checker) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := c.checkFrontend(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (c *Checker) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := c.checkFrontend(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)

		json.NewEncoder(w).Encode(StatusResponse{
			API:     "ok",
			Network: "error",
			Message: "Backend cannot reach the frontend service.",
		})
		return
	}

	missing, err := c.missingNetworkPolicies(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StatusResponse{
			API:     "ok",
			Network: "error",
			Message: "Backend cannot verify Kubernetes network policies: " + err.Error(),
		})
		return
	}
	if len(missing) > 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(StatusResponse{
			API:     "ok",
			Network: "unsafe",
			Message: "Network policies are incomplete. Missing: " + strings.Join(missing, ", ") + ".",
		})
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(StatusResponse{
		API:     "ok",
		Network: "ok",
		Message: "Backend can reach the frontend service and required network policies are active.",
	})
}

func main() {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://frontend-service"
	}
	namespace, err := os.ReadFile(filepath.Join(serviceAccountPath, "namespace"))
	if err != nil {
		panic(fmt.Errorf("read Kubernetes namespace: %w", err))
	}
	token, err := os.ReadFile(filepath.Join(serviceAccountPath, "token"))
	if err != nil {
		panic(fmt.Errorf("read Kubernetes service account token: %w", err))
	}
	certificateAuthority, err := os.ReadFile(filepath.Join(serviceAccountPath, "ca.crt"))
	if err != nil {
		panic(fmt.Errorf("read Kubernetes service account CA certificate: %w", err))
	}
	certificatePool := x509.NewCertPool()
	if !certificatePool.AppendCertsFromPEM(certificateAuthority) {
		panic("parse Kubernetes service account CA certificate")
	}

	checker := &Checker{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		kubernetesClient: &http.Client{
			Timeout: 2 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: certificatePool},
			},
		},
		frontendURL: frontendURL,
		namespace:   strings.TrimSpace(string(namespace)),
		token:       strings.TrimSpace(string(token)),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", checker.readinessHandler)
	mux.HandleFunc("GET /api/status", checker.statusHandler)

	server := &http.Server{
		Addr:              ":8000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Backend listening on http://localhost:8000")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

package main

import (
	"encoding/json"
	"testing"
)

func TestEvaluateNetworkPolicies(t *testing.T) {
	var policies []NetworkPolicy
	if err := json.Unmarshal([]byte(`[
    {"metadata":{"name":"default-deny-ingress"},"spec":{"podSelector":{},"policyTypes":["Ingress"]}},
    {"metadata":{"name":"allow-client-to-frontend"},"spec":{"podSelector":{"matchLabels":{"app":"frontend"}},"policyTypes":["Ingress"],"ingress":[{"ports":[{"protocol":"TCP","port":80}]}]}},
    {"metadata":{"name":"allow-frontend-to-backend"},"spec":{"podSelector":{"matchLabels":{"app":"backend"}},"policyTypes":["Ingress"],"ingress":[{"from":[{"podSelector":{"matchLabels":{"app":"frontend"}}}],"ports":[{"protocol":"TCP","port":8000}]}]}},
    {"metadata":{"name":"allow-backend-to-frontend"},"spec":{"podSelector":{"matchLabels":{"app":"frontend"}},"policyTypes":["Ingress"],"ingress":[{"from":[{"podSelector":{"matchLabels":{"app":"backend"}}}],"ports":[{"protocol":"TCP","port":80}]}]}}
  ]`), &policies); err != nil {
		t.Fatal(err)
	}

	if missing := evaluateNetworkPolicies(policies); len(missing) != 0 {
		t.Fatalf("expected no missing policies, got %v", missing)
	}
}

func TestEvaluateNetworkPoliciesRejectsWrongPort(t *testing.T) {
	var policies []NetworkPolicy
	if err := json.Unmarshal([]byte(`[
    {"metadata":{"name":"default-deny-ingress"},"spec":{"podSelector":{},"policyTypes":["Ingress"]}},
    {"metadata":{"name":"allow-client-to-frontend"},"spec":{"podSelector":{"matchLabels":{"app":"frontend"}},"policyTypes":["Ingress"],"ingress":[{"ports":[{"protocol":"TCP","port":80}]}]}},
    {"metadata":{"name":"allow-frontend-to-backend"},"spec":{"podSelector":{"matchLabels":{"app":"backend"}},"policyTypes":["Ingress"],"ingress":[{"from":[{"podSelector":{"matchLabels":{"app":"frontend"}}}],"ports":[{"protocol":"TCP","port":80}]}]}},
    {"metadata":{"name":"allow-backend-to-frontend"},"spec":{"podSelector":{"matchLabels":{"app":"frontend"}},"policyTypes":["Ingress"],"ingress":[{"from":[{"podSelector":{"matchLabels":{"app":"backend"}}}],"ports":[{"protocol":"TCP","port":80}]}]}}
  ]`), &policies); err != nil {
		t.Fatal(err)
	}

	missing := evaluateNetworkPolicies(policies)
	if len(missing) != 1 || missing[0] != "allow-frontend-to-backend" {
		t.Fatalf("expected frontend-to-backend policy to be missing, got %v", missing)
	}
}

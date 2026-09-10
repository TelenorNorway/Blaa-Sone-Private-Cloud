# Blue Zone: Workshop Instructions

## Before you start

Bring your laptop, curiosity, and a willingness to investigate. You will deploy and troubleshoot a small application in Kubernetes.

### Software requirements

You'll need **Docker** and **Kubernetes** to complete this workshop. Choose the options that work for your operating system.

#### macOS

Choose one container runtime:
- **[Docker Desktop](https://www.docker.com/products/docker-desktop/)** – Official Docker implementation for macOS
- **[Colima](https://github.com/abiosoft/colima)** – Lightweight Docker alternative using Lima

#### Windows

We recommend **[WSL 2](https://learn.microsoft.com/en-us/windows/wsl/)** (Windows Subsystem for Linux):
- [Installation guide](https://learn.microsoft.com/en-us/windows/wsl/install)
- Then install Docker Desktop or one of the Kubernetes options below

#### Linux

Follow the documentation for your distribution on the upstream project sites.

#### Kubernetes cluster

Choose one option to run Kubernetes locally:

| Option | Link | Best for |
| --- | --- | --- |
| **Docker Desktop** | [docker.com](https://www.docker.com/products/docker-desktop/) | Integrated with Docker, easiest setup |
| **Minikube** | [minikube.sigs.k8s.io](https://minikube.sigs.k8s.io/docs/start/) | Testing and learning, supports multiple drivers |
| **K3s** | [k3s.io](https://k3s.io/) | Lightweight, minimal resource usage |

## The challenge

Make the application status page report a healthy private cloud.

The application has two parts:

1. The **frontend** is served by NGINX. Requests to `/api/` are proxied to the backend Kubernetes service.
2. The **backend** is a Go service. Its `/api/status` endpoint checks whether it can reach the frontend Kubernetes service.

The frontend and backend are deliberately connected through Kubernetes networking. Inspect the configuration before changing anything.

## Suggested path

1. Build the frontend and backend container images from `frontend/` and `backend/`.
2. Deploy the Kubernetes manifests in `examples/k8s/`.
3. Make the frontend available locally using your preferred Kubernetes access method.
4. Open the status page and observe the result.
5. Use Kubernetes to find and fix the issue that prevents the services from communicating.
6. Refresh the page. You are done when both the API and internal network are healthy.

## Useful checkpoints

Start broad, then narrow the search.

```bash
kubectl get pods,services
kubectl get deployments
kubectl describe pod <pod-name>
kubectl logs <pod-name>
kubectl get endpoints
```

Look for:

- Pods that are not ready or repeatedly restarting
- Service selectors that do not match pod labels
- Incorrect service names or ports
- Failed readiness or liveness probes
- DNS or network errors in the logs

## Application endpoints

| Endpoint      | Purpose                                                   |
| ------------- | --------------------------------------------------------- |
| `/`           | Frontend status page                                      |
| `/api/status` | Backend status check, accessed through the frontend proxy |
| `/healthz`    | Backend liveness endpoint                                 |
| `/readyz`     | Backend readiness endpoint                                |

## Completion check

You have reached the zone when the page reports:

```text
Backend API: OK
Internal network: Connected
Cluster status: Healthy
```

Explain the root cause to your team. A working deployment is good; understanding why it works is better.

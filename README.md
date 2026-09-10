# Blue Zone: Private Cloud

## Ready to enter the zone?

This is a hands-on Kubernetes challenge for people who want to understand more than the theory.

You will work with a small private-cloud application made up of a frontend, a backend, and the network between them. Your job is not to build everything from scratch. Your job is to look closely, ask good questions, and make the services talk to each other.

No plastic swag. No pretend infrastructure. Just a realistic problem, a team, and the tools to solve it.

## Your mission

Get the status page to report that the cluster is healthy.

The frontend calls the backend through Kubernetes. The backend then checks whether it can reach the frontend service over the internal cluster network. When the page shows a healthy API and network, you have completed the challenge.

## Inside the repository

| Path              | What you will find                                                                 |
| ----------------- | ---------------------------------------------------------------------------------- |
| `frontend/`       | A static status page served by NGINX. It proxies API calls to the backend service. |
| `backend/`        | A small Go service that exposes health, readiness, and status endpoints.           |
| `INSTRUCTIONS.md` | Workshop setup, requirements, and the path through the challenge.                  |

## What good looks like

The status page shows:

- **Backend API:** OK
- **Internal network:** Connected
- **Cluster status:** Healthy

Start with the pods. Then follow the services, ports, probes, logs, and DNS. The answer is in the cluster.

```bash
kubectl get pods,services
```

## Bring curiosity

This challenge is for people who want to explore technology that matters, understand how systems work in practice, and leave with more questions than they arrived with.

Open [`INSTRUCTIONS.md`](INSTRUCTIONS.md) when you are ready to begin.

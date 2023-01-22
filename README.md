# Argo CD Application Updates

A small application to helm visualizing if there are application in ArgoCD that have update of their Helm Chart.

## Run

```bash
docker run --rm -p 8080:8080 \
  ghcr.io/bakito/argocd-app-updates:latest \
  --argo-server https://argocd.foo.com
```

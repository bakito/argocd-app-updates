# Argo CD Application Updates

A small application to helm visualizing if there are application in ArgoCD that have update of their Helm Chart.

## Run

```bash
docker run --rm -p 8080:8080 \
  ghcr.io/bakito/argocd-app-updates:latest \
  --argo-server https://argocd.foo.com
```

## Setup user

### Create an Account

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
data:
  accounts.argocd-updates: apiKey
```

### Add RBAC

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
data:
  policy.csv: |
    g, argocd-updates, role:readonly
```

### Generate a Token

```bash
argocd account generate-token -a argocd-updates --core
```

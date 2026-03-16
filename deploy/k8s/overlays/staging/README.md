# Staging Overlay

This overlay is meant for shared testing or pre-release environments.

Compared with production, the defaults are lighter:

- smaller replica counts
- mock mail provider by default
- mock SMS provider by default
- separate staging hostname and namespace

## First files to update

- `app-configmap-patch.yaml`
- `app-secret-patch.yaml`
- `ingress-patch.yaml`
- `kustomization.yaml`

## Render or apply

```powershell
kubectl kustomize deploy/k8s/overlays/staging
```

First deployment:

```powershell
kubectl apply -k deploy/k8s/overlays/staging/bootstrap
kubectl wait --for=condition=complete job/backend-bootstrap -n nex-backend-staging --timeout=300s
kubectl apply -k deploy/k8s/overlays/staging/runtime
```

## Recommended tags

Use the `sha-<commit>` image tags from `.github/workflows/docker-images.yml` so staging can validate the exact build before production.

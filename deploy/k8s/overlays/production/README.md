# Production Overlay

This overlay keeps the base manifests reusable while moving production-specific changes into one place.

Files you should edit first:

- `app-configmap-patch.yaml`
- `app-secret-patch.yaml`
- `ingress-patch.yaml`
- `kustomization.yaml`

## What this overlay controls

- production hostnames
- image tags
- replica counts
- external MySQL / Redis / object storage addresses
- JWT and SMTP secrets

## Recommended image tag strategy

Use the `sha-<commit>` tags produced by `.github/workflows/docker-images.yml`.

Example:

```yaml
images:
  - name: ghcr.io/gee-coder/template-go-backend-public-api
    newTag: sha-0123456789abcdef
```

This makes rollback and release tracking much easier than `latest`.

## Render or apply

Render the full production manifests locally:

```powershell
kubectl kustomize deploy/k8s/overlays/production
```

For the first deployment, apply bootstrap and runtime separately:

```powershell
kubectl apply -k deploy/k8s/overlays/production/bootstrap
kubectl wait --for=condition=complete job/backend-bootstrap -n nex-backend --timeout=300s
kubectl apply -k deploy/k8s/overlays/production/runtime
kubectl rollout status deployment/public-api -n nex-backend
kubectl rollout status deployment/auth-api -n nex-backend
kubectl rollout status deployment/system-api -n nex-backend
kubectl rollout status deployment/backend-gateway -n nex-backend
```

After the cluster is already initialized, you can still render or re-apply the combined overlay from `deploy/k8s/overlays/production`.

## Notes

- The production overlay keeps `bootstrap` as a Job, so schema setup stays separate from long-running services.
- If you move from MinIO to `aliyun_oss` or `huawei_obs`, switch the provider in `app-configmap-patch.yaml` and complete the matching credentials in `app-secret-patch.yaml`.

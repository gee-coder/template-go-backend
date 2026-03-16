# Release And Rollback

This guide standardizes how to move a backend build from `staging` to `production` using the `sha-<commit>` image tags published by `.github/workflows/docker-images.yml`.

## Release flow

1. Push code to `master`

   The workflow publishes these images:

   - `ghcr.io/gee-coder/template-go-backend-bootstrap:sha-<commit>`
   - `ghcr.io/gee-coder/template-go-backend-public-api:sha-<commit>`
   - `ghcr.io/gee-coder/template-go-backend-auth-api:sha-<commit>`
   - `ghcr.io/gee-coder/template-go-backend-system-api:sha-<commit>`

2. Point `staging` at the exact build

   ```powershell
   powershell -ExecutionPolicy Bypass -File .\scripts\set-k8s-image-tags.ps1 -Overlay staging -Tag sha-<commit>
   kubectl apply -k deploy/k8s/overlays/staging/bootstrap
   kubectl wait --for=condition=complete job/backend-bootstrap -n nex-backend-staging --timeout=300s
   kubectl apply -k deploy/k8s/overlays/staging/runtime
   kubectl rollout status deployment/public-api -n nex-backend-staging
   kubectl rollout status deployment/auth-api -n nex-backend-staging
   kubectl rollout status deployment/system-api -n nex-backend-staging
   kubectl rollout status deployment/backend-gateway -n nex-backend-staging
   ```

3. After staging validation passes, promote the same tags to `production`

   ```powershell
   powershell -ExecutionPolicy Bypass -File .\scripts\set-k8s-image-tags.ps1 -Overlay production -SourceOverlay staging
   kubectl apply -k deploy/k8s/overlays/production/bootstrap
   kubectl wait --for=condition=complete job/backend-bootstrap -n nex-backend --timeout=300s
   kubectl apply -k deploy/k8s/overlays/production/runtime
   kubectl rollout status deployment/public-api -n nex-backend
   kubectl rollout status deployment/auth-api -n nex-backend
   kubectl rollout status deployment/system-api -n nex-backend
   kubectl rollout status deployment/backend-gateway -n nex-backend
   ```

## Rollback flow

If the current release has a problem:

1. Find the previous good commit tag, for example `sha-0123456789abcdef`
2. Point the affected environment back to that exact tag

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\set-k8s-image-tags.ps1 -Overlay production -Tag sha-0123456789abcdef
kubectl apply -k deploy/k8s/overlays/production/runtime
kubectl rollout status deployment/public-api -n nex-backend
kubectl rollout status deployment/auth-api -n nex-backend
kubectl rollout status deployment/system-api -n nex-backend
kubectl rollout status deployment/backend-gateway -n nex-backend
```

If the rollback also needs schema or seed adjustments, rerun:

```powershell
kubectl apply -k deploy/k8s/overlays/production/bootstrap
kubectl wait --for=condition=complete job/backend-bootstrap -n nex-backend --timeout=300s
```

## Notes

- Use `sha-<commit>` tags for release and rollback, not `latest`.
- `-SourceOverlay staging` copies all four service tags from `staging` to `production`, which helps keep bootstrap and runtime images aligned.
- After changing tags with `set-k8s-image-tags.ps1`, commit the overlay changes so the deployed state stays traceable in git.

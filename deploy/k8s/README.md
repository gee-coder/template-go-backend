# Kubernetes Deployment

This directory contains a minimal Kubernetes deployment layout for the split backend services:

- `bootstrap-job.yaml`
- `runtime.yaml`
- `hpa.yaml`
- `ingress.yaml`

It assumes MySQL, Redis, and object storage already exist and are reachable from the cluster.

## What to change first

Before applying the manifests, replace at least these values in:

- `app-secret.yaml`
- `app-configmap.yaml`
- `ingress.yaml`

Important fields:

- `APP_DATABASE_DSN`
- `APP_REDIS_ADDR`
- `APP_REDIS_PASSWORD`
- `APP_JWT_SECRET`
- `APP_STORAGE_PROVIDER`
- `APP_STORAGE_MINIO_*` or OSS / OBS credentials
- `api.nex.local`

## Image assumptions

The manifests expect these images to exist in your registry:

- `ghcr.io/gee-coder/template-go-backend-bootstrap:latest`
- `ghcr.io/gee-coder/template-go-backend-public-api:latest`
- `ghcr.io/gee-coder/template-go-backend-auth-api:latest`
- `ghcr.io/gee-coder/template-go-backend-system-api:latest`

If you use a different registry, replace the image fields before applying.

You can also build the four images locally with:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-images.ps1 -Registry ghcr.io/your-org -Tag latest
```

Add `-Push` if you want the script to push them after build.

The repository also includes `.github/workflows/docker-images.yml`, which publishes the four runtime images to GHCR on every push to `master` or `main`.

## Apply order

Apply the resources in two steps so the database bootstrap finishes before the API Deployments start:

```powershell
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/app-configmap.yaml
kubectl apply -f deploy/k8s/app-secret.yaml
kubectl apply -f deploy/k8s/gateway-configmap.yaml
kubectl apply -f deploy/k8s/bootstrap-job.yaml
kubectl wait --for=condition=complete job/backend-bootstrap -n nex-backend --timeout=300s
kubectl apply -f deploy/k8s/runtime.yaml
kubectl apply -f deploy/k8s/hpa.yaml
kubectl apply -f deploy/k8s/ingress.yaml
```

## Scaling

Each API Deployment is separated so you can scale them independently:

```powershell
kubectl scale deployment/public-api -n nex-backend --replicas=4
kubectl scale deployment/auth-api -n nex-backend --replicas=6
kubectl scale deployment/system-api -n nex-backend --replicas=3
```

If your cluster has `metrics-server`, the HPA resources in `hpa.yaml` can scale them automatically.

## Notes

- The default configuration uses `minio` as the object storage provider.
- If you switch to `local` storage in Kubernetes, add a persistent volume before production use.
- `bootstrap-job.yaml` is safe to rerun after updating schema or seed logic, but avoid deleting it while a migration is running.

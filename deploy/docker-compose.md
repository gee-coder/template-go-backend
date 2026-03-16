# Docker Compose Deployment

This backend stack can be configured without editing `docker-compose.yml`.

## 1. Prepare the env file

```powershell
Copy-Item .env.docker.example .env.docker
```

Edit `.env.docker` and replace at least:

- `MYSQL_ROOT_PASSWORD`
- `MYSQL_PASSWORD`
- `REDIS_PASSWORD`
- `MINIO_ROOT_USER`
- `MINIO_ROOT_PASSWORD`
- `APP_JWT_SECRET`

## 2. Start the stack

```powershell
docker compose --env-file .env.docker up -d --build
```

The startup order is:

1. `mysql`, `redis`, `minio`
2. `minio-init`
3. `bootstrap`
4. `public-api`, `auth-api`, `system-api`
5. `gateway`

## 3. Stop the stack

```powershell
docker compose --env-file .env.docker down
```

## 4. Supported infrastructure switches

- Storage providers: `local`, `minio`, `aliyun_oss`, `huawei_obs`
- Mail providers: `mock`, `smtp`
- SMS provider currently enabled in the template: `mock`

`aliyun` and `huawei` SMS settings are reserved in the env template for future provider activation, but the current template still runs the mock sender only.

## 5. Common endpoints

- Gateway: `http://127.0.0.1:${GATEWAY_PORT}`
- MinIO API: `http://127.0.0.1:${MINIO_API_PORT}`
- MinIO Console: `http://127.0.0.1:${MINIO_CONSOLE_PORT}`

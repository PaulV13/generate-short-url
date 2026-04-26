# generate-short-url

API para crear URLs cortas y redirigir al destino original.

## Requisitos

- Go 1.25+
- PostgreSQL 14+

## Variables de entorno

La app soporta dos formas de configuración para base de datos:

1) `DATABASE_URL` (recomendado para producción)
2) Variables separadas (`DATABASE_HOST`, `DATABASE_PORT`, etc.)

Variables obligatorias:

- `PORT`
- `BASE_URL` (ej: `https://sho.rt/`)
- `CODE_LENGTH`

Si **no** usas `DATABASE_URL`, también son obligatorias:

- `DATABASE_HOST`
- `DATABASE_PORT`
- `DATABASE_USERNAME`
- `DATABASE_PASSWORD`
- `DATABASE_NAME`

Opcionales:

- `DATABASE_SSLMODE` (default: `disable`)

Usa `.env.example` como base.

## Ejecutar local

```bash
go run ./cmd/api
```

Healthcheck:

```bash
curl http://localhost:8080/health
```

## Docker

Build de imagen:

```bash
docker build -t generate-short-url:latest .
```

Run local con variables desde `.env`:

```bash
docker run --rm -p 8080:8080 --env-file .env generate-short-url:latest
```

Si tu Postgres está en tu máquina host, usa `DATABASE_HOST=host.docker.internal` y ejecuta:

```bash
docker run --rm -p 8080:8080 --env-file .env --add-host=host.docker.internal:host-gateway generate-short-url:latest
```

## Docker Compose (API + Postgres + migraciones)

1) Crear archivo de entorno local para compose:

```bash
cp .env.docker.example .env.docker
```

2) Levantar stack completo (Postgres, migrate, api):

```bash
docker compose --env-file .env.docker up --build
```

3) Verificar healthcheck:

```bash
curl http://localhost:8080/health
```

Comandos operativos:

```bash
# levantar en background
docker compose --env-file .env.docker up -d --build

# ver logs solo de API
docker compose --env-file .env.docker logs -f api

# ejecutar migraciones manualmente
docker compose --env-file .env.docker run --rm migrate ./migrate up

# rollback manual
docker compose --env-file .env.docker run --rm migrate ./migrate down

# apagar todo (manteniendo datos)
docker compose --env-file .env.docker down

# apagar y borrar datos de Postgres
docker compose --env-file .env.docker down -v
```

Flujo estable recomendado:

- `postgres` arranca primero y pasa healthcheck.
- `migrate` corre `up` y termina con estado exitoso.
- `api` arranca solo cuando `postgres` está healthy y `migrate` finalizó OK.

## Migraciones

Migración inicial:

- `migrations/001_create_short_urls.sql`
- rollback: `migrations/001_create_short_urls_down.sql`

Ejemplo con `psql`:

```bash
psql "$DATABASE_URL" -f migrations/001_create_short_urls.sql
```

Rollback:

```bash
psql "$DATABASE_URL" -f migrations/001_create_short_urls_down.sql
```

Con el binario de migraciones:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
```

## Endpoints principales

- `POST /api/v1/urls`
- `GET /api/v1/urls`
- `GET /api/v1/urls/:code`
- `PATCH /api/v1/urls/:code/deactivate`
- `PATCH /api/v1/urls/:code/active`
- `GET /:code` (redirección pública)
- `GET /health`

## Seguridad

- No subir `.env` al repositorio.
- Rotar cualquier credencial que haya estado expuesta.
- Para producción, usar Postgres gestionado con TLS (`sslmode=require` o equivalente en `DATABASE_URL`).

## Deploy en Render (base)

- Runtime: Docker.
- Start Command: `./api`
- Pre-Deploy Command: `./migrate up`
- Health Check Path: `/health`

Variables recomendadas en Render:

- `PORT=10000` (o el valor que defina Render para tu servicio)
- `BASE_URL=https://tu-dominio/`
- `CODE_LENGTH=6`
- `DATABASE_URL=<render-postgres-connection-string>`
- `MIGRATIONS_DIR=/app/migrations`

## CI (GitHub Actions)

Workflow: `.github/workflows/ci.yml`

Checks incluidos:

- `test` -> `go test ./...`
- `build` -> `go build ./cmd/api` y `go build ./cmd/migrate`
- `docker` -> `docker build -t generate-short-url:ci .`

Para proteger `main`, en GitHub configura Branch Protection y marca como requeridos:


- `test`
- `build`
- `docker`

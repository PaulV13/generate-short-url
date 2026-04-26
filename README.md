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

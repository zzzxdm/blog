# Blog

Modern blog system using Vue 3 and Go + Gin.

## Development

Start infrastructure:

```bash
docker compose -f deploy/docker-compose.yml up -d postgres redis
```

Start the API:

```bash
cd api
go install github.com/air-verse/air@v1.61.7
air
```

The API automatically loads `api/.env` on startup. Existing system environment variables take precedence over values in that file.

If Air is not installed or not on `PATH`, use `go run ./cmd/server` as a fallback.

Start the web app:

```bash
cd web
npm install
npm run dev
```

Default URLs:

- Web: `http://localhost:5173`
- API: `http://localhost:8080/api/health`

Default development accounts:

- Admin: `admin@example.com` / `password`
- User: `linyi@example.com` / `password`

Turnstile local development:

- If you use a real Cloudflare Turnstile site key locally, add `localhost` and `127.0.0.1` to the widget's allowed hostnames in the Cloudflare dashboard.
- To test the local flow without hostname binding, use Cloudflare's testing keys:
  - Site Key: `1x00000000000000000000AA`
  - Secret Key: `1x0000000000000000000000000000000AA`

## Verification

Backend:

```bash
cd api
go test ./...
```

Frontend:

```bash
cd web
npm run build
```

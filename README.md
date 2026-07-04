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
go run ./cmd/server
```

Start the web app:

```bash
cd web
npm install
npm run dev
```

Default URLs:

- Web: `http://localhost:5173`
- API: `http://localhost:8080/api/health`


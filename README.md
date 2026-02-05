# ShipsGame

## Backend (Go)

Build and run locally:

```bash
make up
make dev
```

Environment variables:

- `PORT` (default: 8080)
- `REDIS_ADDR` (default: localhost:6379)
- `REDIS_PASSWORD`
- `REDIS_DB` (default: 0)
- `JWT_SECRET`
- `POSTGRES_DSN` (default: postgres://ships:ships@localhost:5432/ships?sslmode=disable)
- `CORS_ORIGINS` (default: `*`, comma-separated)

Docker build/run:

```bash
docker build -t ships-backend ./backend
docker run --rm -p 8080:8080 ships-backend
```

## Web (Next.js)

Install and run:

```bash
cd web
npm install
npm run dev
```

Frontend environment variables:

- `NEXT_PUBLIC_API_URL` (e.g. http://localhost:8080)
- `NEXT_PUBLIC_WS_URL` (e.g. ws://localhost:8080/ws)

Docker build/run:

```bash
docker build -t ships-web ./web
docker run --rm -p 3000:3000 ships-web
```

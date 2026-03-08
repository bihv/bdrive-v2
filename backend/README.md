# Backend - 1Drive API

## Development

```bash
# Start dependencies
docker-compose up -d

# Run migrations
migrate -path migrations -database "postgresql://onedrive:onedrive_secret_2026@localhost:5432/onedrive?sslmode=disable" up

# Run server (with hot reload)
air
```

## Environment

Copy `.env.example` to `.env` and update values.

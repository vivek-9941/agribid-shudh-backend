# Agribid Shudh Backend

Go backend for the Agribid Shudh B2B FMCG distribution platform.

## Tech Stack

- **Language**: Go 1.22+
- **Router**: [chi](https://github.com/go-chi/chi)
- **Database**: PostgreSQL 15+
- **DB Driver**: [pgx/v5](https://github.com/jackc/pgx)
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Auth**: JWT (RS256)
- **Config**: [envconfig](https://github.com/kelseyhightower/envconfig)
- **Logging**: [zap](https://go.uber.org/zap)
- **Metrics**: Prometheus

## Repository Structure

```
agribid-shudh-backend/
├── cmd/server/main.go          # Entry point
├── internal/                   # Domain modules (app, auth, partner, catalog, ...)
├── pkg/                        # Shared utility packages
├── migrations/                 # SQL migration files
├── tests/                      # Integration tests and fixtures
├── docs/backend/               # Architecture docs
├── configs/                    # Example config files
├── scripts/                    # Utility scripts (migrate, seed)
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15+ (or use Docker)

### Local Development

1. Copy the example config:
   ```bash
   cp configs/config.example.yaml .env
   ```

2. Start dependencies:
   ```bash
   docker-compose up -d postgres
   ```

3. Run migrations:
   ```bash
   RUN_MIGRATIONS=true go run ./cmd/server
   ```

4. Start the server:
   ```bash
   go run ./cmd/server
   ```

5. Check health:
   ```bash
   curl http://localhost:8080/health
   ```

### Running Tests

```bash
go test ./...
```

For integration tests (requires Docker):
```bash
go test ./tests/integration/...
```

### Docker

```bash
docker-compose up
```

This starts the app and a PostgreSQL 15 instance.

## API

Base path: `/api/v1/`

See [docs/backend/HLD.md](docs/backend/HLD.md) for the full API reference and architecture details.

## Modules

| Module | Package | Description |
|--------|---------|-------------|
| Auth | `internal/auth` | OTP + JWT authentication |
| RBAC | `internal/rbac` | Role-based access control |
| Partner | `internal/partner` | Partner hierarchy and KYC |
| Catalog | `internal/catalog` | Product catalog and categories |
| Pricing | `internal/pricing` | Role-based pricing and schemes |
| Order | `internal/order` | Cart and order management |
| Fulfillment | `internal/fulfillment` | Sell-side order fulfillment |
| Invoice | `internal/invoice` | GST-compliant invoice generation |
| Dispatch | `internal/dispatch` | Shipment and delivery tracking |
| Inventory | `internal/inventory` | Stock management |
| Payment | `internal/payment` | Payments, credit, and ledger |
| Returns | `internal/returns` | Return requests and credit notes |
| Notification | `internal/notification` | Multi-channel notifications |
| Dashboard | `internal/dashboard` | Role-scoped metrics and reports |
| Admin | `internal/admin` | Admin console operations |
| Audit | `internal/audit` | Audit trail logging |
| Integration | `internal/integration` | SMS, email, and payment gateways |

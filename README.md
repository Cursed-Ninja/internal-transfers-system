# Internal Transfers System

_A Go service for managing financial accounts and processing transactions efficiently._

---

## ⚙️ Working

The service is built in Go and exposes REST API endpoints for:

- Creating accounts
- Fetching account details
- Processing transactions between accounts

It follows a layered architecture with separate API and database layers.

1. Incoming requests are tagged with a unique request ID for logging and debugging.
2. Requests are forwarded to the appropriate handlers, which extract and validate the request body.
3. Validated requests are then passed to the database layer to perform the relevant operation.

---

## 📦 Tech Stack & Dependencies

- **Language / Runtime:** Go
- **Database:** PostgreSQL
- **Key Libraries:**
  - `gorilla/mux` – HTTP request routing
  - `zap` – Structured logging
  - `pq` – PostgreSQL driver
  - `shopspring/decimal` – Precise decimal handling for account balances
- **Development Tools:**
  - Docker
  - Makefile

---

## 🚀 Running the Service

### Prerequisites

- Docker
- Make (for local development)
- PostgreSQL (running locally if not using Docker)

### Quick Start

You can run the service either using Docker or locally with Make:

| Method     | Command                 | Notes                                                                                                |
| ---------- | ----------------------- | ---------------------------------------------------------------------------------------------------- |
| **Docker** | `make local-compose-up` | Builds and runs the service along with PostgreSQL                                                    |
| **Local**  | `make run`              | Ensure PostgreSQL is running and the `connstr` in `config.local.yml` points to the correct database. |

---

## 🧪 Tests & Other Commands

### Run Unit Tests

```sh
make unit-test
```

### Format

```sh
make fmt
```

---

## 📁 Project Structure

```
.
├── .dockerignore                 # Files to ignore in Docker builds
├── .gitignore                     # Files to ignore in Git
├── coverage.out                   # Test coverage report
├── docker-compose.local.yml       # Docker Compose for quick start
├── Dockerfile                     # Docker image build instructions
├── go.mod                         # Go module definition
├── go.sum                         # Go module checksums
├── LICENSE                        # Project license
├── Makefile                       # Commands for running, testing, formatting
├── README.md                      # Project documentation
├── cmd/
│   └── server/
│       └── main.go               # Entry point for the service
├── config/
│   ├── config.development.yml     # Dev environment config
│   └── config.local.yml           # Local environment config
└── internal/
    ├── config/
    │   └── config.go              # Config loader and struct definitions
    ├── migrations/
    │   ├── 1763416987_create_accounts.sql  # SQL migration
    │   └── runner.go              # Migration runner
    ├── server/
    │   ├── handler.go             # HTTP handlers
    │   ├── handler_test.go        # Handler tests
    │   ├── middleware.go          # HTTP middleware
    │   ├── routes.go              # Route binding
    │   └── server.go              # Server struct
    ├── storage/
    │   ├── models.go              # Database models
    │   ├── postgres.go            # Postgres DB logic
    │   ├── postgres_test.go       # Postgres tests
    │   ├── storage.go             # Storage interface
    │   └── mocks/
    │       └── storage.go         # Mock implementations for testing
    └── utils/
        └── utils.go               # Helper utilities
```

---

## 🔍 API Endpoints

| Method | Endpoint              | Description                            |
| ------ | --------------------- | -------------------------------------- |
| POST   | /accounts             | Create a new account                   |
| GET    | /accounts/{accountID} | Fetch account details by ID            |
| POST   | /transactions         | Process a transaction between accounts |

### Sample Requests

#### Create Account

```sh
curl -X POST http://localhost:8080/accounts \
     -H "Content-Type: application/json" \
     -d '{
           "account_id": "123",
           "initial_balance": "250.054"
         }'
```

#### Get Account Details

```sh
curl "http://localhost:8080/accounts/123" \
     -H "Accept: application/json"
```

#### Process Transaction

```sh
curl -X POST http://localhost:8080/transactions \
     -H "Content-Type: application/json" \
     -d '{
           "source_account_id": "123",
           "destination_account_id": "456",
           "amount": "250.00"
         }'
```

---

# ✨ Additional Notes

## Assumptions

- Account IDs are unique.
- Negative balances are not allowed during account creation or transaction processing.
- Requests are validated for correctness before processing.
- Field names in requests must exactly match the expected JSON names; no fuzzy matching is allowed.
- Rate limiting and caching are not required, as the system is assumed to handle a small scale of requests.

## Trade-offs

- Strict JSON field matching improves reliability and reduces parsing errors but makes the API less forgiving for clients.
- Validation is performed on every request for correctness, which simplifies error handling but may add slight overhead.
- Rate limiting and caching are omitted to keep the service simple and easy to run, which limits scalability under high load.

# Basilisk

<p align="center">
    <img src="./assets/logo.png" alt="Basilisk Logo" width="150">
</p>

Basilisk is a minimal yet powerful Golang project skeleton that provides essential features out of the box for building production-ready services.

## Features

- **Google Auth Login** -- Google authentication is set up and ready to go out of the box. Just update your Google credentials in the `.env` file and implement your database insert/upsert method.
- **JWT Based Authentication** -- Access and refresh token support with configurable expiry, base64-encoded secrets, and auth middleware for protected routes.
- **Health Check** -- Built-in health check endpoint to monitor service availability.
- **PostgreSQL Database** -- Connection pooling with `database/sql` and the `lib/pq` driver.
- **Middleware Stack** -- Request logging with unique request IDs, panic recovery, CORS, and JWT auth middleware.
- **Configuration Management** -- YAML config with environment variable substitution.
- **Graceful Shutdown** -- Signal handling (SIGINT, SIGTERM) with ordered resource cleanup.
- **TLS Support** -- Automatic HTTPS when certificate files are present.
- **CPU Profiling** -- Built-in `pprof` profiling via `-pprof` flag.
- **Docker/Podman Support** -- Build and run via containers with auto-detected engine.

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL 15+ (or Docker/Podman)

### Clone and install

```sh
git clone https://github.com/shubhvish4495/basilisk.git
cd basilisk
go mod download
```

### Set up environment

Copy the example env file and fill in your values:

```sh
cp .env.example .env
# Edit .env with your actual credentials
source .env
```

### Start the database

A helper script is included to run PostgreSQL in a container:

```sh
./start_local_db.sh start    # Start PostgreSQL container
./start_local_db.sh stop     # Stop container
./start_local_db.sh status   # Check status
./start_local_db.sh logs     # View logs
./start_local_db.sh remove   # Remove container
```

### Run the application

```sh
make run
```

The server starts on **port 4444**. If TLS cert files are configured, HTTPS is enabled automatically.

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the binary to `bin/basilisk` |
| `make build-linux` | Cross-compile for Linux (amd64) |
| `make run` | Build and run the application |
| `make test` | Run tests with race detection and coverage |
| `make lint` | Run `golangci-lint` |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build container image |
| `make docker-run` | Build and run in a container |
| `make create-migration name=<name>` | Create a new migration file pair |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back all migrations |

## Configuration

Basilisk uses a YAML config file (`config/config.yml`) with environment variable substitution. All configuration is driven through environment variables -- see `.env.example` for the full list.

### CPU Profiling

Access profiling data at `http://localhost:4444/debug/pprof/` by running:

```sh
./bin/basilisk -pprof
```

## Project Structure

```
basilisk/
├── cmd/main.go                  # Application entry point
├── config/config.yml            # YAML config with env var substitution
├── pkg/
│   ├── auth/                    # JWT and Google OAuth services
│   ├── config/                  # Configuration loader
│   ├── db/                      # Database connection and models
│   ├── helper/                  # Response formatting, errors, pagination
│   └── rest/                    # Router, middleware, and handlers
├── .env.example                 # Sample environment variables
├── Makefile                     # Build and dev commands
└── start_local_db.sh            # Local PostgreSQL container helper
```

## Contributing

Feel free to open issues and pull requests to improve Basilisk!

## License

This project is licensed under the MIT License.

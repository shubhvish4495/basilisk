# Basilisk

<p align="center">
    <img src="./assets/logo.png" alt="Basilisk Logo" width="150">
</p>

Basilisk is a minimal yet powerful Golang project skeleton that provides essential features out of the box, including:

- **Configuration Management**: Load and manage configurations seamlessly.
- **Graceful Shutdown**: Handle shutdown signals properly to clean up resources.
- **CPU Profiling**: Built-in profiling for performance diagnostics.

## Features

- Structured project layout following best practices.
- Configurable via environment variables and configuration files.
- Signal handling for graceful termination (SIGINT, SIGTERM).
- Profiling enabled via `pprof` for performance monitoring.

## Installation

Clone the repository:

```sh
git clone https://github.com/shubhvish4495/basilisk.git
cd basilisk
```

Install dependencies:

```sh
go mod download
```

## Usage

### Running the Application

```sh
make run
```

### Makefile Commands

We provide a Makefile with useful commands to streamline development:

- `make build`: Build the Go application.
- `make run`: Run the application.
- `make lint`: Run the linter.
- `make clean`: Clean build artifacts.
- `make install-lint`: Install `golangci-lint` if not already installed.

### Configuration

Basilisk supports configuration via environment variables and config files (e.g., JSON, YAML). You can customize settings based on your needs.

### Graceful Shutdown

The application listens for termination signals (SIGINT, SIGTERM) and ensures cleanup before exiting.

### CPU Profiling

CPU profiling is enabled through `pprof`. You can access it via:

```sh
http://localhost:6060/debug/pprof/
```

Enable profiling by running:

```sh
go run main.go -pprof
```

## Project Structure

```
/basilisk
│── /config     # Configuration files
│── /pkg        # Internal package structure
│── /cmd        # Main entry point(s)
│── /cmd/main.go     # Application entry
│── go.mod      # Module definition
```

## Contributing

Feel free to open issues and pull requests to improve Basilisk!

## License

This project is licensed under the MIT License.

---

Happy coding! 🚀

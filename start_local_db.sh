#!/bin/bash

# Docker/Podman PostgreSQL startup script
# Creates a minimal PostgreSQL container with:
# - User: postgres
# - Password: postgres
# - Database: postgres
# - Port: 5432

# Detect container runtime (docker or podman)
if command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
elif command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
else
    echo "Error: Neither docker nor podman found. Please install one of them."
    exit 1
fi

echo "Using container runtime: $CONTAINER_CMD"

CONTAINER_NAME="local-postgres"
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="postgres"
POSTGRES_DB="postgres"
POSTGRES_PORT="5432"

# Function to start PostgreSQL
start_postgres() {
    echo "Starting PostgreSQL container..."

    # Check if container already exists
    if [ $($CONTAINER_CMD ps -a -q -f name=$CONTAINER_NAME) ]; then
        echo "Container $CONTAINER_NAME already exists. Starting it..."
        $CONTAINER_CMD start $CONTAINER_NAME
    else
        echo "Creating new PostgreSQL container..."
        $CONTAINER_CMD run -d \
            --name $CONTAINER_NAME \
            -e POSTGRES_USER=$POSTGRES_USER \
            -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
            -e POSTGRES_DB=$POSTGRES_DB \
            -p $POSTGRES_PORT:5432 \
            postgres:15-alpine
    fi

    echo "PostgreSQL container started successfully!"
    echo "Connection details:"
    echo "  Host: localhost"
    echo "  Port: $POSTGRES_PORT"
    echo "  Database: $POSTGRES_DB"
    echo "  Username: $POSTGRES_USER"
    echo "  Password: $POSTGRES_PASSWORD"
    echo ""
    echo "Connection string: postgresql://$POSTGRES_USER:$POSTGRES_PASSWORD@localhost:$POSTGRES_PORT/$POSTGRES_DB"
}

# Function to stop PostgreSQL
stop_postgres() {
    echo "Stopping PostgreSQL container..."
    $CONTAINER_CMD stop $CONTAINER_NAME
    echo "PostgreSQL container stopped."
}

# Function to remove PostgreSQL container
remove_postgres() {
    echo "Removing PostgreSQL container..."
    $CONTAINER_CMD stop $CONTAINER_NAME 2>/dev/null
    $CONTAINER_CMD rm $CONTAINER_NAME 2>/dev/null
    echo "PostgreSQL container removed."
}

# Function to show container status
status_postgres() {
    echo "PostgreSQL container status:"
    $CONTAINER_CMD ps -a --filter name=$CONTAINER_NAME
}

# Function to show logs
logs_postgres() {
    echo "PostgreSQL container logs:"
    $CONTAINER_CMD logs $CONTAINER_NAME
}

# Main script logic
case $1 in
    "start")
        start_postgres
        ;;
    "stop")
        stop_postgres
        ;;
    "remove")
        remove_postgres
        ;;
    "status")
        status_postgres
        ;;
    "logs")
        logs_postgres
        ;;
    *)
        echo "Usage: $0 {start|stop|remove|status|logs}"
        echo ""
        echo "Commands:"
        echo "  start  - Start PostgreSQL container"
        echo "  stop   - Stop PostgreSQL container"
        echo "  remove - Remove PostgreSQL container"
        echo "  status - Show container status"
        echo "  logs   - Show container logs"
        echo ""
        echo "Default action (no arguments): start"
        echo ""
        if [ $# -eq 0 ]; then
            start_postgres
        fi
        ;;
esac
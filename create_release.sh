#!/bin/bash

# Check if environment parameter is provided
if [ $# -eq 0 ]; then
    echo "❌ Usage: $0 <env>"
    echo "   env: stage or prod"
    exit 1
fi

ENV="$1"

# Validate environment parameter
if [ "$ENV" != "stage" ] && [ "$ENV" != "prod" ]; then
    echo "❌ Error: Environment must be either 'stage' or 'prod'"
    echo "   Usage: $0 <env>"
    exit 1
fi

# Set variables
RELEASE_DIR="release"
BIN_DIR="$RELEASE_DIR"
CONFIG_DIR="$RELEASE_DIR/config"
SWAGGER_DIR="$RELEASE_DIR/swagger"
APP_NAME="clr_apis"  # Change this to your binary name
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🚀 Building release for environment: $ENV"

# Remove existing release directory if it exists
if [ -d "$RELEASE_DIR" ]; then
    echo "🗑️  Removing existing release directory..."
    rm -rf "$RELEASE_DIR"
fi

# Create directory structure
echo "📁 Creating directory structure..."
mkdir -p "$BIN_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$SWAGGER_DIR"

# Build the application for Linux x86_64
echo "🔨 Building application for Linux x86_64..."
make build-linux

# Check if build was successful
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

# Copy binary
echo "📦 Copying binary..."
cp "bin/$APP_NAME" "$BIN_DIR/"

# Copy config files
echo "⚙️  Copying configuration files..."

# Determine which .env file to use based on environment
if [ "$ENV" = "stage" ]; then
    ENV_FILE=".env_dev"
    echo "📋 Using development environment file: $ENV_FILE"
elif [ "$ENV" = "prod" ]; then
    ENV_FILE=".env_prod"
    echo "📋 Using production environment file: $ENV_FILE"
fi

# Check if the environment file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "❌ Error: Environment file $ENV_FILE not found!"
    exit 1
fi

# Copy the appropriate .env file
cp "$ENV_FILE" "$RELEASE_DIR/.env"
cp "config/config.yml" "$CONFIG_DIR/"

# Copy swagger files
echo "📄 Copying Swagger files..."
cp "swagger/apis.yaml" "$SWAGGER_DIR/"

echo "✅ Release created successfully in $RELEASE_DIR for environment: $ENV"
echo "📋 Environment file used: $ENV_FILE"

# scp -i ../clr-key-1.pem -r release/{*,.*} ec2-user@ec2-13-203-125-207.ap-south-1.compute.amazonaws.com:/home/ec2-user/clr 

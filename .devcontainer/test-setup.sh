#!/bin/bash
set -e

echo "=== Testing DevContainer Setup ==="

# Check if Go is available
echo "Checking Go installation..."
go version

# Check if Task is available
echo "Checking Task installation..."
if command -v task &> /dev/null; then
    echo "Task is available"
    task --version
else
    echo "Installing Task..."
    go install github.com/go-task/task/v3/cmd/task@latest
    export PATH=$PATH:$HOME/go/bin
    task --version
fi

# Check if Docker is available
echo "Checking Docker installation..."
docker --version

# Check if Docker Compose is available
echo "Checking Docker Compose installation..."
docker compose version

# Initialize the project
echo "Initializing project..."
task init

# Run short tests
echo "Running short tests..."
task test-short

# Start services and test they are accessible
echo "Starting Docker Compose services..."
task env-up-detach

# Wait a bit for services to start
sleep 10

# Check if services are running
echo "Checking if services are running..."
if docker compose --file=.dev/docker-compose.yml --project-name=wire ps | grep -q "Up"; then
    echo "✅ Services are running"
else
    echo "❌ Services are not running properly"
    docker compose --file=.dev/docker-compose.yml --project-name=wire ps
    exit 1
fi

# Test connectivity to MongoDB
echo "Testing MongoDB connectivity..."
if timeout 5 bash -c 'until docker exec wire-mongodb-1 mongosh --eval "db.runCommand({ping: 1})" 2>/dev/null; do sleep 1; done'; then
    echo "✅ MongoDB is accessible"
else
    echo "❌ MongoDB is not accessible"
fi

# Test connectivity to FerretDB  
echo "Testing FerretDB connectivity..."
if timeout 5 bash -c 'until nc -z localhost 27018; do sleep 1; done'; then
    echo "✅ FerretDB is accessible"
else
    echo "❌ FerretDB is not accessible"
fi

echo "=== DevContainer setup test completed successfully! ==="
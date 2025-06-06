# FerretDB Wire Development Setup

This repository supports multiple development workflows:

## Host-based Development (Current Default)

The traditional setup where code is edited and tools run on the host machine, while databases run in Docker containers.

```bash
# Start MongoDB (existing workflow)
bin/task env-up

# Or start FerretDB
DOCKERFILE=ferretdb bin/task env-up  

# Run tests
bin/task test
```

## Container-based Development (New)

Everything runs inside Docker containers, including the Go development environment. This is perfect for:
- GitHub Codespaces
- VSCode DevContainers
- Contributors who don't want to install Go locally
- Consistent development environments

### Using Docker Compose directly

```bash
# Start the development environment
bin/task env-dev-up

# Run commands in the development container
bin/task env-dev-exec -- go version
bin/task env-dev-exec -- go test -short ./...

# Open a shell in the development container  
bin/task env-dev-shell

# Stop the environment
bin/task env-down
```

### Using VSCode DevContainers

1. Install the "Dev Containers" extension
2. Open the repository in VSCode
3. When prompted, click "Reopen in Container" or:
   - Press `F1` → "Dev Containers: Reopen in Container"
4. VSCode will build and start the development environment automatically
5. Use the integrated terminal to run Go commands

### Using GitHub Codespaces

1. Click the green "Code" button on GitHub
2. Select "Codespaces" tab
3. Click "Create codespace on main"
4. GitHub will automatically set up the development environment
5. Use the terminal to run Go commands

## Architecture

The development setup uses Docker Compose with these services:

- `wire-dev`: Go development environment with source code mounted
- `mongodb`: MongoDB database server (port 27017)
- `ferretdb`: FerretDB server (port 27018)
- `database`: Legacy service for backward compatibility

The `wire-dev` container has:
- Go 1.24.4
- Source code mounted at `/workspace`
- Network access to database services
- Persistent Go module and build caches
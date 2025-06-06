# DevContainer and Codespaces Setup

This directory contains the configuration for running the FerretDB Wire development environment in containers.

## VS Code DevContainers

To use VS Code DevContainers:

1. Install the [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
2. Open the repository in VS Code
3. When prompted, click "Reopen in Container" or use `Ctrl+Shift+P` → "Remote-Containers: Reopen in Container"

The container will automatically:
- Install Go development tools
- Install the Task runner
- Set up the development environment with `task init`
- Start MongoDB and FerretDB services with `task env-up-detach`
- Forward ports 27017 (MongoDB) and 27018 (FerretDB)

## GitHub Codespaces

This configuration also works with GitHub Codespaces:

1. Go to the GitHub repository
2. Click the green "Code" button → "Codespaces" tab → "Create codespace on main"
3. Wait for the environment to set up automatically

## Development Workflow

Once the container is running, you can use the standard development workflow:

```bash
# Run tests (short, no external dependencies)
task test-short

# Run full tests (requires MongoDB services)
task test-all

# Lint code
task lint

# Format code
task fmt

# Start/stop external services
task env-up
task env-down
```

## Port Forwarding

The following ports are automatically forwarded:
- `27017` - MongoDB
- `27018` - FerretDB

## Testing the Setup

You can verify that everything is working correctly by running the test script:

```bash
./.devcontainer/test-setup.sh
```

This script will:
- Check all required tools are installed
- Initialize the project
- Run short tests
- Start services
- Verify connectivity

## Troubleshooting

If services don't start properly, try:
```bash
task env-down
task env-up-detach
```

If the DevContainer doesn't work as expected:
1. Ensure Docker is running on your host machine
2. Try rebuilding the container: `Ctrl+Shift+P` → "Remote-Containers: Rebuild Container"
3. Check the DevContainer logs in VS Code

For more information, see the main project [CONTRIBUTING.md](../CONTRIBUTING.md).
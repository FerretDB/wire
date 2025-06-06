# Contributing

Thank you for your interest in making this package better!

## Development Setup

There are several ways to set up the development environment:

### Option 1: VS Code DevContainers (Recommended)

The easiest way to get started is using VS Code DevContainers:

1. Install [VS Code](https://code.visualstudio.com/) and the [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
2. Open this repository in VS Code
3. When prompted, click "Reopen in Container" or press `Ctrl+Shift+P` → "Remote-Containers: Reopen in Container"

The container will automatically set up the full development environment including:
- Go development tools
- Task runner
- MongoDB and FerretDB services
- All necessary dependencies

### Option 2: GitHub Codespaces

For cloud-based development:

1. Click the green "Code" button on GitHub → "Codespaces" tab → "Create codespace on main"
2. Wait for the environment to set up automatically

### Option 3: Local Development

For local development on your machine:

1. Install [Go 1.24+](https://golang.org/dl/)
2. Install [Task](https://taskfile.dev/installation/)
3. Install [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
4. Clone the repository
5. Run the setup:
   ```bash
   task init
   task env-up-detach  # Start MongoDB and FerretDB services
   ```

### Development Workflow

Once your environment is set up, you can use these commands:

```bash
# Run short tests (no external dependencies)
task test-short

# Run all tests (requires MongoDB services)
task test-all

# Lint and format code
task lint
task fmt

# Start/stop services
task env-up      # Start and follow logs
task env-down    # Stop services

# Generate code
task gen
```

## Contributing code

### Submitting code changes

#### Submitting PR

1. There is no need to use draft pull requests.
   If you want to get feedback on something you are working on,
   please create a normal pull request, even if it is not "fully" ready yet.
2. In the pull request review conversations,
   please either leave a new comment or resolve (close) the conversation,
   which ensures other people can read all comments.
   But do not do that simultaneously.
   Conversations should typically be resolved by the conversation starter, not the PR author.
3. During development in a branch/PR,
   commit messages (both titles and bodies) are not important and can be anything.
   All commits are always squashed on merge by GitHub.
   Please **do not** squash them manually, amend them, and/or force push them -
   that makes the review process harder.
4. Please don't forget to click
   ["re-request review" buttons](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/requesting-a-pull-request-review)
   once PR is ready for re-review.

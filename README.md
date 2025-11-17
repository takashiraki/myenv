# myenv

A CLI tool for rapidly creating and managing containerized development environments with Docker.

## Overview

MyEnv streamlines the process of setting up development environments by automating Docker container creation, proxy network configuration, and project setup. Currently supports PHP projects.

## Features

- **Interactive Setup** - Guided project creation with language and framework selection
- **Automated Container Management** - Docker Compose integration with health checks
- **Proxy Network Management** - Automatic proxy server configuration with localhost subdomains (e.g., `project.localhost`)
- **Editor Integration** - Automatic project opening in development container (supports VS Code, Cursor, devcontainer CLI)
- **Project Templates** - Pre-configured development stacks
- **Modular Architecture** - Add additional services and modules to existing projects

## Installation

```bash
brew install takashiraki/tap/myenv
```

## Usage

### Initial Setup

Before creating your first project, run the setup command:

```bash
myenv setup
```

This creates the necessary configuration files in `~/.config/myenv/` and sets up the required Docker networks.

For a quicker setup that only creates configuration files (without network setup):

```bash
myenv setup --quick
```

Use quick setup when you want to configure MyEnv first and create networks later when running `myenv init`.

### Create a New Project

```bash
myenv init
```

You can also specify the language and framework using flags:

```bash
myenv init -l PHP
myenv init -l PHP -f Laravel
```

This will:

1. Prompt for language selection (if not specified)
2. Prompt for framework selection (if not specified)
3. Ask for container name and proxy domain configuration
4. Clone the appropriate Docker template
5. Generate environment configuration files
6. Build and start the containers with proxy network
7. Create a new project
8. Optionally open the project in your preferred editor (VS Code, Cursor, or devcontainer CLI)

### Add Modules to Existing Projects

Add additional modules or services to your existing development environment:

```bash
myenv add
myenv add -m <module-name>
```

### Start an Existing Project

Start up an existing project's containers:

```bash
myenv up
```

This command is useful when you want to restart a previously created project without recreating it.

### Available Commands

- `myenv setup` - Initial setup with full configuration and network creation (required before first use)
- `myenv setup --quick` - Quick setup with configuration only (networks created on first `myenv init`)
- `myenv init` - Create a new development environment (interactive)
- `myenv init -l PHP` - Create a PHP project directly
- `myenv init -l PHP -f Laravel` - Create a Laravel project directly
- `myenv up` - Start an existing project's containers
- `myenv add` - Add modules to existing environment (interactive)
- `myenv add -m <module>` - Add specific module directly
- `myenv --help` - Show available commands and options
- `myenv --version` or `myenv -v` - Show version information

## Requirements

- Docker and Docker Compose
- Go 1.21 or later (for development)
- Git

## Development

```bash
git clone https://github.com/takashiraki/myenv.git
cd myenv
go mod tidy
go run main.go
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

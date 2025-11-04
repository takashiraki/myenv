# myenv

A CLI tool for rapidly creating and managing containerized development environments with Docker.

## Overview

MyEnv streamlines the process of setting up development environments by automating Docker container creation, port management, and project configuration. Currently supports PHP projects.

## Features

- **Interactive Setup** - Guided project creation with language and framework selection
- **Automated Container Management** - Docker Compose integration with health checks
- **Port Conflict Prevention** - Smart port allocation and validation
- **VS Code Integration** - Automatic project opening in development container
- **Project Templates** - Pre-configured development stacks

## Installation

```bash
brew install takashiraki/tap/myenv
```

## Usage

### Create a New Project

```bash
myenv init
```

You can also specify the language using flags:

```bash
myenv init --lang PHP
```

This will:
1. Prompt for language selection (if not specified)
2. Ask for container name and port configuration
3. Clone the appropriate Docker template
4. Generate environment configuration files
5. Build and start the containers
6. Create a new project
7. Optionally open the project in VS Code

### Available Commands

- `myenv init` - Create a new development environment
- `myenv init -l PHP` - Create a PHP project directly
- `myenv --help` - Show available commands and options
- `myenv --version` - Show version information

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

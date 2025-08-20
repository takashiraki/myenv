# myenv

A CLI tool for rapidly creating and managing containerized development environments with Docker.

## Overview

MyEnv streamlines the process of setting up development environments by automating Docker container creation, port management, and project configuration. Currently supports Laravel projects with plans to expand to other frameworks.

## Features

- **Interactive Setup** - Guided project creation with input validation
- **Automated Container Management** - Docker Compose integration with health checks
- **Port Conflict Prevention** - Smart port allocation and validation
- **VS Code Integration** - Automatic project opening in development environment
- **Project Templates** - Pre-configured development stacks

## Installation

```bash
go install github.com/takashiraki/myenv@latest
```

## Usage

### Create a Laravel Project

```bash
myenv laravel
```

This will:
1. Prompt for container name and port configuration
2. Clone the Laravel Docker template
3. Generate environment configuration files
4. Build and start the containers
5. Create a new Laravel application
6. Open the project in VS Code

### Available Commands

- `myenv laravel` - Create a new Laravel development environment
- `myenv --help` - Show available commands and options

## Requirements

- Docker and Docker Compose
- Go 1.24.5 or later
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

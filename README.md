# Terraform Provider for ALTR SaaS

A Terraform provider for managing ALTR SaaS resources.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [Local Development](#local-development)
- [Running Tests](#running-tests)
- [Documentation](#Documentation)
- [License](#License)
- [Support](#Support)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25

## Installation

### From Terraform Registry
```
hcl
terraform {
required_providers {
altr = {
source  = "altrsoftware/altr"
version = "~> 1.0"
}
}
}

provider "altr" {
# Configuration options
}
```
### Manual Installation

1. Download the latest release from the [releases page](https://github.com/altrsoftware/terraform-provider-altr/releases)
2. Extract the binary to your Terraform plugins directory
3. Run `terraform init` to initialize the provider

## Usage

### Basic Example
```
hcl
resource "altr_repo" "example" {
name = "my-repo"
# Additional configuration
}

resource "altr_sidecar" "example" {
name = "my-sidecar"
# Additional configuration
}
```
For more examples, see the [examples](./examples) directory.

## Local Development

### Prerequisites

- Go 1.25 or later
- Terraform CLI

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/altrsoftware/terraform-provider-altr.git
   cd terraform-provider-altr
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the provider:
   ```bash
   go build -o terraform-provider-altr
   ```

### Development Override

For local development, you can use Terraform's development overrides:

1. Create or update `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "altr/altr" = "/path/to/your/go/bin"
     }
     
     # For all other providers, install them directly from their origin provider
     # registries as normal. If you omit this, Terraform will _only_ use
     # the dev_overrides block, and so no other providers will be available.
     direct {}
   }
   ```

2. Install the provider locally:
   ```bash
   go install .
   ```

3. Navigate to your Go bin directory and create a symlink:
   ```bash
   cd $(go env GOPATH)/bin
   cp terraform-provider-altr terraform-provider-altr_v999.0.0
   ```

4. Run Terraform commands in your test configuration:
   ```bash
   terraform plan
   terraform apply
   ```

## Running Tests

### Unit Tests

Run the unit tests:
```
bash
go test ./...
```
### Acceptance Tests

Run acceptance tests (requires valid credentials):
```
bash
TF_ACC=1 go test ./... -v
```

### Generating Documentation

Generate terraform documentation:
```
bash
cd ./tools
go generate
```

### Linting and Formatting

Ensure code quality:
```
bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

### Development Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`go test ./...`)
6. Run linting and formatting tools
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Code Style Guidelines

- Follow standard Go formatting (`go fmt`)
- Run `go vet` to catch common issues
- Write clear, descriptive commit messages
- Add appropriate documentation for new features
- Include tests for new functionality
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines

### Pull Request Requirements

- All tests must pass
- Code coverage should not decrease
- Include documentation updates for new features
- Follow the existing code style and patterns
- Include example usage where applicable

### Reporting Issues

Please use the [GitHub Issues](https://github.com/altrsoftware/terraform-provider-altr/issues) page to report bugs or request features. When reporting issues, please include:

- Terraform version
- Provider version
- Relevant configuration snippets
- Error messages or logs
- Steps to reproduce

## Documentation

- [Provider Documentation](./docs/)
- [Examples](./examples/)
- [Resource Documentation](./docs/resources/)
- [Data Source Documentation](./docs/data-sources/)

## License

This project is licensed under the [APACHE 2.0 License](LICENSE) - see the LICENSE file for details.

## Support

For support and questions:

- 📖 Check the [documentation](./docs/)
- 🔍 Search [existing issues](https://github.com/altrsoftware/terraform-provider-altr/issues)
- 🆕 Create a [new issue](https://github.com/altrsoftware/terraform-provider-altr/issues/new)

## Additional Resources
https://docs.altr.com/

https://www.youtube.com/channel/UCcqDY0wrRlQ8hQ_mjJNfkAA

https://www.altr.com/resources

---

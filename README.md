[<img src="https://cdn.prod.website-files.com/66b4bc4ca83726f5a87183ab/685136470d15399f106cb13e_adcbd542a834a1942d077cd5c09d3057_GitHub%20cover%20image%201584x396.png">](https://lablabs.io/)

**About us:**</br>
[Labyrinth Labs](https://lablabs.io/) is a one-stop-shop for **DevOps, Cloud & Kubernetes**! We specialize in creating **powerful**, **scalable** and **cloud-native platforms** tailored to elevate your business.

[As a team of experienced DevOps engineers](https://lablabs.io/about/), we know how to help our customers start their journey in the cloud, address the issues they have in their current setups and provide a **strategic solution to transform their infrastructure**.

---

# Terraform Provider LARA Utils

The LARA Utils provider offers a collection of utility functions that extend Terraform's built-in capabilities. Currently, it provides advanced deep merging functionality for complex data structures.

> [!WARNING]
> This provider is in active development and considered experimental. Features and APIs may change without notice. Use with caution.

## Usage

```hcl
terraform {
  required_providers {
    lara-utils = {
      source = "lablabs/lara-utils"
    }
  }
}
```

## Functions

- [deep_merge](docs/functions/deep_merge.md) - Recursively merge nested maps and objects with various merge strategies
- [yaml_deep_merge](docs/functions/yaml_deep_merge.md) - Functionally same as `deep_merge` but for YAML encoded strings

> [!NOTE]
> Terraform map deep merging functionality is taken from <https://github.com/isometry/terraform-provider-deepmerge>. If you are interested in this functionality particually, consider supporting original project.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.8
- [Go](http://www.golang.org) >= 1.25 (for development)
- [Mise](https://mise.jdx.dev/) (for development)

## Installation

See [documentation](docs/index.md) on how to integrate provider into your project.

### From Terraform Registry

The provider is available on the [Terraform Registry](https://registry.terraform.io/providers/lablabs/lara-utils/latest).

### From GitHub Releases

You can also download the provider binary from the [GitHub releases page](https://github.com/lablabs/terraform-provider-lara-utils/releases) and install it manually.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

### Building the Provider

To compile the provider, run:

```shell
mise run build
```

Alternatively, you can install it to your `$GOPATH/bin` directory:

```shell
mise run install
```

### Local Development

For local development and testing, you can use the provider locally by creating a `dev_overrides` block in your `~/.terraformrc` file:

```hcl
provider_installation {
  dev_overrides {
    "lablabs/lara-utils" = "/path/to/your/provider/binary"
  }
  direct {}
}
```

### Documentation

To generate or update documentation, run:

```shell
mise run generate
```

### Testing

To run unit tests:

```shell
mise run test
```

In order to run the full suite of acceptance tests, run:

```shell
mise run testacc
```

> [!NOTE]
> Acceptance tests create real resources, and often cost money to run.

### Code Quality

Before submitting a pull request, ensure your code passes all checks:

```shell
mise run
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](.github/workflows/CONTRIBUTING.md) for details.

### Reporting Issues

If you find a bug or have a feature request, please open an issue on our [GitHub repository](https://github.com/lablabs/terraform-provider-lara-utils/issues).

### Pull Requests

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass
6. Submit a pull request

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation as needed
- Use descriptive commit messages

## Troubleshooting

### Common Issues

#### Provider Not Found

If you encounter "provider not found" errors, ensure that:

1. The provider is properly declared in your `required_providers` block
2. You've run `terraform init` after adding the provider
3. The version constraint allows for available versions

#### Function Not Available

If the functions are not recognized:

1. Verify you're using Terraform >= 1.8 (required for provider-defined functions)
2. Ensure the provider is properly configured
3. Check that you're using the correct function syntax: `provider::lara-utils::deep_merge(...)`

### Getting Help

- Check the [documentation](docs/) for detailed function reference
- Search existing [issues](https://github.com/lablabs/terraform-provider-lara-utils/issues)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

## Support

- **Documentation**: [Provider Documentation](docs/)
- **Issues**: [GitHub Issues](https://github.com/lablabs/terraform-provider-lara-utils/issues)

## Maintainers

This provider is maintained by [Labyrinth Labs](https://lablabs.io/).

## License

This provider is distributed under the [Mozilla Public License 2.0](LICENSE).

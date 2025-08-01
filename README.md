# Terraform Provider Qbee (Terraform Plugin Framework)

This Terraform provider implements (parts of) the qbee API, in order to facilitate configuration of
a [qbee](https://qbee.io) account using Terraform.

## Using the provider

For documentation on how to use this provider, see
the [provider documentation](https://registry.terraform.io/providers/qbee.io/qbee/latest/docs)
at the Terraform Registry.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your
machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary
in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22
- [Goreleaser](https://goreleaser.com/install/) >= 1.20

### Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Acceptance tests

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources. They are dependent on valid qbee credentials being
configured in the environment (QBEE_USERNAME and QBEE_PASSWORD).

```shell
make testacc
```

If you want more control over what's run and output, you can do something like:

```shell
TF_LOG=INFO TF_ACC=1 go test -count=1 -run='TestAccTagFiledistributionGroupResource' -v ./internal/provider
```

Where:

- `TF_LOG` is the log level for Terraform
- `TF_ACC` is the flag to run acceptance tests
- `go test` is the command to run the tests
- `-count=1` is the flag to run the test only once
- `-run='TestAccTagFiledistributionGroupResource'` is the flag to run only the test with the given
  name

You can leave out the `-run` flag to run all tests.

## Installing the provider for local use

See https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers
for the most up-to-date instructions on how to use for local testing.

One suggested workflow which aligns with the way this project is set up is the following:

Build the provider using `goreleaser`:

```shell
# From the directory of terraform-provider-qbee
goreleaser build --single-target --clean --snapshot
```

After this, the binary we built will be in ./dist/terraform-provider-qbee_OS_ARCH. Move the file
to a custom location (e.g. `~/terraform-provider-overrides/terraform-provider-qbee`) with the
correct name for the provider, which is `terraform-provider-qbee` (without the OS and
architecture suffixes).

Next, update your `~/.terraformrc` file to include:

```hcl
provider_installation {
  dev_overrides {
    # Assuming you moved the provider binary to ~/terraform-provider-overrides/terraform-provider-qbee
    "qbee-io/qbee" = "~/terraform-provider-overrides/terraform-provider-qbee
  }
  
  direct {}
}
```

And run `terraform apply` in a project using the provider.

## About

_This repository is built on
the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) template.


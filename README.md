# Terraform Provider Qbee (Terraform Plugin Framework)

> [!IMPORTANT]
> Alpha release, don't use in production.
> Code is contributed by an external entity and is currently undergoing review and testing - breaking changes might be introduced.

This Terraform provider implements (parts of) the qbee API, in order to facilitate configuration of
a [qbee](https://qbee.io) account using Terraform.

## Using the provider

For documentation on how to use this provider, see the [provider documentation](https://registry.terraform.io/providers/lesteenman/qbee-lesteenman/latest/docs)
at the Terraform Registry.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

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
- `-run='TestAccTagFiledistributionGroupResource'` is the flag to run only the test with the given name

You can leave out the `-run` flag to run all tests.

## Installing the provider for local use

If you wish to use the provider locally, you can build it and place it in the correct directory.

```shell
# From the directory of terraform-provider-qbee
goreleaser build --single-target --clean
```

After this, the binary we built will be in ./dist/terraform-provider-qbee_OS_ARCH. Copy that to the
terraform project where you want to use it:

```shell
# From the root of your terraform project (where your *.tf files are stored):
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
TARGET_DIR=".terraform/plugins/github.com/lesteenman/terraform-provider-qbee-lesteenman/$VERSION/${GOOS}_${GOARCH}"

mkdir -p "$TARGET_DIR"
cp <dist> "$TARGET_DIR"
```

After that, initialize the terraform project with `terraform init -plugin-dir=.terraform/plugins`

## About

_This repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) template.


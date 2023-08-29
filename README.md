# Terraform Provider Qbee (Terraform Plugin Framework)

This Terraform provider implements (parts of) the qbee API, in order to facilitate configuration of
a qbee account using Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19
- [Goreleaser](https://goreleaser.com/install/) >= 1.20

## Installing the provider for local use

Because the provider is not yet published to any registry, to use it, you need to build it for your
current OS/ARCH and place it in the correct directory. To do so:

```shell
# From the directory of terraform-provider-qbee
goreleaser build --single-target --clean
```

After this, the binary we built will be in ./dist/terraform-provider-qbee_OS_ARCH. Copy that to the
terraform project where you want to use it:

```shell
# From the root of your terraform project (where your *.tf files are stored):
mkdir -p ./terraform.d/plugins/terraform.local/qbee/qbee/
cp <dist>.zip ./terraform.d/plugins/terraform.local/qbee/qbee/
```

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

Because this provider is hosted on a private bitbucket repository, we need to configure git to fetch it using our configured
SSH credentials: `git config --global url."git@bitbucket.org:booqsoftware".insteadOf "https://bitbucket.org/booqsoftware"`

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

## About

_This repository is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) template.


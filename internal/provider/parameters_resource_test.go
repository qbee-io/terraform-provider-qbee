package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"go.qbee.io/client"
	"go.qbee.io/client/config"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccParametersResource(t *testing.T) {
	var secretsHashState struct{ SecretsHash string }
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
			resource "qbee_parameters" "test" {
			 tag = "terraform:acctest:parameters"
			 extend = true
			 parameters = [
			   {
			     key = "parameter-key-1"
			     value = "parameter-value-1"
			   }
			 ]
			}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
				),
			},
			// Update to a different template
			{
				Config: providerConfig + `
			resource "qbee_parameters" "test" {
			 tag = "terraform:acctest:parameters"
			 extend = false
			 parameters = [
			   {
			     key = "parameter-key-1"
			     value = "parameter-value-1"
			   },
			   {
			     key = "parameter-key-2"
			     value = "parameter-value-2"
			   }
			 ]
			 secrets_wo = [
			   {
			     key   = "secret-key"
			     value = "secret-value"
			   }
			 ]
			}
			`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "false"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.key", "parameter-key-2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.value", "parameter-value-2"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
				),
			},
			// Import tag
			{
				ResourceName:                         "qbee_parameters.test",
				ImportState:                          true,
				ImportStateId:                        "tag:terraform:acctest:parameters",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag",
				ImportStateVerifyIgnore: []string{
					"secrets",
				},
			},
			// Change to only have secrets
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = false
  secrets_wo = [
    {
      key = "secret-key"
      value = "secret-value"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "parameters"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
					testAccCheckSecretsHashIsSet("qbee_parameters.test", &secretsHashState),
				),
			},
			// If we do update secrets_wo, the hash should update
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = false
  secrets_wo = [
    {
      key = "secret-key"
      value = "secret-value-2"
    }
  ]
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
					testAccCheckSecretsHashHasChanged("qbee_parameters.test", &secretsHashState),
				),
			},
			// If we change the secrets, it should still update
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = false
  secrets_wo = [
    {
      key = "secret-key"
      value = "secret-value-2"
    },
    {
      key = "other-key"
      value = "secret-value-3"
    },
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
				),
			},
			// Update to be empty
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
				),
			},
			// Update to be for a node
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  node = "integrationtests"
  extend = true
  parameters = [
    {
      key = "parameter-key-1"
      value = "parameter-value-1"
    },
    {
      key = "parameter-key-2"
      value = "parameter-value-2"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "tag"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "node", "integrationtests"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "true"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.#", "2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.key", "parameter-key-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.0.value", "parameter-value-1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.key", "parameter-key-2"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "parameters.1.value", "parameter-value-2"),
				),
			},
			// Import testing
			{
				ResourceName:                         "qbee_parameters.test",
				ImportState:                          true,
				ImportStateId:                        "node:integrationtests",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "node",
				ImportStateVerifyIgnore: []string{
					"secrets_hash",
				},
			},
		},
	})
}

func TestAccParametersResourceResolvesRemoteSecretsDrift(t *testing.T) {
	username := os.Getenv("QBEE_USERNAME")
	password := os.Getenv("QBEE_PASSWORD")
	baseURL := os.Getenv("QBEE_BASE_URL")

	// Create a qbee client we can use to directly inject drift
	qbeeClient := client.New()
	if baseURL != "" {
		qbeeClient = qbeeClient.WithBaseURL(baseURL)
	}
	err := qbeeClient.Authenticate(context.Background(), username, password)
	if err != nil {
		t.Fatalf("failed to auth drift client: %s", err)
	}

	resourceConfig := providerConfig + `
resource "qbee_parameters" "test" {
    tag = "terraform:acctest:parameters"
    extend = true
    secrets_wo = [
        {
            key = "drift_key"
            value = "value from configuration"
        }
    ]
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: resourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo_version"),
				),
			},
			// Sanity check: no planned changes before injecting drift
			{
				Config:             resourceConfig,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo_version"),
				),
			},
			// Inject drift
			{
				PreConfig: func() {
					ctx := context.Background()
					changeRequest, err := createChangeRequest("parameters", config.Parameters{
						Metadata: config.Metadata{
							Enabled: true,
							Extend:  true,
							Version: "v1",
						},
						Secrets: []config.Parameter{
							{
								Key:   "drift_key",
								Value: "drifted value",
							},
						},
					}, config.EntityTypeTag, "terraform:acctest:parameters")
					if err != nil {
						t.Fatalf("failed to inject drift: %s", err)
					}

					_, err = qbeeClient.CreateConfigurationChange(ctx, changeRequest)
					if err != nil {
						t.Fatalf("failed to inject drift: %s", err)
					}

					_, err = qbeeClient.CommitConfiguration(ctx, "terraform:acctest:parameters")
					if err != nil {
						t.Fatalf("failed to inject drift: %s", err)
					}
				},
				Config:             resourceConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo_version"),
				),
			},
		},
	})
}

func TestAccParametersResourceWithEphemeralSecretSource(t *testing.T) {
	resourceConfig := providerConfig + `
ephemeral "random_password" "rand" {
    length = 16
}

resource "qbee_parameters" "test" {
    tag = "terraform:acctest:parameters"
    extend = true
    // Version gates secret rewrites; ephemeral value may change across runs but won't trigger unless version changes
    secrets_wo_version = 1
    secrets_wo = [
        {
            key = "random_password"
            value = ephemeral.random_password.rand.result
        }
    ]
}
            `

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				Source:            "hashicorp/random",
				VersionConstraint: "3.7.2",
			},
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: resourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets_wo_version", "1"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
				),
			},
			// Since the version does not change, there should be no changes - even though the ephemeral
			// random resource always results in a different value
			{
				Config:             resourceConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Verifies that when secrets_wo_version is specified, changing the secret values alone
// does not trigger a plan, but bumping the version does.
func TestAccParametersResourceSecretsVersionControlsRewrite(t *testing.T) {
	var secretsHashState struct{ SecretsHash string }

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Initial apply with version = 1
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = true
  secrets_wo_version = 1
  secrets_wo = [
    {
      key = "versioned_key"
      value = "value-v1"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "extend", "true"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets_wo_version", "1"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets_wo"),
					testAccCheckSecretsHashIsSet("qbee_parameters.test", &secretsHashState),
				),
			},
			// Change the secret value but keep the same version: no plan expected
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = true
  secrets_wo_version = 1
  secrets_wo = [
    {
      key = "versioned_key"
      value = "value-v1-CHANGED"
    }
  ]
}
`,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets_wo_version", "1"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					testAccCheckSecretsHashIsUnchanged("qbee_parameters.test", &secretsHashState),
				),
			},
			// Bump the version: expect a plan to rewrite secrets
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = true
  secrets_wo_version = 2
  secrets_wo = [
    {
      key = "versioned_key"
      value = "value-v1-CHANGED"
    }
  ]
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets_wo_version", "2"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets_hash"),
					testAccCheckSecretsHashHasChanged("qbee_parameters.test", &secretsHashState),
				),
			},
		},
	})
}

func testAccCheckSecretsHashIsSet(resourceName string, state *struct{ SecretsHash string }) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		secretsHash, ok := rs.Primary.Attributes["secrets_hash"]
		if !ok {
			return fmt.Errorf("secrets_hash not found for %s", resourceName)
		}

		if secretsHash == "" {
			return fmt.Errorf("secrets_hash is not set for %s", resourceName)
		}

		state.SecretsHash = secretsHash
		return nil
	}
}

func testAccCheckSecretsHashHasChanged(resourceName string, previousState *struct{ SecretsHash string }) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		currentHash, ok := rs.Primary.Attributes["secrets_hash"]
		if !ok {
			return fmt.Errorf("secrets_hash not found for %s", resourceName)
		}

		if currentHash == "" {
			return fmt.Errorf("secrets_hash is not set for %s", resourceName)
		}

		if previousState.SecretsHash == "" {
			return fmt.Errorf("previous secrets_hash was not captured")
		}

		if currentHash == previousState.SecretsHash {
			return fmt.Errorf("secrets_hash did not change: both are %s", currentHash)
		}

		return nil
	}
}

func testAccCheckSecretsHashIsUnchanged(resourceName string, previousState *struct{ SecretsHash string }) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		currentHash, ok := rs.Primary.Attributes["secrets_hash"]
		if !ok {
			return fmt.Errorf("secrets_hash not found for %s", resourceName)
		}

		if currentHash == "" {
			return fmt.Errorf("secrets_hash is not set for %s", resourceName)
		}

		if previousState.SecretsHash == "" {
			return fmt.Errorf("previous secrets_hash was not captured")
		}

		if currentHash != previousState.SecretsHash {
			return fmt.Errorf("secrets_hash changed: from %s to %s", previousState.SecretsHash, currentHash)
		}

		return nil
	}
}

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccParametersResource(t *testing.T) {
	secretState := struct {
		SecretId string
	}{
		"",
	}

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
			 secrets = [
			   {
			     key = "secret-key"
			     value_wo = "secret-value"
			     value_wo_version = "1"
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
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.#", "1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.key", "secret-key"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets.0.value_wo"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.value_wo_version", "1"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets.0.secret_id"),
					testAccCheckSecretIdSet("qbee_parameters.test", "secrets.0", &secretState),
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
  secrets = [
    {
      key = "secret-key"
      value_wo = "secret-value"
      value_wo_version = "1"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.#", "1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.key", "secret-key"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets.0.value_wo"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.value_wo_version", "1"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets.0.secret_id"),
					testAccCheckSecretIdSet("qbee_parameters.test", "secrets.0", &secretState),
				),
			},
			// If value_wo_version does not change, it should not update
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = false
  secrets = [
    {
      key = "secret-key"
      value_wo = "secret-value-2"
      value_wo_version = "1"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.#", "1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.key", "secret-key"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets.0.value_wo"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.value_wo_version", "1"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets.0.secret_id"),
					testAccCheckSecretIdUnchanged("qbee_parameters.test", "secrets.0", &secretState),
				),
			},
			// If we do update value_wo_version, it should update
			{
				Config: providerConfig + `
resource "qbee_parameters" "test" {
  tag = "terraform:acctest:parameters"
  extend = false
  secrets = [
    {
      key = "secret-key"
      value_wo = "secret-value-2"
      value_wo_version = "2"
    }
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("qbee_parameters.test", "tag", "terraform:acctest:parameters"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "node"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.#", "1"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.key", "secret-key"),
					resource.TestCheckNoResourceAttr("qbee_parameters.test", "secrets.0.value_wo"),
					resource.TestCheckResourceAttr("qbee_parameters.test", "secrets.0.value_wo_version", "2"),
					resource.TestCheckResourceAttrSet("qbee_parameters.test", "secrets.0.secret_id"),
					testAccCheckSecretIdChanged("qbee_parameters.test", "secrets.0", &secretState),
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
			},
		},
	})
}

func testAccCheckSecretIdSet(resourceName string, resourcePath string, state *struct{ SecretId string }) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		secretId, ok := rs.Primary.Attributes[resourcePath+".secret_id"]
		if !ok {
			return fmt.Errorf("not found: %s.%s.secret_id", resourceName, resourcePath)
		}

		if secretId == "" {
			return fmt.Errorf("secret ID is not set for %s.%s", resourceName, resourcePath)
		}

		state.SecretId = secretId
		return nil
	}
}

func testAccCheckSecretIdUnchanged(resourceName string, resourcePath string, state *struct{ SecretId string }) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		secretId, ok := rs.Primary.Attributes[resourcePath+".secret_id"]
		if !ok {
			return fmt.Errorf("not found: %s.%s.secret_id", resourceName, resourcePath)
		}

		if secretId != state.SecretId {
			return fmt.Errorf("secret ID changed for %s.%s: expected %s, got %s", resourceName, resourcePath, state.SecretId, secretId)
		}

		return nil
	}
}

func testAccCheckSecretIdChanged(resourceName string, resourcePath string, state *struct{ SecretId string }) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		secretId, ok := rs.Primary.Attributes[resourcePath+".secret_id"]
		if !ok {
			return fmt.Errorf("not found: %s.%s.secret_id", resourceName, resourcePath)
		}

		if secretId == state.SecretId {
			return fmt.Errorf("secret ID did not change for %s.%s: expected different from %s, got %s", resourceName, resourcePath, state.SecretId, secretId)
		}

		return nil
	}
}

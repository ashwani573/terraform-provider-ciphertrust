package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// azureVaultTestConfig returns the HCL block for the resource, reading AZURE_CONNECTION_ID
// and AZURE_VAULT_NAME from the environment.
func azureVaultTestConfig(connectionID, vaultName string) string {
	return providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id   = %q
  azure_vault_name = %q
}
`, connectionID, vaultName)
}

func getAzureVaultEnv(t *testing.T) (connectionID, vaultName string) {
	t.Helper()
	connectionID = os.Getenv("AZURE_CONNECTION_ID")
	vaultName = os.Getenv("AZURE_VAULT_NAME")
	if connectionID == "" || vaultName == "" {
		t.Skip("AZURE_CONNECTION_ID and AZURE_VAULT_NAME must be set for Azure vault acceptance tests")
	}
	return
}

func TestCckmAzureVault_BasicCreateRead(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: azureVaultTestConfig(connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_azure_vault.test", "id"),
					resource.TestCheckResourceAttrSet("ciphertrust_azure_vault.test", "name"),
					resource.TestCheckResourceAttrSet("ciphertrust_azure_vault.test", "created_at"),
				),
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestCckmAzureVault_EnableDisableRotation(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
  enable_rotation  = true
}
`, connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ciphertrust_azure_vault.test", "enable_rotation", "true"),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
  enable_rotation  = false
}
`, connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ciphertrust_azure_vault.test", "enable_rotation", "false"),
				),
			},
			{
				Config: azureVaultTestConfig(connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("ciphertrust_azure_vault.test", "enable_rotation"),
				),
			},
		},
	})
}

func TestCckmAzureVault_AclManagement(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	aclUser := os.Getenv("AZURE_ACL_USER")
	if aclUser == "" {
		t.Skip("AZURE_ACL_USER must be set for ACL management tests")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
  acls {
    user_id = %q
    actions = ["read"]
  }
}
`, connectionID, vaultName, aclUser),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ciphertrust_azure_vault.test", "acls.#", "1"),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
  acls {
    user_id = %q
    actions = ["read"]
  }
  acls {
    user_id = %q
    actions = ["write"]
  }
}
`, connectionID, vaultName, aclUser, aclUser+"-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ciphertrust_azure_vault.test", "acls.#", "2"),
				),
			},
			{
				Config: azureVaultTestConfig(connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("ciphertrust_azure_vault.test", "acls.#"),
				),
			},
		},
	})
}

func TestCckmAzureVault_ImmutableConnectionID(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: azureVaultTestConfig(connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_azure_vault.test", "id"),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
}
`, connectionID+"-changed", vaultName),
				ExpectError: regexp.MustCompile(`(?i)immutable|cannot be changed`),
			},
		},
	})
}

func TestCckmAzureVault_ImmutableAzureVaultName(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: azureVaultTestConfig(connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_azure_vault.test", "id"),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
}
`, connectionID, vaultName+"-changed"),
				ExpectError: regexp.MustCompile(`(?i)immutable|cannot be changed`),
			},
		},
	})
}

func TestCckmAzureVault_ImmutableManagedHsm(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
  managed_hsm      = false
}
`, connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_azure_vault.test", "id"),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
  managed_hsm      = true
}
`, connectionID, vaultName),
				ExpectError: regexp.MustCompile(`(?i)immutable|cannot be changed`),
			},
		},
	})
}

func TestCckmAzureVault_ManagedHsmCreate(t *testing.T) {
	RequireCM(t)
	hsmName := os.Getenv("AZURE_MANAGED_HSM_NAME")
	if hsmName == "" {
		t.Skip("AZURE_MANAGED_HSM_NAME must be set for managed HSM tests")
	}
	connectionID, _ := getAzureVaultEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
  managed_hsm      = true
}
`, connectionID, hsmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_azure_vault.test", "id"),
				),
			},
		},
	})
}

func TestCckmAzureVaults_DataSource(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
}

data "ciphertrust_azure_vaults" "test" {
  depends_on = [ciphertrust_azure_vault.test]
  filters {
    name = %q
  }
}
`, connectionID, vaultName, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ciphertrust_azure_vaults.test", "vaults.0.id"),
				),
			},
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_azure_vault" "test" {
  connection_id    = %q
  azure_vault_name = %q
}

data "ciphertrust_azure_vaults" "test" {
  depends_on = [ciphertrust_azure_vault.test]
}
`, connectionID, vaultName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ciphertrust_azure_vaults.test", "vaults.0.id"),
				),
			},
		},
	})
}

func TestCckmAzureVault_DeleteOOB(t *testing.T) {
	RequireCM(t)
	connectionID, vaultName := getAzureVaultEnv(t)

	var capturedID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: azureVaultTestConfig(connectionID, vaultName),
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["ciphertrust_azure_vault.test"]
					if rs == nil {
						return fmt.Errorf("resource not found in state")
					}
					capturedID = rs.Primary.ID
					return nil
				},
			},
			{
				PreConfig: func() {
					client, ok := createCMClient()
					if !ok {
						t.Logf("CM client unavailable — skipping OOB delete step")
						return
					}
					deleteURL := "/api/v1/cckm/azure/vaults/" + capturedID + "/remove-vault"
					_, err := client.PostDataV2(context.Background(), uuid.New().String(), deleteURL, []byte("{}"))
					if err != nil {
						t.Logf("OOB vault delete failed: %v", err)
						return
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

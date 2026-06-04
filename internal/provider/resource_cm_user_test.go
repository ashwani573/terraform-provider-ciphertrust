package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResourceCMUser(t *testing.T) {
	username := fmt.Sprintf("testuser%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_user" "testUser" {
  name     = "%s"
  email    = "%s@local"
  username = "%s"
  password = "CHange01!@"
}
`, username, username, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_user.testUser", "id"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_user" "testUser" {
  name     = "john"
  email    = "john@local"
  username = "%s"
  password = "UPdate02!@"
}
`, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_user.testUser", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestResourceCMUserReadRefresh verifies the user resource Read() method
// (implemented under TFIN-288). After Create, a refresh-only step reconciles
// state against the live CM API. The step asserts that the refreshed state
// keeps the computed/optional attributes populated and, crucially, that the
// write-only "password" attribute is preserved from prior state (the API never
// returns it). A RefreshState step expects an empty plan afterwards by default,
// so this also guards against Read() introducing spurious drift.
func TestResourceCMUserReadRefresh(t *testing.T) {
	username := fmt.Sprintf("testuser_read%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create user with a password.
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_user" "testUserRead" {
  name     = "%s"
  email    = "%s@local"
  username = "%s"
  password = "CHange01!@"
}
`, username, username, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_user.testUserRead", "id"),
					resource.TestCheckResourceAttr("ciphertrust_user.testUserRead", "username", username),
					resource.TestCheckResourceAttr("ciphertrust_user.testUserRead", "password", "CHange01!@"),
				),
			},
			// Step 2: Refresh-only - Read() reconciles state with the API and
			// must not drift. Password must survive the refresh unchanged.
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_user.testUserRead", "id"),
					resource.TestCheckResourceAttr("ciphertrust_user.testUserRead", "username", username),
					resource.TestCheckResourceAttr("ciphertrust_user.testUserRead", "email", username+"@local"),
					resource.TestCheckResourceAttr("ciphertrust_user.testUserRead", "password", "CHange01!@"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestResourceCMUserUpdateWithoutName verifies that a user can be updated
// without providing the optional "name" field. This guards against a regression
// where the provider would send an empty name to the API, causing a 422 error.
func TestResourceCMUserUpdateWithoutName(t *testing.T) {
	username := fmt.Sprintf("testuser_noname%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create user with only required fields (no name)
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_user" "testUserNoName" {
  username = "%s"
  password = "CHange01!@"
}
`, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_user.testUserNoName", "id"),
				),
			},
			// Step 2: Update by adding email without providing name
			{
				Config: providerConfig + fmt.Sprintf(`
resource "ciphertrust_user" "testUserNoName" {
  username = "%s"
  password = "CHange01!@"
  email    = "noname@local"
}
`, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_user.testUserNoName", "id"),
					resource.TestCheckResourceAttr("ciphertrust_user.testUserNoName", "email", "noname@local"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

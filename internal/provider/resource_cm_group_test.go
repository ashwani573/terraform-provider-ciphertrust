package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"

	common "github.com/ThalesGroup/terraform-provider-ciphertrust/internal/provider/common"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceCMGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "ciphertrust_groups" "testGroup" {
  name="TestGroup"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_groups.testGroup", "name"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ciphertrust_groups" "testGroup" {
  description="Updated via TF"
  name="TestGroup"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ciphertrust_groups.testGroup", "name"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccCMGroup_Read_DriftDetection verifies that Read() detects attribute drift:
// when description is changed out-of-band on CM, the refreshed Terraform state reflects
// the live value so that the next plan shows a non-empty diff.
func TestAccCMGroup_Read_DriftDetection(t *testing.T) {
	if os.Getenv("CIPHERTRUST_ADDRESS") == "" {
		t.Skip("CIPHERTRUST_ADDRESS not set; skipping acceptance test")
	}

	groupName := "tf-test-drift-" + uuid.New().String()[:8]
	cfg := providerConfig + fmt.Sprintf(`
resource "ciphertrust_groups" "test" {
  name        = %q
  description = "original description"
}
`, groupName)

	var groupID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create group and verify initial state; capture the resource ID for
			// out-of-band mutation in step 3.
			{
				Config: cfg,
				Check: checkStep(t, "create",
					resource.TestCheckResourceAttr("ciphertrust_groups.test", "name", groupName),
					resource.TestCheckResourceAttr("ciphertrust_groups.test", "description", "original description"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["ciphertrust_groups.test"]
						if !ok {
							return fmt.Errorf("ciphertrust_groups.test not found in state")
						}
						groupID = rs.Primary.ID
						return nil
					},
				),
			},
			// Step 2: confirm no spurious drift on a second plan (Read() idempotency check).
			{
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: mutate description out-of-band using the captured group ID, then
			// refresh state and assert the drifted value is visible in Terraform state.
			{
				PreConfig: func() {
					client, ok := createCMClient()
					if !ok {
						t.Log("createCMClient: skipping OOB update (env vars not set)")
						return
					}
					payload, err := json.Marshal(map[string]interface{}{
						"description": "drifted description",
					})
					if err != nil {
						t.Logf("json.Marshal: %v", err)
						return
					}
					_, err = client.UpdateData(context.Background(), groupID, common.URL_CM_GROUPS, payload, "name")
					if err != nil {
						t.Logf("OOB update failed: %v", err)
					}
				},
				RefreshState: true,
				Check: checkStep(t, "drift detected",
					resource.TestCheckResourceAttr("ciphertrust_groups.test", "description", "drifted description"),
				),
			},
		},
	})
}

// TestAccCMGroup_Read_OutOfBandDelete verifies that Read() detects out-of-band deletion:
// when the group is deleted directly on CM, Read() must call resp.State.RemoveResource so
// Terraform plans a re-create rather than silently reporting "no changes".
func TestAccCMGroup_Read_OutOfBandDelete(t *testing.T) {
	if os.Getenv("CIPHERTRUST_ADDRESS") == "" {
		t.Skip("CIPHERTRUST_ADDRESS not set; skipping acceptance test")
	}

	groupName := "tf-test-oob-" + uuid.New().String()[:8]
	cfg := providerConfig + fmt.Sprintf(`
resource "ciphertrust_groups" "test" {
  name = %q
}
`, groupName)

	var groupID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create group; capture the resource ID for out-of-band deletion in step 3.
			{
				Config: cfg,
				Check: checkStep(t, "create",
					resource.TestCheckResourceAttr("ciphertrust_groups.test", "name", groupName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["ciphertrust_groups.test"]
						if !ok {
							return fmt.Errorf("ciphertrust_groups.test not found in state")
						}
						groupID = rs.Primary.ID
						return nil
					},
				),
			},
			// Step 2: confirm no spurious drift on a second plan.
			{
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: delete the group out-of-band using the captured group ID, then refresh
			// state. ExpectNonEmptyPlan: true asserts the resource was removed from state
			// (Terraform plans a re-create). testAccListResources prints remaining resources
			// for diagnostic visibility.
			{
				PreConfig: func() {
					client, ok := createCMClient()
					if !ok {
						t.Log("createCMClient: skipping OOB delete (env vars not set)")
						return
					}
					_, err := client.DeleteByURL(context.Background(), uuid.NewString(), common.URL_CM_GROUPS+"/"+groupID)
					if err != nil {
						t.Logf("OOB delete failed: %v", err)
					}
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: checkStep(t, "resource absent after OOB delete",
					testAccListResources(),
				),
			},
			// Step 4: with the original config, Terraform must plan a re-create.
			{
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

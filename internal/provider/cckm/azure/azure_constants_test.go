package azure

import (
	"strings"
	"testing"
)

func Test_Azure_Constants(t *testing.T) {
	constants := []string{
		URL_AZURE_GET_SUBSCRIPTIONS,
		URL_AZURE_GET_VAULTS,
		URL_AZURE_GET_MANAGED_HSMS,
		URL_AZURE_ADD_VAULTS,
		URL_AZURE_VAULTS,
	}
	for _, c := range constants {
		if !strings.HasPrefix(c, "/api/v1/cckm/azure/") {
			t.Errorf("expected URL constant to start with /api/v1/cckm/azure/, got: %s", c)
		}
	}
	if subIDPath == "" {
		t.Error("subIDPath must not be empty")
	}
	if addVaultIDPath == "" {
		t.Error("addVaultIDPath must not be empty")
	}
}

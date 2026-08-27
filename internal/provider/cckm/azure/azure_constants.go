package azure

const (
	URL_AZURE_GET_SUBSCRIPTIONS = "/api/v1/cckm/azure/get-subscriptions"
	URL_AZURE_GET_VAULTS        = "/api/v1/cckm/azure/get-vaults"
	URL_AZURE_GET_MANAGED_HSMS  = "/api/v1/cckm/azure/get-managed-hsms"
	URL_AZURE_ADD_VAULTS        = "/api/v1/cckm/azure/add-vaults"
	URL_AZURE_VAULTS            = "/api/v1/cckm/azure/vaults"
)

// subIDPath is the gjson path for the subscription ID in the POST /get-subscriptions response.
// UNVERIFIED STUB — must be confirmed against a live CM instance before merge.
// Create() will fail loudly at runtime with a descriptive error if this path is wrong.
const subIDPath = "resources.0.subscriptionId"

// addVaultIDPath is the gjson path for the vault ID in the POST /add-vaults response.
// UNVERIFIED STUB — must be confirmed against a live CM instance before merge.
// Create() will fail loudly at runtime with a descriptive error if this path is wrong.
const addVaultIDPath = "resources.0.id"

package azure

import "github.com/hashicorp/terraform-plugin-framework/types"

type AzureVaultTFSDK struct {
	ID                           types.String  `tfsdk:"id"`
	URI                          types.String  `tfsdk:"uri"`
	Account                      types.String  `tfsdk:"account"`
	Application                  types.String  `tfsdk:"application"`
	DevAccount                   types.String  `tfsdk:"dev_account"`
	CreatedAt                    types.String  `tfsdk:"created_at"`
	UpdatedAt                    types.String  `tfsdk:"updated_at"`
	SyncedAt                     types.String  `tfsdk:"synced_at"`
	Name                         types.String  `tfsdk:"name"`
	AzureVaultID                 types.String  `tfsdk:"azure_vault_id"`
	AzureName                    types.String  `tfsdk:"azure_name"`
	CloudName                    types.String  `tfsdk:"cloud_name"`
	Location                     types.String  `tfsdk:"location"`
	Type                         types.String  `tfsdk:"type"`
	SubscriptionID               types.String  `tfsdk:"subscription_id"`
	SubscriptionName             types.String  `tfsdk:"subscription_name"`
	VaultConnection              types.String  `tfsdk:"vault_connection"`
	ConnectionName               types.String  `tfsdk:"connection_name"`
	Tags                         types.Map     `tfsdk:"tags"`
	Labels                       types.Map     `tfsdk:"labels"`
	CloudKeyBackupLimit          types.Int64   `tfsdk:"cloud_key_backup_limit"`
	TenantID                     types.String  `tfsdk:"tenant_id"`
	VaultURI                     types.String  `tfsdk:"vault_uri"`
	EnableSoftDelete             types.Bool    `tfsdk:"enable_soft_delete"`
	EnablePurgeProtection        types.Bool    `tfsdk:"enable_purge_protection"`
	SoftDeleteRetentionInDays    types.Int64   `tfsdk:"soft_delete_retention_in_days"`
	EnabledForDeployment         types.Bool    `tfsdk:"enabled_for_deployment"`
	EnabledForDiskEncryption     types.Bool    `tfsdk:"enabled_for_disk_encryption"`
	EnabledForTemplateDeployment types.Bool    `tfsdk:"enabled_for_template_deployment"`
	CreateMode                   types.String  `tfsdk:"create_mode"`
	EnableRbacAuthorization      types.Bool    `tfsdk:"enable_rbac_authorization"`
	Sku                          AzureSkuTFSDK `tfsdk:"sku"`
	EnableSuccessAuditEvent      types.Bool    `tfsdk:"enable_success_audit_event"`
	DiscoverOnly                 types.Bool    `tfsdk:"discover_only"`
	Acls                         types.Set     `tfsdk:"acls"`
	ConnectionID                 types.String  `tfsdk:"connection_id"`
	ManagedHsm                   types.Bool    `tfsdk:"managed_hsm"`
	EnableRotation               types.Bool    `tfsdk:"enable_rotation"`
	RotationJobParams            types.String  `tfsdk:"rotation_job_params"`
	AzureVaultName               types.String  `tfsdk:"azure_vault_name"`
}

type AzureSkuTFSDK struct {
	Name   types.String `tfsdk:"name"`
	Family types.String `tfsdk:"family"`
}

type AzureVaultsListTFSDK struct {
	Filters AzureVaultFiltersTFSDK      `tfsdk:"filters"`
	Vaults  []AzureVaultDataSourceTFSDK `tfsdk:"vaults"`
}

type AzureVaultFiltersTFSDK struct {
	Name           types.String `tfsdk:"name"`
	CloudName      types.String `tfsdk:"cloud_name"`
	SubscriptionID types.String `tfsdk:"subscription_id"`
	Location       types.String `tfsdk:"location"`
}

type AzureVaultDataSourceTFSDK struct {
	ID                           types.String  `tfsdk:"id"`
	URI                          types.String  `tfsdk:"uri"`
	Account                      types.String  `tfsdk:"account"`
	Application                  types.String  `tfsdk:"application"`
	DevAccount                   types.String  `tfsdk:"dev_account"`
	CreatedAt                    types.String  `tfsdk:"created_at"`
	UpdatedAt                    types.String  `tfsdk:"updated_at"`
	SyncedAt                     types.String  `tfsdk:"synced_at"`
	Name                         types.String  `tfsdk:"name"`
	AzureVaultID                 types.String  `tfsdk:"azure_vault_id"`
	AzureName                    types.String  `tfsdk:"azure_name"`
	CloudName                    types.String  `tfsdk:"cloud_name"`
	Location                     types.String  `tfsdk:"location"`
	Type                         types.String  `tfsdk:"type"`
	SubscriptionID               types.String  `tfsdk:"subscription_id"`
	SubscriptionName             types.String  `tfsdk:"subscription_name"`
	VaultConnection              types.String  `tfsdk:"vault_connection"`
	ConnectionName               types.String  `tfsdk:"connection_name"`
	Tags                         types.Map     `tfsdk:"tags"`
	Labels                       types.Map     `tfsdk:"labels"`
	CloudKeyBackupLimit          types.Int64   `tfsdk:"cloud_key_backup_limit"`
	TenantID                     types.String  `tfsdk:"tenant_id"`
	VaultURI                     types.String  `tfsdk:"vault_uri"`
	EnableSoftDelete             types.Bool    `tfsdk:"enable_soft_delete"`
	EnablePurgeProtection        types.Bool    `tfsdk:"enable_purge_protection"`
	SoftDeleteRetentionInDays    types.Int64   `tfsdk:"soft_delete_retention_in_days"`
	EnabledForDeployment         types.Bool    `tfsdk:"enabled_for_deployment"`
	EnabledForDiskEncryption     types.Bool    `tfsdk:"enabled_for_disk_encryption"`
	EnabledForTemplateDeployment types.Bool    `tfsdk:"enabled_for_template_deployment"`
	CreateMode                   types.String  `tfsdk:"create_mode"`
	EnableRbacAuthorization      types.Bool    `tfsdk:"enable_rbac_authorization"`
	Sku                          AzureSkuTFSDK `tfsdk:"sku"`
}

package azure

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tidwall/gjson"

	"github.com/ThalesGroup/terraform-provider-ciphertrust/internal/provider/common"
)

var _ datasource.DataSource = &dataSourceAzureVaultsList{}

type dataSourceAzureVaultsList struct {
	client *common.Client
}

func NewDataSourceAzureVaultsList() datasource.DataSource {
	return &dataSourceAzureVaultsList{}
}

func (d *dataSourceAzureVaultsList) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_vaults"
}

func (d *dataSourceAzureVaultsList) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *dataSourceAzureVaultsList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Azure Key Vaults registered in CipherTrust CCKM.",
		Attributes: map[string]schema.Attribute{
			"filters": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"name":            schema.StringAttribute{Optional: true},
					"cloud_name":      schema.StringAttribute{Optional: true},
					"subscription_id": schema.StringAttribute{Optional: true},
					"location":        schema.StringAttribute{Optional: true},
				},
			},
			"vaults": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                              schema.StringAttribute{Computed: true},
						"uri":                             schema.StringAttribute{Computed: true},
						"account":                         schema.StringAttribute{Computed: true},
						"application":                     schema.StringAttribute{Computed: true},
						"dev_account":                     schema.StringAttribute{Computed: true},
						"created_at":                      schema.StringAttribute{Computed: true},
						"updated_at":                      schema.StringAttribute{Computed: true},
						"synced_at":                       schema.StringAttribute{Computed: true},
						"name":                            schema.StringAttribute{Computed: true},
						"azure_vault_id":                  schema.StringAttribute{Computed: true},
						"azure_name":                      schema.StringAttribute{Computed: true},
						"cloud_name":                      schema.StringAttribute{Computed: true},
						"location":                        schema.StringAttribute{Computed: true},
						"type":                            schema.StringAttribute{Computed: true},
						"subscription_id":                 schema.StringAttribute{Computed: true},
						"subscription_name":               schema.StringAttribute{Computed: true},
						"vault_connection":                schema.StringAttribute{Computed: true},
						"connection_name":                 schema.StringAttribute{Computed: true},
						"tags":                            schema.MapAttribute{Computed: true, ElementType: types.StringType},
						"labels":                          schema.MapAttribute{Computed: true, ElementType: types.StringType},
						"cloud_key_backup_limit":          schema.Int64Attribute{Computed: true},
						"tenant_id":                       schema.StringAttribute{Computed: true},
						"vault_uri":                       schema.StringAttribute{Computed: true},
						"enable_soft_delete":              schema.BoolAttribute{Computed: true},
						"enable_purge_protection":         schema.BoolAttribute{Computed: true},
						"soft_delete_retention_in_days":   schema.Int64Attribute{Computed: true},
						"enabled_for_deployment":          schema.BoolAttribute{Computed: true},
						"enabled_for_disk_encryption":     schema.BoolAttribute{Computed: true},
						"enabled_for_template_deployment": schema.BoolAttribute{Computed: true},
						"create_mode":                     schema.StringAttribute{Computed: true},
						"enable_rbac_authorization":       schema.BoolAttribute{Computed: true},
						"sku": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"name":   schema.StringAttribute{Computed: true},
								"family": schema.StringAttribute{Computed: true},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceAzureVaultsList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	id := uuid.New().String()
	tflog.Trace(ctx, common.MSG_METHOD_START+"[data_source_azure_vaults_list.go -> Read]["+id+"]")
	defer tflog.Trace(ctx, common.MSG_METHOD_END+"[data_source_azure_vaults_list.go -> Read]["+id+"]")

	var state AzureVaultsListTFSDK
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	if !state.Filters.Name.IsNull() {
		params.Set("name", state.Filters.Name.ValueString())
	}
	if !state.Filters.CloudName.IsNull() {
		params.Set("cloud_name", state.Filters.CloudName.ValueString())
	}
	if !state.Filters.SubscriptionID.IsNull() {
		params.Set("subscription_id", state.Filters.SubscriptionID.ValueString())
	}
	if !state.Filters.Location.IsNull() {
		params.Set("location", state.Filters.Location.ValueString())
	}

	listResp, err := d.client.ListWithFilters(ctx, id, URL_AZURE_VAULTS, params)
	if err != nil {
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[data_source_azure_vaults_list.go -> Read]["+id+"]")
		resp.Diagnostics.AddError("Error Reading CipherTrust Azure Vaults", "Error listing Azure vaults: "+err.Error())
		return
	}

	vaults := []AzureVaultDataSourceTFSDK{}
	gjson.Get(listResp, "resources").ForEach(func(_, v gjson.Result) bool {
		var vault AzureVaultDataSourceTFSDK
		hydrateAzureVaultDataSource(ctx, v.Raw, &vault, &resp.Diagnostics)
		vaults = append(vaults, vault)
		return true
	})
	state.Vaults = vaults

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// hydrateAzureVaultDataSource populates all fields of AzureVaultDataSourceTFSDK from the CM API response.
func hydrateAzureVaultDataSource(ctx context.Context, resp string, state *AzureVaultDataSourceTFSDK, diags *diag.Diagnostics) {
	state.ID          = types.StringValue(gjson.Get(resp, "id").String())
	state.URI         = types.StringValue(gjson.Get(resp, "uri").String())
	state.Account     = types.StringValue(gjson.Get(resp, "account").String())
	state.Application = types.StringValue(gjson.Get(resp, "application").String())
	state.DevAccount  = types.StringValue(gjson.Get(resp, "devAccount").String())
	state.CreatedAt   = types.StringValue(gjson.Get(resp, "createdAt").String())
	state.UpdatedAt   = types.StringValue(gjson.Get(resp, "updatedAt").String())
	state.SyncedAt    = types.StringValue(gjson.Get(resp, "synced_at").String())
	state.Name        = types.StringValue(gjson.Get(resp, "name").String())

	state.AzureVaultID     = types.StringValue(gjson.Get(resp, "azure_vault_id").String())
	state.AzureName        = types.StringValue(gjson.Get(resp, "azure_name").String())
	state.CloudName        = types.StringValue(gjson.Get(resp, "cloud_name").String())
	state.Location         = types.StringValue(gjson.Get(resp, "location").String())
	state.Type             = types.StringValue(gjson.Get(resp, "type").String())
	state.SubscriptionID   = types.StringValue(gjson.Get(resp, "subscription_id").String())
	state.SubscriptionName = types.StringValue(gjson.Get(resp, "subscription_name").String())
	state.VaultConnection  = types.StringValue(gjson.Get(resp, "connection").String())
	state.ConnectionName   = types.StringValue(gjson.Get(resp, "connection_name").String())

	state.CloudKeyBackupLimit = types.Int64Value(gjson.Get(resp, "cloud_key_backup_limit").Int())

	if r := gjson.Get(resp, "tags"); r.Exists() && r.Type != gjson.Null {
		tagMap := make(map[string]string)
		r.ForEach(func(k, v gjson.Result) bool { tagMap[k.String()] = v.String(); return true })
		mv, d := types.MapValueFrom(ctx, types.StringType, tagMap)
		diags.Append(d...)
		state.Tags = mv
	} else {
		state.Tags = types.MapNull(types.StringType)
	}

	if r := gjson.Get(resp, "labels"); r.Exists() && r.Type != gjson.Null {
		labelMap := make(map[string]string)
		r.ForEach(func(k, v gjson.Result) bool { labelMap[k.String()] = v.String(); return true })
		mv, d := types.MapValueFrom(ctx, types.StringType, labelMap)
		diags.Append(d...)
		state.Labels = mv
	} else {
		state.Labels = types.MapNull(types.StringType)
	}

	state.TenantID   = types.StringValue(gjson.Get(resp, "properties.tenantId").String())
	state.VaultURI   = types.StringValue(gjson.Get(resp, "properties.vaultUri").String())
	state.CreateMode = types.StringValue(gjson.Get(resp, "properties.createMode").String())

	state.EnableSoftDelete              = types.BoolValue(gjson.Get(resp, "properties.enableSoftDelete").Bool())
	state.EnablePurgeProtection        = types.BoolValue(gjson.Get(resp, "properties.enablePurgeProtection").Bool())
	state.EnabledForDeployment         = types.BoolValue(gjson.Get(resp, "properties.enabledForDeployment").Bool())
	state.EnabledForDiskEncryption     = types.BoolValue(gjson.Get(resp, "properties.enabledForDiskEncryption").Bool())
	state.EnabledForTemplateDeployment = types.BoolValue(gjson.Get(resp, "properties.enabledForTemplateDeployment").Bool())
	state.EnableRbacAuthorization      = types.BoolValue(gjson.Get(resp, "properties.enableRbacAuthorization").Bool())
	state.SoftDeleteRetentionInDays    = types.Int64Value(gjson.Get(resp, "properties.softDeleteRetentionInDays").Int())

	state.Sku = AzureSkuTFSDK{
		Name:   types.StringValue(gjson.Get(resp, "properties.sku.name").String()),
		Family: types.StringValue(gjson.Get(resp, "properties.sku.family").String()),
	}
}

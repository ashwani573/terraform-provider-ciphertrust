package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tidwall/gjson"

	"github.com/ThalesGroup/terraform-provider-ciphertrust/internal/provider/cckm/acls"
	"github.com/ThalesGroup/terraform-provider-ciphertrust/internal/provider/common"
	"github.com/ThalesGroup/terraform-provider-ciphertrust/internal/provider/modifiers"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

// Ensure resourceAzureVault implements resource.Resource and resource.ResourceWithImportState.
var (
	_ resource.Resource                = &resourceAzureVault{}
	_ resource.ResourceWithImportState = &resourceAzureVault{}
)

type resourceAzureVault struct {
	client *common.Client
}

func NewResourceAzureVault() resource.Resource {
	return &resourceAzureVault{}
}

func (r *resourceAzureVault) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_vault"
}

func (r *resourceAzureVault) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*common.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *common.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *resourceAzureVault) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure Key Vault or Managed HSM registered in CipherTrust CCKM.",
		Attributes: map[string]schema.Attribute{
			// --- Computed-only (server-assigned) ---
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"uri": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"account": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"application": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dev_account": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of last update, as returned by CM. Changes on every PATCH.",
			},
			"synced_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of last synchronization with Azure, as returned by CM. Changes on every sync.",
			},
			"name": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"azure_vault_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"azure_name": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cloud_name": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"location": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subscription_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subscription_name": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"vault_connection": schema.StringAttribute{
				Computed:      true,
				Description:   "CM-assigned connection UUID returned by GET /vaults/{id}.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"connection_name": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.MapAttribute{
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
			"labels": schema.MapAttribute{
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			},
			"cloud_key_backup_limit": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"vault_uri": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enable_soft_delete": schema.BoolAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"enable_purge_protection": schema.BoolAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"soft_delete_retention_in_days": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"enabled_for_deployment": schema.BoolAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"enabled_for_disk_encryption": schema.BoolAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"enabled_for_template_deployment": schema.BoolAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"create_mode": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enable_rbac_authorization": schema.BoolAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"sku": schema.SingleNestedAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"name":   schema.StringAttribute{Computed: true},
					"family": schema.StringAttribute{Computed: true},
				},
			},
			// --- Write-only / Optional ---
			"enable_success_audit_event": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable success audit events for the vault. Preserved in state; not returned by CM GET.",
			},
			"discover_only": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, vault is discovered but not actively managed. Preserved in state; not returned by CM GET.",
			},
			"acls": acls.AclsSchema(),
			"connection_id": schema.StringAttribute{
				Required:    true,
				Description: "(Immutable) Azure connection ID used to register this vault in CCKM.",
				PlanModifiers: []planmodifier.String{modifiers.ImmutableString()},
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"managed_hsm": schema.BoolAttribute{
				Optional:    true,
				Description: "(Immutable) If true, discover from get-managed-hsms instead of get-vaults.",
				PlanModifiers: []planmodifier.Bool{modifiers.ImmutableBool()},
			},
			"enable_rotation": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable key rotation job for the vault. Removing this attribute when previously true disables rotation on CM.",
			},
			"rotation_job_params": schema.StringAttribute{
				Optional:    true,
				Description: "JSON string body for the enable-rotation-job request.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(2)},
			},
			"azure_vault_name": schema.StringAttribute{
				Required:    true,
				Description: "(Immutable) Name of the Azure vault to register.",
				PlanModifiers: []planmodifier.String{modifiers.ImmutableString()},
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
		},
	}
}

func (r *resourceAzureVault) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	id := uuid.New().String()
	tflog.Trace(ctx, common.MSG_METHOD_START+"[resource_azure_vault.go -> Create]["+id+"]")
	defer tflog.Trace(ctx, common.MSG_METHOD_END+"[resource_azure_vault.go -> Create]["+id+"]")

	var plan AzureVaultTFSDK
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: resolve subscription ID.
	subPayload, _ := json.Marshal(map[string]string{"connection": plan.ConnectionID.ValueString()})
	subResp, err := r.client.PostDataV2(ctx, id, URL_AZURE_GET_SUBSCRIPTIONS, subPayload)
	if err != nil {
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Create]["+id+"]")
		resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault", "Error getting Azure subscriptions: "+err.Error())
		return
	}
	subID := gjson.Get(subResp, subIDPath).String()
	if subID == "" {
		tflog.Debug(ctx, common.ERR_METHOD_END+"no subscriptionId in response"+"[resource_azure_vault.go -> Create]["+id+"]")
		resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault",
			fmt.Sprintf("get-subscriptions did not return a value at gjson path %q — verify subIDPath constant", subIDPath))
		return
	}

	// Step 2: discover vault entry by name.
	discoverEndpoint := URL_AZURE_GET_VAULTS
	if !plan.ManagedHsm.IsNull() && plan.ManagedHsm.ValueBool() {
		discoverEndpoint = URL_AZURE_GET_MANAGED_HSMS
	}
	discPayload, _ := json.Marshal(map[string]interface{}{
		"subscription_id": subID,
		"connection":      plan.ConnectionID.ValueString(),
	})
	discResp, err := r.client.PostDataV2(ctx, id, discoverEndpoint, discPayload)
	if err != nil {
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Create]["+id+"]")
		resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault", "Error discovering Azure vaults: "+err.Error())
		return
	}
	var vaultEntry gjson.Result
	gjson.Get(discResp, "resources").ForEach(func(_, v gjson.Result) bool {
		if v.Get("name").String() == plan.AzureVaultName.ValueString() {
			vaultEntry = v
			return false
		}
		return true
	})
	if !vaultEntry.Exists() {
		tflog.Debug(ctx, common.ERR_METHOD_END+"vault not found in discovery response"+"[resource_azure_vault.go -> Create]["+id+"]")
		resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault",
			fmt.Sprintf("Azure vault %q not found in subscription %s via %s", plan.AzureVaultName.ValueString(), subID, discoverEndpoint))
		return
	}

	// Step 3: register vault in CCKM.
	// cloud_name is derived from the discovered vault entry to satisfy CCKMAddContainersParams.cloud_name [required].
	cloudName := vaultEntry.Get("cloud_name").String()
	addMap := map[string]interface{}{
		"subscription_id": subID,
		"connection":      plan.ConnectionID.ValueString(),
		"cloud_name":      cloudName,
		"vaults":          []interface{}{json.RawMessage(vaultEntry.Raw)},
	}
	if !plan.EnableSuccessAuditEvent.IsNull() && !plan.EnableSuccessAuditEvent.IsUnknown() {
		addMap["enable_success_audit_event"] = plan.EnableSuccessAuditEvent.ValueBool()
	}
	if !plan.DiscoverOnly.IsNull() && !plan.DiscoverOnly.IsUnknown() {
		addMap["discover_only"] = plan.DiscoverOnly.ValueBool()
	}
	addPayload, _ := json.Marshal(addMap)
	addResp, err := r.client.PostDataV2(ctx, id, URL_AZURE_ADD_VAULTS, addPayload)
	if err != nil {
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Create]["+id+"]")
		resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault", "Error adding Azure vault: "+err.Error())
		return
	}
	vaultID := gjson.Get(addResp, addVaultIDPath).String()
	if vaultID == "" {
		tflog.Debug(ctx, common.ERR_METHOD_END+"no vault ID in add-vaults response"+"[resource_azure_vault.go -> Create]["+id+"]")
		resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault",
			fmt.Sprintf("add-vaults did not return a value at gjson path %q — verify addVaultIDPath constant", addVaultIDPath))
		return
	}

	// Step 4: GET to hydrate all Computed fields.
	getResp, err := r.client.GetById(ctx, id, vaultID, URL_AZURE_VAULTS)
	if err != nil {
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Create]["+id+"]")
		resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault", "Error reading Azure vault after create: "+err.Error())
		return
	}
	hydrateAzureVaultState(ctx, getResp, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 5: ACLs — sent on create if configured.
	if !plan.Acls.IsNull() && !plan.Acls.IsUnknown() {
		aclPayload, diags := buildAclPayload(ctx, plan.Acls)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		aclURL := URL_AZURE_VAULTS + "/" + vaultID + "/update-acls"
		_, err = r.client.PostDataV2(ctx, id, aclURL, aclPayload)
		if err != nil {
			tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Create]["+id+"]")
			resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault", "Error setting ACLs on Azure vault: "+err.Error())
			return
		}
	}

	// Step 6: Rotation — enable on create if requested.
	if !plan.EnableRotation.IsNull() && plan.EnableRotation.ValueBool() {
		rotURL := URL_AZURE_VAULTS + "/" + vaultID + "/enable-rotation-job"
		rotBody := []byte("{}")
		if !plan.RotationJobParams.IsNull() {
			rotBody = []byte(plan.RotationJobParams.ValueString())
		}
		_, err = r.client.PostDataV2(ctx, id, rotURL, rotBody)
		if err != nil {
			tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Create]["+id+"]")
			resp.Diagnostics.AddError("Error Creating CipherTrust Azure Vault", "Error enabling rotation job on Azure vault: "+err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAzureVault) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	id := uuid.New().String()
	tflog.Trace(ctx, common.MSG_METHOD_START+"[resource_azure_vault.go -> Read]["+id+"]")
	defer tflog.Trace(ctx, common.MSG_METHOD_END+"[resource_azure_vault.go -> Read]["+id+"]")

	var state AzureVaultTFSDK
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.client.GetById(ctx, id, state.ID.ValueString(), URL_AZURE_VAULTS)
	if err != nil {
		if strings.Contains(err.Error(), "status: 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Read]["+id+"]")
		resp.Diagnostics.AddError("Error Reading CipherTrust Azure Vault",
			"Could not read Azure vault "+state.ID.ValueString()+": "+err.Error())
		return
	}

	hydrateAzureVaultState(ctx, getResp, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAzureVault) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	id := uuid.New().String()
	tflog.Trace(ctx, common.MSG_METHOD_START+"[resource_azure_vault.go -> Update]["+id+"]")
	defer tflog.Trace(ctx, common.MSG_METHOD_END+"[resource_azure_vault.go -> Update]["+id+"]")

	var plan, state AzureVaultTFSDK
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vaultID := state.ID.ValueString()

	patchMap := map[string]interface{}{
		"connection": plan.ConnectionID.ValueString(),
	}
	if !plan.EnableSuccessAuditEvent.IsNull() && !plan.EnableSuccessAuditEvent.IsUnknown() {
		patchMap["enable_success_audit_event"] = plan.EnableSuccessAuditEvent.ValueBool()
	}
	if !plan.DiscoverOnly.IsNull() && !plan.DiscoverOnly.IsUnknown() {
		patchMap["discover_only"] = plan.DiscoverOnly.ValueBool()
	}
	patchBody, _ := json.Marshal(patchMap)

	_, err := r.client.UpdateData(ctx, vaultID, URL_AZURE_VAULTS, patchBody, "updatedAt")
	if err != nil {
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Update]["+id+"]")
		resp.Diagnostics.AddError("Error Updating CipherTrust Azure Vault",
			"Could not update Azure vault "+vaultID+": "+err.Error())
		return
	}

	// ACL update — only when acls changed AND plan acls is non-null.
	// When plan.Acls is null, ACLs are NOT cleared on CM; preserved in TF state only (documented limitation).
	if !plan.Acls.IsNull() && !plan.Acls.Equal(state.Acls) {
		aclPayload, diags := buildAclPayload(ctx, plan.Acls)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		aclURL := URL_AZURE_VAULTS + "/" + vaultID + "/update-acls"
		_, err = r.client.PostDataV2(ctx, id, aclURL, aclPayload)
		if err != nil {
			tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Update]["+id+"]")
			resp.Diagnostics.AddError("Error Updating CipherTrust Azure Vault",
				"Error updating ACLs on Azure vault "+vaultID+": "+err.Error())
			return
		}
	}

	// Rotation toggle.
	planRot := !plan.EnableRotation.IsNull() && plan.EnableRotation.ValueBool()
	stateRot := !state.EnableRotation.IsNull() && state.EnableRotation.ValueBool()
	if planRot != stateRot {
		if planRot {
			rotURL := URL_AZURE_VAULTS + "/" + vaultID + "/enable-rotation-job"
			rotBody := []byte("{}")
			if !plan.RotationJobParams.IsNull() {
				rotBody = []byte(plan.RotationJobParams.ValueString())
			}
			_, err = r.client.PostDataV2(ctx, id, rotURL, rotBody)
			if err != nil {
				tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Update]["+id+"]")
				resp.Diagnostics.AddError("Error Updating CipherTrust Azure Vault",
					"Error enabling rotation job on Azure vault "+vaultID+": "+err.Error())
				return
			}
		} else {
			disURL := URL_AZURE_VAULTS + "/" + vaultID + "/disable-rotation-job"
			_, err = r.client.PostDataV2(ctx, id, disURL, []byte("{}"))
			if err != nil {
				tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Update]["+id+"]")
				resp.Diagnostics.AddError("Error Updating CipherTrust Azure Vault",
					"Error disabling rotation job on Azure vault "+vaultID+": "+err.Error())
				return
			}
		}
	}

	// Preserve write-only Saved=No Optional fields not returned by CM GET.
	if plan.Acls.IsNull() {
		plan.Acls = state.Acls
	}
	if plan.RotationJobParams.IsNull() {
		plan.RotationJobParams = state.RotationJobParams
	}
	if plan.EnableSuccessAuditEvent.IsNull() {
		plan.EnableSuccessAuditEvent = state.EnableSuccessAuditEvent
	}
	if plan.DiscoverOnly.IsNull() {
		plan.DiscoverOnly = state.DiscoverOnly
	}
	// enable_rotation: null plan value means "disable rotation" (already acted on above).
	// Null in state after disable is the intended terminal state.

	// Read back to hydrate Computed fields after PATCH.
	getResp, err := r.client.GetById(ctx, id, vaultID, URL_AZURE_VAULTS)
	if err != nil {
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Update]["+id+"]")
		resp.Diagnostics.AddError("Error Updating CipherTrust Azure Vault",
			"Error reading Azure vault after update: "+err.Error())
		return
	}
	hydrateAzureVaultState(ctx, getResp, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAzureVault) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	id := uuid.New().String()
	tflog.Trace(ctx, common.MSG_METHOD_START+"[resource_azure_vault.go -> Delete]["+id+"]")
	defer tflog.Trace(ctx, common.MSG_METHOD_END+"[resource_azure_vault.go -> Delete]["+id+"]")

	var state AzureVaultTFSDK
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteURL := URL_AZURE_VAULTS + "/" + state.ID.ValueString() + "/remove-vault"
	_, err := r.client.PostDataV2(ctx, id, deleteURL, []byte("{}"))
	if err != nil {
		if strings.Contains(err.Error(), "status: 404") {
			return
		}
		tflog.Debug(ctx, common.ERR_METHOD_END+err.Error()+"[resource_azure_vault.go -> Delete]["+id+"]")
		resp.Diagnostics.AddError("Error Deleting CipherTrust Azure Vault",
			"Could not delete Azure vault "+state.ID.ValueString()+": "+err.Error())
	}
}

func (r *resourceAzureVault) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Import Not Supported",
		"ciphertrust_azure_vault does not support import. The 'connection_id' and 'azure_vault_name' "+
			"fields are write-only and cannot be recovered from the CipherTrust API. "+
			"Recreate the resource using Terraform instead.",
	)
}

// hydrateAzureVaultState populates all Computed-only (Saved=Yes) fields from the CM API response.
// Write-only fields (ConnectionID, ManagedHsm, AzureVaultName, EnableRotation, RotationJobParams,
// EnableSuccessAuditEvent, DiscoverOnly, Acls) are never assigned here.
func hydrateAzureVaultState(ctx context.Context, resp string, state *AzureVaultTFSDK, diags *diag.Diagnostics) {
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

// buildAclPayload converts the ACL set to a JSON payload for the update-acls endpoint.
func buildAclPayload(ctx context.Context, aclSet types.Set) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	var rawElems []acls.AclTFSDK
	diags.Append(aclSet.ElementsAs(ctx, &rawElems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	type aclPayloadItem struct {
		UserID  string   `json:"user_id,omitempty"`
		Group   string   `json:"group,omitempty"`
		Actions []string `json:"actions"`
	}

	items := make([]aclPayloadItem, 0, len(rawElems))
	for _, acl := range rawElems {
		var ops []string
		for _, el := range acl.Actions.Elements() {
			if sv, ok := el.(types.String); ok {
				ops = append(ops, sv.ValueString())
			}
		}
		items = append(items, aclPayloadItem{
			UserID:  acl.UserID.ValueString(),
			Group:   acl.Group.ValueString(),
			Actions: ops,
		})
	}
	payload, err := json.Marshal(map[string]interface{}{"acls": items})
	if err != nil {
		diags.AddError("Error Marshaling ACL Payload", err.Error())
		return nil, diags
	}
	return payload, diags
}

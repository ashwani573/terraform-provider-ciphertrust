package acls

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ContainerAclJSON struct {
	UserID  string   `json:"user_id,omitempty"`
	Group   string   `json:"group,omitempty"`
	Permit  bool     `json:"permit"`
	Actions []string `json:"actions"`
}

type BaseAclsJSON struct {
	ContainerAcls []ContainerAclJSON `json:"acls"`
}

type AclTFSDK struct {
	UserID  types.String `tfsdk:"user_id"`
	Group   types.String `tfsdk:"group"`
	Actions types.Set    `tfsdk:"actions"`
}

var AclAttributes = map[string]attr.Type{
	"user_id": types.StringType,
	"group":   types.StringType,
	"actions": types.SetType{ElemType: types.StringType},
}

// AclsSchema returns the standard Optional ACL set attribute for resource schemas.
func AclsSchema() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Optional:    true,
		Description: "Access control list entries.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"user_id": schema.StringAttribute{
					Optional:    true,
					Description: "CipherTrust Manager user ID.",
				},
				"group": schema.StringAttribute{
					Optional:    true,
					Description: "CipherTrust Manager group.",
				},
				"actions": schema.SetAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "Permitted actions.",
				},
			},
		},
	}
}

// AclElemType returns the Terraform object type for a single ACL entry.
func AclElemType() types.ObjectType {
	return types.ObjectType{AttrTypes: AclAttributes}
}

type AclJSON struct {
	UserID  string   `json:"user_id"`
	Group   string   `json:"group"`
	Actions []string `json:"actions"`
}

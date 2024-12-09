package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type containerResourceModel struct {
	Name         types.String `tfsdk:"name"`
	Image        types.String `tfsdk:"image"`
	EnvFile      types.String `tfsdk:"env_file"`
	Command      types.String `tfsdk:"command"`
	PreCondition types.String `tfsdk:"pre_condition"`
}

type registryAuthResourceModel struct {
	Server   types.String `tfsdk:"server"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

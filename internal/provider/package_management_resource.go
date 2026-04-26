package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &packageManagementResourceModel{}
	_ resource.Resource                     = &packageManagementResource{}
	_ resource.ResourceWithConfigure        = &packageManagementResource{}
	_ resource.ResourceWithConfigValidators = &packageManagementResource{}
	_ resource.ResourceWithImportState      = &packageManagementResource{}
)

// NewPackageManagementResource is a helper function to simplify the provider implementation.
func NewPackageManagementResource() resource.Resource {
	return &packageManagementResource{
		configurationResource: configurationResource{
			resourceBase: newResourceBase(config.PackageManagementBundle),
			modelFactory: func() any {
				return new(packageManagementResourceModel)
			},
		},
	}
}

type packageManagementResource struct {
	configurationResource
}

func (r *packageManagementResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return append(
		r.configurationResource.ConfigValidators(ctx),
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("full_upgrade"),
			path.MatchRoot("packages"),
		),
	)
}

// Schema defines the schema for the resource.
func (r *packageManagementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "PackageManagement controls system packages.",
		Attributes: map[string]schema.Attribute{
			"tag": schema.StringAttribute{
				Optional:      true,
				Description:   "The tag for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"node": schema.StringAttribute{
				Optional:      true,
				Description:   "The node for which to set the configuration. Either tag or node is required.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"extend": schema.BoolAttribute{
				Required: true,
				Description: "If the configuration should extend configuration from the parent nodes of the node " +
					"the configuration is applied to. If set to false, configuration from parent nodes is ignored.",
			},
			"pre_condition": schema.StringAttribute{
				Optional:    true,
				Description: "If set, will be executed before package maintenance. If the command returns a non-zero exit code, the package maintenance will be skipped.",
			},
			"reboot_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Defines whether the system should be rebooted after package maintenance or not.",
				Default:     stringdefault.StaticString("never"),
				Validators: []validator.String{
					stringvalidator.OneOf("never", "always"),
				},
			},
			"full_upgrade": schema.BoolAttribute{
				Optional:    true,
				Description: "If set to true, will perform a full system upgrade.",
			},
			"packages": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of packages to be maintained.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the package to be maintained.",
						},
						"version": schema.StringAttribute{
							Required:    true,
							Description: "Version of the package to be maintained.",
						},
					},
				},
			},
		},
	}
}

type packageManagementResourceModel struct {
	configurationResourceModel
	PreCondition types.String     `tfsdk:"pre_condition"`
	RebootMode   types.String     `tfsdk:"reboot_mode"`
	FullUpgrade  types.Bool       `tfsdk:"full_upgrade"`
	Packages     []managedPackage `tfsdk:"packages"`
}

type managedPackage struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

func (m packageManagementResourceModel) getConfigBundle() config.Bundle {
	return config.PackageManagementBundle
}

// Read refreshes the Terraform state with the latest data.
func (m *packageManagementResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.PackageManagement
	if data == nil {
		return fmt.Errorf("package_management bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Extend)

	m.RebootMode = types.StringValue(string(data.RebootMode))

	if data.PreCondition != "" {
		m.PreCondition = types.StringValue(data.PreCondition)
	}

	if data.Packages != nil {
		for _, p := range data.Packages {
			m.Packages = append(m.Packages, managedPackage{
				Name:    types.StringValue(p.Name),
				Version: types.StringValue(p.Version),
			})
		}
	} else {
		m.FullUpgrade = types.BoolValue(data.FullUpgrade)
	}

	return nil
}

func (m packageManagementResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.PackageManagement{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	bundleData.RebootMode = config.RebootMode(m.RebootMode.ValueString())

	if !m.PreCondition.IsNull() {
		bundleData.PreCondition = m.PreCondition.ValueString()
	}

	if m.Packages != nil {
		for _, p := range m.Packages {
			bundleData.Packages = append(bundleData.Packages, config.Package{
				Name:    p.Name.ValueString(),
				Version: p.Version.ValueString(),
			})
		}
	} else {
		bundleData.FullUpgrade = m.FullUpgrade.ValueBool()
	}

	return bundleData
}

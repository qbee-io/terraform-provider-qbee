package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client/config"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resourceModelManager                  = &softwareManagementResourceModel{}
	_ resource.Resource                     = &softwaremanagementResource{}
	_ resource.ResourceWithConfigure        = &softwaremanagementResource{}
	_ resource.ResourceWithConfigValidators = &softwaremanagementResource{}
	_ resource.ResourceWithImportState      = &softwaremanagementResource{}
)

func NewSoftwareManagementResource() resource.Resource {
	return &softwaremanagementResource{
		configurationResource: configurationResource{
			// for backwards compatibility, use softwaremanagement instead of config.SoftwareManagementBundle
			resourceBase: newResourceBase("softwaremanagement"),
			modelFactory: func() any {
				return new(softwareManagementResourceModel)
			},
		},
	}
}

type softwaremanagementResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *softwaremanagementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SoftwareManagement controls software in the system.",
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
			"items": schema.ListNestedAttribute{
				Required:    true,
				Description: "The filesets that must be distributed",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"package": schema.StringAttribute{
							Required: true,
							Description: "Package name (with .deb) from package in File manager or package name " +
								"(without .deb ending) to install it from a apt repository configured on the device " +
								"(e.g. mc for midnight commander will install from repository)",
						},
						"service_name": schema.StringAttribute{
							Optional: true,
							Description: "Define a service name if it differs from the package name. " +
								"If empty then service name will be assumed to be the same as the package name",
						},
						"pre_condition": schema.StringAttribute{
							Optional: true,
							Description: "Script/executable that needs to return successfully before software " +
								"package is installed. We expect 0 or true. We assume true if left empty. For " +
								"example, call: /bin/true or finish with exit(0)",
						},
						"config_files": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"template": schema.StringAttribute{
										Required:    true,
										Description: "The source of the file. Must correspond to a file in the qbee filemanager.",
									},
									"location": schema.StringAttribute{
										Required:    true,
										Description: "The destination of the file on the target device.",
									},
								},
							},
						},
						"parameters": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Required: true,
									},
									"value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type softwareManagementResourceModel struct {
	configurationResourceModel
	Items []softwareManagementItemModel `tfsdk:"items"`
}

type softwareManagementItemModel struct {
	Package      types.String                      `tfsdk:"package"`
	ServiceName  types.String                      `tfsdk:"service_name"`
	PreCondition types.String                      `tfsdk:"pre_condition"`
	ConfigFiles  []softwareManagementItemFile      `tfsdk:"config_files"`
	Parameters   []softwareManagementItemParameter `tfsdk:"parameters"`
}

type softwareManagementItemFile struct {
	Template types.String `tfsdk:"template"`
	Location types.String `tfsdk:"location"`
}

type softwareManagementItemParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (m softwareManagementResourceModel) getConfigBundle() config.Bundle {
	return config.SoftwareManagementBundle
}

func (m *softwareManagementResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.SoftwareManagement
	if data == nil {
		return fmt.Errorf("softwaremanagement bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, softwarePackage := range data.Items {
		softwarePackageModel := softwareManagementItemModel{
			Package:      nullableStringValue(softwarePackage.Package),
			ServiceName:  nullableStringValue(softwarePackage.ServiceName),
			PreCondition: nullableStringValue(softwarePackage.PreCondition),
		}

		for _, configFile := range softwarePackage.ConfigFiles {
			softwarePackageModel.ConfigFiles = append(softwarePackageModel.ConfigFiles, softwareManagementItemFile{
				Template: nullableStringValue(configFile.ConfigTemplate),
				Location: nullableStringValue(configFile.ConfigLocation),
			})
		}

		for _, parameter := range softwarePackage.Parameters {
			softwarePackageModel.Parameters = append(softwarePackageModel.Parameters, softwareManagementItemParameter{
				Key:   nullableStringValue(parameter.Key),
				Value: nullableStringValue(parameter.Value),
			})
		}

		m.Items = append(m.Items, softwarePackageModel)
	}

	return nil
}

func (m softwareManagementResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.SoftwareManagement{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, item := range m.Items {
		softwarePackage := config.SoftwarePackage{
			Package:      item.Package.ValueString(),
			ServiceName:  item.ServiceName.ValueString(),
			PreCondition: item.PreCondition.ValueString(),
		}

		for _, model := range item.ConfigFiles {
			softwarePackage.ConfigFiles = append(softwarePackage.ConfigFiles, config.ConfigurationFile{
				ConfigTemplate: model.Template.ValueString(),
				ConfigLocation: model.Location.ValueString(),
			})
		}

		for _, model := range item.Parameters {
			softwarePackage.Parameters = append(softwarePackage.Parameters, config.ConfigurationFileParameter{
				Key:   model.Key.ValueString(),
				Value: model.Value.ValueString(),
			})
		}

		bundleData.Items = append(bundleData.Items, softwarePackage)
	}

	return bundleData
}

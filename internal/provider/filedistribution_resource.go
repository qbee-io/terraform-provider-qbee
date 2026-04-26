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
	_ resourceModelManager                  = &filedistributionResourceModel{}
	_ resource.Resource                     = &filedistributionResource{}
	_ resource.ResourceWithConfigure        = &filedistributionResource{}
	_ resource.ResourceWithConfigValidators = &filedistributionResource{}
	_ resource.ResourceWithImportState      = &filedistributionResource{}
)

func NewFiledistributionResource() resource.Resource {
	return &filedistributionResource{
		configurationResource: configurationResource{
			// for backwards compatibility, use filedistribution instead of config.FileDistributionBundle
			resourceBase: newResourceBase("filedistribution"),
			modelFactory: func() any {
				return new(filedistributionResourceModel)
			},
		},
	}
}

type filedistributionResource struct {
	configurationResource
}

// Schema defines the schema for the resource.
func (r *filedistributionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Defines a file set to be maintained in the system.",
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
			"files": schema.ListNestedAttribute{
				Required:    true,
				Description: "The filesets to distribute.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"label": schema.StringAttribute{
							Optional:    true,
							Description: "An optional label for the fileset.",
						},
						"templates": schema.ListNestedAttribute{
							Required:    true,
							Description: "Defines files to be created in the filesystem.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"source": schema.StringAttribute{
										Required: true,
										Description: "The source of the file. Must correspond to a file in the qbee " +
											"filemanager.",
									},
									"destination": schema.StringAttribute{
										Required:    true,
										Description: "The destination of the file on the target device.",
									},
									"is_template": schema.BoolAttribute{
										Required: true,
										Description: "If this file is a template. If set to true, template " +
											"substitution of '\\{\\{ pattern \\}\\}' will be performed in the file contents, " +
											"using the parameters defined in this filedistribution config.",
									},
								},
							},
						},
						"parameters": schema.ListNestedAttribute{
							Optional:    true,
							Description: "Define values to be used for template files.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Description: "Key of the parameter used in files.",
										Required:    true,
									},
									"value": schema.StringAttribute{
										Description: "Value of the parameter which will replace Key placeholders.",
										Required:    true,
									},
								},
							},
						},
						"pre_condition": schema.StringAttribute{
							Optional: true,
							Description: "A command that must successfully execute on the device (return a non-zero " +
								"exit code) before this fileset can be distributed. Example: `/bin/true`.",
						},
						"command": schema.StringAttribute{
							Optional: true,
							Description: "A command that will be run on the device after this fileset is " +
								"distributed. Example: `/bin/true`.",
						},
					},
				},
			},
		},
	}
}

type filedistributionResourceModel struct {
	configurationResourceModel
	Files []file `tfsdk:"files"`
}

type file struct {
	Label        types.String        `tfsdk:"label"`
	Command      types.String        `tfsdk:"command"`
	PreCondition types.String        `tfsdk:"pre_condition"`
	Templates    []template          `tfsdk:"templates"`
	Parameters   []templateParameter `tfsdk:"parameters"`
}

func (f file) toBundleData() config.FileSet {
	var files []config.File
	for _, t := range f.Templates {
		files = append(files, config.File{
			Source:      t.Source.ValueString(),
			Destination: t.Destination.ValueString(),
			IsTemplate:  t.IsTemplate.ValueBool(),
		})
	}

	var templateParameters []config.TemplateParameter
	for _, p := range f.Parameters {
		templateParameters = append(templateParameters, config.TemplateParameter{
			Key:   p.Key.ValueString(),
			Value: p.Value.ValueString(),
		})
	}

	return config.FileSet{
		Label:              f.Label.ValueString(),
		AfterCommand:       f.Command.ValueString(),
		PreCondition:       f.PreCondition.ValueString(),
		Files:              files,
		TemplateParameters: templateParameters,
	}
}

type templateParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type template struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	IsTemplate  types.Bool   `tfsdk:"is_template"`
}

func (m filedistributionResourceModel) getConfigBundle() config.Bundle {
	return config.FileDistributionBundle
}

func (m *filedistributionResourceModel) fromBundleData(bundleData config.BundleData) error {
	data := bundleData.FileDistribution
	if data == nil {
		return fmt.Errorf("file_distribution bundle data is nil")
	}

	m.Extend = types.BoolValue(data.Metadata.Extend)

	for _, fileSet := range data.FileSets {
		var templates []template
		for _, t := range fileSet.Files {
			templates = append(templates, template{
				Source:      types.StringValue(t.Source),
				Destination: types.StringValue(t.Destination),
				IsTemplate:  types.BoolValue(t.IsTemplate),
			})
		}

		var parameters []templateParameter
		for _, p := range fileSet.TemplateParameters {
			parameters = append(parameters, templateParameter{
				Key:   types.StringValue(p.Key),
				Value: types.StringValue(p.Value),
			})
		}

		var label types.String
		if fileSet.Label == "" {
			label = types.StringNull()
		} else {
			label = types.StringValue(fileSet.Label)
		}

		var command types.String
		if fileSet.AfterCommand == "" {
			command = types.StringNull()
		} else {
			command = types.StringValue(fileSet.AfterCommand)
		}

		var precondition types.String
		if fileSet.PreCondition == "" {
			precondition = types.StringNull()
		} else {
			precondition = types.StringValue(fileSet.PreCondition)
		}

		m.Files = append(m.Files, file{
			Label:        label,
			Command:      command,
			PreCondition: precondition,
			Templates:    templates,
			Parameters:   parameters,
		})
	}

	return nil
}

func (m filedistributionResourceModel) toBundleData(metadata config.Metadata) any {
	bundleData := config.FileDistribution{
		Metadata: metadata,
	}

	if metadata.Reset {
		return bundleData
	}

	for _, fileSet := range m.Files {
		bundleData.FileSets = append(bundleData.FileSets, fileSet.toBundleData())
	}

	return bundleData
}

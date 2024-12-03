package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qbee-io/terraform-provider-qbee/internal/qbee"
	"strings"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &filedistributionResource{}
	_ resource.ResourceWithConfigure        = &filedistributionResource{}
	_ resource.ResourceWithConfigValidators = &filedistributionResource{}
	_ resource.ResourceWithImportState      = &filedistributionResource{}
)

func NewFiledistributionResource() resource.Resource {
	return &filedistributionResource{}
}

type filedistributionResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *filedistributionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filedistribution"
}

// Configure adds the provider configured client to the resource.
func (r *filedistributionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *filedistributionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder ID value",
			},
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
				Description: "The filesets that must be distributed",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"templates": schema.ListNestedAttribute{
							Required: true,
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

func (r *filedistributionResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("tag"),
			path.MatchRoot("node"),
		),
	}
}

type filedistributionResourceModel struct {
	Node   types.String `tfsdk:"node"`
	Tag    types.String `tfsdk:"tag"`
	ID     types.String `tfsdk:"id"`
	Extend types.Bool   `tfsdk:"extend"`
	Files  []file       `tfsdk:"files"`
}

func (m filedistributionResourceModel) typeAndIdentifier() (qbee.ConfigType, string) {
	return typeAndIdentifier(m.Tag, m.Node)
}

type file struct {
	Command      types.String `tfsdk:"command"`
	PreCondition types.String `tfsdk:"pre_condition"`
	Templates    []template   `tfsdk:"templates"`
	Parameters   []parameter  `tfsdk:"parameters"`
}

type parameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type template struct {
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	IsTemplate  types.Bool   `tfsdk:"is_template"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *filedistributionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan filedistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeFiledistribution(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *filedistributionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from the plan
	var plan filedistributionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.writeFiledistribution(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue("placeholder")

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *filedistributionResource) writeFiledistribution(ctx context.Context, plan filedistributionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	configType, identifier := plan.typeAndIdentifier()
	extend := plan.Extend.ValueBool()

	var mappedFiles []qbee.FiledistributionFile
	for _, f := range plan.Files {
		var mappedTemplates []qbee.FiledistributionTemplate
		for _, t := range f.Templates {
			mappedTemplates = append(mappedTemplates, qbee.FiledistributionTemplate{
				Source:      t.Source.ValueString(),
				Destination: t.Destination.ValueString(),
				IsTemplate:  t.IsTemplate.ValueBool(),
			})
		}

		var mappedParameters []qbee.FiledistributionParameter
		for _, p := range f.Parameters {
			mappedParameters = append(mappedParameters, qbee.FiledistributionParameter{
				Key:   p.Key.ValueString(),
				Value: p.Value.ValueString(),
			})
		}

		mappedFiles = append(mappedFiles, qbee.FiledistributionFile{
			Command:      f.Command.ValueString(),
			PreCondition: f.PreCondition.ValueString(),
			Templates:    mappedTemplates,
			Parameters:   mappedParameters,
		})
	}

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating file distribution for %v %v with %v filesets", configType.String(), identifier, len(mappedFiles)))
	createResponse, err := r.client.FileDistribution.Create(configType, identifier, mappedFiles, extend)
	if err != nil {
		diags.AddError(
			"Error creating filedistribution",
			fmt.Sprintf("Error creating a filedistribution resource with qbee: %v", err),
		)
		return diags
	}

	_, err = r.client.Configuration.Commit("terraform: create filedistribution_resource")
	if err != nil {
		err = fmt.Errorf("error creating a commit for the filedistribution: %w", err)

		err = r.client.Configuration.DeleteUncommitted(createResponse.Sha)
		if err != nil {
			err = fmt.Errorf("error deleting uncommitted filedistribution changes: %w", err)
		}

		diags.AddError(
			"Error creating firewall",
			fmt.Sprintf("Error while committing firewall change: %v", err),
		)
		return diags
	}

	return nil
}

// Read refreshes the Terraform state with the latest data.
func (r *filedistributionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state *filedistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configType, identifier := state.typeAndIdentifier()

	// Read the real status
	currentFiledistribution, err := r.client.FileDistribution.Get(configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError("Could not read filedistribution",
			"error reading the filedistribution resource: "+err.Error())

		return
	}

	if currentFiledistribution == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the current state
	var files []file
	for _, f := range currentFiledistribution.FiledistributionFiles {
		var templates []template
		for _, t := range f.Templates {
			templates = append(templates, template{
				Source:      types.StringValue(t.Source),
				Destination: types.StringValue(t.Destination),
				IsTemplate:  types.BoolValue(t.IsTemplate),
			})
		}

		var parameters []parameter
		for _, p := range f.Parameters {
			parameters = append(parameters, parameter{
				Key:   types.StringValue(p.Key),
				Value: types.StringValue(p.Value),
			})
		}

		var command types.String
		if f.Command == "" {
			command = types.StringNull()
		} else {
			command = types.StringValue(f.Command)
		}

		var precondition types.String
		if f.PreCondition == "" {
			precondition = types.StringNull()
		} else {
			precondition = types.StringValue(f.PreCondition)
		}

		files = append(files, file{
			Command:      command,
			PreCondition: precondition,
			Templates:    templates,
			Parameters:   parameters,
		})
	}

	state.ID = types.StringValue("placeholder")
	state.Extend = types.BoolValue(currentFiledistribution.Extend)
	state.Files = files

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *filedistributionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state filedistributionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	configType, identifier := state.typeAndIdentifier()
	tflog.Info(ctx, fmt.Sprintf("Deleting filedistribution for %v %v", configType.String(), identifier))

	deleteResponse, err := r.client.FileDistribution.Clear(configType, identifier)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting filedistribution",
			"could not delete filedistribution, unexpected error: "+err.Error())
		return
	}

	_, err = r.client.Configuration.Commit("terraform: create filedistribution_resource")
	if err != nil {
		resp.Diagnostics.AddError("Could not commit deletion of filedistribution",
			"error creating a commit to delete the filedistribution resource: "+err.Error())

		err = r.client.Configuration.DeleteUncommitted(deleteResponse.Sha)
		if err != nil {
			resp.Diagnostics.AddError("Could not revert uncommitted filedistribution changes",
				"error deleting uncommitted filedistribution changes: "+err.Error())
		}

		return
	}
}

func (r *filedistributionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	configType, identifier, found := strings.Cut(req.ID, ":")
	if !found || configType == "" || identifier == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: type:identifier. Got: %q", req.ID),
		)
		return

	}

	if configType == "tag" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), identifier)...)
	} else if configType == "node" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node"), identifier)...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Import type must be either 'node' or 'tag'. Got: %q", configType),
		)
		return
	}
}

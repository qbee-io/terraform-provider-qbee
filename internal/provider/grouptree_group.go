package provider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"go.qbee.io/terraform/internal/qbee"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &grouptreeGroupResource{}
	_ resource.ResourceWithConfigure   = &grouptreeGroupResource{}
	_ resource.ResourceWithImportState = &grouptreeGroupResource{}
)

func NewGrouptreeGroupResource() resource.Resource {
	return &grouptreeGroupResource{}
}

type grouptreeGroupResource struct {
	client *qbee.HttpClient
}

// Metadata returns the resource type name.
func (r *grouptreeGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_grouptree_group"
}

// Configure adds the provider configured client to the resource.
func (r *grouptreeGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*qbee.HttpClient)
}

// Schema defines the schema for the resource.
func (r *grouptreeGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "NodeId of the group",
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Title of the group",
			},
			"ancestor": schema.StringAttribute{
				Required:      true,
				Description:   "node_id of the direct ancestor of the group",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of tags to add to the group",
			},
		},
	}
}

type grouptreeGroupResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Title    types.String `tfsdk:"title"`
	Ancestor types.String `tfsdk:"ancestor"`
	Tags     types.List   `tfsdk:"tags"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *grouptreeGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from the plan
	var plan grouptreeGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	ancestor := plan.Ancestor.ValueString()
	title := plan.Title.ValueString()

	// Create the resource
	tflog.Info(ctx, fmt.Sprintf("Creating grouptree %v (title=%v), with ancestor %v", id, title, ancestor))
	err := r.client.Grouptree.Create(id, ancestor, title)
	if err != nil {
		fmt.Printf("error: %v", err)
		resp.Diagnostics.AddError("Error creating Grouptree resource", "could not create grouptree resource: "+err.Error())
		return
	}

	var tags []string
	if !plan.Tags.IsNull() {
		tags = make([]string, len(plan.Tags.Elements()))
		diags = plan.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	err = r.client.Grouptree.SetTags(id, tags)
	if err != nil {
		fmt.Printf("error while settings tags to %+v: %v", tags, err)
		resp.Diagnostics.AddError("Error creating Grouptree source", "could not set tags on resource: "+err.Error())
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *grouptreeGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state
	var state grouptreeGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	// Read the real status
	readResponse, err := r.client.Grouptree.Get(id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Grouptree resource", "could not read grouptree resource: "+err.Error())
		return
	}

	state.Title = types.StringValue(readResponse.Title)

	state.Tags, diags = types.ListValueFrom(ctx, types.StringType, readResponse.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ancestors := readResponse.Ancestors
	lastAncestor := ancestors[len(ancestors)-2]
	state.Ancestor = types.StringValue(lastAncestor)

	// Update the current state
	resp.State.Set(ctx, state)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *grouptreeGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get the current state
	var state grouptreeGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan grouptreeGroupResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	// Check if we should move the group
	oldAncestor := state.Ancestor.ValueString()
	currentAncestor := plan.Ancestor.ValueString()
	if oldAncestor != currentAncestor {
		resp.Diagnostics.AddError(
			"Error updating grouptree_group",
			"grouptree_group does not support changing ancestors (moving a group)")
		return
	}

	// Check if we should rename the group
	oldTitle := state.Title.ValueString()
	currentTitle := plan.Title.ValueString()
	if oldTitle != currentTitle {
		tflog.Info(ctx, fmt.Sprintf("Renaming grouptree %v from '%v' to '%v'", id, oldTitle, currentTitle))
		err := r.client.Grouptree.Rename(id, currentAncestor, currentTitle)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error response while renaming grouptree_group: %v", err))
			resp.Diagnostics.AddError("Could not rename group", "Error while renaming grouptree_group resource: "+err.Error())
			return
		}
	}

	// Check if we should update the tags
	oldTags := state.Tags
	newTags := plan.Tags
	if !reflect.DeepEqual(oldTags, newTags) {
		var tags []string
		if !plan.Tags.IsNull() {
			tags = make([]string, len(plan.Tags.Elements()))
			diags = plan.Tags.ElementsAs(ctx, &tags, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}

		err := r.client.Grouptree.SetTags(id, tags)
		if err != nil {
			fmt.Printf("error while settings tags to %+v: %v", tags, err)
			resp.Diagnostics.AddError("Error creating Grouptree source", "could not set tags on resource: "+err.Error())
			return
		}
	}

	resp.State.Set(ctx, plan)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *grouptreeGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from the state
	var state grouptreeGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the resource
	id := state.ID.ValueString()
	ancestor := state.Ancestor.ValueString()

	tflog.Info(ctx, fmt.Sprintf("Deleting grouptree node %v with ancestor %v", id, ancestor))
	err := r.client.Grouptree.Delete(id, ancestor)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Grouptree resource", "could not delete grouptree resource: "+err.Error())
	}
}

func (r *grouptreeGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

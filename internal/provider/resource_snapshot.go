package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &SnapshotResource{}
var _ resource.ResourceWithImportState = &SnapshotResource{}

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

type SnapshotResource struct {
	client *truenas.Client
}

type SnapshotResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Dataset   types.String `tfsdk:"dataset"`
	Name      types.String `tfsdk:"name"`
	Recursive types.Bool   `tfsdk:"recursive"`
	VMSync    types.String `tfsdk:"vmware_sync"`
	Properties types.Map   `tfsdk:"properties"`
}

func (r *SnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (r *SnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ZFS snapshot on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Snapshot identifier (dataset@name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset": schema.StringAttribute{
				MarkdownDescription: "Dataset to snapshot (e.g., tank/mydata)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Snapshot name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"recursive": schema.BoolAttribute{
				MarkdownDescription: "Create recursive snapshot of all children",
				Optional:            true,
				Computed:            true,
			},
			"vmware_sync": schema.StringAttribute{
				MarkdownDescription: "VMware sync option (NONE, CONTINUE, FAIL)",
				Optional:            true,
				Computed:            true,
			},
			"properties": schema.MapAttribute{
				MarkdownDescription: "Snapshot properties",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (r *SnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*truenas.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *truenas.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SnapshotResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"dataset": data.Dataset.ValueString(),
		"name":    data.Name.ValueString(),
	}

	if !data.Recursive.IsNull() {
		createReq["recursive"] = data.Recursive.ValueBool()
	}
	if !data.VMSync.IsNull() {
		createReq["vmware_sync"] = data.VMSync.ValueString()
	}

	respBody, err := r.client.Post("/zfs/snapshot", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create snapshot, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Snapshot ID is dataset@name
	snapshotID := fmt.Sprintf("%s@%s", data.Dataset.ValueString(), data.Name.ValueString())
	data.ID = types.StringValue(snapshotID)

	r.readSnapshot(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapshotResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSnapshot(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Snapshots are immutable, so update is not supported
	resp.Diagnostics.AddError("Update Not Supported", "Snapshots are immutable and cannot be updated")
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SnapshotResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/zfs/snapshot/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete snapshot, got error: %s", err))
		return
	}
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: dataset@snapshotname
	parts := strings.Split(req.ID, "@")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected format: dataset@snapshotname")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dataset"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func (r *SnapshotResource) readSnapshot(ctx context.Context, data *SnapshotResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/zfs/snapshot/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read snapshot, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if name, ok := result["name"].(string); ok {
		// name is in format dataset@snapshot
		parts := strings.Split(name, "@")
		if len(parts) == 2 {
			data.Dataset = types.StringValue(parts[0])
			data.Name = types.StringValue(parts[1])
			data.ID = types.StringValue(name)
		}
	}

	if properties, ok := result["properties"].(map[string]interface{}); ok {
		propsMap := make(map[string]string)
		for k, v := range properties {
			if propMap, ok := v.(map[string]interface{}); ok {
				if value, ok := propMap["value"].(string); ok {
					propsMap[k] = value
				}
			}
		}
		propTypes, _ := types.MapValueFrom(ctx, types.StringType, propsMap)
		data.Properties = propTypes
	}
}


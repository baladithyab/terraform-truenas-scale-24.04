package provider

import (
	"net/url"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DatasetResource{}
var _ resource.ResourceWithImportState = &DatasetResource{}

func NewDatasetResource() resource.Resource {
	return &DatasetResource{}
}

// DatasetResource defines the resource implementation.
type DatasetResource struct {
	client *truenas.Client
}

// DatasetResourceModel describes the resource data model.
type DatasetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Comments    types.String `tfsdk:"comments"`
	Compression types.String `tfsdk:"compression"`
	Atime       types.String `tfsdk:"atime"`
	Quota       types.Int64  `tfsdk:"quota"`
	RefQuota    types.Int64  `tfsdk:"refquota"`
	Reservation types.Int64  `tfsdk:"reservation"`
	RefReserv   types.Int64  `tfsdk:"refreservation"`
	Dedup       types.String `tfsdk:"deduplication"`
	ReadOnly    types.String `tfsdk:"readonly"`
	Exec        types.String `tfsdk:"exec"`
	Sync        types.String `tfsdk:"sync"`
	SnapDir     types.String `tfsdk:"snapdir"`
	Copies      types.Int64  `tfsdk:"copies"`
	RecordSize  types.String `tfsdk:"recordsize"`
	Volsize     types.Int64  `tfsdk:"volsize"` // Volume size in bytes (required for VOLUME type)
}

func (r *DatasetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (r *DatasetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ZFS dataset on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Dataset identifier (same as name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Full path of the dataset (e.g., pool/dataset)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Dataset type (FILESYSTEM or VOLUME)",
				Optional:            true,
				Computed:            true,
			},
			"comments": schema.StringAttribute{
				MarkdownDescription: "Comments about the dataset",
				Optional:            true,
				Computed:            true,
			},
			"compression": schema.StringAttribute{
				MarkdownDescription: "Compression algorithm (e.g., LZ4, GZIP, OFF)",
				Optional:            true,
				Computed:            true,
			},
			"atime": schema.StringAttribute{
				MarkdownDescription: "Access time updates (ON or OFF)",
				Optional:            true,
				Computed:            true,
			},
			"quota": schema.Int64Attribute{
				MarkdownDescription: "Quota in bytes (0 for unlimited)",
				Optional:            true,
				Computed:            true,
			},
			"refquota": schema.Int64Attribute{
				MarkdownDescription: "Reference quota in bytes (0 for unlimited)",
				Optional:            true,
				Computed:            true,
			},
			"reservation": schema.Int64Attribute{
				MarkdownDescription: "Reservation in bytes (0 for none)",
				Optional:            true,
				Computed:            true,
			},
			"refreservation": schema.Int64Attribute{
				MarkdownDescription: "Reference reservation in bytes (0 for none)",
				Optional:            true,
				Computed:            true,
			},
			"deduplication": schema.StringAttribute{
				MarkdownDescription: "Deduplication (ON or OFF)",
				Optional:            true,
				Computed:            true,
			},
			"readonly": schema.StringAttribute{
				MarkdownDescription: "Read-only (ON or OFF)",
				Optional:            true,
				Computed:            true,
			},
			"exec": schema.StringAttribute{
				MarkdownDescription: "Execute permissions (ON or OFF)",
				Optional:            true,
				Computed:            true,
			},
			"sync": schema.StringAttribute{
				MarkdownDescription: "Sync mode (STANDARD, ALWAYS, DISABLED)",
				Optional:            true,
				Computed:            true,
			},
			"snapdir": schema.StringAttribute{
				MarkdownDescription: "Snapshot directory visibility (VISIBLE or HIDDEN)",
				Optional:            true,
				Computed:            true,
			},
			"copies": schema.Int64Attribute{
				MarkdownDescription: "Number of copies (1-3)",
				Optional:            true,
				Computed:            true,
			},
			"recordsize": schema.StringAttribute{
				MarkdownDescription: "Record size (e.g., 128K)",
				Optional:            true,
				Computed:            true,
			},
			"volsize": schema.Int64Attribute{
				MarkdownDescription: "Volume size in bytes (required for VOLUME type datasets, not applicable to FILESYSTEM type)",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *DatasetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*truenas.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *truenas.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DatasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatasetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate volsize based on dataset type
	datasetType := "FILESYSTEM"
	if !data.Type.IsNull() {
		datasetType = data.Type.ValueString()
	}

	if datasetType == "VOLUME" && data.Volsize.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"volsize is required when type is VOLUME. Please specify the volume size in bytes.",
		)
		return
	}

	// Fix Bug 1: Check both IsNull() and IsUnknown() to prevent false positives
	if datasetType == "FILESYSTEM" && !data.Volsize.IsNull() && !data.Volsize.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid Attribute",
			"volsize is not valid for FILESYSTEM type datasets. Remove the volsize attribute or change type to VOLUME.",
		)
		return
	}

	// Build the request body
	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
		"type": datasetType,
	}

	// Properties valid for BOTH FILESYSTEM and VOLUME
	if !data.Comments.IsNull() && data.Comments.ValueString() != "" {
		createReq["comments"] = data.Comments.ValueString()
	}
	if !data.Compression.IsNull() && data.Compression.ValueString() != "" {
		createReq["compression"] = data.Compression.ValueString()
	}
	if !data.Sync.IsNull() && data.Sync.ValueString() != "" {
		createReq["sync"] = data.Sync.ValueString()
	}
	if !data.Dedup.IsNull() && data.Dedup.ValueString() != "" {
		createReq["deduplication"] = data.Dedup.ValueString()
	}
	if !data.ReadOnly.IsNull() && data.ReadOnly.ValueString() != "" {
		createReq["readonly"] = data.ReadOnly.ValueString()
	}
	if !data.Copies.IsNull() && data.Copies.ValueInt64() > 0 {
		createReq["copies"] = data.Copies.ValueInt64()
	}
	if !data.Reservation.IsNull() && data.Reservation.ValueInt64() > 0 {
		createReq["reservation"] = data.Reservation.ValueInt64()
	}
	if !data.RefReserv.IsNull() && data.RefReserv.ValueInt64() > 0 {
		createReq["refreservation"] = data.RefReserv.ValueInt64()
	}

	// VOLUME-specific properties
	if datasetType == "VOLUME" {
		if !data.Volsize.IsNull() && !data.Volsize.IsUnknown() && data.Volsize.ValueInt64() > 0 {
			createReq["volsize"] = data.Volsize.ValueInt64()
		}
	}

	// FILESYSTEM-specific properties
	if datasetType == "FILESYSTEM" {
		if !data.Atime.IsNull() && data.Atime.ValueString() != "" {
			createReq["atime"] = data.Atime.ValueString()
		}
		if !data.Exec.IsNull() && data.Exec.ValueString() != "" {
			createReq["exec"] = data.Exec.ValueString()
		}
		if !data.RecordSize.IsNull() && data.RecordSize.ValueString() != "" {
			createReq["recordsize"] = data.RecordSize.ValueString()
		}
		if !data.Quota.IsNull() && data.Quota.ValueInt64() > 0 {
			createReq["quota"] = data.Quota.ValueInt64()
		}
		if !data.RefQuota.IsNull() && data.RefQuota.ValueInt64() > 0 {
			createReq["refquota"] = data.RefQuota.ValueInt64()
		}
		if !data.SnapDir.IsNull() && data.SnapDir.ValueString() != "" {
			createReq["snapdir"] = data.SnapDir.ValueString()
		}
	}

	respBody, err := r.client.Post("/pool/dataset", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create dataset, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	data.ID = types.StringValue(data.Name.ValueString())

	// Read back the created dataset to get computed values
	r.readDataset(ctx, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatasetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readDataset(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatasetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine dataset type for conditional property updates
	datasetType := "FILESYSTEM"
	if !data.Type.IsNull() {
		datasetType = data.Type.ValueString()
	}

	// Build the update request body
	updateReq := make(map[string]interface{})

	// Properties valid for BOTH FILESYSTEM and VOLUME
	if !data.Comments.IsNull() && data.Comments.ValueString() != "" {
		updateReq["comments"] = data.Comments.ValueString()
	}
	if !data.Compression.IsNull() && data.Compression.ValueString() != "" {
		updateReq["compression"] = data.Compression.ValueString()
	}
	if !data.Sync.IsNull() && data.Sync.ValueString() != "" {
		updateReq["sync"] = data.Sync.ValueString()
	}
	if !data.Dedup.IsNull() && data.Dedup.ValueString() != "" {
		updateReq["deduplication"] = data.Dedup.ValueString()
	}
	if !data.ReadOnly.IsNull() && data.ReadOnly.ValueString() != "" {
		updateReq["readonly"] = data.ReadOnly.ValueString()
	}
	if !data.Copies.IsNull() && data.Copies.ValueInt64() > 0 {
		updateReq["copies"] = data.Copies.ValueInt64()
	}
	if !data.Reservation.IsNull() && data.Reservation.ValueInt64() > 0 {
		updateReq["reservation"] = data.Reservation.ValueInt64()
	}
	if !data.RefReserv.IsNull() && data.RefReserv.ValueInt64() > 0 {
		updateReq["refreservation"] = data.RefReserv.ValueInt64()
	}

	// VOLUME-specific properties
	if datasetType == "VOLUME" {
		if !data.Volsize.IsNull() && !data.Volsize.IsUnknown() && data.Volsize.ValueInt64() > 0 {
			updateReq["volsize"] = data.Volsize.ValueInt64()
		}
	}

	// FILESYSTEM-specific properties
	if datasetType == "FILESYSTEM" {
		if !data.Atime.IsNull() && data.Atime.ValueString() != "" {
			updateReq["atime"] = data.Atime.ValueString()
		}
		if !data.Exec.IsNull() && data.Exec.ValueString() != "" {
			updateReq["exec"] = data.Exec.ValueString()
		}
		if !data.RecordSize.IsNull() && data.RecordSize.ValueString() != "" {
			updateReq["recordsize"] = data.RecordSize.ValueString()
		}
		if !data.Quota.IsNull() && data.Quota.ValueInt64() > 0 {
			updateReq["quota"] = data.Quota.ValueInt64()
		}
		if !data.RefQuota.IsNull() && data.RefQuota.ValueInt64() > 0 {
			updateReq["refquota"] = data.RefQuota.ValueInt64()
		}
		if !data.SnapDir.IsNull() && data.SnapDir.ValueString() != "" {
			updateReq["snapdir"] = data.SnapDir.ValueString()
		}
	}

	endpoint := fmt.Sprintf("/pool/dataset/id/%s", url.PathEscape(data.ID.ValueString()))
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update dataset, got error: %s", err))
		return
	}

	// Read back the updated dataset
	r.readDataset(ctx, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatasetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/pool/dataset/id/%s", url.PathEscape(data.ID.ValueString()))
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete dataset, got error: %s", err))
		return
	}
}

func (r *DatasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to read dataset data from the API
func (r *DatasetResource) readDataset(ctx context.Context, data *DatasetResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/pool/dataset/id/%s", url.PathEscape(data.ID.ValueString()))
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read dataset, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Update the model with values from the API
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
		data.ID = types.StringValue(name)
	}

	// Determine dataset type first
	datasetType := "FILESYSTEM"
	if dtype, ok := result["type"].(string); ok {
		datasetType = dtype
		data.Type = types.StringValue(dtype)
	}

	// Read properties valid for BOTH FILESYSTEM and VOLUME
	if comments, ok := result["comments"].(map[string]interface{}); ok {
		if value, ok := comments["value"].(string); ok {
			data.Comments = types.StringValue(value)
		}
	}
	if compression, ok := result["compression"].(map[string]interface{}); ok {
		if value, ok := compression["value"].(string); ok {
			data.Compression = types.StringValue(value)
		}
	}

	// Read properties based on dataset type
	if datasetType == "VOLUME" {
		// Read VOLUME-specific properties
		if volsize, ok := result["volsize"].(map[string]interface{}); ok {
			if value, ok := volsize["parsed"].(float64); ok {
				data.Volsize = types.Int64Value(int64(value))
			}
		}

		// Set FILESYSTEM-only properties to null for VOLUME datasets
		data.Atime = types.StringNull()
		data.Exec = types.StringNull()
		data.RecordSize = types.StringNull()
		data.Quota = types.Int64Null()
		data.RefQuota = types.Int64Null()
		data.SnapDir = types.StringNull()
	} else {
		// Read FILESYSTEM-specific properties
		if atime, ok := result["atime"].(map[string]interface{}); ok {
			if value, ok := atime["value"].(string); ok {
				data.Atime = types.StringValue(value)
			}
		}

		// Set VOLUME-only properties to null for FILESYSTEM datasets
		data.Volsize = types.Int64Null()
	}
}


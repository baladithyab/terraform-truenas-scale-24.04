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

var _ resource.Resource = &PeriodicSnapshotTaskResource{}
var _ resource.ResourceWithImportState = &PeriodicSnapshotTaskResource{}

// Helper function to parse cron schedule to JSON format expected by TrueNAS
func parseCronSchedule(cronStr string) (map[string]interface{}, error) {
	parts := strings.Fields(cronStr)
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid cron format: expected 5 fields (minute hour dom month dow), got %d", len(parts))
	}

	return map[string]interface{}{
		"minute": parts[0],
		"hour":   parts[1],
		"dom":    parts[2], // day of month
		"month":  parts[3],
		"dow":    parts[4], // day of week
	}, nil
}

// Helper function to convert JSON schedule to cron format
func scheduleToCron(schedule map[string]interface{}) string {
	return fmt.Sprintf("%s %s %s %s %s",
		schedule["minute"],
		schedule["hour"],
		schedule["dom"],
		schedule["month"],
		schedule["dow"],
	)
}

func NewPeriodicSnapshotTaskResource() resource.Resource {
	return &PeriodicSnapshotTaskResource{}
}

type PeriodicSnapshotTaskResource struct {
	client *truenas.Client
}

type PeriodicSnapshotTaskResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Dataset       types.String `tfsdk:"dataset"`
	Recursive     types.Bool   `tfsdk:"recursive"`
	Exclude       types.List   `tfsdk:"exclude"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	NamingSchema  types.String `tfsdk:"naming_schema"`
	Schedule      types.String `tfsdk:"schedule"`
	LifetimeValue types.Int64  `tfsdk:"lifetime_value"`
	LifetimeUnit  types.String `tfsdk:"lifetime_unit"`
	AllowEmpty    types.Bool   `tfsdk:"allow_empty"`
}

func (r *PeriodicSnapshotTaskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_periodic_snapshot_task"
}

func (r *PeriodicSnapshotTaskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a periodic snapshot task on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Task identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset": schema.StringAttribute{
				MarkdownDescription: "Dataset to snapshot (e.g., tank/mydata)",
				Required:            true,
			},
			"recursive": schema.BoolAttribute{
				MarkdownDescription: "Create recursive snapshots of all children",
				Optional:            true,
				Computed:            true,
			},
			"exclude": schema.ListAttribute{
				MarkdownDescription: "List of child datasets to exclude from recursive snapshots",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable this snapshot task",
				Optional:            true,
				Computed:            true,
			},
			"naming_schema": schema.StringAttribute{
				MarkdownDescription: "Naming schema for snapshots (e.g., auto-%Y-%m-%d_%H-%M)",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "Cron schedule in JSON format",
				Required:            true,
			},
			"lifetime_value": schema.Int64Attribute{
				MarkdownDescription: "How long to keep snapshots",
				Required:            true,
			},
			"lifetime_unit": schema.StringAttribute{
				MarkdownDescription: "Lifetime unit (HOUR, DAY, WEEK, MONTH, YEAR)",
				Required:            true,
			},
			"allow_empty": schema.BoolAttribute{
				MarkdownDescription: "Allow taking empty snapshots",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *PeriodicSnapshotTaskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PeriodicSnapshotTaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PeriodicSnapshotTaskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"dataset":        data.Dataset.ValueString(),
		"naming_schema":  data.NamingSchema.ValueString(),
		"lifetime_value": data.LifetimeValue.ValueInt64(),
		"lifetime_unit":  data.LifetimeUnit.ValueString(),
	}

	// Parse schedule from cron format to JSON
	schedule, err := parseCronSchedule(data.Schedule.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Schedule Format", fmt.Sprintf("Unable to parse schedule: %s", err))
		return
	}
	createReq["schedule"] = schedule

	if !data.Recursive.IsNull() {
		createReq["recursive"] = data.Recursive.ValueBool()
	}
	if !data.Enabled.IsNull() {
		createReq["enabled"] = data.Enabled.ValueBool()
	}
	if !data.AllowEmpty.IsNull() {
		createReq["allow_empty"] = data.AllowEmpty.ValueBool()
	}
	if !data.Exclude.IsNull() {
		var exclude []string
		data.Exclude.ElementsAs(ctx, &exclude, false)
		createReq["exclude"] = exclude
	}

	respBody, err := r.client.Post("/pool/snapshottask", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create periodic snapshot task, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(fmt.Sprintf("%d", int(id)))
	}

	r.readPeriodicSnapshotTask(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PeriodicSnapshotTaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PeriodicSnapshotTaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readPeriodicSnapshotTask(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PeriodicSnapshotTaskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PeriodicSnapshotTaskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Dataset.IsNull() {
		updateReq["dataset"] = data.Dataset.ValueString()
	}
	if !data.NamingSchema.IsNull() {
		updateReq["naming_schema"] = data.NamingSchema.ValueString()
	}
	if !data.LifetimeValue.IsNull() {
		updateReq["lifetime_value"] = data.LifetimeValue.ValueInt64()
	}
	if !data.LifetimeUnit.IsNull() {
		updateReq["lifetime_unit"] = data.LifetimeUnit.ValueString()
	}
	if !data.Recursive.IsNull() {
		updateReq["recursive"] = data.Recursive.ValueBool()
	}
	if !data.Enabled.IsNull() {
		updateReq["enabled"] = data.Enabled.ValueBool()
	}
	if !data.AllowEmpty.IsNull() {
		updateReq["allow_empty"] = data.AllowEmpty.ValueBool()
	}

	// Parse schedule from cron format to JSON
	if !data.Schedule.IsNull() {
		schedule, err := parseCronSchedule(data.Schedule.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid Schedule Format", fmt.Sprintf("Unable to parse schedule: %s", err))
			return
		}
		updateReq["schedule"] = schedule
	}

	if !data.Exclude.IsNull() {
		var exclude []string
		data.Exclude.ElementsAs(ctx, &exclude, false)
		updateReq["exclude"] = exclude
	}

	endpoint := fmt.Sprintf("/pool/snapshottask/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update periodic snapshot task, got error: %s", err))
		return
	}

	r.readPeriodicSnapshotTask(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PeriodicSnapshotTaskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PeriodicSnapshotTaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/pool/snapshottask/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete periodic snapshot task, got error: %s", err))
		return
	}
}

func (r *PeriodicSnapshotTaskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PeriodicSnapshotTaskResource) readPeriodicSnapshotTask(ctx context.Context, data *PeriodicSnapshotTaskResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/pool/snapshottask/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read periodic snapshot task, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(fmt.Sprintf("%d", int(id)))
	}
	if dataset, ok := result["dataset"].(string); ok {
		data.Dataset = types.StringValue(dataset)
	}
	if recursive, ok := result["recursive"].(bool); ok {
		data.Recursive = types.BoolValue(recursive)
	}
	if enabled, ok := result["enabled"].(bool); ok {
		data.Enabled = types.BoolValue(enabled)
	}
	if namingSchema, ok := result["naming_schema"].(string); ok {
		data.NamingSchema = types.StringValue(namingSchema)
	}
	if lifetimeValue, ok := result["lifetime_value"].(float64); ok {
		data.LifetimeValue = types.Int64Value(int64(lifetimeValue))
	}
	if lifetimeUnit, ok := result["lifetime_unit"].(string); ok {
		data.LifetimeUnit = types.StringValue(lifetimeUnit)
	}
	if allowEmpty, ok := result["allow_empty"].(bool); ok {
		data.AllowEmpty = types.BoolValue(allowEmpty)
	}
	if schedule, ok := result["schedule"].(map[string]interface{}); ok {
		// Convert JSON schedule back to cron format
		cronStr := scheduleToCron(schedule)
		data.Schedule = types.StringValue(cronStr)
	}
}

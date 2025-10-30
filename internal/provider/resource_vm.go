package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &VMResource{}
var _ resource.ResourceWithImportState = &VMResource{}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

type VMResource struct {
	client *truenas.Client
}

type VMResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	VCPUs       types.Int64  `tfsdk:"vcpus"`
	Cores       types.Int64  `tfsdk:"cores"`
	Threads     types.Int64  `tfsdk:"threads"`
	Memory      types.Int64  `tfsdk:"memory"`
	MinMemory   types.Int64  `tfsdk:"min_memory"`
	Autostart   types.Bool   `tfsdk:"autostart"`
	Bootloader  types.String `tfsdk:"bootloader"`
	CPUMode     types.String `tfsdk:"cpu_mode"`
	CPUModel    types.String `tfsdk:"cpu_model"`
	MachineType types.String `tfsdk:"machine_type"`
	ArchType    types.String `tfsdk:"arch_type"`
	Time        types.String `tfsdk:"time"`
	Status      types.String `tfsdk:"status"`
}

func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a virtual machine on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "VM identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VM name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "VM description",
				Optional:            true,
				Computed:            true,
			},
			"vcpus": schema.Int64Attribute{
				MarkdownDescription: "Number of virtual CPUs",
				Optional:            true,
				Computed:            true,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: "Number of cores per socket",
				Optional:            true,
				Computed:            true,
			},
			"threads": schema.Int64Attribute{
				MarkdownDescription: "Number of threads per core",
				Optional:            true,
				Computed:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "Memory in MB",
				Required:            true,
			},
			"min_memory": schema.Int64Attribute{
				MarkdownDescription: "Minimum memory in MB (for memory ballooning)",
				Optional:            true,
				Computed:            true,
			},
			"autostart": schema.BoolAttribute{
				MarkdownDescription: "Start VM automatically on boot",
				Optional:            true,
				Computed:            true,
			},
			"bootloader": schema.StringAttribute{
				MarkdownDescription: "Bootloader type (UEFI, UEFI_CSM, GRUB)",
				Optional:            true,
				Computed:            true,
			},
			"cpu_mode": schema.StringAttribute{
				MarkdownDescription: "CPU mode (CUSTOM, HOST-MODEL, HOST-PASSTHROUGH)",
				Optional:            true,
				Computed:            true,
			},
			"cpu_model": schema.StringAttribute{
				MarkdownDescription: "CPU model (when cpu_mode is CUSTOM)",
				Optional:            true,
				Computed:            true,
			},
			"machine_type": schema.StringAttribute{
				MarkdownDescription: "Machine type (e.g., q35, pc)",
				Optional:            true,
				Computed:            true,
			},
			"arch_type": schema.StringAttribute{
				MarkdownDescription: "Architecture type",
				Optional:            true,
				Computed:            true,
			},
			"time": schema.StringAttribute{
				MarkdownDescription: "Time synchronization (LOCAL or UTC)",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current VM status",
				Computed:            true,
			},
		},
	}
}

func (r *VMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":   data.Name.ValueString(),
		"memory": data.Memory.ValueInt64(),
	}

	// Handle optional string fields - only send if non-empty
	if !data.Description.IsNull() && data.Description.ValueString() != "" {
		createReq["description"] = data.Description.ValueString()
	}

	// Handle optional integer fields - only send if > 0
	if !data.VCPUs.IsNull() && data.VCPUs.ValueInt64() > 0 {
		createReq["vcpus"] = data.VCPUs.ValueInt64()
	}
	if !data.Cores.IsNull() && data.Cores.ValueInt64() > 0 {
		createReq["cores"] = data.Cores.ValueInt64()
	}
	if !data.Threads.IsNull() && data.Threads.ValueInt64() > 0 {
		createReq["threads"] = data.Threads.ValueInt64()
	}
	if !data.MinMemory.IsNull() && data.MinMemory.ValueInt64() > 0 {
		createReq["min_memory"] = data.MinMemory.ValueInt64()
	}

	// Handle optional boolean fields
	if !data.Autostart.IsNull() {
		createReq["autostart"] = data.Autostart.ValueBool()
	}

	// Handle optional string fields - only send if non-empty
	if !data.Bootloader.IsNull() && data.Bootloader.ValueString() != "" {
		createReq["bootloader"] = data.Bootloader.ValueString()
	}
	if !data.CPUMode.IsNull() && data.CPUMode.ValueString() != "" {
		createReq["cpu_mode"] = data.CPUMode.ValueString()
	}
	if !data.CPUModel.IsNull() && data.CPUModel.ValueString() != "" {
		createReq["cpu_model"] = data.CPUModel.ValueString()
	}
	if !data.MachineType.IsNull() && data.MachineType.ValueString() != "" {
		createReq["machine_type"] = data.MachineType.ValueString()
	}
	if !data.ArchType.IsNull() && data.ArchType.ValueString() != "" {
		createReq["arch_type"] = data.ArchType.ValueString()
	}
	if !data.Time.IsNull() && data.Time.ValueString() != "" {
		createReq["time"] = data.Time.ValueString()
	}

	respBody, err := r.client.Post("/vm", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create VM, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(strconv.Itoa(int(id)))
	}

	r.readVM(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readVM(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Name.IsNull() {
		updateReq["name"] = data.Name.ValueString()
	}
	if !data.Description.IsNull() {
		updateReq["description"] = data.Description.ValueString()
	}
	if !data.Memory.IsNull() {
		updateReq["memory"] = data.Memory.ValueInt64()
	}
	if !data.VCPUs.IsNull() {
		updateReq["vcpus"] = data.VCPUs.ValueInt64()
	}
	if !data.Cores.IsNull() {
		updateReq["cores"] = data.Cores.ValueInt64()
	}
	if !data.Threads.IsNull() {
		updateReq["threads"] = data.Threads.ValueInt64()
	}
	if !data.MinMemory.IsNull() {
		updateReq["min_memory"] = data.MinMemory.ValueInt64()
	}
	if !data.Autostart.IsNull() {
		updateReq["autostart"] = data.Autostart.ValueBool()
	}
	if !data.Bootloader.IsNull() {
		updateReq["bootloader"] = data.Bootloader.ValueString()
	}
	if !data.CPUMode.IsNull() {
		updateReq["cpu_mode"] = data.CPUMode.ValueString()
	}
	if !data.CPUModel.IsNull() {
		updateReq["cpu_model"] = data.CPUModel.ValueString()
	}

	endpoint := fmt.Sprintf("/vm/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update VM, got error: %s", err))
		return
	}

	r.readVM(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/vm/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete VM, got error: %s", err))
		return
	}
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VMResource) readVM(ctx context.Context, data *VMResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/vm/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read VM, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Read ID
	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(strconv.Itoa(int(id)))
	}

	// Read name
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}

	// Read description (optional)
	if description, ok := result["description"].(string); ok && description != "" {
		data.Description = types.StringValue(description)
	} else {
		data.Description = types.StringNull()
	}

	// Read integer properties
	if vcpus, ok := result["vcpus"].(float64); ok {
		data.VCPUs = types.Int64Value(int64(vcpus))
	}
	if cores, ok := result["cores"].(float64); ok {
		data.Cores = types.Int64Value(int64(cores))
	}
	if threads, ok := result["threads"].(float64); ok {
		data.Threads = types.Int64Value(int64(threads))
	}
	if memory, ok := result["memory"].(float64); ok {
		data.Memory = types.Int64Value(int64(memory))
	}

	// Read min_memory (optional)
	if minMemory, ok := result["min_memory"].(float64); ok && minMemory > 0 {
		data.MinMemory = types.Int64Value(int64(minMemory))
	} else {
		data.MinMemory = types.Int64Null()
	}

	// Read autostart
	if autostart, ok := result["autostart"].(bool); ok {
		data.Autostart = types.BoolValue(autostart)
	}

	// Read bootloader (optional)
	if bootloader, ok := result["bootloader"].(string); ok && bootloader != "" {
		data.Bootloader = types.StringValue(bootloader)
	} else {
		data.Bootloader = types.StringNull()
	}

	// Read cpu_mode (optional)
	if cpuMode, ok := result["cpu_mode"].(string); ok && cpuMode != "" {
		data.CPUMode = types.StringValue(cpuMode)
	} else {
		data.CPUMode = types.StringNull()
	}

	// Read cpu_model (optional)
	if cpuModel, ok := result["cpu_model"].(string); ok && cpuModel != "" {
		data.CPUModel = types.StringValue(cpuModel)
	} else {
		data.CPUModel = types.StringNull()
	}

	// Read machine_type (optional)
	if machineType, ok := result["machine_type"].(string); ok && machineType != "" {
		data.MachineType = types.StringValue(machineType)
	} else {
		data.MachineType = types.StringNull()
	}

	// Read arch_type (optional)
	if archType, ok := result["arch_type"].(string); ok && archType != "" {
		data.ArchType = types.StringValue(archType)
	} else {
		data.ArchType = types.StringNull()
	}

	// Read time (optional)
	if time, ok := result["time"].(string); ok && time != "" {
		data.Time = types.StringValue(time)
	} else {
		data.Time = types.StringNull()
	}

	// Read status
	if status, ok := result["status"].(map[string]interface{}); ok {
		if state, ok := status["state"].(string); ok {
			data.Status = types.StringValue(state)
		}
	}
}


package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/baladithyab/terraform-provider-truenas/internal/truenas"
)

var _ datasource.DataSource = &VMDataSource{}

func NewVMDataSource() datasource.DataSource {
	return &VMDataSource{}
}

type VMDataSource struct {
	client *truenas.Client
}

type VMDataSourceModel struct {
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
	Status      types.String `tfsdk:"status"`
}

func (d *VMDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (d *VMDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a specific VM by name or ID",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VM ID (numeric) - specify either id or name",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VM name - specify either id or name",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "VM description",
				Computed:            true,
			},
			"vcpus": schema.Int64Attribute{
				MarkdownDescription: "Number of virtual CPUs",
				Computed:            true,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: "Number of cores per socket",
				Computed:            true,
			},
			"threads": schema.Int64Attribute{
				MarkdownDescription: "Number of threads per core",
				Computed:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "Memory in MiB",
				Computed:            true,
			},
			"min_memory": schema.Int64Attribute{
				MarkdownDescription: "Minimum memory in MiB",
				Computed:            true,
			},
			"autostart": schema.BoolAttribute{
				MarkdownDescription: "Whether VM starts automatically on boot",
				Computed:            true,
			},
			"bootloader": schema.StringAttribute{
				MarkdownDescription: "Bootloader type (UEFI, GRUB, etc.)",
				Computed:            true,
			},
			"cpu_mode": schema.StringAttribute{
				MarkdownDescription: "CPU mode (HOST-PASSTHROUGH, etc.)",
				Computed:            true,
			},
			"cpu_model": schema.StringAttribute{
				MarkdownDescription: "CPU model",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "VM status (RUNNING, STOPPED, etc.)",
				Computed:            true,
			},
		},
	}
}

func (d *VMDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*truenas.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *truenas.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *VMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Must specify either ID or name
	if data.ID.IsNull() && data.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified",
		)
		return
	}

	var result map[string]interface{}

	// If ID is specified, query by ID
	if !data.ID.IsNull() {
		endpoint := fmt.Sprintf("/vm/id/%s", data.ID.ValueString())
		respBody, err := d.client.Get(endpoint)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read VM by ID, got error: %s", err))
			return
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
			return
		}
	} else {
		// Query by name - list all VMs and find matching name
		respBody, err := d.client.Get("/vm")
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list VMs, got error: %s", err))
			return
		}

		var vms []map[string]interface{}
		if err := json.Unmarshal(respBody, &vms); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
			return
		}

		// Find VM by name
		vmName := data.Name.ValueString()
		for _, vm := range vms {
			if name, ok := vm["name"].(string); ok && name == vmName {
				result = vm
				break
			}
		}

		if result == nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("VM with name %q not found", vmName))
			return
		}
	}

	// Parse VM data
	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(strconv.Itoa(int(id)))
	}

	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}

	if description, ok := result["description"].(string); ok && description != "" {
		data.Description = types.StringValue(description)
	} else {
		data.Description = types.StringNull()
	}

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

	if minMemory, ok := result["min_memory"].(float64); ok {
		data.MinMemory = types.Int64Value(int64(minMemory))
	} else {
		data.MinMemory = types.Int64Null()
	}

	if autostart, ok := result["autostart"].(bool); ok {
		data.Autostart = types.BoolValue(autostart)
	}

	if bootloader, ok := result["bootloader"].(string); ok && bootloader != "" {
		data.Bootloader = types.StringValue(bootloader)
	} else {
		data.Bootloader = types.StringNull()
	}

	if cpuMode, ok := result["cpu_mode"].(string); ok && cpuMode != "" {
		data.CPUMode = types.StringValue(cpuMode)
	} else {
		data.CPUMode = types.StringNull()
	}

	if cpuModel, ok := result["cpu_model"].(string); ok && cpuModel != "" {
		data.CPUModel = types.StringValue(cpuModel)
	} else {
		data.CPUModel = types.StringNull()
	}

	// Get status from status object
	if status, ok := result["status"].(map[string]interface{}); ok {
		if state, ok := status["state"].(string); ok {
			data.Status = types.StringValue(state)
		} else {
			data.Status = types.StringNull()
		}
	} else {
		data.Status = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

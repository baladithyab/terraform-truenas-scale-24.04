package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/baladithyab/terraform-provider-truenas/internal/truenas"
)

var _ datasource.DataSource = &VMsDataSource{}

func NewVMsDataSource() datasource.DataSource {
	return &VMsDataSource{}
}

type VMsDataSource struct {
	client *truenas.Client
}

type VMsDataSourceModel struct {
	VMs types.List `tfsdk:"vms"`
}

type VMInfoModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	VCPUs       types.Int64  `tfsdk:"vcpus"`
	Cores       types.Int64  `tfsdk:"cores"`
	Threads     types.Int64  `tfsdk:"threads"`
	Memory      types.Int64  `tfsdk:"memory"`
	Autostart   types.Bool   `tfsdk:"autostart"`
	Status      types.String `tfsdk:"status"`
}

func (d *VMsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vms"
}

func (d *VMsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about all VMs on the TrueNAS system",
		Attributes: map[string]schema.Attribute{
			"vms": schema.ListNestedAttribute{
				MarkdownDescription: "List of VMs",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "VM ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "VM name",
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
						"autostart": schema.BoolAttribute{
							MarkdownDescription: "Whether VM starts automatically on boot",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "VM status (RUNNING, STOPPED, etc.)",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *VMsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VMsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all VMs
	respBody, err := d.client.Get("/vm")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read VMs, got error: %s", err))
		return
	}

	var apiVMs []map[string]interface{}
	if err := json.Unmarshal(respBody, &apiVMs); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Convert API response to model
	var vms []VMInfoModel
	for _, apiVM := range apiVMs {
		vm := VMInfoModel{}

		if id, ok := apiVM["id"].(float64); ok {
			vm.ID = types.StringValue(strconv.Itoa(int(id)))
		}

		if name, ok := apiVM["name"].(string); ok {
			vm.Name = types.StringValue(name)
		} else {
			vm.Name = types.StringNull()
		}

		if description, ok := apiVM["description"].(string); ok && description != "" {
			vm.Description = types.StringValue(description)
		} else {
			vm.Description = types.StringNull()
		}

		if vcpus, ok := apiVM["vcpus"].(float64); ok {
			vm.VCPUs = types.Int64Value(int64(vcpus))
		}

		if cores, ok := apiVM["cores"].(float64); ok {
			vm.Cores = types.Int64Value(int64(cores))
		}

		if threads, ok := apiVM["threads"].(float64); ok {
			vm.Threads = types.Int64Value(int64(threads))
		}

		if memory, ok := apiVM["memory"].(float64); ok {
			vm.Memory = types.Int64Value(int64(memory))
		}

		if autostart, ok := apiVM["autostart"].(bool); ok {
			vm.Autostart = types.BoolValue(autostart)
		}

		// Get status from status object
		if status, ok := apiVM["status"].(map[string]interface{}); ok {
			if state, ok := status["state"].(string); ok {
				vm.Status = types.StringValue(state)
			} else {
				vm.Status = types.StringNull()
			}
		} else {
			vm.Status = types.StringNull()
		}

		vms = append(vms, vm)
	}

	// Convert VMs to types.List
	vmsList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":          types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"vcpus":       types.Int64Type,
			"cores":       types.Int64Type,
			"threads":     types.Int64Type,
			"memory":      types.Int64Type,
			"autostart":   types.BoolType,
			"status":      types.StringType,
		},
	}, vms)

	if diagErr.HasError() {
		resp.Diagnostics.Append(diagErr...)
		return
	}

	data.VMs = vmsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

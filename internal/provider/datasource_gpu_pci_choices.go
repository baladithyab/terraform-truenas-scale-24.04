package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/baladithyab/terraform-provider-truenas/internal/truenas"
)

var _ datasource.DataSource = &GPUPCIChoicesDataSource{}

func NewGPUPCIChoicesDataSource() datasource.DataSource {
	return &GPUPCIChoicesDataSource{}
}

type GPUPCIChoicesDataSource struct {
	client *truenas.Client
}

type GPUPCIChoicesDataSourceModel struct {
	ID      types.String            `tfsdk:"id"`
	Choices map[string]types.String `tfsdk:"choices"`
}

func (d *GPUPCIChoicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gpu_pci_choices"
}

func (d *GPUPCIChoicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches available GPU PCI device choices from the TrueNAS system. Returns a map of GPU descriptions to PCI addresses.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier (always 'gpu_pci_choices')",
				Computed:            true,
			},
			"choices": schema.MapAttribute{
				MarkdownDescription: "Map of GPU descriptions to PCI addresses (e.g., 'NVIDIA Corporation Device 2584' -> '0000:3b:00.0')",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *GPUPCIChoicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*truenas.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *truenas.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *GPUPCIChoicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GPUPCIChoicesDataSourceModel

	endpoint := "/device/gpu_pci_ids_choices"
	respBody, err := d.client.Get(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read GPU PCI choices, got error: %s", err))
		return
	}

	var result map[string]string
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse GPU PCI choices response: %s", err))
		return
	}

	// Convert map[string]string to map[string]types.String
	choices := make(map[string]types.String)
	for k, v := range result {
		choices[k] = types.StringValue(v)
	}

	data.ID = types.StringValue("gpu_pci_choices")
	data.Choices = choices

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

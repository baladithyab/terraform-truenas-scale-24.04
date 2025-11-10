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

var _ datasource.DataSource = &VMIOMMUEnabledDataSource{}

func NewVMIOMMUEnabledDataSource() datasource.DataSource {
	return &VMIOMMUEnabledDataSource{}
}

type VMIOMMUEnabledDataSource struct {
	client *truenas.Client
}

type VMIOMMUEnabledDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (d *VMIOMMUEnabledDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_iommu_enabled"
}

func (d *VMIOMMUEnabledDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks if IOMMU (Intel VT-d / AMD-Vi) is enabled on the TrueNAS system. IOMMU must be enabled for PCI passthrough to work.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier (always 'vm_iommu_enabled')",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether IOMMU is enabled on the system",
				Computed:            true,
			},
		},
	}
}

func (d *VMIOMMUEnabledDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VMIOMMUEnabledDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMIOMMUEnabledDataSourceModel

	endpoint := "/vm/device/iommu_enabled"
	respBody, err := d.client.Get(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to check IOMMU status, got error: %s", err))
		return
	}

	var enabled bool
	if err := json.Unmarshal(respBody, &enabled); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse IOMMU status response: %s", err))
		return
	}

	data.ID = types.StringValue("vm_iommu_enabled")
	data.Enabled = types.BoolValue(enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ datasource.DataSource = &DatasetDataSource{}

func NewDatasetDataSource() datasource.DataSource {
	return &DatasetDataSource{}
}

type DatasetDataSource struct {
	client *truenas.Client
}

type DatasetDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Pool        types.String `tfsdk:"pool"`
	Compression types.String `tfsdk:"compression"`
	Available   types.Int64  `tfsdk:"available"`
	Used        types.Int64  `tfsdk:"used"`
}

func (d *DatasetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (d *DatasetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a ZFS dataset",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Dataset identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Full path of the dataset",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Dataset type",
				Computed:            true,
			},
			"pool": schema.StringAttribute{
				MarkdownDescription: "Pool name",
				Computed:            true,
			},
			"compression": schema.StringAttribute{
				MarkdownDescription: "Compression algorithm",
				Computed:            true,
			},
			"available": schema.Int64Attribute{
				MarkdownDescription: "Available space in bytes",
				Computed:            true,
			},
			"used": schema.Int64Attribute{
				MarkdownDescription: "Used space in bytes",
				Computed:            true,
			},
		},
	}
}

func (d *DatasetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatasetDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/pool/dataset/id/%s", data.ID.ValueString())
	respBody, err := d.client.Get(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read dataset, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if dtype, ok := result["type"].(string); ok {
		data.Type = types.StringValue(dtype)
	}
	if pool, ok := result["pool"].(string); ok {
		data.Pool = types.StringValue(pool)
	}
	if compression, ok := result["compression"].(map[string]interface{}); ok {
		if value, ok := compression["value"].(string); ok {
			data.Compression = types.StringValue(value)
		}
	}
	if available, ok := result["available"].(map[string]interface{}); ok {
		if parsed, ok := available["parsed"].(float64); ok {
			data.Available = types.Int64Value(int64(parsed))
		}
	}
	if used, ok := result["used"].(map[string]interface{}); ok {
		if parsed, ok := used["parsed"].(float64); ok {
			data.Used = types.Int64Value(int64(parsed))
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

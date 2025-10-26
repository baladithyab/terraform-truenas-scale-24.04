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

var _ datasource.DataSource = &PoolDataSource{}

func NewPoolDataSource() datasource.DataSource {
	return &PoolDataSource{}
}

type PoolDataSource struct {
	client *truenas.Client
}

type PoolDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	Healthy   types.Bool   `tfsdk:"healthy"`
	Path      types.String `tfsdk:"path"`
	Available types.Int64  `tfsdk:"available"`
	Size      types.Int64  `tfsdk:"size"`
}

func (d *PoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (d *PoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a ZFS pool",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Pool identifier (pool name)",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Pool name",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Pool status",
				Computed:            true,
			},
			"healthy": schema.BoolAttribute{
				MarkdownDescription: "Whether the pool is healthy",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Pool mount path",
				Computed:            true,
			},
			"available": schema.Int64Attribute{
				MarkdownDescription: "Available space in bytes",
				Computed:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Total pool size in bytes",
				Computed:            true,
			},
		},
	}
}

func (d *PoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PoolDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/pool/id/%s", data.ID.ValueString())
	respBody, err := d.client.Get(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read pool, got error: %s", err))
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
	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if healthy, ok := result["healthy"].(bool); ok {
		data.Healthy = types.BoolValue(healthy)
	}
	if path, ok := result["path"].(string); ok {
		data.Path = types.StringValue(path)
	}
	if topology, ok := result["topology"].(map[string]interface{}); ok {
		if stats, ok := topology["data"].([]interface{}); ok && len(stats) > 0 {
			if stat, ok := stats[0].(map[string]interface{}); ok {
				if statsData, ok := stat["stats"].(map[string]interface{}); ok {
					if size, ok := statsData["size"].(float64); ok {
						data.Size = types.Int64Value(int64(size))
					}
					if allocated, ok := statsData["allocated"].(float64); ok {
						if size, ok := statsData["size"].(float64); ok {
							data.Available = types.Int64Value(int64(size - allocated))
						}
					}
				}
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}


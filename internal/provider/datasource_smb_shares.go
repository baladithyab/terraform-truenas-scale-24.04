package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ datasource.DataSource = &SMBSharesDataSource{}

func NewSMBSharesDataSource() datasource.DataSource {
	return &SMBSharesDataSource{}
}

type SMBSharesDataSource struct {
	client *truenas.Client
}

type SMBSharesDataSourceModel struct {
	Shares types.List `tfsdk:"shares"`
}

type SMBShareModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Path        types.String `tfsdk:"path"`
	Name        types.String `tfsdk:"name"`
	Comment     types.String `tfsdk:"comment"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	ReadOnly    types.Bool   `tfsdk:"readonly"`
	Browsable   types.Bool   `tfsdk:"browsable"`
	GuestOK     types.Bool   `tfsdk:"guestok"`
	Recyclebin  types.Bool   `tfsdk:"recyclebin"`
	Purpose     types.String `tfsdk:"purpose"`
	Home        types.Bool   `tfsdk:"home"`
	Timemachine types.Bool   `tfsdk:"timemachine"`
}

func (d *SMBSharesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smb_shares"
}

func (d *SMBSharesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about all SMB/CIFS shares on the TrueNAS system",
		Attributes: map[string]schema.Attribute{
			"shares": schema.ListNestedAttribute{
				MarkdownDescription: "List of SMB shares",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "SMB share ID",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Path to be shared",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Share name",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Share comment/description",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the share is enabled",
							Computed:            true,
						},
						"readonly": schema.BoolAttribute{
							MarkdownDescription: "Whether the share is read-only",
							Computed:            true,
						},
						"browsable": schema.BoolAttribute{
							MarkdownDescription: "Whether the share is browsable",
							Computed:            true,
						},
						"guestok": schema.BoolAttribute{
							MarkdownDescription: "Whether guest access is allowed",
							Computed:            true,
						},
						"recyclebin": schema.BoolAttribute{
							MarkdownDescription: "Whether recycle bin is enabled",
							Computed:            true,
						},
						"purpose": schema.StringAttribute{
							MarkdownDescription: "Share purpose (e.g., DEFAULT_SHARE, ENHANCED_TIMEMACHINE)",
							Computed:            true,
						},
						"home": schema.BoolAttribute{
							MarkdownDescription: "Whether this is a home share",
							Computed:            true,
						},
						"timemachine": schema.BoolAttribute{
							MarkdownDescription: "Whether Time Machine support is enabled",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SMBSharesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SMBSharesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SMBSharesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all SMB shares
	respBody, err := d.client.Get("/sharing/smb")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read SMB shares, got error: %s", err))
		return
	}

	var apiShares []map[string]interface{}
	if err := json.Unmarshal(respBody, &apiShares); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Convert API response to model
	var shares []SMBShareModel
	for _, apiShare := range apiShares {
		share := SMBShareModel{}

		if id, ok := apiShare["id"].(float64); ok {
			share.ID = types.Int64Value(int64(id))
		}

		if path, ok := apiShare["path"].(string); ok {
			share.Path = types.StringValue(path)
		} else {
			share.Path = types.StringNull()
		}

		if name, ok := apiShare["name"].(string); ok {
			share.Name = types.StringValue(name)
		} else {
			share.Name = types.StringNull()
		}

		if comment, ok := apiShare["comment"].(string); ok && comment != "" {
			share.Comment = types.StringValue(comment)
		} else {
			share.Comment = types.StringNull()
		}

		if enabled, ok := apiShare["enabled"].(bool); ok {
			share.Enabled = types.BoolValue(enabled)
		}

		if ro, ok := apiShare["ro"].(bool); ok {
			share.ReadOnly = types.BoolValue(ro)
		}

		if browsable, ok := apiShare["browsable"].(bool); ok {
			share.Browsable = types.BoolValue(browsable)
		}

		if guestok, ok := apiShare["guestok"].(bool); ok {
			share.GuestOK = types.BoolValue(guestok)
		}

		if recyclebin, ok := apiShare["recyclebin"].(bool); ok {
			share.Recyclebin = types.BoolValue(recyclebin)
		}

		if purpose, ok := apiShare["purpose"].(string); ok && purpose != "" {
			share.Purpose = types.StringValue(purpose)
		} else {
			share.Purpose = types.StringNull()
		}

		if home, ok := apiShare["home"].(bool); ok {
			share.Home = types.BoolValue(home)
		}

		if timemachine, ok := apiShare["timemachine"].(bool); ok {
			share.Timemachine = types.BoolValue(timemachine)
		}

		shares = append(shares, share)
	}

	// Convert shares to types.List
	sharesList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":          types.Int64Type,
			"path":        types.StringType,
			"name":        types.StringType,
			"comment":     types.StringType,
			"enabled":     types.BoolType,
			"readonly":    types.BoolType,
			"browsable":   types.BoolType,
			"guestok":     types.BoolType,
			"recyclebin":  types.BoolType,
			"purpose":     types.StringType,
			"home":        types.BoolType,
			"timemachine": types.BoolType,
		},
	}, shares)

	if diagErr.HasError() {
		resp.Diagnostics.Append(diagErr...)
		return
	}

	data.Shares = sharesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}


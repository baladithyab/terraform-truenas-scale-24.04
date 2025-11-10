package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/baladithyab/terraform-provider-truenas/internal/truenas"
)

var _ datasource.DataSource = &NFSSharesDataSource{}

func NewNFSSharesDataSource() datasource.DataSource {
	return &NFSSharesDataSource{}
}

type NFSSharesDataSource struct {
	client *truenas.Client
}

type NFSSharesDataSourceModel struct {
	Shares types.List `tfsdk:"shares"`
}

type NFSShareModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Path         types.String `tfsdk:"path"`
	Comment      types.String `tfsdk:"comment"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	ReadOnly     types.Bool   `tfsdk:"readonly"`
	MaprootUser  types.String `tfsdk:"maproot_user"`
	MaprootGroup types.String `tfsdk:"maproot_group"`
	MapallUser   types.String `tfsdk:"mapall_user"`
	MapallGroup  types.String `tfsdk:"mapall_group"`
	Networks     types.List   `tfsdk:"networks"`
	Hosts        types.List   `tfsdk:"hosts"`
}

func (d *NFSSharesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nfs_shares"
}

func (d *NFSSharesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about all NFS shares on the TrueNAS system",
		Attributes: map[string]schema.Attribute{
			"shares": schema.ListNestedAttribute{
				MarkdownDescription: "List of NFS shares",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "NFS share ID",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Path to be exported",
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
						"maproot_user": schema.StringAttribute{
							MarkdownDescription: "Map root user to this user",
							Computed:            true,
						},
						"maproot_group": schema.StringAttribute{
							MarkdownDescription: "Map root group to this group",
							Computed:            true,
						},
						"mapall_user": schema.StringAttribute{
							MarkdownDescription: "Map all users to this user",
							Computed:            true,
						},
						"mapall_group": schema.StringAttribute{
							MarkdownDescription: "Map all groups to this group",
							Computed:            true,
						},
						"networks": schema.ListAttribute{
							MarkdownDescription: "Allowed networks (CIDR notation)",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"hosts": schema.ListAttribute{
							MarkdownDescription: "Allowed hosts",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *NFSSharesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NFSSharesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NFSSharesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all NFS shares
	respBody, err := d.client.Get("/sharing/nfs")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read NFS shares, got error: %s", err))
		return
	}

	var apiShares []map[string]interface{}
	if err := json.Unmarshal(respBody, &apiShares); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Convert API response to model
	var shares []NFSShareModel
	for _, apiShare := range apiShares {
		share := NFSShareModel{}

		if id, ok := apiShare["id"].(float64); ok {
			share.ID = types.Int64Value(int64(id))
		}

		if path, ok := apiShare["path"].(string); ok {
			share.Path = types.StringValue(path)
		} else {
			share.Path = types.StringNull()
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

		if maprootUser, ok := apiShare["maproot_user"].(string); ok && maprootUser != "" {
			share.MaprootUser = types.StringValue(maprootUser)
		} else {
			share.MaprootUser = types.StringNull()
		}

		if maprootGroup, ok := apiShare["maproot_group"].(string); ok && maprootGroup != "" {
			share.MaprootGroup = types.StringValue(maprootGroup)
		} else {
			share.MaprootGroup = types.StringNull()
		}

		if mapallUser, ok := apiShare["mapall_user"].(string); ok && mapallUser != "" {
			share.MapallUser = types.StringValue(mapallUser)
		} else {
			share.MapallUser = types.StringNull()
		}

		if mapallGroup, ok := apiShare["mapall_group"].(string); ok && mapallGroup != "" {
			share.MapallGroup = types.StringValue(mapallGroup)
		} else {
			share.MapallGroup = types.StringNull()
		}

		// Parse networks
		if networks, ok := apiShare["networks"].([]interface{}); ok && len(networks) > 0 {
			networkStrs := make([]string, 0, len(networks))
			for _, net := range networks {
				if netStr, ok := net.(string); ok {
					networkStrs = append(networkStrs, netStr)
				}
			}
			networkList, diagErr := types.ListValueFrom(ctx, types.StringType, networkStrs)
			if diagErr.HasError() {
				resp.Diagnostics.Append(diagErr...)
			} else {
				share.Networks = networkList
			}
		} else {
			share.Networks = types.ListNull(types.StringType)
		}

		// Parse hosts
		if hosts, ok := apiShare["hosts"].([]interface{}); ok && len(hosts) > 0 {
			hostStrs := make([]string, 0, len(hosts))
			for _, host := range hosts {
				if hostStr, ok := host.(string); ok {
					hostStrs = append(hostStrs, hostStr)
				}
			}
			hostList, diagErr := types.ListValueFrom(ctx, types.StringType, hostStrs)
			if diagErr.HasError() {
				resp.Diagnostics.Append(diagErr...)
			} else {
				share.Hosts = hostList
			}
		} else {
			share.Hosts = types.ListNull(types.StringType)
		}

		shares = append(shares, share)
	}

	// Convert shares to types.List
	sharesList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":            types.Int64Type,
			"path":          types.StringType,
			"comment":       types.StringType,
			"enabled":       types.BoolType,
			"readonly":      types.BoolType,
			"maproot_user":  types.StringType,
			"maproot_group": types.StringType,
			"mapall_user":   types.StringType,
			"mapall_group":  types.StringType,
			"networks":      types.ListType{ElemType: types.StringType},
			"hosts":         types.ListType{ElemType: types.StringType},
		},
	}, shares)

	if diagErr.HasError() {
		resp.Diagnostics.Append(diagErr...)
		return
	}

	data.Shares = sharesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

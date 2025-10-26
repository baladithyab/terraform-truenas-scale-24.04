package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &ISCSIPortalResource{}
var _ resource.ResourceWithImportState = &ISCSIPortalResource{}

func NewISCSIPortalResource() resource.Resource {
	return &ISCSIPortalResource{}
}

type ISCSIPortalResource struct {
	client *truenas.Client
}

type ISCSIPortalResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Comment              types.String `tfsdk:"comment"`
	DiscoveryAuthMethod  types.String `tfsdk:"discovery_authmethod"`
	DiscoveryAuthGroup   types.Int64  `tfsdk:"discovery_authgroup"`
	Listen               types.List   `tfsdk:"listen"`
}

type ISCSIPortalListen struct {
	IP   types.String `tfsdk:"ip"`
	Port types.Int64  `tfsdk:"port"`
}

func (r *ISCSIPortalResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_portal"
}

func (r *ISCSIPortalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an iSCSI portal (network listener) on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Portal identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Portal comment/description",
				Optional:            true,
				Computed:            true,
			},
			"discovery_authmethod": schema.StringAttribute{
				MarkdownDescription: "Discovery authentication method (NONE, CHAP, CHAP_MUTUAL)",
				Optional:            true,
				Computed:            true,
			},
			"discovery_authgroup": schema.Int64Attribute{
				MarkdownDescription: "Discovery authentication group ID",
				Optional:            true,
				Computed:            true,
			},
			"listen": schema.ListNestedAttribute{
				MarkdownDescription: "List of IP addresses and ports to listen on",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "IP address to listen on (0.0.0.0 for all)",
							Required:            true,
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "Port to listen on (default: 3260)",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (r *ISCSIPortalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ISCSIPortalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ISCSIPortalResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{}

	if !data.Comment.IsNull() {
		createReq["comment"] = data.Comment.ValueString()
	}
	if !data.DiscoveryAuthMethod.IsNull() {
		createReq["discovery_authmethod"] = data.DiscoveryAuthMethod.ValueString()
	}
	if !data.DiscoveryAuthGroup.IsNull() {
		createReq["discovery_authgroup"] = data.DiscoveryAuthGroup.ValueInt64()
	}

	// Parse listen addresses
	var listenList []ISCSIPortalListen
	data.Listen.ElementsAs(ctx, &listenList, false)
	
	listen := make([]map[string]interface{}, 0, len(listenList))
	for _, l := range listenList {
		entry := map[string]interface{}{
			"ip": l.IP.ValueString(),
		}
		if !l.Port.IsNull() {
			entry["port"] = l.Port.ValueInt64()
		} else {
			entry["port"] = 3260
		}
		listen = append(listen, entry)
	}
	createReq["listen"] = listen

	respBody, err := r.client.Post("/iscsi/portal", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create iSCSI portal, got error: %s", err))
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

	r.readISCSIPortal(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSIPortalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ISCSIPortalResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readISCSIPortal(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSIPortalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ISCSIPortalResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Comment.IsNull() {
		updateReq["comment"] = data.Comment.ValueString()
	}
	if !data.DiscoveryAuthMethod.IsNull() {
		updateReq["discovery_authmethod"] = data.DiscoveryAuthMethod.ValueString()
	}
	if !data.DiscoveryAuthGroup.IsNull() {
		updateReq["discovery_authgroup"] = data.DiscoveryAuthGroup.ValueInt64()
	}

	// Parse listen addresses
	var listenList []ISCSIPortalListen
	data.Listen.ElementsAs(ctx, &listenList, false)
	
	listen := make([]map[string]interface{}, 0, len(listenList))
	for _, l := range listenList {
		entry := map[string]interface{}{
			"ip": l.IP.ValueString(),
		}
		if !l.Port.IsNull() {
			entry["port"] = l.Port.ValueInt64()
		} else {
			entry["port"] = 3260
		}
		listen = append(listen, entry)
	}
	updateReq["listen"] = listen

	endpoint := fmt.Sprintf("/iscsi/portal/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update iSCSI portal, got error: %s", err))
		return
	}

	r.readISCSIPortal(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSIPortalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ISCSIPortalResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/iscsi/portal/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete iSCSI portal, got error: %s", err))
		return
	}
}

func (r *ISCSIPortalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ISCSIPortalResource) readISCSIPortal(ctx context.Context, data *ISCSIPortalResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/iscsi/portal/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read iSCSI portal, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if comment, ok := result["comment"].(string); ok {
		data.Comment = types.StringValue(comment)
	}
	if authmethod, ok := result["discovery_authmethod"].(string); ok {
		data.DiscoveryAuthMethod = types.StringValue(authmethod)
	}
	if authgroup, ok := result["discovery_authgroup"].(float64); ok {
		data.DiscoveryAuthGroup = types.Int64Value(int64(authgroup))
	}
	
	if listen, ok := result["listen"].([]interface{}); ok {
		listenList := make([]ISCSIPortalListen, 0, len(listen))
		for _, l := range listen {
			if listenMap, ok := l.(map[string]interface{}); ok {
				entry := ISCSIPortalListen{}
				if ip, ok := listenMap["ip"].(string); ok {
					entry.IP = types.StringValue(ip)
				}
				if port, ok := listenMap["port"].(float64); ok {
					entry.Port = types.Int64Value(int64(port))
				}
				listenList = append(listenList, entry)
			}
		}
		list, _ := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"ip":   types.StringType,
				"port": types.Int64Type,
			},
		}, listenList)
		data.Listen = list
	}
}


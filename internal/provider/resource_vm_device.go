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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/baladithyab/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &VMDeviceResource{}
var _ resource.ResourceWithImportState = &VMDeviceResource{}

func NewVMDeviceResource() resource.Resource {
	return &VMDeviceResource{}
}

type VMDeviceResource struct {
	client *truenas.Client
}

type VMDeviceResourceModel struct {
	ID            types.String `tfsdk:"id"`
	VMID          types.String `tfsdk:"vm_id"`
	DeviceType    types.String `tfsdk:"device_type"`
	Order         types.Int64  `tfsdk:"order"`
	NICConfig     types.List   `tfsdk:"nic_config"`
	DiskConfig    types.List   `tfsdk:"disk_config"`
	CDROMConfig   types.List   `tfsdk:"cdrom_config"`
	PCIConfig     types.List   `tfsdk:"pci_config"`
	USBConfig     types.List   `tfsdk:"usb_config"`
	DisplayConfig types.List   `tfsdk:"display_config"`
	RAWConfig     types.List   `tfsdk:"raw_config"`
}

type NICConfigModel struct {
	Type                types.String `tfsdk:"type"`
	MAC                 types.String `tfsdk:"mac"`
	NICAttach           types.String `tfsdk:"nic_attach"`
	TrustGuestRxFilters types.Bool   `tfsdk:"trust_guest_rx_filters"`
}

type DiskConfigModel struct {
	Path               types.String `tfsdk:"path"`
	Type               types.String `tfsdk:"type"`
	IOType             types.String `tfsdk:"iotype"`
	PhysicalSectorSize types.Int64  `tfsdk:"physical_sectorsize"`
	LogicalSectorSize  types.Int64  `tfsdk:"logical_sectorsize"`
}

type CDROMConfigModel struct {
	Path types.String `tfsdk:"path"`
}

type PCIConfigModel struct {
	PPTDev types.String `tfsdk:"pptdev"`
}

type USBConfigModel struct {
	Controller types.String `tfsdk:"controller"`
	Device     types.String `tfsdk:"device"`
}

type DisplayConfigModel struct {
	Port       types.Int64  `tfsdk:"port"`
	Bind       types.String `tfsdk:"bind"`
	Password   types.String `tfsdk:"password"`
	Web        types.Bool   `tfsdk:"web"`
	Type       types.String `tfsdk:"type"`
	Resolution types.String `tfsdk:"resolution"`
	WebPort    types.Int64  `tfsdk:"web_port"`
	Wait       types.Bool   `tfsdk:"wait"`
}

type RAWConfigModel struct {
	Path types.String `tfsdk:"path"`
	Size types.Int64  `tfsdk:"size"`
	Boot types.Bool   `tfsdk:"boot"`
}

func (r *VMDeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_device"
}

func (r *VMDeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VM device independently on TrueNAS. This resource allows you to manage VM devices (NICs, Disks, CDROMs, PCI devices, USB devices, Displays, and RAW devices) separately from the VM itself.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Device identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vm_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VM this device belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"device_type": schema.StringAttribute{
				MarkdownDescription: "Type of device: NIC, DISK, CDROM, PCI, USB, DISPLAY, or RAW",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"order": schema.Int64Attribute{
				MarkdownDescription: "Boot order for this device. Lower values boot first.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nic_config": schema.ListNestedAttribute{
				MarkdownDescription: "Configuration for NIC device (required when device_type is NIC)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "NIC type (VIRTIO, E1000, etc.). Default: VIRTIO",
							Optional:            true,
							Computed:            true,
						},
						"mac": schema.StringAttribute{
							MarkdownDescription: "MAC address (leave empty for auto-generation)",
							Optional:            true,
							Computed:            true,
						},
						"nic_attach": schema.StringAttribute{
							MarkdownDescription: "Physical network interface to attach to (e.g., eno1, br0)",
							Required:            true,
						},
						"trust_guest_rx_filters": schema.BoolAttribute{
							MarkdownDescription: "Trust guest RX filters. Default: false",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"disk_config": schema.ListNestedAttribute{
				MarkdownDescription: "Configuration for DISK device (required when device_type is DISK)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Path to disk (e.g., /dev/zvol/pool/vm-disk0)",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Disk type (VIRTIO, AHCI, etc.). Default: VIRTIO",
							Optional:            true,
							Computed:            true,
						},
						"iotype": schema.StringAttribute{
							MarkdownDescription: "IO type (THREADS, NATIVE). Default: THREADS",
							Optional:            true,
							Computed:            true,
						},
						"physical_sectorsize": schema.Int64Attribute{
							MarkdownDescription: "Physical sector size in bytes",
							Optional:            true,
							Computed:            true,
						},
						"logical_sectorsize": schema.Int64Attribute{
							MarkdownDescription: "Logical sector size in bytes",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"cdrom_config": schema.ListNestedAttribute{
				MarkdownDescription: "Configuration for CDROM device (required when device_type is CDROM)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Path to ISO file (e.g., /mnt/pool/isos/ubuntu.iso)",
							Required:            true,
						},
					},
				},
			},
			"pci_config": schema.ListNestedAttribute{
				MarkdownDescription: "Configuration for PCI device (required when device_type is PCI)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pptdev": schema.StringAttribute{
							MarkdownDescription: "PCI device ID to pass through (e.g., 'pci_0000_3b_00_0')",
							Required:            true,
						},
					},
				},
			},
			"usb_config": schema.ListNestedAttribute{
				MarkdownDescription: "Configuration for USB device (required when device_type is USB)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"controller": schema.StringAttribute{
							MarkdownDescription: "USB controller type",
							Optional:            true,
							Computed:            true,
						},
						"device": schema.StringAttribute{
							MarkdownDescription: "USB device identifier",
							Required:            true,
						},
					},
				},
			},
			"display_config": schema.ListNestedAttribute{
				MarkdownDescription: "Configuration for DISPLAY device (required when device_type is DISPLAY)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							MarkdownDescription: "Port number for the display server",
							Optional:            true,
							Computed:            true,
						},
						"bind": schema.StringAttribute{
							MarkdownDescription: "IP address to bind to",
							Optional:            true,
							Computed:            true,
						},
						"password": schema.StringAttribute{
							MarkdownDescription: "Password for display access",
							Optional:            true,
							Sensitive:           true,
						},
						"web": schema.BoolAttribute{
							MarkdownDescription: "Enable web access",
							Optional:            true,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Display type: SPICE or VNC",
							Optional:            true,
							Computed:            true,
						},
						"resolution": schema.StringAttribute{
							MarkdownDescription: "Display resolution",
							Optional:            true,
							Computed:            true,
						},
						"web_port": schema.Int64Attribute{
							MarkdownDescription: "Port for web access",
							Optional:            true,
							Computed:            true,
						},
						"wait": schema.BoolAttribute{
							MarkdownDescription: "Wait for client connection",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"raw_config": schema.ListNestedAttribute{
				MarkdownDescription: "Configuration for RAW device (required when device_type is RAW)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Path to raw file",
							Required:            true,
						},
						"size": schema.Int64Attribute{
							MarkdownDescription: "Size in bytes",
							Optional:            true,
							Computed:            true,
						},
						"boot": schema.BoolAttribute{
							MarkdownDescription: "Whether this is a boot device",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (r *VMDeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VMDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VMDeviceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build device request
	deviceReq := map[string]interface{}{
		"vm":    data.VMID.ValueString(),
		"dtype": data.DeviceType.ValueString(),
	}

	// Set order if specified, otherwise use default
	if !data.Order.IsNull() {
		deviceReq["order"] = int(data.Order.ValueInt64())
	} else {
		deviceReq["order"] = 1000
	}

	// Build attributes based on device type
	attributes := make(map[string]interface{})

	switch data.DeviceType.ValueString() {
	case "NIC":
		if !data.NICConfig.IsNull() && !data.NICConfig.IsUnknown() {
			var nicConfigs []NICConfigModel
			diags := data.NICConfig.ElementsAs(ctx, &nicConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(nicConfigs) > 0 {
				nic := nicConfigs[0]
				attributes["nic_attach"] = nic.NICAttach.ValueString()
				if !nic.Type.IsNull() && nic.Type.ValueString() != "" {
					attributes["type"] = nic.Type.ValueString()
				} else {
					attributes["type"] = "VIRTIO"
				}
				if !nic.MAC.IsNull() && nic.MAC.ValueString() != "" {
					attributes["mac"] = nic.MAC.ValueString()
				}
				if !nic.TrustGuestRxFilters.IsNull() {
					attributes["trust_guest_rx_filters"] = nic.TrustGuestRxFilters.ValueBool()
				} else {
					attributes["trust_guest_rx_filters"] = false
				}
			}
		}

	case "DISK":
		if !data.DiskConfig.IsNull() && !data.DiskConfig.IsUnknown() {
			var diskConfigs []DiskConfigModel
			diags := data.DiskConfig.ElementsAs(ctx, &diskConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(diskConfigs) > 0 {
				disk := diskConfigs[0]
				attributes["path"] = disk.Path.ValueString()
				if !disk.Type.IsNull() && disk.Type.ValueString() != "" {
					attributes["type"] = disk.Type.ValueString()
				} else {
					attributes["type"] = "VIRTIO"
				}
				if !disk.IOType.IsNull() && disk.IOType.ValueString() != "" {
					attributes["iotype"] = disk.IOType.ValueString()
				}
				if !disk.PhysicalSectorSize.IsNull() && disk.PhysicalSectorSize.ValueInt64() > 0 {
					attributes["physical_sectorsize"] = disk.PhysicalSectorSize.ValueInt64()
				}
				if !disk.LogicalSectorSize.IsNull() && disk.LogicalSectorSize.ValueInt64() > 0 {
					attributes["logical_sectorsize"] = disk.LogicalSectorSize.ValueInt64()
				}
			}
		}

	case "CDROM":
		if !data.CDROMConfig.IsNull() && !data.CDROMConfig.IsUnknown() {
			var cdromConfigs []CDROMConfigModel
			diags := data.CDROMConfig.ElementsAs(ctx, &cdromConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(cdromConfigs) > 0 {
				attributes["path"] = cdromConfigs[0].Path.ValueString()
			}
		}

	case "PCI":
		if !data.PCIConfig.IsNull() && !data.PCIConfig.IsUnknown() {
			var pciConfigs []PCIConfigModel
			diags := data.PCIConfig.ElementsAs(ctx, &pciConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(pciConfigs) > 0 {
				attributes["pptdev"] = pciConfigs[0].PPTDev.ValueString()
			}
		}

	case "USB":
		if !data.USBConfig.IsNull() && !data.USBConfig.IsUnknown() {
			var usbConfigs []USBConfigModel
			diags := data.USBConfig.ElementsAs(ctx, &usbConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(usbConfigs) > 0 {
				usb := usbConfigs[0]
				attributes["device"] = usb.Device.ValueString()
				if !usb.Controller.IsNull() && usb.Controller.ValueString() != "" {
					attributes["controller"] = usb.Controller.ValueString()
				}
			}
		}

	case "DISPLAY":
		if !data.DisplayConfig.IsNull() && !data.DisplayConfig.IsUnknown() {
			var displayConfigs []DisplayConfigModel
			diags := data.DisplayConfig.ElementsAs(ctx, &displayConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(displayConfigs) > 0 {
				display := displayConfigs[0]
				if !display.Port.IsNull() {
					attributes["port"] = int(display.Port.ValueInt64())
				}
				if !display.Bind.IsNull() {
					attributes["bind"] = display.Bind.ValueString()
				}
				if !display.Password.IsNull() {
					attributes["password"] = display.Password.ValueString()
				}
				if !display.Web.IsNull() {
					attributes["web"] = display.Web.ValueBool()
				}
				if !display.Type.IsNull() {
					attributes["type"] = display.Type.ValueString()
				}
				if !display.Resolution.IsNull() {
					attributes["resolution"] = display.Resolution.ValueString()
				}
				if !display.WebPort.IsNull() {
					attributes["web_port"] = int(display.WebPort.ValueInt64())
				}
				if !display.Wait.IsNull() {
					attributes["wait"] = display.Wait.ValueBool()
				}
			}
		}

	case "RAW":
		if !data.RAWConfig.IsNull() && !data.RAWConfig.IsUnknown() {
			var rawConfigs []RAWConfigModel
			diags := data.RAWConfig.ElementsAs(ctx, &rawConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(rawConfigs) > 0 {
				raw := rawConfigs[0]
				attributes["path"] = raw.Path.ValueString()
				if !raw.Size.IsNull() {
					attributes["size"] = raw.Size.ValueInt64()
				}
				if !raw.Boot.IsNull() {
					attributes["boot"] = raw.Boot.ValueBool()
				}
			}
		}
	}

	deviceReq["attributes"] = attributes

	respBody, err := r.client.CreateVMDevice(deviceReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create VM device, got error: %s", err))
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

	// Read back the created device to populate computed values
	r.readVMDevice(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VMDeviceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readVMDevice(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VMDeviceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := map[string]interface{}{}

	// Set order if specified
	if !data.Order.IsNull() {
		updateReq["order"] = int(data.Order.ValueInt64())
	}

	// Build attributes based on device type
	attributes := make(map[string]interface{})

	switch data.DeviceType.ValueString() {
	case "NIC":
		if !data.NICConfig.IsNull() && !data.NICConfig.IsUnknown() {
			var nicConfigs []NICConfigModel
			diags := data.NICConfig.ElementsAs(ctx, &nicConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(nicConfigs) > 0 {
				nic := nicConfigs[0]
				attributes["nic_attach"] = nic.NICAttach.ValueString()
				if !nic.Type.IsNull() {
					attributes["type"] = nic.Type.ValueString()
				}
				if !nic.MAC.IsNull() {
					attributes["mac"] = nic.MAC.ValueString()
				}
				if !nic.TrustGuestRxFilters.IsNull() {
					attributes["trust_guest_rx_filters"] = nic.TrustGuestRxFilters.ValueBool()
				}
			}
		}

	case "DISK":
		if !data.DiskConfig.IsNull() && !data.DiskConfig.IsUnknown() {
			var diskConfigs []DiskConfigModel
			diags := data.DiskConfig.ElementsAs(ctx, &diskConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(diskConfigs) > 0 {
				disk := diskConfigs[0]
				attributes["path"] = disk.Path.ValueString()
				if !disk.Type.IsNull() {
					attributes["type"] = disk.Type.ValueString()
				}
				if !disk.IOType.IsNull() {
					attributes["iotype"] = disk.IOType.ValueString()
				}
				if !disk.PhysicalSectorSize.IsNull() {
					attributes["physical_sectorsize"] = disk.PhysicalSectorSize.ValueInt64()
				}
				if !disk.LogicalSectorSize.IsNull() {
					attributes["logical_sectorsize"] = disk.LogicalSectorSize.ValueInt64()
				}
			}
		}

	case "CDROM":
		if !data.CDROMConfig.IsNull() && !data.CDROMConfig.IsUnknown() {
			var cdromConfigs []CDROMConfigModel
			diags := data.CDROMConfig.ElementsAs(ctx, &cdromConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(cdromConfigs) > 0 {
				attributes["path"] = cdromConfigs[0].Path.ValueString()
			}
		}

	case "PCI":
		if !data.PCIConfig.IsNull() && !data.PCIConfig.IsUnknown() {
			var pciConfigs []PCIConfigModel
			diags := data.PCIConfig.ElementsAs(ctx, &pciConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(pciConfigs) > 0 {
				attributes["pptdev"] = pciConfigs[0].PPTDev.ValueString()
			}
		}

	case "USB":
		if !data.USBConfig.IsNull() && !data.USBConfig.IsUnknown() {
			var usbConfigs []USBConfigModel
			diags := data.USBConfig.ElementsAs(ctx, &usbConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(usbConfigs) > 0 {
				usb := usbConfigs[0]
				attributes["device"] = usb.Device.ValueString()
				if !usb.Controller.IsNull() {
					attributes["controller"] = usb.Controller.ValueString()
				}
			}
		}

	case "DISPLAY":
		if !data.DisplayConfig.IsNull() && !data.DisplayConfig.IsUnknown() {
			var displayConfigs []DisplayConfigModel
			diags := data.DisplayConfig.ElementsAs(ctx, &displayConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(displayConfigs) > 0 {
				display := displayConfigs[0]
				if !display.Port.IsNull() {
					attributes["port"] = int(display.Port.ValueInt64())
				}
				if !display.Bind.IsNull() {
					attributes["bind"] = display.Bind.ValueString()
				}
				if !display.Password.IsNull() {
					attributes["password"] = display.Password.ValueString()
				}
				if !display.Web.IsNull() {
					attributes["web"] = display.Web.ValueBool()
				}
				if !display.Type.IsNull() {
					attributes["type"] = display.Type.ValueString()
				}
				if !display.Resolution.IsNull() {
					attributes["resolution"] = display.Resolution.ValueString()
				}
				if !display.WebPort.IsNull() {
					attributes["web_port"] = int(display.WebPort.ValueInt64())
				}
				if !display.Wait.IsNull() {
					attributes["wait"] = display.Wait.ValueBool()
				}
			}
		}

	case "RAW":
		if !data.RAWConfig.IsNull() && !data.RAWConfig.IsUnknown() {
			var rawConfigs []RAWConfigModel
			diags := data.RAWConfig.ElementsAs(ctx, &rawConfigs, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if len(rawConfigs) > 0 {
				raw := rawConfigs[0]
				attributes["path"] = raw.Path.ValueString()
				if !raw.Size.IsNull() {
					attributes["size"] = raw.Size.ValueInt64()
				}
				if !raw.Boot.IsNull() {
					attributes["boot"] = raw.Boot.ValueBool()
				}
			}
		}
	}

	if len(attributes) > 0 {
		updateReq["attributes"] = attributes
	}

	_, err := r.client.UpdateVMDevice(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update VM device, got error: %s", err))
		return
	}

	r.readVMDevice(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VMDeviceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteVMDevice(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete VM device, got error: %s", err))
		return
	}
}

func (r *VMDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VMDeviceResource) readVMDevice(ctx context.Context, data *VMDeviceResourceModel, diags *diag.Diagnostics) {
	respBody, err := r.client.GetVMDevice(data.ID.ValueString())
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read VM device, got error: %s", err))
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

	// Read VM ID
	if vm, ok := result["vm"].(float64); ok {
		data.VMID = types.StringValue(strconv.Itoa(int(vm)))
	}

	// Read device type
	if dtype, ok := result["dtype"].(string); ok {
		data.DeviceType = types.StringValue(dtype)
	}

	// Read order
	if order, ok := result["order"].(float64); ok {
		data.Order = types.Int64Value(int64(order))
	}

	// Read attributes based on device type
	if attributes, ok := result["attributes"].(map[string]interface{}); ok {
		switch data.DeviceType.ValueString() {
		case "NIC":
			nic := NICConfigModel{}
			if nicType, ok := attributes["type"].(string); ok {
				nic.Type = types.StringValue(nicType)
			}
			if mac, ok := attributes["mac"].(string); ok {
				nic.MAC = types.StringValue(mac)
			}
			if nicAttach, ok := attributes["nic_attach"].(string); ok {
				nic.NICAttach = types.StringValue(nicAttach)
			}
			if trustRx, ok := attributes["trust_guest_rx_filters"].(bool); ok {
				nic.TrustGuestRxFilters = types.BoolValue(trustRx)
			}
			nicList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type":                   types.StringType,
					"mac":                    types.StringType,
					"nic_attach":             types.StringType,
					"trust_guest_rx_filters": types.BoolType,
				},
			}, []NICConfigModel{nic})
			if diagErr.HasError() {
				diags.Append(diagErr...)
			} else {
				data.NICConfig = nicList
			}

		case "DISK":
			disk := DiskConfigModel{}
			if path, ok := attributes["path"].(string); ok {
				disk.Path = types.StringValue(path)
			}
			if diskType, ok := attributes["type"].(string); ok {
				disk.Type = types.StringValue(diskType)
			}
			if iotype, ok := attributes["iotype"].(string); ok {
				disk.IOType = types.StringValue(iotype)
			}
			if physSector, ok := attributes["physical_sectorsize"].(float64); ok {
				disk.PhysicalSectorSize = types.Int64Value(int64(physSector))
			}
			if logSector, ok := attributes["logical_sectorsize"].(float64); ok {
				disk.LogicalSectorSize = types.Int64Value(int64(logSector))
			}
			diskList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"path":                types.StringType,
					"type":                types.StringType,
					"iotype":              types.StringType,
					"physical_sectorsize": types.Int64Type,
					"logical_sectorsize":  types.Int64Type,
				},
			}, []DiskConfigModel{disk})
			if diagErr.HasError() {
				diags.Append(diagErr...)
			} else {
				data.DiskConfig = diskList
			}

		case "CDROM":
			cdrom := CDROMConfigModel{}
			if path, ok := attributes["path"].(string); ok {
				cdrom.Path = types.StringValue(path)
			}
			cdromList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"path": types.StringType,
				},
			}, []CDROMConfigModel{cdrom})
			if diagErr.HasError() {
				diags.Append(diagErr...)
			} else {
				data.CDROMConfig = cdromList
			}

		case "PCI":
			pci := PCIConfigModel{}
			if pptdev, ok := attributes["pptdev"].(string); ok {
				pci.PPTDev = types.StringValue(pptdev)
			}
			pciList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"pptdev": types.StringType,
				},
			}, []PCIConfigModel{pci})
			if diagErr.HasError() {
				diags.Append(diagErr...)
			} else {
				data.PCIConfig = pciList
			}

		case "USB":
			usb := USBConfigModel{}
			if controller, ok := attributes["controller"].(string); ok {
				usb.Controller = types.StringValue(controller)
			}
			if device, ok := attributes["device"].(string); ok {
				usb.Device = types.StringValue(device)
			}
			usbList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"controller": types.StringType,
					"device":     types.StringType,
				},
			}, []USBConfigModel{usb})
			if diagErr.HasError() {
				diags.Append(diagErr...)
			} else {
				data.USBConfig = usbList
			}

		case "DISPLAY":
			display := DisplayConfigModel{}
			if port, ok := attributes["port"].(float64); ok {
				display.Port = types.Int64Value(int64(port))
			}
			if bind, ok := attributes["bind"].(string); ok {
				display.Bind = types.StringValue(bind)
			}
			if password, ok := attributes["password"].(string); ok {
				display.Password = types.StringValue(password)
			}
			if web, ok := attributes["web"].(bool); ok {
				display.Web = types.BoolValue(web)
			}
			if displayType, ok := attributes["type"].(string); ok {
				display.Type = types.StringValue(displayType)
			}
			if resolution, ok := attributes["resolution"].(string); ok {
				display.Resolution = types.StringValue(resolution)
			}
			if webPort, ok := attributes["web_port"].(float64); ok {
				display.WebPort = types.Int64Value(int64(webPort))
			}
			if wait, ok := attributes["wait"].(bool); ok {
				display.Wait = types.BoolValue(wait)
			}
			displayList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"port":       types.Int64Type,
					"bind":       types.StringType,
					"password":   types.StringType,
					"web":        types.BoolType,
					"type":       types.StringType,
					"resolution": types.StringType,
					"web_port":   types.Int64Type,
					"wait":       types.BoolType,
				},
			}, []DisplayConfigModel{display})
			if diagErr.HasError() {
				diags.Append(diagErr...)
			} else {
				data.DisplayConfig = displayList
			}

		case "RAW":
			raw := RAWConfigModel{}
			if path, ok := attributes["path"].(string); ok {
				raw.Path = types.StringValue(path)
			}
			if size, ok := attributes["size"].(float64); ok {
				raw.Size = types.Int64Value(int64(size))
			}
			if boot, ok := attributes["boot"].(bool); ok {
				raw.Boot = types.BoolValue(boot)
			}
			rawList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"path": types.StringType,
					"size": types.Int64Type,
					"boot": types.BoolType,
				},
			}, []RAWConfigModel{raw})
			if diagErr.HasError() {
				diags.Append(diagErr...)
			} else {
				data.RAWConfig = rawList
			}
		}
	}
}

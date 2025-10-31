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

var _ resource.Resource = &VMResource{}
var _ resource.ResourceWithImportState = &VMResource{}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

type VMResource struct {
	client *truenas.Client
}

type VMResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	VCPUs         types.Int64  `tfsdk:"vcpus"`
	Cores         types.Int64  `tfsdk:"cores"`
	Threads       types.Int64  `tfsdk:"threads"`
	Memory        types.Int64  `tfsdk:"memory"`
	MinMemory     types.Int64  `tfsdk:"min_memory"`
	Autostart     types.Bool   `tfsdk:"autostart"`
	StartOnCreate types.Bool   `tfsdk:"start_on_create"`
	Bootloader    types.String `tfsdk:"bootloader"`
	CPUMode       types.String `tfsdk:"cpu_mode"`
	CPUModel      types.String `tfsdk:"cpu_model"`
	MachineType   types.String `tfsdk:"machine_type"`
	ArchType      types.String `tfsdk:"arch_type"`
	Time          types.String `tfsdk:"time"`
	Status        types.String `tfsdk:"status"`
	MACAddresses  types.List   `tfsdk:"mac_addresses"`
	NICDevices    types.List   `tfsdk:"nic_devices"`
	DiskDevices   types.List   `tfsdk:"disk_devices"`
	CDROMDevices  types.List   `tfsdk:"cdrom_devices"`
}

type NICDeviceModel struct {
	Type                 types.String `tfsdk:"type"`
	MAC                  types.String `tfsdk:"mac"`
	NICAttach            types.String `tfsdk:"nic_attach"`
	TrustGuestRxFilters  types.Bool   `tfsdk:"trust_guest_rx_filters"`
}

type DiskDeviceModel struct {
	Path                types.String `tfsdk:"path"`
	Type                types.String `tfsdk:"type"`
	IOType              types.String `tfsdk:"iotype"`
	PhysicalSectorSize  types.Int64  `tfsdk:"physical_sectorsize"`
	LogicalSectorSize   types.Int64  `tfsdk:"logical_sectorsize"`
}

type CDROMDeviceModel struct {
	Path types.String `tfsdk:"path"`
}

func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a virtual machine on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "VM identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VM name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "VM description",
				Optional:            true,
				Computed:            true,
			},
			"vcpus": schema.Int64Attribute{
				MarkdownDescription: "Number of virtual CPUs",
				Optional:            true,
				Computed:            true,
			},
			"cores": schema.Int64Attribute{
				MarkdownDescription: "Number of cores per socket",
				Optional:            true,
				Computed:            true,
			},
			"threads": schema.Int64Attribute{
				MarkdownDescription: "Number of threads per core",
				Optional:            true,
				Computed:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "Memory in MB",
				Required:            true,
			},
			"min_memory": schema.Int64Attribute{
				MarkdownDescription: "Minimum memory in MB (for memory ballooning)",
				Optional:            true,
				Computed:            true,
			},
			"autostart": schema.BoolAttribute{
				MarkdownDescription: "Start VM automatically on boot",
				Optional:            true,
				Computed:            true,
			},
			"start_on_create": schema.BoolAttribute{
				MarkdownDescription: "Start VM immediately after creation (default: false)",
				Optional:            true,
			},
			"bootloader": schema.StringAttribute{
				MarkdownDescription: "Bootloader type (UEFI, UEFI_CSM, GRUB)",
				Optional:            true,
				Computed:            true,
			},
			"cpu_mode": schema.StringAttribute{
				MarkdownDescription: "CPU mode (CUSTOM, HOST-MODEL, HOST-PASSTHROUGH)",
				Optional:            true,
				Computed:            true,
			},
			"cpu_model": schema.StringAttribute{
				MarkdownDescription: "CPU model (when cpu_mode is CUSTOM)",
				Optional:            true,
				Computed:            true,
			},
			"machine_type": schema.StringAttribute{
				MarkdownDescription: "Machine type (e.g., q35, pc)",
				Optional:            true,
				Computed:            true,
			},
			"arch_type": schema.StringAttribute{
				MarkdownDescription: "Architecture type",
				Optional:            true,
				Computed:            true,
			},
			"time": schema.StringAttribute{
				MarkdownDescription: "Time synchronization (LOCAL or UTC)",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current VM status",
				Computed:            true,
			},
			"mac_addresses": schema.ListAttribute{
				MarkdownDescription: "List of MAC addresses for all NIC devices attached to this VM",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"nic_devices": schema.ListNestedAttribute{
				MarkdownDescription: "Network interface devices to attach to the VM",
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
			"disk_devices": schema.ListNestedAttribute{
				MarkdownDescription: "Disk devices to attach to the VM",
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
			"cdrom_devices": schema.ListNestedAttribute{
				MarkdownDescription: "CD-ROM devices to attach to the VM",
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
		},
	}
}

func (r *VMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":   data.Name.ValueString(),
		"memory": data.Memory.ValueInt64(),
	}

	// Handle optional string fields - only send if non-empty
	if !data.Description.IsNull() && data.Description.ValueString() != "" {
		createReq["description"] = data.Description.ValueString()
	}

	// Handle optional integer fields - only send if > 0
	if !data.VCPUs.IsNull() && data.VCPUs.ValueInt64() > 0 {
		createReq["vcpus"] = data.VCPUs.ValueInt64()
	}
	if !data.Cores.IsNull() && data.Cores.ValueInt64() > 0 {
		createReq["cores"] = data.Cores.ValueInt64()
	}
	if !data.Threads.IsNull() && data.Threads.ValueInt64() > 0 {
		createReq["threads"] = data.Threads.ValueInt64()
	}
	if !data.MinMemory.IsNull() && data.MinMemory.ValueInt64() > 0 {
		createReq["min_memory"] = data.MinMemory.ValueInt64()
	}

	// Handle optional boolean fields
	if !data.Autostart.IsNull() {
		createReq["autostart"] = data.Autostart.ValueBool()
	}

	// Handle optional string fields - only send if non-empty
	if !data.Bootloader.IsNull() && data.Bootloader.ValueString() != "" {
		createReq["bootloader"] = data.Bootloader.ValueString()
	}
	if !data.CPUMode.IsNull() && data.CPUMode.ValueString() != "" {
		createReq["cpu_mode"] = data.CPUMode.ValueString()
	}
	if !data.CPUModel.IsNull() && data.CPUModel.ValueString() != "" {
		createReq["cpu_model"] = data.CPUModel.ValueString()
	}
	if !data.MachineType.IsNull() && data.MachineType.ValueString() != "" {
		createReq["machine_type"] = data.MachineType.ValueString()
	}
	if !data.ArchType.IsNull() && data.ArchType.ValueString() != "" {
		createReq["arch_type"] = data.ArchType.ValueString()
	}
	if !data.Time.IsNull() && data.Time.ValueString() != "" {
		createReq["time"] = data.Time.ValueString()
	}

	respBody, err := r.client.Post("/vm", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create VM, got error: %s", err))
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

	// Create devices after VM creation
	r.createDevices(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Start VM if start_on_create is true
	if !data.StartOnCreate.IsNull() && data.StartOnCreate.ValueBool() {
		startEndpoint := fmt.Sprintf("/vm/id/%s/start", data.ID.ValueString())
		_, err := r.client.Post(startEndpoint, nil)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"VM Start Warning",
				fmt.Sprintf("VM created successfully but failed to start: %s. You can start it manually.", err),
			)
		}
	}

	r.readVM(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readVM(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Name.IsNull() {
		updateReq["name"] = data.Name.ValueString()
	}
	if !data.Description.IsNull() {
		updateReq["description"] = data.Description.ValueString()
	}
	if !data.Memory.IsNull() {
		updateReq["memory"] = data.Memory.ValueInt64()
	}
	if !data.VCPUs.IsNull() {
		updateReq["vcpus"] = data.VCPUs.ValueInt64()
	}
	if !data.Cores.IsNull() {
		updateReq["cores"] = data.Cores.ValueInt64()
	}
	if !data.Threads.IsNull() {
		updateReq["threads"] = data.Threads.ValueInt64()
	}
	if !data.MinMemory.IsNull() {
		updateReq["min_memory"] = data.MinMemory.ValueInt64()
	}
	if !data.Autostart.IsNull() {
		updateReq["autostart"] = data.Autostart.ValueBool()
	}
	if !data.Bootloader.IsNull() {
		updateReq["bootloader"] = data.Bootloader.ValueString()
	}
	if !data.CPUMode.IsNull() {
		updateReq["cpu_mode"] = data.CPUMode.ValueString()
	}
	if !data.CPUModel.IsNull() {
		updateReq["cpu_model"] = data.CPUModel.ValueString()
	}

	endpoint := fmt.Sprintf("/vm/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update VM, got error: %s", err))
		return
	}

	r.readVM(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/vm/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete VM, got error: %s", err))
		return
	}
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VMResource) readVM(ctx context.Context, data *VMResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/vm/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read VM, got error: %s", err))
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

	// Read name
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}

	// Read description (optional)
	if description, ok := result["description"].(string); ok && description != "" {
		data.Description = types.StringValue(description)
	} else {
		data.Description = types.StringNull()
	}

	// Read integer properties
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

	// Read min_memory (optional)
	if minMemory, ok := result["min_memory"].(float64); ok && minMemory > 0 {
		data.MinMemory = types.Int64Value(int64(minMemory))
	} else {
		data.MinMemory = types.Int64Null()
	}

	// Read autostart
	if autostart, ok := result["autostart"].(bool); ok {
		data.Autostart = types.BoolValue(autostart)
	}

	// Read bootloader (optional)
	if bootloader, ok := result["bootloader"].(string); ok && bootloader != "" {
		data.Bootloader = types.StringValue(bootloader)
	} else {
		data.Bootloader = types.StringNull()
	}

	// Read cpu_mode (optional)
	if cpuMode, ok := result["cpu_mode"].(string); ok && cpuMode != "" {
		data.CPUMode = types.StringValue(cpuMode)
	} else {
		data.CPUMode = types.StringNull()
	}

	// Read cpu_model (optional)
	if cpuModel, ok := result["cpu_model"].(string); ok && cpuModel != "" {
		data.CPUModel = types.StringValue(cpuModel)
	} else {
		data.CPUModel = types.StringNull()
	}

	// Read machine_type (optional)
	if machineType, ok := result["machine_type"].(string); ok && machineType != "" {
		data.MachineType = types.StringValue(machineType)
	} else {
		data.MachineType = types.StringNull()
	}

	// Read arch_type (optional)
	if archType, ok := result["arch_type"].(string); ok && archType != "" {
		data.ArchType = types.StringValue(archType)
	} else {
		data.ArchType = types.StringNull()
	}

	// Read time (optional)
	if time, ok := result["time"].(string); ok && time != "" {
		data.Time = types.StringValue(time)
	} else {
		data.Time = types.StringNull()
	}

	// Read status
	if status, ok := result["status"].(map[string]interface{}); ok {
		if state, ok := status["state"].(string); ok {
			data.Status = types.StringValue(state)
		}
	}

	// Read devices from API response
	macAddresses := []string{}
	var nics []NICDeviceModel
	var disks []DiskDeviceModel
	var cdroms []CDROMDeviceModel

	if devices, ok := result["devices"].([]interface{}); ok {
		for _, device := range devices {
			if deviceMap, ok := device.(map[string]interface{}); ok {
				dtype, _ := deviceMap["dtype"].(string)
				attributes, _ := deviceMap["attributes"].(map[string]interface{})

				switch dtype {
				case "NIC":
					nic := NICDeviceModel{}
					if nicType, ok := attributes["type"].(string); ok {
						nic.Type = types.StringValue(nicType)
					} else {
						nic.Type = types.StringNull()
					}
					if mac, ok := attributes["mac"].(string); ok && mac != "" {
						nic.MAC = types.StringValue(mac)
						macAddresses = append(macAddresses, mac)
					} else {
						nic.MAC = types.StringNull()
					}
					if nicAttach, ok := attributes["nic_attach"].(string); ok {
						nic.NICAttach = types.StringValue(nicAttach)
					} else {
						nic.NICAttach = types.StringNull()
					}
					if trustRx, ok := attributes["trust_guest_rx_filters"].(bool); ok {
						nic.TrustGuestRxFilters = types.BoolValue(trustRx)
					} else {
						nic.TrustGuestRxFilters = types.BoolNull()
					}
					nics = append(nics, nic)

				case "DISK":
					disk := DiskDeviceModel{}
					if path, ok := attributes["path"].(string); ok {
						disk.Path = types.StringValue(path)
					} else {
						disk.Path = types.StringNull()
					}
					if diskType, ok := attributes["type"].(string); ok {
						disk.Type = types.StringValue(diskType)
					} else {
						disk.Type = types.StringNull()
					}
					if iotype, ok := attributes["iotype"].(string); ok {
						disk.IOType = types.StringValue(iotype)
					} else {
						disk.IOType = types.StringNull()
					}
					if physSector, ok := attributes["physical_sectorsize"].(float64); ok {
						disk.PhysicalSectorSize = types.Int64Value(int64(physSector))
					} else {
						disk.PhysicalSectorSize = types.Int64Null()
					}
					if logSector, ok := attributes["logical_sectorsize"].(float64); ok {
						disk.LogicalSectorSize = types.Int64Value(int64(logSector))
					} else {
						disk.LogicalSectorSize = types.Int64Null()
					}
					disks = append(disks, disk)

				case "CDROM":
					cdrom := CDROMDeviceModel{}
					if path, ok := attributes["path"].(string); ok {
						cdrom.Path = types.StringValue(path)
					} else {
						cdrom.Path = types.StringNull()
					}
					cdroms = append(cdroms, cdrom)
				}
			}
		}
	}

	// Convert MAC addresses to types.List
	if len(macAddresses) > 0 {
		macList, diagErr := types.ListValueFrom(ctx, types.StringType, macAddresses)
		if diagErr.HasError() {
			diags.Append(diagErr...)
		} else {
			data.MACAddresses = macList
		}
	} else {
		data.MACAddresses = types.ListNull(types.StringType)
	}

	// Convert device lists to types.List
	if len(nics) > 0 {
		nicList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":                    types.StringType,
				"mac":                     types.StringType,
				"nic_attach":              types.StringType,
				"trust_guest_rx_filters":  types.BoolType,
			},
		}, nics)
		if diagErr.HasError() {
			diags.Append(diagErr...)
		} else {
			data.NICDevices = nicList
		}
	} else {
		data.NICDevices = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type":                    types.StringType,
				"mac":                     types.StringType,
				"nic_attach":              types.StringType,
				"trust_guest_rx_filters":  types.BoolType,
			},
		})
	}

	if len(disks) > 0 {
		diskList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"path":                  types.StringType,
				"type":                  types.StringType,
				"iotype":                types.StringType,
				"physical_sectorsize":   types.Int64Type,
				"logical_sectorsize":    types.Int64Type,
			},
		}, disks)
		if diagErr.HasError() {
			diags.Append(diagErr...)
		} else {
			data.DiskDevices = diskList
		}
	} else {
		data.DiskDevices = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"path":                  types.StringType,
				"type":                  types.StringType,
				"iotype":                types.StringType,
				"physical_sectorsize":   types.Int64Type,
				"logical_sectorsize":    types.Int64Type,
			},
		})
	}

	if len(cdroms) > 0 {
		cdromList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"path": types.StringType,
			},
		}, cdroms)
		if diagErr.HasError() {
			diags.Append(diagErr...)
		} else {
			data.CDROMDevices = cdromList
		}
	} else {
		data.CDROMDevices = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"path": types.StringType,
			},
		})
	}
}

// createDevices creates NIC, disk, and CDROM devices for a VM
func (r *VMResource) createDevices(ctx context.Context, data *VMResourceModel, diags *diag.Diagnostics) {
	vmID := data.ID.ValueString()
	deviceOrder := 1000 // Starting order for devices

	// Create NIC devices
	if !data.NICDevices.IsNull() && !data.NICDevices.IsUnknown() {
		var nics []NICDeviceModel
		diagErr := data.NICDevices.ElementsAs(ctx, &nics, false)
		if diagErr.HasError() {
			diags.Append(diagErr...)
			return
		}

		for _, nic := range nics {
			deviceReq := map[string]interface{}{
				"vm":    vmID,
				"dtype": "NIC",
				"order": deviceOrder,
				"attributes": map[string]interface{}{
					"nic_attach": nic.NICAttach.ValueString(),
				},
			}

			// Add optional NIC attributes
			if !nic.Type.IsNull() && nic.Type.ValueString() != "" {
				deviceReq["attributes"].(map[string]interface{})["type"] = nic.Type.ValueString()
			} else {
				deviceReq["attributes"].(map[string]interface{})["type"] = "VIRTIO"
			}

			if !nic.MAC.IsNull() && nic.MAC.ValueString() != "" {
				deviceReq["attributes"].(map[string]interface{})["mac"] = nic.MAC.ValueString()
			}

			if !nic.TrustGuestRxFilters.IsNull() {
				deviceReq["attributes"].(map[string]interface{})["trust_guest_rx_filters"] = nic.TrustGuestRxFilters.ValueBool()
			} else {
				deviceReq["attributes"].(map[string]interface{})["trust_guest_rx_filters"] = false
			}

			_, err := r.client.Post("/vm/device", deviceReq)
			if err != nil {
				diags.AddError("Device Creation Error", fmt.Sprintf("Unable to create NIC device: %s", err))
				return
			}
			deviceOrder++
		}
	}

	// Create Disk devices
	if !data.DiskDevices.IsNull() && !data.DiskDevices.IsUnknown() {
		var disks []DiskDeviceModel
		diagErr := data.DiskDevices.ElementsAs(ctx, &disks, false)
		if diagErr.HasError() {
			diags.Append(diagErr...)
			return
		}

		for _, disk := range disks {
			deviceReq := map[string]interface{}{
				"vm":    vmID,
				"dtype": "DISK",
				"order": deviceOrder,
				"attributes": map[string]interface{}{
					"path": disk.Path.ValueString(),
				},
			}

			// Add optional disk attributes
			if !disk.Type.IsNull() && disk.Type.ValueString() != "" {
				deviceReq["attributes"].(map[string]interface{})["type"] = disk.Type.ValueString()
			} else {
				deviceReq["attributes"].(map[string]interface{})["type"] = "VIRTIO"
			}

			if !disk.IOType.IsNull() && disk.IOType.ValueString() != "" {
				deviceReq["attributes"].(map[string]interface{})["iotype"] = disk.IOType.ValueString()
			}

			if !disk.PhysicalSectorSize.IsNull() && disk.PhysicalSectorSize.ValueInt64() > 0 {
				deviceReq["attributes"].(map[string]interface{})["physical_sectorsize"] = disk.PhysicalSectorSize.ValueInt64()
			}

			if !disk.LogicalSectorSize.IsNull() && disk.LogicalSectorSize.ValueInt64() > 0 {
				deviceReq["attributes"].(map[string]interface{})["logical_sectorsize"] = disk.LogicalSectorSize.ValueInt64()
			}

			_, err := r.client.Post("/vm/device", deviceReq)
			if err != nil {
				diags.AddError("Device Creation Error", fmt.Sprintf("Unable to create disk device: %s", err))
				return
			}
			deviceOrder++
		}
	}

	// Create CDROM devices
	if !data.CDROMDevices.IsNull() && !data.CDROMDevices.IsUnknown() {
		var cdroms []CDROMDeviceModel
		diagErr := data.CDROMDevices.ElementsAs(ctx, &cdroms, false)
		if diagErr.HasError() {
			diags.Append(diagErr...)
			return
		}

		for _, cdrom := range cdroms {
			deviceReq := map[string]interface{}{
				"vm":    vmID,
				"dtype": "CDROM",
				"order": deviceOrder,
				"attributes": map[string]interface{}{
					"path": cdrom.Path.ValueString(),
				},
			}

			_, err := r.client.Post("/vm/device", deviceReq)
			if err != nil {
				diags.AddError("Device Creation Error", fmt.Sprintf("Unable to create CDROM device: %s", err))
				return
			}
			deviceOrder++
		}
	}
}


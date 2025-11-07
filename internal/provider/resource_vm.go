package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	VCPUs               types.Int64  `tfsdk:"vcpus"`
	Cores               types.Int64  `tfsdk:"cores"`
	Threads             types.Int64  `tfsdk:"threads"`
	Memory              types.Int64  `tfsdk:"memory"`
	MinMemory           types.Int64  `tfsdk:"min_memory"`
	Autostart           types.Bool   `tfsdk:"autostart"`
	StartOnCreate       types.Bool   `tfsdk:"start_on_create"`
	DesiredState        types.String `tfsdk:"desired_state"`
	Bootloader          types.String `tfsdk:"bootloader"`
	CPUMode             types.String `tfsdk:"cpu_mode"`
	CPUModel            types.String `tfsdk:"cpu_model"`
	MachineType         types.String `tfsdk:"machine_type"`
	ArchType            types.String `tfsdk:"arch_type"`
	Time                types.String `tfsdk:"time"`
	Status              types.String `tfsdk:"status"`
	MACAddresses        types.List   `tfsdk:"mac_addresses"`
	HideFromMSR         types.Bool   `tfsdk:"hide_from_msr"`
	EnsureDisplayDevice types.Bool   `tfsdk:"ensure_display_device"`
	NICDevices          types.List   `tfsdk:"nic_devices"`
	DiskDevices         types.List   `tfsdk:"disk_devices"`
	CDROMDevices        types.List   `tfsdk:"cdrom_devices"`
	DisplayDevices      types.List   `tfsdk:"display_devices"`
	PCIDevices          types.List   `tfsdk:"pci_devices"`
}

type NICDeviceModel struct {
	Type                 types.String `tfsdk:"type"`
	MAC                  types.String `tfsdk:"mac"`
	NICAttach            types.String `tfsdk:"nic_attach"`
	TrustGuestRxFilters  types.Bool   `tfsdk:"trust_guest_rx_filters"`
	Order                types.Int64  `tfsdk:"order"`
}

type DiskDeviceModel struct {
	Path                types.String `tfsdk:"path"`
	Type                types.String `tfsdk:"type"`
	IOType              types.String `tfsdk:"iotype"`
	PhysicalSectorSize  types.Int64  `tfsdk:"physical_sectorsize"`
	LogicalSectorSize   types.Int64  `tfsdk:"logical_sectorsize"`
	Order               types.Int64  `tfsdk:"order"`
}

type CDROMDeviceModel struct {
	Path  types.String `tfsdk:"path"`
	Order types.Int64  `tfsdk:"order"`
}

type PCIDeviceModel struct {
	PPTDev types.String `tfsdk:"pptdev"`
	Order  types.Int64  `tfsdk:"order"`
}

type DisplayDeviceModel struct {
	Port       types.Int64  `tfsdk:"port"`
	Bind       types.String `tfsdk:"bind"`
	Password   types.String `tfsdk:"password"`
	Web        types.Bool   `tfsdk:"web"`
	Type       types.String `tfsdk:"type"`
	Resolution types.String `tfsdk:"resolution"`
	WebPort    types.Int64  `tfsdk:"web_port"`
	Wait       types.Bool   `tfsdk:"wait"`
	Order      types.Int64  `tfsdk:"order"`
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
				MarkdownDescription: "Minimum memory in MB (for memory ballooning). Defaults to the value of memory if not specified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"autostart": schema.BoolAttribute{
				MarkdownDescription: "Start VM automatically on boot",
				Optional:            true,
				Computed:            true,
			},
			"start_on_create": schema.BoolAttribute{
				MarkdownDescription: "**Deprecated:** Use `desired_state` instead. Start VM immediately after creation (default: false). This attribute is deprecated in favor of `desired_state` which provides more control over VM lifecycle.",
				Optional:            true,
				DeprecationMessage:  "Use desired_state instead for more granular control over VM lifecycle state",
			},
			"desired_state": schema.StringAttribute{
				MarkdownDescription: "Desired state of the VM. Valid values: `RUNNING`, `STOPPED`, `SUSPENDED`. If not specified, defaults to `STOPPED`. This attribute controls the VM's lifecycle state and takes precedence over `start_on_create` if both are specified.",
				Optional:            true,
				Computed:            true,
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
						"order": schema.Int64Attribute{
							MarkdownDescription: "Boot order for this device. Lower values boot first. If not specified, devices are ordered by type (NICs, then disks, then CDROMs)",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
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
						"order": schema.Int64Attribute{
							MarkdownDescription: "Boot order for this device. Lower values boot first. If not specified, devices are ordered by type (NICs, then disks, then CDROMs)",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
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
						"order": schema.Int64Attribute{
							MarkdownDescription: "Boot order for this device. Lower values boot first. If not specified, devices are ordered by type (NICs, then disks, then CDROMs)",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"display_devices": schema.ListNestedAttribute{
				MarkdownDescription: "Display devices (SPICE/VNC) to attach to the VM for console access",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							MarkdownDescription: "Port number for the display server (e.g., 5900 for VNC, 5902 for SPICE)",
							Optional:            true,
							Computed:            true,
						},
						"bind": schema.StringAttribute{
							MarkdownDescription: "IP address to bind to (default: 0.0.0.0)",
							Optional:            true,
							Computed:            true,
						},
						"password": schema.StringAttribute{
							MarkdownDescription: "Password for display access",
							Optional:            true,
							Sensitive:           true,
						},
						"web": schema.BoolAttribute{
							MarkdownDescription: "Enable web access (default: true)",
							Optional:            true,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Display type: SPICE or VNC (default: SPICE)",
							Optional:            true,
							Computed:            true,
						},
						"resolution": schema.StringAttribute{
							MarkdownDescription: "Display resolution (e.g., 1024x768, 1920x1080)",
							Optional:            true,
							Computed:            true,
						},
						"web_port": schema.Int64Attribute{
							MarkdownDescription: "Port for web access",
							Optional:            true,
							Computed:            true,
						},
						"wait": schema.BoolAttribute{
							MarkdownDescription: "Wait for client connection before starting VM",
							Optional:            true,
							Computed:            true,
						},
						"order": schema.Int64Attribute{
							MarkdownDescription: "Boot order for this device",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"pci_devices": schema.ListNestedAttribute{
				MarkdownDescription: "PCI passthrough devices to attach to the VM (requires IOMMU enabled)",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pptdev": schema.StringAttribute{
							MarkdownDescription: "PCI device ID to pass through (e.g., 'pci_0000_3b_00_0'). Use truenas_vm_pci_passthrough_devices data source to discover available devices.",
							Required:            true,
						},
						"order": schema.Int64Attribute{
							MarkdownDescription: "Boot order for this device. Lower values boot first.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"hide_from_msr": schema.BoolAttribute{
				MarkdownDescription: "Hide KVM hypervisor from MSR discovery. Useful for GPU passthrough to avoid detection. Default: false",
				Optional:            true,
				Computed:            true,
			},
			"ensure_display_device": schema.BoolAttribute{
				MarkdownDescription: "Ensure a virtual display device is attached. Set to false when using GPU passthrough. Default: true",
				Optional:            true,
				Computed:            true,
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
	// Set min_memory: if explicitly provided, use that value; otherwise default to memory value
	// This disables memory ballooning by default, preventing "virtio_balloon: Out of puff!" errors
	if !data.MinMemory.IsNull() && data.MinMemory.ValueInt64() > 0 {
		createReq["min_memory"] = data.MinMemory.ValueInt64()
	} else {
		// Default to memory value to disable ballooning
		createReq["min_memory"] = data.Memory.ValueInt64()
	}

	// Handle optional boolean fields
	if !data.Autostart.IsNull() {
		createReq["autostart"] = data.Autostart.ValueBool()
	}
	if !data.HideFromMSR.IsNull() {
		createReq["hide_from_msr"] = data.HideFromMSR.ValueBool()
	}
	if !data.EnsureDisplayDevice.IsNull() {
		createReq["ensure_display_device"] = data.EnsureDisplayDevice.ValueBool()
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

	// Determine desired state for the VM
	// Priority: desired_state > start_on_create > default (STOPPED)
	desiredState := "STOPPED"
	
	if !data.DesiredState.IsNull() && data.DesiredState.ValueString() != "" {
		// Use explicit desired_state if provided
		desiredState = data.DesiredState.ValueString()
	} else if !data.StartOnCreate.IsNull() && data.StartOnCreate.ValueBool() {
		// Fall back to start_on_create for backward compatibility
		desiredState = "RUNNING"
	}
	
	// Transition VM to desired state
	r.transitionVMState(ctx, data.ID.ValueString(), desiredState, &resp.Diagnostics)
	
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
	// Set min_memory: if explicitly provided, use that value; otherwise default to memory value
	// This disables memory ballooning by default, preventing "virtio_balloon: Out of puff!" errors
	if !data.MinMemory.IsNull() && data.MinMemory.ValueInt64() > 0 {
		updateReq["min_memory"] = data.MinMemory.ValueInt64()
	} else {
		// Default to memory value to disable ballooning
		updateReq["min_memory"] = data.Memory.ValueInt64()
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

	// Handle state transitions if desired_state is specified
	if !data.DesiredState.IsNull() && data.DesiredState.ValueString() != "" {
		r.transitionVMState(ctx, data.ID.ValueString(), data.DesiredState.ValueString(), &resp.Diagnostics)
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

// getCurrentVMState retrieves the current state of the VM
func (r *VMResource) getCurrentVMState(vmID string) (string, error) {
	result, err := r.client.GetVMStatus(vmID)
	if err != nil {
		return "", err
	}
	
	if status, ok := result["status"].(map[string]interface{}); ok {
		if state, ok := status["state"].(string); ok {
			return strings.ToUpper(state), nil
		}
	}
	
	return "", fmt.Errorf("unable to determine VM state")
}

// transitionVMState transitions the VM to the desired state with retry logic
func (r *VMResource) transitionVMState(ctx context.Context, vmID string, desiredState string, diags *diag.Diagnostics) {
	if desiredState == "" {
		// No desired state specified, don't attempt transition
		return
	}
	
	// Normalize desired state to uppercase
	desiredState = strings.ToUpper(desiredState)
	
	// Validate desired state
	validStates := map[string]bool{
		"RUNNING":   true,
		"STOPPED":   true,
		"SUSPENDED": true,
	}
	
	if !validStates[desiredState] {
		diags.AddError(
			"Invalid Desired State",
			fmt.Sprintf("desired_state must be one of: RUNNING, STOPPED, SUSPENDED. Got: %s", desiredState),
		)
		return
	}
	
	// Get current state
	currentState, err := r.getCurrentVMState(vmID)
	if err != nil {
		diags.AddWarning(
			"VM State Check Warning",
			fmt.Sprintf("Unable to check current VM state: %s", err),
		)
		return
	}
	
	// If already in desired state, nothing to do
	if currentState == desiredState {
		return
	}
	
	// Define state transition logic
	var transitionErr error
	maxRetries := 3
	retryDelay := 5 * time.Second
	timeout := 5 * time.Minute
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}
		
		// Re-check current state before attempting transition
		currentState, err = r.getCurrentVMState(vmID)
		if err != nil {
			transitionErr = err
			continue
		}
		
		// If already in desired state, we're done
		if currentState == desiredState {
			return
		}
		
		// Perform state transition based on current and desired states
		switch {
		case desiredState == "RUNNING" && currentState == "STOPPED":
			_, transitionErr = r.client.StartVM(vmID)
		case desiredState == "RUNNING" && currentState == "SUSPENDED":
			_, transitionErr = r.client.ResumeVM(vmID)
		case desiredState == "STOPPED" && currentState == "RUNNING":
			// Try graceful stop first
			_, transitionErr = r.client.StopVM(vmID)
			if transitionErr != nil {
				// If graceful stop fails, try power off
				time.Sleep(2 * time.Second)
				_, transitionErr = r.client.PowerOffVM(vmID)
			}
		case desiredState == "STOPPED" && currentState == "SUSPENDED":
			// Resume first, then stop
			_, err = r.client.ResumeVM(vmID)
			if err != nil {
				transitionErr = err
				continue
			}
			time.Sleep(2 * time.Second)
			_, transitionErr = r.client.StopVM(vmID)
			if transitionErr != nil {
				time.Sleep(2 * time.Second)
				_, transitionErr = r.client.PowerOffVM(vmID)
			}
		case desiredState == "SUSPENDED" && currentState == "RUNNING":
			_, transitionErr = r.client.SuspendVM(vmID)
		case desiredState == "SUSPENDED" && currentState == "STOPPED":
			// Start first, then suspend
			_, err = r.client.StartVM(vmID)
			if err != nil {
				transitionErr = err
				continue
			}
			time.Sleep(2 * time.Second)
			_, transitionErr = r.client.SuspendVM(vmID)
		default:
			diags.AddWarning(
				"Unsupported State Transition",
				fmt.Sprintf("Cannot transition from %s to %s", currentState, desiredState),
			)
			return
		}
		
		if transitionErr != nil {
			continue
		}
		
		// Wait for state transition to complete
		startTime := time.Now()
		for time.Since(startTime) < timeout {
			time.Sleep(2 * time.Second)
			
			newState, err := r.getCurrentVMState(vmID)
			if err != nil {
				continue
			}
			
			if newState == desiredState {
				return
			}
		}
		
		transitionErr = fmt.Errorf("timeout waiting for VM to reach %s state", desiredState)
	}
	
	// If we get here, all retries failed
	if transitionErr != nil {
		diags.AddWarning(
			"VM State Transition Warning",
			fmt.Sprintf("Unable to transition VM to %s state after %d attempts: %s. You may need to manually manage the VM state.", desiredState, maxRetries, transitionErr),
		)
	}
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

	// Read hide_from_msr
	if hideFromMSR, ok := result["hide_from_msr"].(bool); ok {
		data.HideFromMSR = types.BoolValue(hideFromMSR)
	}

	// Read ensure_display_device
	if ensureDisplayDevice, ok := result["ensure_display_device"].(bool); ok {
		data.EnsureDisplayDevice = types.BoolValue(ensureDisplayDevice)
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
			
			// Set DesiredState to current state if not explicitly set
			// This ensures the computed value reflects the actual state
			if data.DesiredState.IsNull() || data.DesiredState.IsUnknown() {
				normalizedState := strings.ToUpper(state)
				data.DesiredState = types.StringValue(normalizedState)
			}
		}
	}

	// Read devices from API response
	macAddresses := []string{}
	var nics []NICDeviceModel
	var disks []DiskDeviceModel
	var cdroms []CDROMDeviceModel
	var displays []DisplayDeviceModel
	var pciDevices []PCIDeviceModel

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
					// Read order from device level (not attributes)
					if order, ok := deviceMap["order"].(float64); ok {
						nic.Order = types.Int64Value(int64(order))
					} else {
						nic.Order = types.Int64Null()
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
					// Read order from device level (not attributes)
					if order, ok := deviceMap["order"].(float64); ok {
						disk.Order = types.Int64Value(int64(order))
					} else {
						disk.Order = types.Int64Null()
					}
					disks = append(disks, disk)

				case "CDROM":
					cdrom := CDROMDeviceModel{}
					if path, ok := attributes["path"].(string); ok {
						cdrom.Path = types.StringValue(path)
					} else {
						cdrom.Path = types.StringNull()
					}
					// Read order from device level (not attributes)
					if order, ok := deviceMap["order"].(float64); ok {
						cdrom.Order = types.Int64Value(int64(order))
					} else {
						cdrom.Order = types.Int64Null()
					}
					cdroms = append(cdroms, cdrom)

				case "DISPLAY":
					display := DisplayDeviceModel{}
					if port, ok := attributes["port"].(float64); ok {
						display.Port = types.Int64Value(int64(port))
					} else {
						display.Port = types.Int64Null()
					}
					if bind, ok := attributes["bind"].(string); ok {
						display.Bind = types.StringValue(bind)
					} else {
						display.Bind = types.StringNull()
					}
					if password, ok := attributes["password"].(string); ok {
						display.Password = types.StringValue(password)
					} else {
						display.Password = types.StringNull()
					}
					if web, ok := attributes["web"].(bool); ok {
						display.Web = types.BoolValue(web)
					} else {
						display.Web = types.BoolNull()
					}
					if displayType, ok := attributes["type"].(string); ok {
						display.Type = types.StringValue(displayType)
					} else {
						display.Type = types.StringNull()
					}
					if resolution, ok := attributes["resolution"].(string); ok {
						display.Resolution = types.StringValue(resolution)
					} else {
						display.Resolution = types.StringNull()
					}
					if webPort, ok := attributes["web_port"].(float64); ok {
						display.WebPort = types.Int64Value(int64(webPort))
					} else {
						display.WebPort = types.Int64Null()
					}
					if wait, ok := attributes["wait"].(bool); ok {
						display.Wait = types.BoolValue(wait)
					} else {
						display.Wait = types.BoolNull()
					}
					// Read order from device level (not attributes)
					if order, ok := deviceMap["order"].(float64); ok {
						display.Order = types.Int64Value(int64(order))
					} else {
						display.Order = types.Int64Null()
					}
					displays = append(displays, display)

				case "PCI":
					pci := PCIDeviceModel{}
					if pptdev, ok := attributes["pptdev"].(string); ok {
						pci.PPTDev = types.StringValue(pptdev)
					} else {
						pci.PPTDev = types.StringNull()
					}
					// Read order from device level (not attributes)
					if order, ok := deviceMap["order"].(float64); ok {
						pci.Order = types.Int64Value(int64(order))
					} else {
						pci.Order = types.Int64Null()
					}
					pciDevices = append(pciDevices, pci)
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
				"order":                   types.Int64Type,
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
				"order":                   types.Int64Type,
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
				"order":                 types.Int64Type,
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
				"order":                 types.Int64Type,
			},
		})
	}

	if len(cdroms) > 0 {
		cdromList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"path":  types.StringType,
				"order": types.Int64Type,
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
				"path":  types.StringType,
				"order": types.Int64Type,
			},
		})
	}

	if len(displays) > 0 {
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
				"order":      types.Int64Type,
			},
		}, displays)
		if diagErr.HasError() {
			diags.Append(diagErr...)
		} else {
			data.DisplayDevices = displayList
		}
	} else {
		data.DisplayDevices = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"port":       types.Int64Type,
				"bind":       types.StringType,
				"password":   types.StringType,
				"web":        types.BoolType,
				"type":       types.StringType,
				"resolution": types.StringType,
				"web_port":   types.Int64Type,
				"wait":       types.BoolType,
				"order":      types.Int64Type,
			},
		})
	}

	if len(pciDevices) > 0 {
		pciList, diagErr := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"pptdev": types.StringType,
				"order":  types.Int64Type,
			},
		}, pciDevices)
		if diagErr.HasError() {
			diags.Append(diagErr...)
		} else {
			data.PCIDevices = pciList
		}
	} else {
		data.PCIDevices = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"pptdev": types.StringType,
				"order":  types.Int64Type,
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
			// Use user-specified order if provided, otherwise use auto-incrementing order
			order := deviceOrder
			if !nic.Order.IsNull() && nic.Order.ValueInt64() > 0 {
				order = int(nic.Order.ValueInt64())
			}

			deviceReq := map[string]interface{}{
				"vm":    vmID,
				"dtype": "NIC",
				"order": order,
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
			// Use user-specified order if provided, otherwise use auto-incrementing order
			order := deviceOrder
			if !disk.Order.IsNull() && disk.Order.ValueInt64() > 0 {
				order = int(disk.Order.ValueInt64())
			}

			deviceReq := map[string]interface{}{
				"vm":    vmID,
				"dtype": "DISK",
				"order": order,
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
			// Use user-specified order if provided, otherwise use auto-incrementing order
			order := deviceOrder
			if !cdrom.Order.IsNull() && cdrom.Order.ValueInt64() > 0 {
				order = int(cdrom.Order.ValueInt64())
			}

			deviceReq := map[string]interface{}{
				"vm":    vmID,
				"dtype": "CDROM",
				"order": order,
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

	// Create Display devices
	if !data.DisplayDevices.IsNull() && !data.DisplayDevices.IsUnknown() {
		var displays []DisplayDeviceModel
		diagErr := data.DisplayDevices.ElementsAs(ctx, &displays, false)
		if diagErr.HasError() {
			diags.Append(diagErr...)
			return
		}

		for _, display := range displays {
			// Use user-specified order if provided, otherwise use auto-incrementing order
			order := deviceOrder
			if !display.Order.IsNull() && display.Order.ValueInt64() > 0 {
				order = int(display.Order.ValueInt64())
			}

			attributes := make(map[string]interface{})

			// Add optional attributes only if they're specified
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

			deviceReq := map[string]interface{}{
				"vm":         vmID,
				"dtype":      "DISPLAY",
				"order":      order,
				"attributes": attributes,
			}

			_, err := r.client.Post("/vm/device", deviceReq)
			if err != nil {
				diags.AddError("Device Creation Error", fmt.Sprintf("Unable to create DISPLAY device: %s", err))
				return
			}
			deviceOrder++
		}
	}

	// Create PCI passthrough devices
	if !data.PCIDevices.IsNull() && !data.PCIDevices.IsUnknown() {
		var pciDevices []PCIDeviceModel
		diagErr := data.PCIDevices.ElementsAs(ctx, &pciDevices, false)
		if diagErr.HasError() {
			diags.Append(diagErr...)
			return
		}

		for _, pci := range pciDevices {
			// Use user-specified order if provided, otherwise use auto-incrementing order
			order := deviceOrder
			if !pci.Order.IsNull() && pci.Order.ValueInt64() > 0 {
				order = int(pci.Order.ValueInt64())
			}

			deviceReq := map[string]interface{}{
				"vm":    vmID,
				"dtype": "PCI",
				"order": order,
				"attributes": map[string]interface{}{
					"pptdev": pci.PPTDev.ValueString(),
				},
			}

			_, err := r.client.Post("/vm/device", deviceReq)
			if err != nil {
				diags.AddError("Device Creation Error", fmt.Sprintf("Unable to create PCI passthrough device: %s", err))
				return
			}
			deviceOrder++
		}
	}
}


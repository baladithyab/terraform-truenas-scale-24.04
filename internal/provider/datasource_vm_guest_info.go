package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VMGuestInfoDataSource{}

func NewVMGuestInfoDataSource() datasource.DataSource {
	return &VMGuestInfoDataSource{}
}

type VMGuestInfoDataSource struct{}

type VMGuestInfoDataSourceModel struct {
	VMName       types.String `tfsdk:"vm_name"`
	TrueNASHost  types.String `tfsdk:"truenas_host"`
	SSHUser      types.String `tfsdk:"ssh_user"`
	SSHKeyPath   types.String `tfsdk:"ssh_key_path"`
	IPAddresses  types.List   `tfsdk:"ip_addresses"`
	Hostname     types.String `tfsdk:"hostname"`
	OSName       types.String `tfsdk:"os_name"`
	OSVersion    types.String `tfsdk:"os_version"`
}

func (d *VMGuestInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_guest_info"
}

func (d *VMGuestInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Retrieves guest information from a VM using QEMU Guest Agent via SSH.

**Requirements:**
- QEMU Guest Agent must be installed and running in the VM
- SSH access to the TrueNAS host
- SSH key-based authentication configured

**Note:** This data source queries the TrueNAS host directly via SSH to run virsh commands.
It does not use the TrueNAS API because the API does not expose guest agent information.

**Example Usage:**

` + "```hcl" + `
data "truenas_vm_guest_info" "ubuntu" {
  vm_name        = "ubuntu-vm"
  truenas_host   = "10.0.0.83"
  ssh_user       = "root"
  ssh_key_path   = "~/.ssh/id_rsa"
}

output "ubuntu_ips" {
  value = data.truenas_vm_guest_info.ubuntu.ip_addresses
}
` + "```",
		Attributes: map[string]schema.Attribute{
			"vm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the VM to query (must match the VM name in TrueNAS)",
				Required:            true,
			},
			"truenas_host": schema.StringAttribute{
				MarkdownDescription: "TrueNAS host IP address or hostname for SSH connection",
				Required:            true,
			},
			"ssh_user": schema.StringAttribute{
				MarkdownDescription: "SSH username for connecting to TrueNAS (default: root)",
				Optional:            true,
			},
			"ssh_key_path": schema.StringAttribute{
				MarkdownDescription: "Path to SSH private key for authentication (default: ~/.ssh/id_rsa)",
				Optional:            true,
			},
			"ip_addresses": schema.ListAttribute{
				MarkdownDescription: "List of IP addresses reported by the guest agent",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname reported by the guest agent",
				Computed:            true,
			},
			"os_name": schema.StringAttribute{
				MarkdownDescription: "Operating system name reported by the guest agent",
				Computed:            true,
			},
			"os_version": schema.StringAttribute{
				MarkdownDescription: "Operating system version reported by the guest agent",
				Computed:            true,
			},
		},
	}
}

func (d *VMGuestInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMGuestInfoDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	sshUser := "root"
	if !data.SSHUser.IsNull() {
		sshUser = data.SSHUser.ValueString()
	}

	sshKeyPath := "~/.ssh/id_rsa"
	if !data.SSHKeyPath.IsNull() {
		sshKeyPath = data.SSHKeyPath.ValueString()
	}

	vmName := data.VMName.ValueString()
	trueNASHost := data.TrueNASHost.ValueString()

	// Query guest agent for network interfaces
	networkCmd := fmt.Sprintf(
		`ssh -i %s -o StrictHostKeyChecking=no %s@%s "virsh qemu-agent-command %s '{\"execute\":\"guest-network-get-interfaces\"}'"`,
		sshKeyPath, sshUser, trueNASHost, vmName,
	)

	output, err := exec.Command("sh", "-c", networkCmd).Output()
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Guest Agent Query Failed",
			fmt.Sprintf("Failed to query guest agent for network info: %s\nCommand: %s\nMake sure:\n1. QEMU Guest Agent is installed and running in the VM\n2. SSH access is configured\n3. VM name matches exactly", err, networkCmd),
		)
		// Set empty values
		data.IPAddresses = types.ListNull(types.StringType)
		data.Hostname = types.StringNull()
		data.OSName = types.StringNull()
		data.OSVersion = types.StringNull()
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Parse network interfaces
	var networkResult map[string]interface{}
	if err := json.Unmarshal(output, &networkResult); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Failed to parse guest agent response: %s\nOutput: %s", err, string(output)),
		)
		return
	}

	// Extract IP addresses
	ipAddresses := []string{}
	if returnData, ok := networkResult["return"].([]interface{}); ok {
		for _, iface := range returnData {
			if ifaceMap, ok := iface.(map[string]interface{}); ok {
				// Skip loopback interface
				if name, ok := ifaceMap["name"].(string); ok && name == "lo" {
					continue
				}

				// Get IP addresses
				if ipAddrs, ok := ifaceMap["ip-addresses"].([]interface{}); ok {
					for _, ipAddr := range ipAddrs {
						if ipMap, ok := ipAddr.(map[string]interface{}); ok {
							if ip, ok := ipMap["ip-address"].(string); ok {
								// Skip link-local addresses
								if !strings.HasPrefix(ip, "fe80:") && !strings.HasPrefix(ip, "169.254.") {
									ipAddresses = append(ipAddresses, ip)
								}
							}
						}
					}
				}
			}
		}
	}

	// Convert IP addresses to types.List
	if len(ipAddresses) > 0 {
		ipList, diagErr := types.ListValueFrom(ctx, types.StringType, ipAddresses)
		if diagErr.HasError() {
			resp.Diagnostics.Append(diagErr...)
		} else {
			data.IPAddresses = ipList
		}
	} else {
		data.IPAddresses = types.ListNull(types.StringType)
	}

	// Query guest agent for OS info
	osInfoCmd := fmt.Sprintf(
		`ssh -i %s -o StrictHostKeyChecking=no %s@%s "virsh qemu-agent-command %s '{\"execute\":\"guest-get-osinfo\"}'"`,
		sshKeyPath, sshUser, trueNASHost, vmName,
	)

	osOutput, err := exec.Command("sh", "-c", osInfoCmd).Output()
	if err == nil {
		var osResult map[string]interface{}
		if err := json.Unmarshal(osOutput, &osResult); err == nil {
			if returnData, ok := osResult["return"].(map[string]interface{}); ok {
				if name, ok := returnData["name"].(string); ok {
					data.OSName = types.StringValue(name)
				}
				if version, ok := returnData["version"].(string); ok {
					data.OSVersion = types.StringValue(version)
				}
			}
		}
	}

	// If OS info not available, set to null
	if data.OSName.IsNull() {
		data.OSName = types.StringNull()
	}
	if data.OSVersion.IsNull() {
		data.OSVersion = types.StringNull()
	}

	// Query for hostname
	hostnameCmd := fmt.Sprintf(
		`ssh -i %s -o StrictHostKeyChecking=no %s@%s "virsh qemu-agent-command %s '{\"execute\":\"guest-get-host-name\"}'"`,
		sshKeyPath, sshUser, trueNASHost, vmName,
	)

	hostnameOutput, err := exec.Command("sh", "-c", hostnameCmd).Output()
	if err == nil {
		var hostnameResult map[string]interface{}
		if err := json.Unmarshal(hostnameOutput, &hostnameResult); err == nil {
			if returnData, ok := hostnameResult["return"].(map[string]interface{}); ok {
				if hostname, ok := returnData["host-name"].(string); ok {
					data.Hostname = types.StringValue(hostname)
				}
			}
		}
	}

	// If hostname not available, set to null
	if data.Hostname.IsNull() {
		data.Hostname = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}


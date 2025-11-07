package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VMGuestInfoDataSource{}
var _ datasource.DataSourceWithConfigValidators = &VMGuestInfoDataSource{}

func NewVMGuestInfoDataSource() datasource.DataSource {
	return &VMGuestInfoDataSource{}
}

type VMGuestInfoDataSource struct{}

type VMGuestInfoDataSourceModel struct {
	VMName                    types.String `tfsdk:"vm_name"`
	TrueNASHost               types.String `tfsdk:"truenas_host"`
	SSHUser                   types.String `tfsdk:"ssh_user"`
	SSHKeyPath                types.String `tfsdk:"ssh_key_path"`
	SSHPassword               types.String `tfsdk:"ssh_password"`
	SSHStrictHostKeyChecking  types.Bool   `tfsdk:"ssh_strict_host_key_checking"`
	SSHTimeoutSeconds         types.Int64  `tfsdk:"ssh_timeout_seconds"`
	IPAddresses               types.List   `tfsdk:"ip_addresses"`
	Hostname                  types.String `tfsdk:"hostname"`
	OSName                    types.String `tfsdk:"os_name"`
	OSVersion                 types.String `tfsdk:"os_version"`
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
- SSH key-based authentication OR password authentication (at least one required)
- For password authentication: sshpass must be installed on the Terraform execution host

**Note:** This data source queries the TrueNAS host directly via SSH to run virsh commands.
It does not use the TrueNAS API because the API does not expose guest agent information.

**Example Usage (with SSH key):**

` + "```hcl" + `
data "truenas_vm_guest_info" "ubuntu" {
  vm_name                      = "ubuntu-vm"
  truenas_host                 = "10.0.0.83"
  ssh_user                     = "root"
  ssh_key_path                 = "~/.ssh/id_rsa"
  ssh_strict_host_key_checking = true
  ssh_timeout_seconds          = 30
}

output "ubuntu_ips" {
  value = data.truenas_vm_guest_info.ubuntu.ip_addresses
}
` + "```" + `

**Example Usage (with password):**

` + "```hcl" + `
data "truenas_vm_guest_info" "talos" {
  vm_name                      = "talos-vm"
  truenas_host                 = "10.0.0.83"
  ssh_user                     = "root"
  ssh_password                 = var.truenas_ssh_password
  ssh_strict_host_key_checking = false
  ssh_timeout_seconds          = 60
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
				MarkdownDescription: "Path to SSH private key for authentication (default: ~/.ssh/id_rsa). Either ssh_key_path or ssh_password must be provided.",
				Optional:            true,
			},
			"ssh_password": schema.StringAttribute{
				MarkdownDescription: "SSH password for authentication. Requires sshpass to be installed. Either ssh_key_path or ssh_password must be provided.",
				Optional:            true,
				Sensitive:           true,
			},
			"ssh_strict_host_key_checking": schema.BoolAttribute{
				MarkdownDescription: "Enable strict host key checking for SSH connections. When true, SSH will verify the host key against known_hosts. When false, SSH will accept any host key (less secure but useful for dynamic environments). Default: true",
				Optional:            true,
			},
			"ssh_timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "SSH connection timeout in seconds. Default: 30",
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

// ConfigValidators returns validators for the data source configuration
func (d *VMGuestInfoDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("ssh_key_path"),
			path.MatchRoot("ssh_password"),
		),
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

	strictHostKeyChecking := true
	if !data.SSHStrictHostKeyChecking.IsNull() {
		strictHostKeyChecking = data.SSHStrictHostKeyChecking.ValueBool()
	}

	sshTimeout := int64(30)
	if !data.SSHTimeoutSeconds.IsNull() {
		sshTimeout = data.SSHTimeoutSeconds.ValueInt64()
	}

	vmName := data.VMName.ValueString()
	trueNASHost := data.TrueNASHost.ValueString()

	// Build SSH options
	strictHostKeyOpt := "yes"
	if !strictHostKeyChecking {
		strictHostKeyOpt = "no"
	}

	// Build SSH command based on authentication method
	var sshCmd string
	usePassword := !data.SSHPassword.IsNull() && data.SSHPassword.ValueString() != ""
	
	if usePassword {
		// Check if sshpass is available
		if _, err := exec.LookPath("sshpass"); err != nil {
			resp.Diagnostics.AddError(
				"sshpass Not Found",
				"Password authentication requires 'sshpass' to be installed on the system where Terraform is running.\n\n"+
					"Install sshpass:\n"+
					"  - Debian/Ubuntu: sudo apt-get install sshpass\n"+
					"  - RHEL/CentOS: sudo yum install sshpass\n"+
					"  - macOS: brew install hudochenkov/sshpass/sshpass\n"+
					"  - Alpine: apk add sshpass\n\n"+
					"Alternatively, use ssh_key_path for key-based authentication instead.",
			)
			return
		}

		// Use password authentication with sshpass
		sshPassword := data.SSHPassword.ValueString()
		sshCmd = fmt.Sprintf(
			`sshpass -p '%s' ssh -o StrictHostKeyChecking=%s -o ConnectTimeout=%d %s@%s`,
			sshPassword, strictHostKeyOpt, sshTimeout, sshUser, trueNASHost,
		)
	} else {
		// Use key-based authentication
		sshKeyPath := "~/.ssh/id_rsa"
		if !data.SSHKeyPath.IsNull() {
			sshKeyPath = data.SSHKeyPath.ValueString()
		}
		sshCmd = fmt.Sprintf(
			`ssh -i %s -o StrictHostKeyChecking=%s -o ConnectTimeout=%d %s@%s`,
			sshKeyPath, strictHostKeyOpt, sshTimeout, sshUser, trueNASHost,
		)
	}

	// Query guest agent for network interfaces
	networkCmd := fmt.Sprintf(
		`%s "virsh qemu-agent-command %s '{\"execute\":\"guest-network-get-interfaces\"}'"`,
		sshCmd, vmName,
	)

	output, err := exec.Command("sh", "-c", networkCmd).CombinedOutput()
	if err != nil {
		// Provide better error diagnostics
		outputStr := string(output)
		errorMsg := err.Error()
		
		var diagTitle, diagDetail string
		
		// Analyze the error to provide specific guidance
		if strings.Contains(errorMsg, "connection timed out") || strings.Contains(outputStr, "Connection timed out") {
			diagTitle = "SSH Connection Timeout"
			diagDetail = fmt.Sprintf(
				"SSH connection to %s@%s timed out after %d seconds.\n\n"+
					"Possible causes:\n"+
					"  1. TrueNAS host is not reachable from this machine\n"+
					"  2. SSH service is not running on TrueNAS\n"+
					"  3. Firewall blocking SSH connection\n"+
					"  4. Incorrect hostname or IP address\n\n"+
					"Try increasing ssh_timeout_seconds if the network is slow.",
				sshUser, trueNASHost, sshTimeout,
			)
		} else if strings.Contains(errorMsg, "Permission denied") || strings.Contains(outputStr, "Permission denied") {
			diagTitle = "SSH Authentication Failed"
			if usePassword {
				diagDetail = "SSH password authentication failed. Verify the password is correct."
			} else {
				sshKeyPath := "~/.ssh/id_rsa"
				if !data.SSHKeyPath.IsNull() {
					sshKeyPath = data.SSHKeyPath.ValueString()
				}
				diagDetail = fmt.Sprintf(
					"SSH key authentication failed.\n\n"+
						"Verify:\n"+
						"  1. SSH key exists at: %s\n"+
						"  2. Public key is in authorized_keys on TrueNAS\n"+
						"  3. SSH key has correct permissions (600)\n"+
						"  4. SSH user '%s' is correct",
					sshKeyPath, sshUser,
				)
			}
		} else if strings.Contains(outputStr, "Host key verification failed") {
			diagTitle = "SSH Host Key Verification Failed"
			diagDetail = fmt.Sprintf(
				"SSH host key verification failed for %s.\n\n"+
					"Options:\n"+
					"  1. Add the host key to ~/.ssh/known_hosts manually\n"+
					"  2. Set ssh_strict_host_key_checking = false (less secure)\n"+
					"  3. Remove conflicting key: ssh-keygen -R %s",
				trueNASHost, trueNASHost,
			)
		} else if strings.Contains(outputStr, "domain") && strings.Contains(outputStr, "not found") {
			diagTitle = "VM Not Found"
			diagDetail = fmt.Sprintf(
				"VM '%s' was not found on TrueNAS host.\n\n"+
					"Verify:\n"+
					"  1. VM name matches exactly (case-sensitive)\n"+
					"  2. VM exists on the TrueNAS host\n"+
					"  3. Run 'virsh list --all' on TrueNAS to see available VMs",
				vmName,
			)
		} else if strings.Contains(outputStr, "QEMU guest agent is not connected") || strings.Contains(outputStr, "not connected") {
			diagTitle = "QEMU Guest Agent Not Running"
			diagDetail = fmt.Sprintf(
				"QEMU Guest Agent is not running in VM '%s'.\n\n"+
					"Install and start the guest agent in the VM:\n"+
					"  - Debian/Ubuntu: sudo apt-get install qemu-guest-agent && sudo systemctl start qemu-guest-agent\n"+
					"  - RHEL/CentOS: sudo yum install qemu-guest-agent && sudo systemctl start qemu-guest-agent\n"+
					"  - Windows: Install virtio-win-guest-tools.exe\n\n"+
					"After installation, restart the VM or start the service.",
				vmName,
			)
		} else if strings.Contains(outputStr, "virsh: command not found") {
			diagTitle = "virsh Command Not Found"
			diagDetail = "The 'virsh' command is not available on the TrueNAS host. This is unexpected. Ensure TrueNAS SCALE is properly installed."
		} else {
			diagTitle = "Guest Agent Query Failed"
			diagDetail = fmt.Sprintf(
				"Failed to query QEMU guest agent for VM '%s'.\n\n"+
					"Error: %s\n"+
					"Output: %s\n\n"+
					"Command: %s\n\n"+
					"Troubleshooting:\n"+
					"  1. Verify QEMU Guest Agent is installed and running in the VM\n"+
					"  2. Check SSH connection: ssh %s@%s\n"+
					"  3. Test virsh command: ssh %s@%s 'virsh list --all'\n"+
					"  4. Verify VM name matches exactly",
				vmName, errorMsg, outputStr, networkCmd, sshUser, trueNASHost, sshUser, trueNASHost,
			)
		}
		
		resp.Diagnostics.AddError(diagTitle, diagDetail)
		
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
		`%s "virsh qemu-agent-command %s '{\"execute\":\"guest-get-osinfo\"}'"`,
		sshCmd, vmName,
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
		`%s "virsh qemu-agent-command %s '{\"execute\":\"guest-get-host-name\"}'"`,
		sshCmd, vmName,
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

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

var _ datasource.DataSource = &VMPCIPassthroughDevicesDataSource{}

func NewVMPCIPassthroughDevicesDataSource() datasource.DataSource {
	return &VMPCIPassthroughDevicesDataSource{}
}

type VMPCIPassthroughDevicesDataSource struct {
	client *truenas.Client
}

type VMPCIPassthroughDevicesDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	AvailableOnly   types.Bool   `tfsdk:"available_only"`
	Devices         types.Map    `tfsdk:"devices"`
}

type PCIPassthroughDevice struct {
	PCIAddress       types.String `tfsdk:"pci_address"`
	Description      types.String `tfsdk:"description"`
	ControllerType   types.String `tfsdk:"controller_type"`
	Available        types.Bool   `tfsdk:"available"`
	Critical         types.Bool   `tfsdk:"critical"`
	IOMMUGroup       types.Int64  `tfsdk:"iommu_group"`
	Vendor           types.String `tfsdk:"vendor"`
	Product          types.String `tfsdk:"product"`
}

func (d *VMPCIPassthroughDevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_pci_passthrough_devices"
}

func (d *VMPCIPassthroughDevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches available PCI passthrough devices for VM attachment. Optionally filter to only available devices.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier (always 'vm_pci_passthrough_devices')",
				Computed:            true,
			},
			"available_only": schema.BoolAttribute{
				MarkdownDescription: "If true, only return devices where available=true (default: true)",
				Optional:            true,
			},
			"devices": schema.MapAttribute{
				MarkdownDescription: "Map of PCI device IDs to device information objects",
				Computed:            true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"pci_address":     types.StringType,
						"description":     types.StringType,
						"controller_type": types.StringType,
						"available":       types.BoolType,
						"critical":        types.BoolType,
						"iommu_group":     types.Int64Type,
						"vendor":          types.StringType,
						"product":         types.StringType,
					},
				},
			},
		},
	}
}

func (d *VMPCIPassthroughDevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*truenas.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *truenas.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *VMPCIPassthroughDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMPCIPassthroughDevicesDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default to available_only = true
	availableOnly := true
	if !data.AvailableOnly.IsNull() {
		availableOnly = data.AvailableOnly.ValueBool()
	}

	endpoint := "/vm/device/passthrough_device_choices"
	respBody, err := d.client.Get(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read PCI passthrough devices, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse PCI passthrough devices response: %s", err))
		return
	}

	// Build the devices map
	devices := make(map[string]attr.Value)
	
	for deviceID, deviceData := range result {
		deviceMap, ok := deviceData.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if device is available
		available := false
		if avail, ok := deviceMap["available"].(bool); ok {
			available = avail
		}

		// Skip if filtering for available only and device is not available
		if availableOnly && !available {
			continue
		}

		// Extract capability info
		var vendor, product, pciAddress string
		if capability, ok := deviceMap["capability"].(map[string]interface{}); ok {
			if v, ok := capability["vendor"].(string); ok {
				vendor = v
			}
			if p, ok := capability["product"].(string); ok {
				product = p
			}
			// Build PCI address from domain:bus:slot.function
			domain := "0000"
			bus := "00"
			slot := "00"
			function := "0"
			if d, ok := capability["domain"].(string); ok {
				domain = d
			}
			if b, ok := capability["bus"].(string); ok {
				bus = b
			}
			if s, ok := capability["slot"].(string); ok {
				slot = s
			}
			if f, ok := capability["function"].(string); ok {
				function = f
			}
			pciAddress = fmt.Sprintf("%s:%s:%s.%s", domain, bus, slot, function)
		}

		// Extract other fields
		description := ""
		if desc, ok := deviceMap["description"].(string); ok {
			description = desc
		}

		controllerType := ""
		if ct, ok := deviceMap["controller_type"].(string); ok {
			controllerType = ct
		}

		critical := false
		if crit, ok := deviceMap["critical"].(bool); ok {
			critical = crit
		}

		iommuGroup := int64(0)
		if iommu, ok := deviceMap["iommu_group"].(map[string]interface{}); ok {
			if num, ok := iommu["number"].(float64); ok {
				iommuGroup = int64(num)
			}
		}

		// Create device object
		deviceObj := map[string]attr.Value{
			"pci_address":     types.StringValue(pciAddress),
			"description":     types.StringValue(description),
			"controller_type": types.StringValue(controllerType),
			"available":       types.BoolValue(available),
			"critical":        types.BoolValue(critical),
			"iommu_group":     types.Int64Value(iommuGroup),
			"vendor":          types.StringValue(vendor),
			"product":         types.StringValue(product),
		}

		objValue, diag := types.ObjectValue(
			map[string]attr.Type{
				"pci_address":     types.StringType,
				"description":     types.StringType,
				"controller_type": types.StringType,
				"available":       types.BoolType,
				"critical":        types.BoolType,
				"iommu_group":     types.Int64Type,
				"vendor":          types.StringType,
				"product":         types.StringType,
			},
			deviceObj,
		)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		devices[deviceID] = objValue
	}

	devicesMap, diag := types.MapValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"pci_address":     types.StringType,
				"description":     types.StringType,
				"controller_type": types.StringType,
				"available":       types.BoolType,
				"critical":        types.BoolType,
				"iommu_group":     types.Int64Type,
				"vendor":          types.StringType,
				"product":         types.StringType,
			},
		},
		devices,
	)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("vm_pci_passthrough_devices")
	data.AvailableOnly = types.BoolValue(availableOnly)
	data.Devices = devicesMap

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}


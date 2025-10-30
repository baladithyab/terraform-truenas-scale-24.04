# Contributing to TrueNAS Terraform Provider

Thank you for your interest in contributing! This guide will help you add new resources to the provider.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your feature
4. Make your changes
5. Test thoroughly
6. Submit a pull request

## Development Environment

### Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- TrueNAS Scale 24.04 server for testing
- TrueNAS API key

### Setup

```bash
git clone https://github.com/YOUR_USERNAME/terraform-provider-truenas.git
cd terraform-provider-truenas
go mod download
```

## Adding a New Resource

### Step 1: Research the API

First, understand the TrueNAS API endpoint you want to implement:

```bash
# Download the OpenAPI spec
curl http://10.0.0.83:81/api/v2.0 > openapi.json

# Find your endpoint
cat openapi.json | jq '.paths | keys[] | select(contains("your_endpoint"))'

# View endpoint details
cat openapi.json | jq '.paths["/your/endpoint"]'

# View schema
cat openapi.json | jq '.components.schemas.your_schema'
```

Or browse the interactive docs at: `http://your-truenas-ip/api/docs/`

### Step 2: Create the Resource File

Create a new file in `internal/provider/` named `resource_<name>.go`:

```go
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &YourResource{}
var _ resource.ResourceWithImportState = &YourResource{}

func NewYourResource() resource.Resource {
	return &YourResource{}
}

type YourResource struct {
	client *truenas.Client
}

type YourResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// Add other fields based on API schema
}

func (r *YourResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_your_resource"
}

func (r *YourResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a <resource> on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name",
				Required:            true,
			},
			// Add other attributes
		},
	}
}

func (r *YourResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *YourResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data YourResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request
	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
	}

	// Make API call
	respBody, err := r.client.Post("/your/endpoint", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create resource, got error: %s", err))
		return
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Set ID
	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(strconv.Itoa(int(id)))
	}

	// Read back to get computed values
	r.readResource(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *YourResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data YourResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readResource(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *YourResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data YourResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := map[string]interface{}{}
	// Add fields to update

	endpoint := fmt.Sprintf("/your/endpoint/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update resource, got error: %s", err))
		return
	}

	r.readResource(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *YourResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data YourResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/your/endpoint/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete resource, got error: %s", err))
		return
	}
}

func (r *YourResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *YourResource) readResource(ctx context.Context, data *YourResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/your/endpoint/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read resource, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Map API response to model
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	// Map other fields
}
```

### Step 3: Register the Resource

Add your resource to `internal/provider/provider.go`:

```go
func (p *TruenasProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// ... existing resources ...
		NewYourResource,
	}
}
```

### Step 4: Create Examples

Create `examples/resources/truenas_your_resource/resource.tf`:

```hcl
# Basic example
resource "truenas_your_resource" "example" {
  name = "example"
  # Add other required fields
}

# Advanced example
resource "truenas_your_resource" "advanced" {
  name = "advanced"
  # Add optional fields
}

# Import example
# terraform import truenas_your_resource.existing <id>
```

### Step 5: Update Documentation

1. Update `README.md` - Add to the resources list
2. Update `API_COVERAGE.md` - Mark as implemented
3. Add inline documentation in the schema

### Step 6: Build and Test

```bash
# Build
go build -o terraform-provider-truenas

# Install locally
make install

# Test
cd test-directory
terraform init
terraform plan
terraform apply
terraform destroy
```

### Step 7: Test Import

```bash
# Create resource manually in TrueNAS
# Then import it
terraform import truenas_your_resource.test <id>
terraform plan  # Should show no changes
```

## Code Style Guidelines

- Follow Go conventions
- Use meaningful variable names
- Add comments for complex logic
- Handle errors appropriately
- Use types from terraform-plugin-framework
- Keep functions focused and small

## Testing Checklist

Before submitting a PR, ensure:

- [ ] Resource creates successfully
- [ ] Resource reads correctly
- [ ] Resource updates work
- [ ] Resource deletes cleanly
- [ ] Import functionality works
- [ ] All attributes are properly mapped
- [ ] Computed values are handled
- [ ] Required vs optional attributes are correct
- [ ] Error messages are helpful
- [ ] Code compiles without warnings
- [ ] Examples are provided
- [ ] Documentation is updated

## Common Patterns

### Handling Nested Objects

```go
if nested, ok := result["nested"].(map[string]interface{}); ok {
	if value, ok := nested["field"].(string); ok {
		data.Field = types.StringValue(value)
	}
}
```

### Handling Lists

```go
if items, ok := result["items"].([]interface{}); ok {
	var list []string
	for _, item := range items {
		if str, ok := item.(string); ok {
			list = append(list, str)
		}
	}
	// Convert to types.List
}
```

### Handling Optional Fields

```go
if !data.OptionalField.IsNull() {
	createReq["optional_field"] = data.OptionalField.ValueString()
}
```

## Submitting a Pull Request

1. Ensure all tests pass
2. Update CHANGELOG.md
3. Write a clear PR description
4. Reference any related issues
5. Be responsive to feedback

## Questions?

- Open an issue on GitHub
- Check existing resources for examples
- Review the Terraform Plugin Framework docs

## Resources

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [TrueNAS API Docs](http://your-truenas-ip/api/docs/)
- [Go Documentation](https://go.dev/doc/)


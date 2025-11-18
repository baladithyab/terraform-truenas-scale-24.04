# Cloud-Init Support Design for TrueNAS Scale Terraform Provider

## Overview
This document outlines the design for adding Cloud-Init support to the `truenas_vm` resource. TrueNAS Scale does not natively support Cloud-Init configuration fields in its VM API. The standard workaround is to generate a "NoCloud" datasource ISO containing the configuration files (`user-data`, `meta-data`, `network-config`), upload it to the TrueNAS filesystem, and attach it as a CD-ROM device to the VM.

## 1. Schema Changes (`internal/provider/resource_vm.go`)

We will modify the `truenas_vm` resource schema to include a new `cloud_init` block.

### New Attributes

```go
"cloud_init": schema.SingleNestedAttribute{
    MarkdownDescription: "Cloud-Init configuration. If specified, an ISO image will be generated and attached as a CD-ROM.",
    Optional:            true,
    Attributes: map[string]schema.Attribute{
        "user_data": schema.StringAttribute{
            MarkdownDescription: "Cloud-init user-data configuration (YAML format).",
            Optional:            true,
        },
        "meta_data": schema.StringAttribute{
            MarkdownDescription: "Cloud-init meta-data configuration (YAML format).",
            Optional:            true,
        },
        "network_config": schema.StringAttribute{
            MarkdownDescription: "Cloud-init network-config configuration (YAML format).",
            Optional:            true,
        },
        "image_storage_path": schema.StringAttribute{
            MarkdownDescription: "Directory path on TrueNAS where the generated Cloud-Init ISO will be stored (e.g., /mnt/pool/isos).",
            Required:            true,
        },
        "filename": schema.StringAttribute{
            MarkdownDescription: "Optional custom filename for the ISO. If not provided, one will be generated based on the VM name.",
            Optional:            true,
            Computed:            true,
        },
    },
}
```

## 2. ISO Generation

We will use a pure Go library to generate the ISO 9660 image in-memory to avoid external dependencies like `genisoimage` or `mkisofs`.

**Library Recommendation:** `github.com/kdomanski/iso9660` (or `github.com/hooklift/iso9660`)
*   **Reason:** It provides a simple API for creating ISO images and writing files to them.
*   **Implementation:**
    *   Create a helper function `GenerateCloudInitISO(userData, metaData, networkConfig string) ([]byte, error)`.
    *   The ISO will contain:
        *   `user-data` (if provided)
        *   `meta-data` (if provided, otherwise default to `{instance-id: <vm-id>, local-hostname: <vm-name>}`)
        *   `network-config` (if provided)
    *   The volume label should be `cidata` (standard for NoCloud datasource).

## 3. TrueNAS API Client Updates (`internal/truenas/client.go`)

We need to add methods to handle file operations on the TrueNAS filesystem.

### New Methods

1.  **`UploadFile(path string, content []byte) error`**
    *   **Endpoint:** `/filesystem/put`
    *   **Method:** POST
    *   **Payload:** Multipart form data.
    *   **Parameters:**
        *   `path`: Full path to the destination file (e.g., `/mnt/pool/isos/vm-1-cloud-init.iso`).
        *   `file`: The file content.

2.  **`DeleteFile(path string) error`**
    *   **Endpoint:** `/filesystem/delete` (Need to verify exact endpoint, likely `filesystem/delete` taking a path).
    *   **Method:** POST (usually)
    *   **Payload:** `{"path": "/path/to/file"}`

## 4. Resource Lifecycle Management

### Create
1.  **Validate:** Check if `cloud_init` block is present.
2.  **Generate:** Call `GenerateCloudInitISO` with the provided config.
3.  **Determine Path:**
    *   If `filename` is unset, generate: `{vm_name}-cloud-init.iso`.
    *   Full path: `{image_storage_path}/{filename}`.
4.  **Upload:** Call `client.UploadFile` to send the ISO to TrueNAS.
5.  **Attach:**
    *   Create a new `CDROMDeviceModel` entry.
    *   Set `path` to the uploaded ISO path.
    *   Append this to the `cdrom_devices` list before creating the VM devices.
    *   *Note:* Ensure this CD-ROM is ordered correctly (usually first or explicitly ordered) if boot order matters, though Cloud-Init usually just scans all attached block devices.

### Read
*   The provider cannot easily read the *content* of the remote ISO to verify it matches the Terraform config.
*   We will rely on Terraform state. If the `user_data`, `meta_data`, or `network_config` in the plan differs from the state, it triggers an update.

### Update
1.  **Check Changes:** If any `cloud_init` fields change:
2.  **Re-generate:** Generate new ISO content.
3.  **Re-upload:** Overwrite the existing file at the path.
4.  **VM Restart:** Changing the CD-ROM content might not be picked up by a running VM until reboot. The provider generally doesn't force restart unless necessary, but we should document this behavior.

### Delete
1.  **Cleanup:** After deleting the VM and its devices, call `client.DeleteFile` to remove the generated ISO.

## 5. Implementation Steps

1.  **Dependency:** Add `github.com/kdomanski/iso9660` to `go.mod`.
2.  **Client:** Implement `UploadFile` and `DeleteFile` in `internal/truenas/client.go`.
3.  **Helper:** Create `internal/provider/iso_gen.go` for ISO generation logic.
4.  **Resource:**
    *   Update `VMResourceModel` struct.
    *   Update `Schema` method.
    *   Update `Create` method to handle generation and upload.
    *   Update `Update` method to handle regeneration.
    *   Update `Delete` method to handle cleanup.

## 6. Questions / Risks
*   **Permissions:** Does the API user have permission to write to the specified `image_storage_path`? (User responsibility).
*   **Endpoint Verification:** Need to confirm `/filesystem/put` behavior (multipart field names).
*   **Boot Order:** Does the Cloud-Init ISO need to be the *first* CD-ROM? Usually not, but good to keep in mind.
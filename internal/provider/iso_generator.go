package provider

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kdomanski/iso9660"
)

// GenerateCloudInitISO creates an ISO image containing user-data and meta-data files
// for Cloud-Init NoCloud datasource.
func GenerateCloudInitISO(userData, metaData string) ([]byte, error) {
	writer, err := iso9660.NewWriter()
	if err != nil {
		return nil, fmt.Errorf("failed to create ISO writer: %w", err)
	}
	defer writer.Cleanup()

	// Add user-data
	if err := writer.AddFile(strings.NewReader(userData), "user-data"); err != nil {
		return nil, fmt.Errorf("failed to add user-data to ISO: %w", err)
	}

	// Add meta-data
	if err := writer.AddFile(strings.NewReader(metaData), "meta-data"); err != nil {
		return nil, fmt.Errorf("failed to add meta-data to ISO: %w", err)
	}

	// Write to buffer with volume label "cidata"
	output := &bytes.Buffer{}
	if err := writer.WriteTo(output, "cidata"); err != nil {
		return nil, fmt.Errorf("failed to write ISO to buffer: %w", err)
	}

	return output.Bytes(), nil
}
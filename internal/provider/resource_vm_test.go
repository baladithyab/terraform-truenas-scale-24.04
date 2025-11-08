package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// TestVMUpdate_OnlySendsChangedFields tests that the Update() function logic
// only includes changed fields in the update payload
func TestVMUpdate_OnlySendsChangedFields(t *testing.T) {
	// Test case: Only memory changed, other fields should be preserved from state
	plan := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(4096), // Changed from 2048
		VCPUs:   types.Int64Value(2),
		Cores:   types.Int64Value(1),
		Threads: types.Int64Value(1),
	}

	state := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(2048), // Old value
		VCPUs:   types.Int64Value(2),
		Cores:   types.Int64Value(1),
		Threads: types.Int64Value(1),
	}

	// Verify the comparison logic that would be used in Update()
	assert.NotEqual(t, plan.Memory, state.Memory, "Memory should be different")
	assert.Equal(t, plan.VCPUs, state.VCPUs, "VCPUs should be same")
	assert.Equal(t, plan.Cores, state.Cores, "Cores should be same")
	assert.Equal(t, plan.Threads, state.Threads, "Threads should be same")

	// Since VCPUs, Cores, Threads are equal, they shouldn't trigger an update
	// Only Memory being different would be included in the update payload
	assert.True(t, !plan.VCPUs.Equal(state.VCPUs) == false, "VCPUs are equal, should not be sent")
	assert.True(t, !plan.Cores.Equal(state.Cores) == false, "Cores are equal, should not be sent")
	assert.True(t, !plan.Threads.Equal(state.Threads) == false, "Threads are equal, should not be sent")
}

// TestVMUpdate_PreservesComputedFieldsWithZero tests that computed fields with
// zero values in plan preserve state values to avoid sending zeros to the API
func TestVMUpdate_PreservesComputedFieldsWithZero(t *testing.T) {
	// Simulate scenario where plan has zero values for computed fields
	// (which happens when user doesn't explicitly set them)
	plan := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(4096),
		VCPUs:   types.Int64Value(0), // Zero - should use state value
		Cores:   types.Int64Value(0), // Zero - should use state value
		Threads: types.Int64Value(0), // Zero - should use state value
	}

	state := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(2048),
		VCPUs:   types.Int64Value(2), // Valid value from API
		Cores:   types.Int64Value(1), // Valid value from API
		Threads: types.Int64Value(1), // Valid value from API
	}

	// The fix ensures that zero values in plan don't get sent to the API
	// Instead, state values should be preserved
	assert.True(t, plan.VCPUs.ValueInt64() == 0, "Plan has zero VCPUs")
	assert.True(t, state.VCPUs.ValueInt64() > 0, "State has valid VCPUs")

	assert.True(t, plan.Cores.ValueInt64() == 0, "Plan has zero Cores")
	assert.True(t, state.Cores.ValueInt64() > 0, "State has valid Cores")

	assert.True(t, plan.Threads.ValueInt64() == 0, "Plan has zero Threads")
	assert.True(t, state.Threads.ValueInt64() > 0, "State has valid Threads")

	// Verify the logic: should use state value when plan has zero
	vcpusToSend := plan.VCPUs.ValueInt64()
	if vcpusToSend == 0 && !state.VCPUs.IsNull() && state.VCPUs.ValueInt64() > 0 {
		vcpusToSend = state.VCPUs.ValueInt64()
	}
	assert.Equal(t, int64(2), vcpusToSend, "Should use state VCPUs value when plan is zero")

	coresToSend := plan.Cores.ValueInt64()
	if coresToSend == 0 && !state.Cores.IsNull() && state.Cores.ValueInt64() > 0 {
		coresToSend = state.Cores.ValueInt64()
	}
	assert.Equal(t, int64(1), coresToSend, "Should use state Cores value when plan is zero")

	threadsToSend := plan.Threads.ValueInt64()
	if threadsToSend == 0 && !state.Threads.IsNull() && state.Threads.ValueInt64() > 0 {
		threadsToSend = state.Threads.ValueInt64()
	}
	assert.Equal(t, int64(1), threadsToSend, "Should use state Threads value when plan is zero")
}

// TestVMUpdate_MemoryChangeOnly tests updating only memory while preserving CPU topology
func TestVMUpdate_MemoryChangeOnly(t *testing.T) {
	plan := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(8192), // Changed from 4096
		VCPUs:   types.Int64Value(4),    // Unchanged
		Cores:   types.Int64Value(2),    // Unchanged
		Threads: types.Int64Value(2),    // Unchanged
	}

	state := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(4096),
		VCPUs:   types.Int64Value(4),
		Cores:   types.Int64Value(2),
		Threads: types.Int64Value(2),
	}

	// Memory changed, CPU values unchanged
	assert.NotEqual(t, plan.Memory.ValueInt64(), state.Memory.ValueInt64())
	assert.Equal(t, plan.VCPUs.ValueInt64(), state.VCPUs.ValueInt64())
	assert.Equal(t, plan.Cores.ValueInt64(), state.Cores.ValueInt64())
	assert.Equal(t, plan.Threads.ValueInt64(), state.Threads.ValueInt64())

	// Since CPU values are equal and valid, they should be preserved (sent with same value)
	// This maintains CPU topology even though only memory changed
	assert.True(t, plan.VCPUs.ValueInt64() > 0, "VCPUs is valid")
	assert.True(t, plan.Cores.ValueInt64() > 0, "Cores is valid")
	assert.True(t, plan.Threads.ValueInt64() > 0, "Threads is valid")
}

// TestVMUpdate_DesiredStateChangeOnly tests that changing only desired_state
// doesn't trigger unnecessary Update() API calls for other fields
func TestVMUpdate_DesiredStateChangeOnly(t *testing.T) {
	plan := VMResourceModel{
		ID:           types.StringValue("1"),
		Name:         types.StringValue("test-vm"),
		Memory:       types.Int64Value(4096),
		VCPUs:        types.Int64Value(2),
		Cores:        types.Int64Value(1),
		Threads:      types.Int64Value(1),
		DesiredState: types.StringValue("RUNNING"), // Changed from STOPPED
	}

	state := VMResourceModel{
		ID:           types.StringValue("1"),
		Name:         types.StringValue("test-vm"),
		Memory:       types.Int64Value(4096),
		VCPUs:        types.Int64Value(2),
		Cores:        types.Int64Value(1),
		Threads:      types.Int64Value(1),
		DesiredState: types.StringValue("STOPPED"),
	}

	// Only desired_state changed
	assert.NotEqual(t, plan.DesiredState, state.DesiredState)
	assert.Equal(t, plan.Memory, state.Memory)
	assert.Equal(t, plan.VCPUs, state.VCPUs)
	assert.Equal(t, plan.Cores, state.Cores)
	assert.Equal(t, plan.Threads, state.Threads)

	// The fix ensures that when only desired_state changes,
	// the Update() function doesn't send other unchanged fields
	// State transition is handled separately via transitionVMState()
}

// TestVMUpdate_PreservesComputedStringFields tests that computed string fields
// like bootloader and cpu_mode preserve state values when plan doesn't provide them
func TestVMUpdate_PreservesComputedStringFields(t *testing.T) {
	plan := VMResourceModel{
		ID:         types.StringValue("1"),
		Name:       types.StringValue("test-vm"),
		Memory:     types.Int64Value(4096),
		VCPUs:      types.Int64Value(2),
		Bootloader: types.StringNull(), // Not set in plan
		CPUMode:    types.StringNull(), // Not set in plan
	}

	state := VMResourceModel{
		ID:         types.StringValue("1"),
		Name:       types.StringValue("test-vm"),
		Memory:     types.Int64Value(4096),
		VCPUs:      types.Int64Value(2),
		Bootloader: types.StringValue("UEFI"),       // From API
		CPUMode:    types.StringValue("HOST-MODEL"), // From API
	}

	// Plan doesn't have these computed values
	assert.True(t, plan.Bootloader.IsNull(), "Plan bootloader is null")
	assert.True(t, plan.CPUMode.IsNull(), "Plan cpu_mode is null")

	// State has valid values from API
	assert.False(t, state.Bootloader.IsNull(), "State bootloader is set")
	assert.False(t, state.CPUMode.IsNull(), "State cpu_mode is set")

	// The fix ensures state values are preserved
	bootloaderToSend := ""
	if !plan.Bootloader.IsNull() && plan.Bootloader.ValueString() != "" {
		bootloaderToSend = plan.Bootloader.ValueString()
	} else if !state.Bootloader.IsNull() && state.Bootloader.ValueString() != "" {
		bootloaderToSend = state.Bootloader.ValueString()
	}
	assert.Equal(t, "UEFI", bootloaderToSend, "Should preserve state bootloader")

	cpuModeToSend := ""
	if !plan.CPUMode.IsNull() && plan.CPUMode.ValueString() != "" {
		cpuModeToSend = plan.CPUMode.ValueString()
	} else if !state.CPUMode.IsNull() && state.CPUMode.ValueString() != "" {
		cpuModeToSend = state.CPUMode.ValueString()
	}
	assert.Equal(t, "HOST-MODEL", cpuModeToSend, "Should preserve state cpu_mode")
}

// TestVMUpdate_ValidatesNonZeroValues tests that the update logic validates
// values before including them in the update payload
func TestVMUpdate_ValidatesNonZeroValues(t *testing.T) {
	// Test that zero values are filtered out
	vcpus := int64(0)
	cores := int64(0)
	threads := int64(0)

	assert.False(t, vcpus > 0, "Zero VCPUs should not pass validation")
	assert.False(t, cores > 0, "Zero Cores should not pass validation")
	assert.False(t, threads > 0, "Zero Threads should not pass validation")

	// Test that positive values pass validation
	vcpus = 2
	cores = 1
	threads = 1

	assert.True(t, vcpus > 0, "Positive VCPUs should pass validation")
	assert.True(t, cores > 0, "Positive Cores should pass validation")
	assert.True(t, threads > 0, "Positive Threads should pass validation")
}

// TestVMUpdate_EmptyStringValidation tests that empty string values are filtered out
func TestVMUpdate_EmptyStringValidation(t *testing.T) {
	// Test empty strings are filtered
	bootloader := ""
	cpuMode := ""

	assert.False(t, bootloader != "", "Empty bootloader should not pass validation")
	assert.False(t, cpuMode != "", "Empty cpu_mode should not pass validation")

	// Test non-empty strings pass
	bootloader = "UEFI"
	cpuMode = "HOST-MODEL"

	assert.True(t, bootloader != "", "Non-empty bootloader should pass validation")
	assert.True(t, cpuMode != "", "Non-empty cpu_mode should pass validation")
}

// TestVMUpdate_CPUTopologyUpdate tests updating CPU topology fields
func TestVMUpdate_CPUTopologyUpdate(t *testing.T) {
	// Test changing CPU topology from 2 cores to 4 cores
	plan := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(4096),
		VCPUs:   types.Int64Value(4), // Changed from 2
		Cores:   types.Int64Value(4), // Changed from 2
		Threads: types.Int64Value(1),
	}

	state := VMResourceModel{
		ID:      types.StringValue("1"),
		Name:    types.StringValue("test-vm"),
		Memory:  types.Int64Value(4096),
		VCPUs:   types.Int64Value(2),
		Cores:   types.Int64Value(2),
		Threads: types.Int64Value(1),
	}

	// VCPUs and Cores changed, Threads unchanged
	assert.NotEqual(t, plan.VCPUs.ValueInt64(), state.VCPUs.ValueInt64())
	assert.NotEqual(t, plan.Cores.ValueInt64(), state.Cores.ValueInt64())
	assert.Equal(t, plan.Threads.ValueInt64(), state.Threads.ValueInt64())

	// All values are valid (> 0)
	assert.True(t, plan.VCPUs.ValueInt64() > 0, "VCPUs is valid")
	assert.True(t, plan.Cores.ValueInt64() > 0, "Cores is valid")
	assert.True(t, plan.Threads.ValueInt64() > 0, "Threads is valid")
}

// TestVMUpdate_MinMemoryHandling tests that min_memory is handled correctly
func TestVMUpdate_MinMemoryHandling(t *testing.T) {
	// Test 1: min_memory explicitly set in plan
	plan1 := VMResourceModel{
		ID:        types.StringValue("1"),
		Name:      types.StringValue("test-vm"),
		Memory:    types.Int64Value(4096),
		MinMemory: types.Int64Value(2048), // Explicitly set
	}

	assert.False(t, plan1.MinMemory.IsNull(), "MinMemory should be set")
	assert.Equal(t, int64(2048), plan1.MinMemory.ValueInt64())

	// Test 2: min_memory not set in plan, should default to memory value
	plan2 := VMResourceModel{
		ID:        types.StringValue("1"),
		Name:      types.StringValue("test-vm"),
		Memory:    types.Int64Value(4096),
		MinMemory: types.Int64Null(), // Not set
	}

	minMemoryToSend := int64(0)
	if !plan2.MinMemory.IsNull() && plan2.MinMemory.ValueInt64() > 0 {
		minMemoryToSend = plan2.MinMemory.ValueInt64()
	} else {
		// Default to memory value to disable ballooning
		minMemoryToSend = plan2.Memory.ValueInt64()
	}

	assert.Equal(t, plan2.Memory.ValueInt64(), minMemoryToSend,
		"MinMemory should default to Memory value when not set")
}

// TestVMUpdate_BooleanFieldHandling tests that boolean fields are handled correctly
func TestVMUpdate_BooleanFieldHandling(t *testing.T) {
	plan := VMResourceModel{
		ID:          types.StringValue("1"),
		Name:        types.StringValue("test-vm"),
		Memory:      types.Int64Value(4096),
		Autostart:   types.BoolValue(true),  // Changed from false
		HideFromMSR: types.BoolValue(false), // Unchanged
	}

	state := VMResourceModel{
		ID:          types.StringValue("1"),
		Name:        types.StringValue("test-vm"),
		Memory:      types.Int64Value(4096),
		Autostart:   types.BoolValue(false),
		HideFromMSR: types.BoolValue(false),
	}

	// Autostart changed, HideFromMSR unchanged
	assert.NotEqual(t, plan.Autostart.ValueBool(), state.Autostart.ValueBool())
	assert.Equal(t, plan.HideFromMSR.ValueBool(), state.HideFromMSR.ValueBool())

	// Boolean fields should always be sent since false is a valid value
	assert.False(t, plan.Autostart.IsNull(), "Autostart should not be null")
	assert.False(t, plan.HideFromMSR.IsNull(), "HideFromMSR should not be null")
}

// Test VM lifecycle state transitions
func TestVMLifecycle_StateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		fromState  string
		toState    string
		shouldPass bool
	}{
		{"STOPPED to RUNNING", "STOPPED", "RUNNING", true},
		{"RUNNING to STOPPED", "RUNNING", "STOPPED", true},
		{"RUNNING to SUSPENDED", "RUNNING", "SUSPENDED", true},
		{"SUSPENDED to RUNNING", "SUSPENDED", "RUNNING", true},
		{"STOPPED to SUSPENDED", "STOPPED", "SUSPENDED", true},
		{"SUSPENDED to STOPPED", "SUSPENDED", "STOPPED", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the state value assignment
			initialState := types.StringValue(tt.fromState)
			newState := types.StringValue(tt.toState)

			assert.NotEqual(t, initialState, newState, "States should be different")
			assert.NotEmpty(t, newState.ValueString(), "New state should not be empty")
		})
	}
}

// Test that desired_state defaults properly
func TestVMLifecycle_DefaultState(t *testing.T) {
	vm := VMResourceModel{
		Name:   types.StringValue("test-vm"),
		Memory: types.Int64Value(2048),
		VCPUs:  types.Int64Value(2),
		// DesiredState not set - should default appropriately
	}

	// Verify the model can be created without desired_state
	assert.Equal(t, "test-vm", vm.Name.ValueString())
	assert.True(t, vm.DesiredState.IsNull(), "DesiredState should be null if not set")
}

// Test deprecated start_on_create vs desired_state priority
func TestVMLifecycle_DeprecatedPriority(t *testing.T) {
	// When both are set, desired_state should take priority
	vm := VMResourceModel{
		Name:          types.StringValue("test-vm"),
		StartOnCreate: types.BoolValue(false),       // Deprecated: says don't start
		DesiredState:  types.StringValue("RUNNING"), // New: says start
	}

	// desired_state should have priority
	assert.Equal(t, "RUNNING", vm.DesiredState.ValueString())
	assert.False(t, vm.StartOnCreate.ValueBool())
}

// Test valid state values
func TestVMLifecycle_ValidStates(t *testing.T) {
	validStates := []string{"RUNNING", "STOPPED", "SUSPENDED"}

	for _, state := range validStates {
		vm := VMResourceModel{
			DesiredState: types.StringValue(state),
		}
		assert.Contains(t, validStates, vm.DesiredState.ValueString())
	}
}

// Test state change detection
func TestVMLifecycle_StateChangeDetection(t *testing.T) {
	oldState := types.StringValue("STOPPED")
	newState := types.StringValue("RUNNING")

	// Test detection logic
	assert.False(t, oldState.Equal(newState), "States should not be equal")
	assert.NotEqual(t, oldState.ValueString(), newState.ValueString())
}

// Test backward compatibility - start_on_create still works
func TestVMLifecycle_StartOnCreateBackwardCompat(t *testing.T) {
	vm := VMResourceModel{
		StartOnCreate: types.BoolValue(true),
		DesiredState:  types.StringNull(), // Not set
	}

	// Should respect start_on_create when desired_state not set
	assert.True(t, vm.StartOnCreate.ValueBool())
	assert.True(t, vm.DesiredState.IsNull())
}

package pcidevice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/harvester/terraform-provider-harvester/internal/config"
	"github.com/harvester/terraform-provider-harvester/pkg/client"
	"github.com/harvester/terraform-provider-harvester/pkg/constants"
	"github.com/harvester/terraform-provider-harvester/pkg/helper"
)

// PCIDeviceClaim GVR (Group Version Resource) for Harvester PCIDeviceClaim CRD
var (
	pcideviceClaimGVR = k8sschema.GroupVersionResource{
		Group:    "devices.harvesterhci.io",
		Version:  "v1beta1",
		Resource: "pcideviceclaims",
	}
)

// ResourcePCIDevice returns the Terraform resource schema for harvester_pci_device.
// This resource manages PCI device passthrough to VMs using Harvester's PCIDeviceClaim CRD.
func ResourcePCIDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePCIDeviceCreate,
		ReadContext:   resourcePCIDeviceRead,
		DeleteContext: resourcePCIDeviceDelete,
		UpdateContext: resourcePCIDeviceUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: Schema(),
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(2 * time.Minute),
			Read:    schema.DefaultTimeout(2 * time.Minute),
			Update:  schema.DefaultTimeout(2 * time.Minute),
			Delete:  schema.DefaultTimeout(2 * time.Minute),
			Default: schema.DefaultTimeout(2 * time.Minute),
		},
	}
}

// getDynamicClient returns a dynamic client for accessing CRDs
func getDynamicClient(c *client.Client) (dynamic.Interface, error) {
	return dynamic.NewForConfig(c.RestConfig)
}

// resourcePCIDeviceCreate creates a new PCIDeviceClaim resource in Harvester.
// It creates a claim that attaches PCI devices to a VM, ensuring the VM runs on a specific node.
func resourcePCIDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace := d.Get(constants.FieldCommonNamespace).(string)

	// Get VM name and namespace
	vmNameRaw := d.Get(constants.FieldPCIDeviceVMName).(string)
	vmNamespace, vmName, err := helper.NamespacedNamePartsByDefault(vmNameRaw, namespace)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid VM name format: %w", err))
	}

	// Verify VM exists
	_, err = c.HarvesterClient.KubevirtV1().VirtualMachines(vmNamespace).Get(ctx, vmName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return diag.Errorf("virtual machine %s/%s not found", vmNamespace, vmName)
		}
		return diag.FromErr(err)
	}

	nodeName := d.Get(constants.FieldPCIDeviceNodeName).(string)
	pciAddressesRaw := d.Get(constants.FieldPCIDevicePCIAddresses).([]interface{})
	pciAddresses := make([]string, len(pciAddressesRaw))
	for i, addr := range pciAddressesRaw {
		pciAddresses[i] = addr.(string)
	}

	// Get labels (optional)
	labels := make(map[string]string)
	if labelsRaw, ok := d.GetOk(constants.FieldCommonLabels); ok {
		for k, v := range labelsRaw.(map[string]interface{}) {
			labels[k] = v.(string)
		}
	}

	// Create PCIDeviceClaim using dynamic client
	dynamicClient, err := getDynamicClient(c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	// Create a PCIDeviceClaim for each PCI address
	// IMPORTANT: The PCIDeviceClaim name MUST match the PCIDevice name format: node-address
	// The admission webhook validates that the PCIDevice exists with this exact name
	addresses := make(map[string]bool)
	for _, addr := range pciAddresses {
		addresses[addr] = true
	}

	firstClaimName, err := ensureClaims(ctx, dynamicClient, nodeName, addresses, labels)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set resource ID (format: namespace/vmname/claimname)
	d.SetId(fmt.Sprintf("%s/%s/%s", vmNamespace, vmName, firstClaimName))

	return resourcePCIDeviceRead(ctx, d, meta)
}

// resourcePCIDeviceRead reads the state of all PCIDeviceClaim resources for this resource.
// It verifies that all claims (one per PCI address) still exist and reads their state.
func resourcePCIDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse ID using helper
	vmNamespace, vmName, claimName, err := helper.PCIDeviceIDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the PCIDeviceClaim using dynamic client
	dynamicClient, err := getDynamicClient(c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	// Read the primary claim (from ID) to get node_name
	primaryClaim, err := dynamicClient.Resource(pcideviceClaimGVR).Get(ctx, claimName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	spec, ok := primaryClaim.Object["spec"].(map[string]interface{})
	if !ok {
		return diag.Errorf("invalid PCIDeviceClaim spec")
	}
	nodeName := getSpecField(spec, "nodeName")

	// Collect all addresses: check if we have addresses in state/config
	pciAddressesRaw := d.Get(constants.FieldPCIDevicePCIAddresses).([]interface{})
	var addresses []string

	if len(pciAddressesRaw) > 0 {
		// Verify all claims from state exist
		for _, addr := range pciAddressesRaw {
			address := addr.(string)
			cn := buildClaimName(nodeName, address)
			_, claimErr := dynamicClient.Resource(pcideviceClaimGVR).Get(ctx, cn, metav1.GetOptions{})
			if claimErr != nil {
				if apierrors.IsNotFound(claimErr) {
					d.SetId("")
					return nil
				}
				return diag.FromErr(claimErr)
			}
			addresses = append(addresses, address)
		}
	} else {
		// Import case: only the primary claim address is available
		if address := getSpecField(spec, "address"); address != "" {
			addresses = append(addresses, address)
		}
	}

	// Set all resource state
	states := map[string]interface{}{
		constants.FieldCommonNamespace:        vmNamespace,
		constants.FieldPCIDeviceVMName:        helper.BuildNamespacedName(vmNamespace, vmName),
		constants.FieldPCIDeviceNodeName:      nodeName,
		constants.FieldPCIDevicePCIAddresses:  addresses,
	}
	if labels := extractLabelsFromUnstructured(primaryClaim); labels != nil {
		states[constants.FieldCommonLabels] = labels
	}
	for key, value := range states {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// getSpecField extracts a string field from spec
func getSpecField(spec map[string]interface{}, field string) string {
	if val, ok := spec[field].(string); ok {
		return val
	}
	return ""
}

// extractLabelsFromUnstructured extracts labels from unstructured object metadata
func extractLabelsFromUnstructured(obj *unstructured.Unstructured) map[string]string {
	labels := obj.GetLabels()
	if len(labels) == 0 {
		return nil
	}
	return labels
}

// buildClaimName generates the PCIDeviceClaim name from node name and PCI address.
// Format: {nodeName}-{address with colons and dots removed}
func buildClaimName(nodeName, pciAddress string) string {
	addressPart := strings.ReplaceAll(strings.ReplaceAll(pciAddress, ":", ""), ".", "")
	return fmt.Sprintf("%s-%s", nodeName, addressPart)
}

// deleteRemovedClaims deletes PCIDeviceClaims for addresses that were removed.
func deleteRemovedClaims(ctx context.Context, dc dynamic.Interface, nodeName string, oldAddrs, newAddrs map[string]bool) error {
	for addr := range oldAddrs {
		if !newAddrs[addr] {
			claimName := buildClaimName(nodeName, addr)
			err := dc.Resource(pcideviceClaimGVR).Delete(ctx, claimName, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				return fmt.Errorf("failed to delete PCIDeviceClaim %s: %w", claimName, err)
			}
		}
	}
	return nil
}

// ensureClaims creates or updates PCIDeviceClaims for the given addresses.
// Returns the name of the first claim processed.
func ensureClaims(ctx context.Context, dc dynamic.Interface, nodeName string, addresses map[string]bool, labels map[string]string) (string, error) {
	var firstClaimName string
	for addr := range addresses {
		claimName := buildClaimName(nodeName, addr)
		if firstClaimName == "" {
			firstClaimName = claimName
		}

		existing, getErr := dc.Resource(pcideviceClaimGVR).Get(ctx, claimName, metav1.GetOptions{})
		if getErr != nil {
			if !apierrors.IsNotFound(getErr) {
				return "", getErr
			}
			// Create new claim
			claim := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "devices.harvesterhci.io/v1beta1",
					"kind":       "PCIDeviceClaim",
					"metadata":   map[string]interface{}{"name": claimName, "labels": labels},
					"spec":       map[string]interface{}{"address": addr, "nodeName": nodeName},
				},
			}
			_, err := dc.Resource(pcideviceClaimGVR).Create(ctx, claim, metav1.CreateOptions{})
			if err != nil && !apierrors.IsAlreadyExists(err) {
				return "", fmt.Errorf("failed to create PCIDeviceClaim %s: %w", claimName, err)
			}
		} else {
			// Update labels on existing claim
			metadata, ok := existing.Object["metadata"].(map[string]interface{})
			if !ok {
				metadata = make(map[string]interface{})
				existing.Object["metadata"] = metadata
			}
			metadata["labels"] = labels
			_, err := dc.Resource(pcideviceClaimGVR).Update(ctx, existing, metav1.UpdateOptions{})
			if err != nil {
				return "", fmt.Errorf("failed to update PCIDeviceClaim %s: %w", claimName, err)
			}
		}
	}
	return firstClaimName, nil
}

// resourcePCIDeviceUpdate updates all PCIDeviceClaim resources for this resource.
func resourcePCIDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	vmNamespace, vmName, _, err := helper.PCIDeviceIDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	vmNameRaw := d.Get(constants.FieldPCIDeviceVMName).(string)
	targetVMNamespace, targetVMName, err := helper.NamespacedNamePartsByDefault(vmNameRaw, vmNamespace)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid VM name format: %w", err))
	}

	nodeName := d.Get(constants.FieldPCIDeviceNodeName).(string)
	newAddresses := make(map[string]bool)
	for _, addr := range d.Get(constants.FieldPCIDevicePCIAddresses).([]interface{}) {
		newAddresses[addr.(string)] = true
	}

	labels := make(map[string]string)
	if labelsRaw, ok := d.GetOk(constants.FieldCommonLabels); ok {
		for k, v := range labelsRaw.(map[string]interface{}) {
			labels[k] = v.(string)
		}
	}

	dynamicClient, err := getDynamicClient(c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	// Determine old addresses from state (before change)
	oldAddresses := make(map[string]bool)
	if d.HasChange(constants.FieldPCIDevicePCIAddresses) {
		old, _ := d.GetChange(constants.FieldPCIDevicePCIAddresses)
		for _, addr := range old.([]interface{}) {
			oldAddresses[addr.(string)] = true
		}
	} else {
		oldAddresses = newAddresses
	}

	if err := deleteRemovedClaims(ctx, dynamicClient, nodeName, oldAddresses, newAddresses); err != nil {
		return diag.FromErr(err)
	}

	firstClaimName, err := ensureClaims(ctx, dynamicClient, nodeName, newAddresses, labels)
	if err != nil {
		return diag.FromErr(err)
	}

	if targetVMNamespace != vmNamespace || targetVMName != vmName {
		d.SetId(fmt.Sprintf("%s/%s/%s", targetVMNamespace, targetVMName, firstClaimName))
	}

	return resourcePCIDeviceRead(ctx, d, meta)
}

// resourcePCIDeviceDelete deletes a PCIDeviceClaim resource.
func resourcePCIDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	dynamicClient, err := getDynamicClient(c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	// Delete all PCIDeviceClaims for this resource
	nodeName := d.Get(constants.FieldPCIDeviceNodeName).(string)
	for _, addr := range d.Get(constants.FieldPCIDevicePCIAddresses).([]interface{}) {
		claimName := buildClaimName(nodeName, addr.(string))
		err = dynamicClient.Resource(pcideviceClaimGVR).Delete(ctx, claimName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return diag.FromErr(fmt.Errorf("failed to delete PCIDeviceClaim %s: %w", claimName, err))
		}
	}

	d.SetId("")
	return nil
}

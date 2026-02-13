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
	// Format: {nodeName}-{address with colons replaced by dashes}
	// Example: harv1.home.lo-0000001f3 for address 0000:00:1f.3 on node harv1.home.lo
	createdClaimNames := []string{}

	for _, pciAddress := range pciAddresses {
		// Generate claim name: must match PCIDevice name format
		// Convert address 0000:00:1f.3 to 0000001f3 (remove colons and dots)
		addressPart := strings.ReplaceAll(strings.ReplaceAll(pciAddress, ":", ""), ".", "")
		claimName := fmt.Sprintf("%s-%s", nodeName, addressPart)

		// Build PCIDeviceClaim object
		pcideviceClaim := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "devices.harvesterhci.io/v1beta1",
				"kind":       "PCIDeviceClaim",
				"metadata": map[string]interface{}{
					"name":   claimName,
					"labels": labels,
				},
				"spec": map[string]interface{}{
					"address":  pciAddress, // Single address per claim
					"nodeName": nodeName,
				},
			},
		}

		// Create the PCIDeviceClaim (cluster-scoped, no namespace)
		created, err := dynamicClient.Resource(pcideviceClaimGVR).Create(ctx, pcideviceClaim, metav1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				// Claim already exists, use existing one
				createdClaimNames = append(createdClaimNames, claimName)
				continue
			}
			return diag.FromErr(fmt.Errorf("failed to create PCIDeviceClaim %s (GVR: %s/%s/%s): %w", claimName, pcideviceClaimGVR.Group, pcideviceClaimGVR.Version, pcideviceClaimGVR.Resource, err))
		}

		createdClaimNames = append(createdClaimNames, created.GetName())
	}

	// Set resource ID (format: namespace/vmname/claimname)
	// Use the first claim name as the primary identifier
	d.SetId(fmt.Sprintf("%s/%s/%s", vmNamespace, vmName, createdClaimNames[0]))

	return resourcePCIDeviceRead(ctx, d, meta)
}

// resourcePCIDeviceRead reads the state of an existing PCIDeviceClaim resource.
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

	// PCIDeviceClaim is cluster-scoped, not namespaced
	pcideviceClaim, err := dynamicClient.Resource(pcideviceClaimGVR).Get(ctx, claimName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Extract and set resource state
	return setPCIDeviceResourceData(d, pcideviceClaim, vmNamespace, vmName)
}

// setPCIDeviceResourceData sets all resource data from PCIDeviceClaim
func setPCIDeviceResourceData(d *schema.ResourceData, claim *unstructured.Unstructured, vmNamespace, vmName string) diag.Diagnostics {
	spec, ok := claim.Object["spec"].(map[string]interface{})
	if !ok {
		return diag.Errorf("invalid PCIDeviceClaim spec")
	}

	// Build state map
	// Note: We intentionally do NOT set the "name" field here.
	// The PCIDeviceClaim name is auto-generated (format: nodename-address) and differs
	// from the user-provided Terraform resource name. Setting it would cause drift.
	states := map[string]interface{}{
		constants.FieldCommonNamespace:   vmNamespace,
		constants.FieldPCIDeviceVMName:   helper.BuildNamespacedName(vmNamespace, vmName),
		constants.FieldPCIDeviceNodeName: getSpecField(spec, "nodeName"),
	}

	// Add address if present
	if address := getSpecField(spec, "address"); address != "" {
		states[constants.FieldPCIDevicePCIAddresses] = []string{address}
	}

	// Add labels if present
	if labels := extractLabelsFromUnstructured(claim); labels != nil {
		states[constants.FieldCommonLabels] = labels
	}

	// Set all states
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

// resourcePCIDeviceUpdate updates an existing PCIDeviceClaim resource.
func resourcePCIDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse ID using helper
	vmNamespace, vmName, claimName, err := helper.PCIDeviceIDParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Get updated values
	vmNameRaw := d.Get(constants.FieldPCIDeviceVMName).(string)
	targetVMNamespace, targetVMName, err := helper.NamespacedNamePartsByDefault(vmNameRaw, vmNamespace)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid VM name format: %w", err))
	}

	nodeName := d.Get(constants.FieldPCIDeviceNodeName).(string)
	pciAddressesRaw := d.Get(constants.FieldPCIDevicePCIAddresses).([]interface{})
	pciAddresses := make([]string, len(pciAddressesRaw))
	for i, addr := range pciAddressesRaw {
		pciAddresses[i] = addr.(string)
	}

	labels := make(map[string]string)
	if labelsRaw, ok := d.GetOk(constants.FieldCommonLabels); ok {
		for k, v := range labelsRaw.(map[string]interface{}) {
			labels[k] = v.(string)
		}
	}

	// Get existing PCIDeviceClaim
	dynamicClient, err := getDynamicClient(c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	// PCIDeviceClaim is cluster-scoped, not namespaced
	existing, err := dynamicClient.Resource(pcideviceClaimGVR).Get(ctx, claimName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Update the PCIDeviceClaim spec
	existing.Object["spec"] = map[string]interface{}{
		"address":  pciAddresses[0], // PCIDeviceClaim uses single address
		"nodeName": nodeName,
	}

	metadata, ok := existing.Object["metadata"].(map[string]interface{})
	if !ok {
		metadata = make(map[string]interface{})
		existing.Object["metadata"] = metadata
	}
	metadata["labels"] = labels

	// PCIDeviceClaim is cluster-scoped, not namespaced
	_, err = dynamicClient.Resource(pcideviceClaimGVR).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update PCIDeviceClaim: %w", err))
	}

	// Update ID if VM changed
	if targetVMNamespace != vmNamespace || targetVMName != vmName {
		d.SetId(fmt.Sprintf("%s/%s/%s", targetVMNamespace, targetVMName, claimName))
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
	pciAddressesRaw := d.Get(constants.FieldPCIDevicePCIAddresses).([]interface{})
	for _, addr := range pciAddressesRaw {
		addressPart := strings.ReplaceAll(strings.ReplaceAll(addr.(string), ":", ""), ".", "")
		claimName := fmt.Sprintf("%s-%s", nodeName, addressPart)
		err = dynamicClient.Resource(pcideviceClaimGVR).Delete(ctx, claimName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return diag.FromErr(fmt.Errorf("failed to delete PCIDeviceClaim %s: %w", claimName, err))
		}
	}

	d.SetId("")
	return nil
}

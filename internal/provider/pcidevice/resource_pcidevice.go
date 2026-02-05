// Package pcidevice provides the Terraform resource for managing Harvester PCI device passthrough.
// This resource creates and manages PCIDeviceClaim CRDs which enable PCI device passthrough
// to VirtualMachines in Harvester.
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

	// Get node name (required)
	nodeName := d.Get(constants.FieldPCIDeviceNodeName).(string)
	if nodeName == "" {
		return diag.Errorf("node_name is required to ensure the VM runs on the correct node where PCI devices are available")
	}

	// Get PCI addresses (required)
	pciAddressesRaw := d.Get(constants.FieldPCIDevicePCIAddresses).([]interface{})
	if len(pciAddressesRaw) == 0 {
		return diag.Errorf("at least one PCI address must be specified")
	}

	pciAddresses := make([]string, len(pciAddressesRaw))
	for i, addr := range pciAddressesRaw {
		pciAddresses[i] = addr.(string)
	}

	// Get labels (optional)
	labels := make(map[string]string)
	if labelsRaw, ok := d.GetOk(constants.FieldPCIDeviceLabels); ok {
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
		// First, try to get the resource to check if API is accessible
		_, getErr := dynamicClient.Resource(pcideviceClaimGVR).Get(ctx, claimName, metav1.GetOptions{})
		if getErr == nil {
			// Claim already exists, use existing one
			createdClaimNames = append(createdClaimNames, claimName)
			continue
		}
		
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

	// Parse ID: format is namespace/vmname/claimname
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
		return diag.Errorf("invalid resource ID format: %s (expected namespace/vmname/claimname)", id)
	}

	vmNamespace := parts[0]
	vmName := parts[1]
	claimName := parts[2]

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

	// Extract spec data
	spec, ok := pcideviceClaim.Object["spec"].(map[string]interface{})
	if !ok {
		return diag.Errorf("invalid PCIDeviceClaim spec")
	}

	// Set resource data
	if err := d.Set(constants.FieldCommonNamespace, vmNamespace); err != nil {
		return diag.FromErr(err)
	}
	// Note: We intentionally do NOT set the "name" field here.
	// The PCIDeviceClaim name is auto-generated (format: nodename-address) and differs
	// from the user-provided Terraform resource name. Setting it would cause drift.

	if err := d.Set(constants.FieldPCIDeviceVMName, helper.BuildNamespacedName(vmNamespace, vmName)); err != nil {
		return diag.FromErr(err)
	}

	nodeName, _ := spec["nodeName"].(string)
	if err := d.Set(constants.FieldPCIDeviceNodeName, nodeName); err != nil {
		return diag.FromErr(err)
	}

	// PCIDeviceClaim stores a single address per claim. When multiple PCI addresses
	// are specified, multiple claims are created. Here we read the address from this claim.
	address, ok := spec["address"].(string)
	if ok && address != "" {
		if err := d.Set(constants.FieldPCIDevicePCIAddresses, []string{address}); err != nil {
			return diag.FromErr(err)
		}
	}

	metadata, ok := pcideviceClaim.Object["metadata"].(map[string]interface{})
	if ok {
		if labels, ok := metadata["labels"].(map[string]interface{}); ok && len(labels) > 0 {
			labelMap := make(map[string]string)
			for k, v := range labels {
				if str, ok := v.(string); ok {
					labelMap[k] = str
				}
			}
			if len(labelMap) > 0 {
				if err := d.Set(constants.FieldPCIDeviceLabels, labelMap); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return nil
}

// resourcePCIDeviceUpdate updates an existing PCIDeviceClaim resource.
func resourcePCIDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := meta.(*config.Config).K8sClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse ID
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
		return diag.Errorf("invalid resource ID format: %s (expected namespace/vmname/claimname)", id)
	}

	vmNamespace := parts[0]
	vmName := parts[1]
	claimName := parts[2]

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
	if labelsRaw, ok := d.GetOk(constants.FieldPCIDeviceLabels); ok {
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

	// Parse ID
	id := d.Id()
	parts := strings.Split(id, "/")
	if len(parts) < 3 {
		return diag.Errorf("invalid resource ID format: %s (expected namespace/vmname/claimname)", id)
	}

	claimName := parts[2]

	// Delete the PCIDeviceClaim
	dynamicClient, err := getDynamicClient(c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create dynamic client: %w", err))
	}

	// PCIDeviceClaim is cluster-scoped, not namespaced
	err = dynamicClient.Resource(pcideviceClaimGVR).Delete(ctx, claimName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return diag.FromErr(fmt.Errorf("failed to delete PCIDeviceClaim: %w", err))
	}

	d.SetId("")
	return nil
}


# Virtual GPU Devices

This document describes how the different CRDs work together to configure and use virtual GPUs.

## Basics

GPUs are not just processors, but more or less entire computers. This means they also have memory.
Virtual GPUs mirror this, but they represent only a share of the resources of the physical GPU.
This means that each vGPU looks to the virtual machine like any other GPU, but its compute power is
only a time-share of the physical GPUs compute and the vGPU's memory is only a share of the physical
GPU's memory.
How big these shares are and what other features are included (supported screen sizes, frame rates,
etc.) depends on the vGPU type. Which vGPU types can be configured depends on which mode the
physical GPU operates in.

## SRIOVGPUDevice

The `SRIOVGPUDevice` CRD is the resource which configures the physical GPU. It determines the
operation mode, the number an sizes of available vGPUs and the available vGPU types.
At the time of writing this, only the equal-size mode is supported, meaning that all vGPUs will have
the same sizes.

## VGPUDevice

The `VGPUDevice` CRD configures a specific vGPU. It specifies which exact vGPU type is used, which
in turn determines the exact memory and compute resources assigned to the vGPU, as well as other
features (available screen resolutions, maximum frame rates, etc.).
Each vGPU is exposed as a PCI device, which can be attached to a VM much like any other PCI device.

## Diagram


          ┌─────────────────────────────────┐
          │ SR-IOV GPU Device CRD           │
          │     │                           │
          │     │                           │
          └─────┼───────────────────────────┘
                │configures mode
                │
                │          ┌─────────────────────────────┐
                │          │ vGPU Device CRD             │
┌───────────────┼─────────►│  │                          │
│               │          │  │                          │
│               │          └──┼──────────────────────────┘
│enables        │             │
│               │             │
│               │             │configures type
│               ▼             │
│         ┌─────────────┐─────┼────────────────────────────────────────────────────────┐
│         │Physical GPU │     │                                                        │
│         └─────────────┘     │                                                        │
│         │ - mode determines │                                                        │
│         │   available types │                                                        │
│         │ - compute and mem-│ory is divided between vGPUs                            │
│         │ - number and types│of vGPUs depdends on vendor, model and mode             │
│         │                   ▼                                                        │
│         ┌─────────────────────────────────┌───────────────┌────────────────┐ ────────┐
│         │ vGPU 0                          │ vGPU 1        │ vGPU 2         │         │
│         │ - share of memory               │               │                │         │
│         │ - timeshare of compute          │               │                │ . . .   │
│         │ - other features depend on type │               │                │         │
│         └──┬──────────────────────────────└───────────────└────────────────┘ ────────┘
│            │
│            ▼
│         ┌─────────────────────────────────┐
│         │ PCI Device CRD                  │                 ┌────────────────────────┐
│         │                                 │ attaches to     │ Virtual Machine        │
│         │                                 ├────────────────►│                        │
│         └─────────────────────────────────┘ as host device  │                        │
│            ▲                                                │                        │
│            │                                                │                        │
│            │                                                │                        │
│            │enables PCI passthrough                         │                        │
│         ┌──┴──────────────────────────────┐                 │                        │
│         │ PCI Device Claim CRD            │                 │                        │
└─────────┤                                 │                 │                        │
          │                                 │                 │                        │
          └─────────────────────────────────┘                 └────────────────────────┘

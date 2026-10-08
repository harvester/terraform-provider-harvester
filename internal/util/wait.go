package util

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/harvester/terraform-provider-harvester/pkg/constants"
)

// pollInterval is a variable so that tests do not wait between polls.
var pollInterval = constants.PollInterval

// WaitForReady polls ready until it reports true or returns an error. The SDK
// gives every operation a context that expires with the timeout the user set
// in the timeouts block, so the wait ends with it.
func WaitForReady(ctx context.Context, ready func(context.Context) (bool, error)) error {
	return wait.PollUntilContextCancel(ctx, pollInterval, true, ready)
}

// WaitForDeletion polls get until it returns a NotFound error, so that a
// resource whose delete returned before the object is gone (finalizers still
// running) does not let a dependent deletion start too early.
func WaitForDeletion(ctx context.Context, get func(context.Context) error) error {
	return WaitForReady(ctx, func(ctx context.Context) (bool, error) {
		err := get(ctx)
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
}

// RetryDelete calls del until the API accepts the delete or reports the
// object as already gone. kube-ovn webhooks refuse to delete an object while
// others still use it (a VPC with subnets, a subnet with allocated IPs), and
// those are often being released by the same destroy; any other refusal is
// returned once the operation times out.
func RetryDelete(ctx context.Context, del func(context.Context) error) error {
	var lastErr error
	err := WaitForReady(ctx, func(ctx context.Context) (bool, error) {
		lastErr = del(ctx)
		return lastErr == nil || apierrors.IsNotFound(lastErr), nil
	})
	if err != nil && lastErr != nil {
		return fmt.Errorf("%w, last error: %w", err, lastErr)
	}
	return err
}

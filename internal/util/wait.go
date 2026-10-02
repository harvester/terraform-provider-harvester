package util

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// WaitForDeletion polls get until it returns a NotFound error, so that a
// resource whose delete returned before the object is gone (finalizers still
// running) does not let a dependent deletion start too early.
func WaitForDeletion(ctx context.Context, timeout, interval time.Duration, get func(context.Context) error) error {
	deadline := time.Now().Add(timeout)
	for {
		err := get(ctx)
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for the object to be deleted", timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

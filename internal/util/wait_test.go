package util

import (
	"context"
	"errors"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var notFound = apierrors.NewNotFound(schema.GroupResource{Group: "kubeovn.io", Resource: "iptables-dnat-rules"}, "r")

func TestWaitForDeletionReturnsOnceGone(t *testing.T) {
	calls := 0
	err := WaitForDeletion(context.Background(), time.Second, time.Millisecond, func(context.Context) error {
		calls++
		if calls < 3 {
			return nil // still there, being finalized
		}
		return notFound
	})
	if err != nil || calls != 3 {
		t.Errorf("err = %v, calls = %d, want nil and 3", err, calls)
	}
}

func TestWaitForDeletionTimesOut(t *testing.T) {
	err := WaitForDeletion(context.Background(), 5*time.Millisecond, time.Millisecond, func(context.Context) error { return nil })
	if err == nil {
		t.Error("expected a timeout error")
	}
}

func TestWaitForDeletionReturnsOtherErrors(t *testing.T) {
	boom := errors.New("boom")
	if err := WaitForDeletion(context.Background(), time.Second, time.Millisecond, func(context.Context) error { return boom }); !errors.Is(err, boom) {
		t.Errorf("err = %v, want boom", err)
	}
}

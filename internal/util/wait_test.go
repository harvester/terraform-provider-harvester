package util

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var subnetsResource = schema.GroupResource{Group: kubeovnv1.SchemeGroupVersion.Group, Resource: "subnets"}

type waitResponse struct {
	ready bool
	err   error
}

func TestWaitForReady(t *testing.T) {
	pollInterval = time.Millisecond
	boom := errors.New("boom")
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, cancelExpired := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancelExpired()

	testcases := []struct {
		name          string
		ctx           context.Context
		responses     []waitResponse // returned by successive polls, the last one repeats
		expectedErr   string         // empty when the wait must succeed
		expectedCalls int
	}{
		{
			name:          "already ready",
			ctx:           context.Background(),
			responses:     []waitResponse{{ready: true}},
			expectedCalls: 1,
		},
		{
			name:          "ready after several polls",
			ctx:           context.Background(),
			responses:     []waitResponse{{}, {}, {ready: true}},
			expectedCalls: 3,
		},
		{
			name:          "errors stop the wait",
			ctx:           context.Background(),
			responses:     []waitResponse{{}, {err: boom}},
			expectedErr:   boom.Error(),
			expectedCalls: 2,
		},
		{
			name:          "timeout of the operation already passed",
			ctx:           expired,
			responses:     []waitResponse{{}},
			expectedErr:   context.DeadlineExceeded.Error(),
			expectedCalls: 1,
		},
		{
			name:          "cancelled context stops the wait",
			ctx:           cancelled,
			responses:     []waitResponse{{}},
			expectedErr:   context.Canceled.Error(),
			expectedCalls: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			err := WaitForReady(tc.ctx, func(context.Context) (bool, error) {
				calls++
				response := tc.responses[min(calls, len(tc.responses))-1]
				return response.ready, response.err
			})
			checkWaitResult(t, err, tc.expectedErr, calls, tc.expectedCalls)
		})
	}
}

func TestWaitForDeletion(t *testing.T) {
	pollInterval = time.Millisecond
	notFound := apierrors.NewNotFound(subnetsResource, "s")
	boom := errors.New("boom")
	expired, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	testcases := []struct {
		name          string
		ctx           context.Context
		responses     []error // returned by successive gets, the last one repeats
		expectedErr   string  // empty when the wait must succeed
		expectedCalls int
	}{
		{
			name:          "already gone",
			ctx:           context.Background(),
			responses:     []error{notFound},
			expectedCalls: 1,
		},
		{
			name:          "gone after several polls",
			ctx:           context.Background(),
			responses:     []error{nil, nil, notFound},
			expectedCalls: 3,
		},
		{
			name:          "other errors are returned",
			ctx:           context.Background(),
			responses:     []error{nil, boom},
			expectedErr:   boom.Error(),
			expectedCalls: 2,
		},
		{
			name:          "timeout of the operation already passed",
			ctx:           expired,
			responses:     []error{nil},
			expectedErr:   context.DeadlineExceeded.Error(),
			expectedCalls: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			err := WaitForDeletion(tc.ctx, func(context.Context) error {
				calls++
				return tc.responses[min(calls, len(tc.responses))-1]
			})
			checkWaitResult(t, err, tc.expectedErr, calls, tc.expectedCalls)
		})
	}
}

func TestRetryDelete(t *testing.T) {
	pollInterval = time.Millisecond
	notFound := apierrors.NewNotFound(subnetsResource, "s")
	inUse := apierrors.NewForbidden(subnetsResource, "s", errors.New("can't delete subnet when any IPs in Using"))
	expired, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	testcases := []struct {
		name          string
		ctx           context.Context
		responses     []error // returned by successive deletes, the last one repeats
		expectedErr   string  // empty when the delete must succeed
		expectedCalls int
	}{
		{
			name:          "accepted at once",
			ctx:           context.Background(),
			responses:     []error{nil},
			expectedCalls: 1,
		},
		{
			name:          "already gone",
			ctx:           context.Background(),
			responses:     []error{notFound},
			expectedCalls: 1,
		},
		{
			name:          "accepted once the IPs are released",
			ctx:           context.Background(),
			responses:     []error{inUse, inUse, nil},
			expectedCalls: 3,
		},
		{
			name:          "still refused when the operation times out",
			ctx:           expired,
			responses:     []error{inUse},
			expectedErr:   "IPs in Using",
			expectedCalls: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			err := RetryDelete(tc.ctx, func(context.Context) error {
				calls++
				return tc.responses[min(calls, len(tc.responses))-1]
			})
			checkWaitResult(t, err, tc.expectedErr, calls, tc.expectedCalls)
			if tc.expectedErr != "" && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err = %v, want it to wrap the expired deadline", err)
			}
		})
	}
}

func checkWaitResult(t *testing.T, err error, expectedErr string, calls, expectedCalls int) {
	t.Helper()
	if expectedErr == "" && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if expectedErr != "" && (err == nil || !strings.Contains(err.Error(), expectedErr)) {
		t.Errorf("err = %v, want an error containing %q", err, expectedErr)
	}
	if calls != expectedCalls {
		t.Errorf("calls = %d, want %d", calls, expectedCalls)
	}
}

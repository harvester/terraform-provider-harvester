package constants

import "time"

// PollInterval is the delay between two reads of an object while waiting for
// it to become ready or to be deleted. The wait itself ends with the timeout
// the user configured for the operation.
const PollInterval = 2 * time.Second

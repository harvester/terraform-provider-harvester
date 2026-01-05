package volumebackup

// RecurringJobSpec represents a Longhorn recurring job specification
type RecurringJobSpec struct {
	Name        string            `json:"name"`
	Task        string            `json:"task"`
	Groups      []string          `json:"groups,omitempty"`
	Cron        string            `json:"cron"`
	Retain      int               `json:"retain"`
	Concurrency int               `json:"concurrency"`
	Labels      map[string]string `json:"labels,omitempty"`
}

package tile

import (
	"sync"
	"time"
)

// JobState is the render lifecycle state of one tile.
type JobState uint8

const (
	JobPending JobState = iota
	JobRendering
	JobPublished
)

// String returns a stable state name.
func (s JobState) String() string {
	switch s {
	case JobPending:
		return "pending"
	case JobRendering:
		return "rendering"
	case JobPublished:
		return "published"
	default:
		return "unknown"
	}
}

// Job tracks one tile render through the pending -> rendering -> published
// state machine.
type Job struct {
	Key       string
	State     JobState
	StartedAt time.Time
}

// JobTracker owns the render jobs for all tile keys.
type JobTracker struct {
	mu   sync.Mutex
	jobs map[string]*Job
}

// NewJobTracker builds an empty tracker.
func NewJobTracker() *JobTracker {
	return &JobTracker{jobs: make(map[string]*Job)}
}

// Begin moves a key from pending to rendering. It returns nil when the key is
// already being rendered by another caller.
func (t *JobTracker) Begin(key string) *Job {
	t.mu.Lock()
	defer t.mu.Unlock()
	if job, ok := t.jobs[key]; ok && job.State == JobRendering {
		return nil
	}
	job := &Job{Key: key, State: JobRendering, StartedAt: time.Now()}
	t.jobs[key] = job
	return job
}

// Publish marks a key as published.
func (t *JobTracker) Publish(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if job, ok := t.jobs[key]; ok {
		job.State = JobPublished
	}
}

// Reset returns a published key to pending, normally after invalidation.
func (t *JobTracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if job, ok := t.jobs[key]; ok {
		job.State = JobPending
	}
}

// ResetAll returns every job to pending.
func (t *JobTracker) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, job := range t.jobs {
		job.State = JobPending
	}
}

// Stats summarizes the job states.
func (t *JobTracker) Stats() map[JobState]int {
	t.mu.Lock()
	defer t.mu.Unlock()
	stats := map[JobState]int{JobPending: 0, JobRendering: 0, JobPublished: 0}
	for _, job := range t.jobs {
		stats[job.State]++
	}
	return stats
}

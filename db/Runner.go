package db

import (
	"encoding/base64"
	"slices"
	"time"

	"github.com/gorilla/securecookie"
)

type RunnerState string

// RunnerStatus reports whether a runner is currently reachable.
type RunnerStatus string

const (
	RunnerStatusOnline  RunnerStatus = "online"
	RunnerStatusOffline RunnerStatus = "offline"
)

// RunnerCacheCleanupStatus is the durable state of a runner cache cleanup
// request. A cleanup is a confirmable, at-least-once operation: the server
// keeps the request in one of the active states until the runner reports a
// successful deletion carrying the matching request ID.
type RunnerCacheCleanupStatus string

const (
	// RunnerCacheCleanupWaiting means the request was accepted but the runner
	// still has in-scope jobs (running or queued) using the cache directories.
	// The runner does not accept new matching jobs while in this state.
	RunnerCacheCleanupWaiting RunnerCacheCleanupStatus = "waiting"
	// RunnerCacheCleanupCleaning means the runner is deleting the cache
	// directories of the request scope.
	RunnerCacheCleanupCleaning RunnerCacheCleanupStatus = "cleaning"
	// RunnerCacheCleanupFailed means the last deletion attempt failed. The
	// runner keeps refusing new jobs and retries the deletion automatically.
	RunnerCacheCleanupFailed RunnerCacheCleanupStatus = "failed"
	// RunnerCacheCleanupDone means the runner confirmed a successful deletion
	// for the request. It is kept on the row as the result of the latest
	// request until a new one supersedes it.
	RunnerCacheCleanupDone RunnerCacheCleanupStatus = "done"
)

// Valid reports whether the value is a cleanup state the API accepts.
func (s RunnerCacheCleanupStatus) Valid() bool {
	switch s {
	case RunnerCacheCleanupWaiting, RunnerCacheCleanupCleaning, RunnerCacheCleanupFailed, RunnerCacheCleanupDone:
		return true
	default:
		return false
	}
}

// IsActive reports whether a cleanup is still pending (waiting for jobs,
// deleting, or retrying after a failure). A runner with an active cleanup is
// excluded from task dispatch.
func (s RunnerCacheCleanupStatus) IsActive() bool {
	return s == RunnerCacheCleanupWaiting || s == RunnerCacheCleanupCleaning || s == RunnerCacheCleanupFailed
}

type RunnerTagFilterMode string

const (
	RunnerFilterTagCompleteMatch RunnerTagFilterMode = "complete_match"
	RunnerFilterHasAnyTag        RunnerTagFilterMode = "has_any_tag"
	RunnerFilterIgnoreTags       RunnerTagFilterMode = "ignore_tags"
	RunnerFilterIsDefault        RunnerTagFilterMode = "is_default"
)

type Runner struct {
	ID                int        `db:"id" json:"id"`
	Token             string     `db:"token" json:"-" backup:"-"`
	ProjectID         *int       `db:"project_id" json:"project_id"`
	Webhook           string     `db:"webhook" json:"webhook"`
	MaxParallelTasks  int        `db:"max_parallel_tasks" json:"max_parallel_tasks"`
	Active            bool       `db:"active" json:"active"`
	IsDefault         bool       `db:"is_default" json:"is_default"`
	Name              string     `db:"name" json:"name"`
	Tags              []string   `db:"-" json:"tags" backup:"tags"`
	Touched           *time.Time `db:"touched" json:"touched"`
	CleaningRequested *time.Time `db:"cleaning_requested" json:"cleaning_requested"`

	// Cache cleanup request state. A DELETE .../cache request does not delete
	// anything itself: it creates a cleanup request identified by
	// CleaningRequestID in state RunnerCacheCleanupWaiting. The polled runner
	// drains its in-scope jobs, performs the deletion and confirms the result
	// with the same request ID. CleaningStatus stays active until that
	// confirmation, so restarts and poll retries resume the request and a
	// late confirmation of an older request cannot finish a newer one.
	CleaningRequestID  *string                   `db:"cleaning_request_id" json:"cleaning_request_id"`
	CleaningStatus     *string                   `db:"cleaning_status" json:"cleaning_status"`
	CleaningError      *string                   `db:"cleaning_error" json:"cleaning_error"`
	CleaningUpdatedAt  *time.Time                `db:"cleaning_updated_at" json:"cleaning_updated_at"`

	// StartedAt is the runner process's start time, reported by the runner on
	// every poll (X-Runner-Started-At header) and persisted next to Touched.
	// It changes on every restart, which is how the server detects that a
	// runner lost its in-memory job pool while still polling.
	StartedAt *time.Time `db:"started_at" json:"started_at"`

	PublicKey *string `db:"public_key" json:"-"`

	// Registered is a transient flag (never persisted) used at creation time to
	// request a runner without an auth token. Such a runner gets a one-time,
	// short-lived registration token instead and must be registered later by
	// presenting that token to `semaphore runner register`.
	Registered bool `db:"-" json:"registered"`

	// Status is a transient field (never persisted) reporting whether the runner
	// is currently online or offline, derived from heartbeat liveness. It is
	// populated for API responses via FillStatus.
	Status RunnerStatus `db:"-" json:"status"`

	// RegistrationTokenHash is the stored SHA-256 hash of the one-time registration
	// token (the plaintext is never persisted). It is issued on demand via
	// RegenerateRegistrationToken, not at creation time.
	RegistrationTokenHash      *string    `db:"registration_token" json:"-" backup:"-"`
	RegistrationTokenExpiresAt *time.Time `db:"registration_token_expires_at" json:"-" backup:"-"`
}

// IsRegistered reports whether the runner has been registered (has a token).
func (r Runner) IsRegistered() bool {
	return r.Token != ""
}

// GenerateRunnerToken creates a new runner authentication token.
func GenerateRunnerToken() string {
	return base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
}

// HasTag reports whether the runner is tagged with the given tag.
func (r Runner) HasTag(tag string) bool {
	return slices.Contains(r.Tags, tag)
}

// IsOnline reports whether the runner is considered reachable for dispatch.
// A poll-based runner is online while its last poll (Touched) is within
// offlineTimeout. Webhook-driven runners do not poll, so heartbeat staleness
// does not apply to them — they are always dispatch candidates.
func (r Runner) IsOnline(now time.Time, offlineTimeout time.Duration) bool {
	if r.Webhook != "" {
		return true
	}
	return r.Touched != nil && now.Sub(*r.Touched) <= offlineTimeout
}

// HasFreeCapacity reports whether the runner can take one more task.
// MaxParallelTasks == 0 means unlimited.
func (r Runner) HasFreeCapacity(runningTasks int) bool {
	return r.MaxParallelTasks == 0 || runningTasks < r.MaxParallelTasks
}

// FillStatus populates the transient Status field from heartbeat liveness.
func (r *Runner) FillStatus(now time.Time, offlineTimeout time.Duration) {
	if r.IsOnline(now, offlineTimeout) {
		r.Status = RunnerStatusOnline
	} else {
		r.Status = RunnerStatusOffline
	}
}

// CacheCleanupActive reports whether a cleanup request is currently in
// progress (waiting for jobs, deleting or retrying after a failure). Such a
// runner is paused for new task dispatch.
func (r Runner) CacheCleanupActive() bool {
	return r.CleaningStatus != nil && RunnerCacheCleanupStatus(*r.CleaningStatus).IsActive()
}

// CacheCleanupScope returns "project" for a project runner (only its project
// directory is deleted) and "all" for a global runner (the whole cache
// directory is deleted, preserving the historical full-cache semantics).
func (r Runner) CacheCleanupScope() string {
	if r.ProjectID != nil {
		return "project"
	}
	return "all"
}

type RunnerTag struct {
	Tag             string `db:"-" json:"tag"`
	NumberOfRunners int    `db:"-" json:"number_of_runners"`
}

package runners

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// errCacheCleanupMissingProject is reported when a project-scoped cleanup
// directive carries no project ID. It fails the attempt (and triggers a retry)
// instead of falling back to deleting the whole cache.
var errCacheCleanupMissingProject = errors.New("project-scoped cache cleanup requested without a project ID")

// errCacheCleanupUnknownScope is reported when the server sends a scope this
// runner does not understand. Failing the attempt is safer than guessing: the
// default must never be to delete the whole cache directory.
var errCacheCleanupUnknownScope = errors.New("cache cleanup requested with an unknown scope")

// cacheCleanupRetryInterval is the delay between deletion attempts after a
// failure. A transient file-system error (locked file, permission glitch,
// antivirus scan holding a handle) must not finish the request as failed;
// the runner keeps the request pending, stays paused for new jobs and retries.
const cacheCleanupRetryInterval = 30 * time.Second

// cacheCleanupState is the runner-side state of one cleanup request. It is
// re-created from the server directive after a restart (the directive is
// durable server-side and repeated on every poll), so no on-disk state is
// needed on the runner.
type cacheCleanupState struct {
	requestID string
	scope     CacheCleanupScope
	projectID *int

	status  db.RunnerCacheCleanupStatus
	errMsg  string
	running bool

	// lastAttempt anchors the failure backoff.
	lastAttempt time.Time
}

// cacheCleanup tracks the single cleanup request the runner is handling.
// The zero value is ready to use.
type cacheCleanup struct {
	mu    sync.Mutex
	state *cacheCleanupState
}

// handleDirective reconciles the local state with the server directive from a
// poll. A nil directive means the server has no pending request for this
// runner anymore (completed, superseded-and-done or deleted): the local state
// is dropped. A directive with a new request ID replaces any previous state
// (including a late local "done"); the same request ID keeps the local
// progress so retries and restarts continue from where the runner was.
func (c *cacheCleanup) handleDirective(directive *CacheCleanupDirective) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if directive == nil {
		if c.state != nil {
			log.WithFields(log.Fields{
				"context":         "cache_cleanup",
				"cleanup_request": c.state.requestID,
				"status":          string(c.state.status),
			}).Debug("Cache cleanup request is no longer pending on the server")
			c.state = nil
		}
		return
	}

	if c.state == nil {
		c.state = &cacheCleanupState{
			requestID: directive.RequestID,
			scope:     directive.Scope,
			projectID: directive.ProjectID,
			status:    db.RunnerCacheCleanupWaiting,
		}
		log.WithFields(log.Fields{
			"context":         "cache_cleanup",
			"cleanup_request": directive.RequestID,
			"scope":           string(directive.Scope),
		}).Info("Cache cleanup requested; runner paused for new jobs, waiting for assigned jobs to finish")
		return
	}

	if c.state.requestID != directive.RequestID {
		log.WithFields(log.Fields{
			"context":     "cache_cleanup",
			"old_request": c.state.requestID,
			"new_request": directive.RequestID,
		}).Info("Replacing superseded cache cleanup request")
		c.state = &cacheCleanupState{
			requestID: directive.RequestID,
			scope:     directive.Scope,
			projectID: directive.ProjectID,
			status:    db.RunnerCacheCleanupWaiting,
		}
	}
}

// progress returns the report to attach to the next progress PUT, or nil when
// there is no local cleanup state.
func (c *cacheCleanup) progress() *CacheCleanupProgress {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == nil {
		return nil
	}

	return &CacheCleanupProgress{
		RequestID: c.state.requestID,
		Status:    c.state.status,
		Error:     c.state.errMsg,
	}
}

// active reports whether the runner must stay paused for new jobs.
func (c *cacheCleanup) active() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state != nil && c.state.status != db.RunnerCacheCleanupDone
}

// tick advances the state machine: once no in-scope job is left, a waiting
// request starts deleting; a failed request retries after the backoff. The
// deletion itself runs in its own goroutine so progress reports (the runner
// heartbeat) continue while a large directory tree is being removed.
func (c *cacheCleanup) tick(pool *JobPool) {
	c.mu.Lock()
	state := c.state
	if state == nil || state.running {
		c.mu.Unlock()
		return
	}

	switch state.status {
	case db.RunnerCacheCleanupWaiting:
		if pool.activeJobsForCleanup() > 0 {
			c.mu.Unlock()
			return
		}
	case db.RunnerCacheCleanupFailed:
		if time.Since(state.lastAttempt) < cacheCleanupRetryInterval {
			c.mu.Unlock()
			return
		}
		if pool.activeJobsForCleanup() > 0 {
			c.mu.Unlock()
			return
		}
	default:
		// cleaning (goroutine in flight, guarded by state.running) and done
		// (waiting for the server to withdraw the directive) have nothing to do.
		c.mu.Unlock()
		return
	}

	state.status = db.RunnerCacheCleanupCleaning
	state.errMsg = ""
	state.running = true
	requestID := state.requestID
	scope := state.scope
	projectID := state.projectID
	c.mu.Unlock()

	log.WithFields(log.Fields{
		"context":         "cache_cleanup",
		"cleanup_request": requestID,
		"scope":           string(scope),
	}).Info("All assigned jobs released cache resources; deleting cache directories")

	go c.runDeletion(requestID, scope, projectID, pool)
}

// runDeletion performs the actual removal and records the outcome on the
// state. A failure never finishes the request: the state goes to "failed"
// and tick retries it. Tasks are never touched by a cleanup failure.
func (c *cacheCleanup) runDeletion(requestID string, scope CacheCleanupScope, projectID *int, pool *JobPool) {
	err := pool.performCacheCleanup(scope, projectID)

	c.mu.Lock()
	defer c.mu.Unlock()

	// The request was replaced or withdrawn while the deletion was running.
	// Its result belongs to the old request and must not drive the new state.
	if c.state == nil || c.state.requestID != requestID {
		log.WithFields(log.Fields{
			"context":         "cache_cleanup",
			"cleanup_request": requestID,
		}).Debug("Discarding cleanup result: request was superseded while deletion was running")
		return
	}

	c.state.running = false

	if err != nil {
		c.state.status = db.RunnerCacheCleanupFailed
		c.state.errMsg = err.Error()
		c.state.lastAttempt = time.Now()
		log.WithError(err).WithFields(log.Fields{
			"context":         "cache_cleanup",
			"cleanup_request": requestID,
		}).Error("Cache cleanup failed; runner stays paused and will retry")
		return
	}

	c.state.status = db.RunnerCacheCleanupDone
	c.state.errMsg = ""

	log.WithFields(log.Fields{
		"context":         "cache_cleanup",
		"cleanup_request": requestID,
	}).Info("Cache cleanup completed; waiting for server confirmation")
}

// performCacheCleanup deletes the directories of the request scope. A project
// runner removes only its project directory; a global runner removes the whole
// cache directory (full-cache semantics). An unknown scope fails instead of
// deleting anything.
func (p *JobPool) performCacheCleanup(scope CacheCleanupScope, projectID *int) error {
	switch scope {
	case CacheCleanupProject:
		if projectID == nil {
			return errCacheCleanupMissingProject
		}
		return util.Config.ClearProjectTmpDir(*projectID)
	case CacheCleanupAll:
		return util.Config.ClearTmpDir()
	default:
		return fmt.Errorf("%w: %q", errCacheCleanupUnknownScope, scope)
	}
}

// activeJobsForCleanup counts jobs that still hold cache resources in the
// cleanup scope. A project runner only ever receives its own project's jobs
// and a global cleanup covers all projects, so every non-finished running job
// and every queued job (which will create working directories as soon as it
// starts) is in scope.
func (p *JobPool) activeJobsForCleanup() int {
	count := 0

	for _, j := range p.snapshotRunningJobs() {
		if !j.getStatus().IsFinished() {
			count++
		}
	}

	p.mu.Lock()
	for _, j := range p.queue {
		if j.status != task_logger.TaskFailStatus {
			count++
		}
	}
	p.mu.Unlock()

	return count
}

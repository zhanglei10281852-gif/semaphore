package runners

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCleanupTestPool(t *testing.T) *JobPool {
	initConfig(t)
	util.Config.TmpPath = t.TempDir()
	p := NewJobPool(nil)
	return p
}

func TestCacheCleanup_DirectiveArmDisarmReplace(t *testing.T) {
	p := newCleanupTestPool(t)
	pid := 7

	assert.Nil(t, p.cleanup.progress())
	assert.False(t, p.cleanup.active())

	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "r1", Scope: CacheCleanupProject, ProjectID: &pid})

	prog := p.cleanup.progress()
	require.NotNil(t, prog)
	assert.Equal(t, "r1", prog.RequestID)
	assert.Equal(t, db.RunnerCacheCleanupWaiting, prog.Status)
	assert.True(t, p.cleanup.active())

	// Same request ID keeps the local state.
	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "r1", Scope: CacheCleanupProject, ProjectID: &pid})
	assert.Equal(t, "r1", p.cleanup.progress().RequestID)

	// A newer request supersedes the old one and resets it to waiting.
	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "r2", Scope: CacheCleanupAll})
	prog = p.cleanup.progress()
	require.NotNil(t, prog)
	assert.Equal(t, "r2", prog.RequestID)
	assert.Equal(t, db.RunnerCacheCleanupWaiting, prog.Status)

	// Server withdrawal (no directive, e.g. completed) clears local state.
	p.cleanup.handleDirective(nil)
	assert.Nil(t, p.cleanup.progress())
	assert.False(t, p.cleanup.active())
}

func TestCacheCleanup_WaitsForActiveJobsThenDeletesProjectDir(t *testing.T) {
	p := newCleanupTestPool(t)
	pid := 7

	projectDir := util.Config.GetProjectTmpDir(pid)
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "repository"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "repository", "checkout"), []byte("x"), 0o644))

	// Untouched content outside the project scope must survive.
	otherDir := util.Config.GetProjectTmpDir(8)
	require.NoError(t, os.MkdirAll(filepath.Join(otherDir, "repository"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(otherDir, "repository", "keep"), []byte("x"), 0o644))

	p.cleanup.handleDirective(&CacheCleanupDirective{
		RequestID: "req-project",
		Scope:     CacheCleanupProject,
		ProjectID: &pid,
	})

	// A running job keeps the cleanup waiting: nothing is deleted.
	rj := &runningJob{job: &tasks.LocalExecutor{Task: db.Task{ID: 42, ProjectID: pid}}, status: task_logger.TaskRunningStatus}
	p.addRunningJob(42, rj)
	p.cleanup.tick(p)
	assert.Equal(t, db.RunnerCacheCleanupWaiting, p.cleanup.progress().Status)
	assert.DirExists(t, projectDir)

	// A queued (not yet running) job also holds the scope: it would create
	// directories the moment it starts. The running job has now finished...
	p.addRunningJob(42, &runningJob{job: rj.job, status: task_logger.TaskSuccessStatus})
	p.enqueue(newTestJob(43))
	p.cleanup.tick(p)
	assert.Equal(t, db.RunnerCacheCleanupWaiting, p.cleanup.progress().Status)

	// Once the last job is gone, deletion runs and completes.
	_ = p.clearQueue()
	p.cleanup.tick(p)

	require.Eventually(t, func() bool {
		return p.cleanup.progress().Status == db.RunnerCacheCleanupDone
	}, 2*time.Second, 10*time.Millisecond)

	// ClearProjectTmpDir removes the contents of the project directory but
	// keeps the directory itself (same semantics as the synchronous clear).
	assert.NoDirExists(t, filepath.Join(projectDir, "repository"))
	assert.DirExists(t, projectDir)
	assert.FileExists(t, filepath.Join(otherDir, "repository", "keep"), "project-scoped cleanup must not touch other projects")
}

func TestCacheCleanup_GlobalScopeDeletesEverything(t *testing.T) {
	p := newCleanupTestPool(t)

	for _, pid := range []int{1, 2} {
		dir := util.Config.GetProjectTmpDir(pid)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "repository"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "repository", "f"), []byte("x"), 0o644))
	}

	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "req-all", Scope: CacheCleanupAll})
	p.cleanup.tick(p)

	require.Eventually(t, func() bool {
		return p.cleanup.progress().Status == db.RunnerCacheCleanupDone
	}, 2*time.Second, 10*time.Millisecond)

	assert.NoDirExists(t, util.Config.GetProjectTmpDir(1))
	assert.NoDirExists(t, util.Config.GetProjectTmpDir(2))
}

func TestCacheCleanup_ProjectScopeWithoutProjectIDFailsAndRetries(t *testing.T) {
	p := newCleanupTestPool(t)

	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "req-bad", Scope: CacheCleanupProject})
	p.cleanup.tick(p)

	require.Eventually(t, func() bool {
		return p.cleanup.progress().Status == db.RunnerCacheCleanupFailed
	}, 2*time.Second, 10*time.Millisecond)
	assert.NotEmpty(t, p.cleanup.progress().Error)

	// The runner stays paused while failed.
	assert.True(t, p.cleanup.active())

	// Immediately ticking must not retry: the failure backoff is active.
	p.cleanup.tick(p)
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, db.RunnerCacheCleanupFailed, p.cleanup.progress().Status)

	// Once the backoff elapsed (and the directive becomes valid), it retries.
	p.cleanup.mu.Lock()
	p.cleanup.state.lastAttempt = time.Now().Add(-cacheCleanupRetryInterval - time.Second)
	pid := 9
	p.cleanup.state.projectID = &pid
	p.cleanup.mu.Unlock()

	require.NoError(t, os.MkdirAll(util.Config.GetProjectTmpDir(pid), 0o755))
	p.cleanup.tick(p)
	require.Eventually(t, func() bool {
		return p.cleanup.progress().Status == db.RunnerCacheCleanupDone
	}, 2*time.Second, 10*time.Millisecond)
}

func TestCacheCleanup_FailureDoesNotFinishRequest(t *testing.T) {
	initConfig(t)

	// TmpPath pointing at a regular file makes the removal fail (ReadDir on a
	// file), modelling a transient file-system failure.
	path := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
	util.Config.TmpPath = path

	p := NewJobPool(nil)
	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "req-fail", Scope: CacheCleanupAll})
	p.cleanup.tick(p)

	require.Eventually(t, func() bool {
		return p.cleanup.progress().Status == db.RunnerCacheCleanupFailed
	}, 2*time.Second, 10*time.Millisecond)
	assert.NotEmpty(t, p.cleanup.progress().Error)
	assert.True(t, p.cleanup.active(), "failed cleanup must keep the runner paused")
}

func TestCacheCleanup_StaleResultAfterSupersedeDiscarded(t *testing.T) {
	p := newCleanupTestPool(t)

	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "old", Scope: CacheCleanupAll})
	// A newer request arrives before the old deletion goroutine reports back.
	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "new", Scope: CacheCleanupAll})

	// The old goroutine finishes successfully; its result belongs to "old".
	p.cleanup.runDeletion("old", CacheCleanupAll, nil, p)

	prog := p.cleanup.progress()
	require.NotNil(t, prog)
	assert.Equal(t, "new", prog.RequestID)
	assert.Equal(t, db.RunnerCacheCleanupWaiting, prog.Status, "old result must not drive the new request")
}

func TestCacheCleanup_ActiveJobsCount(t *testing.T) {
	p := newCleanupTestPool(t)

	assert.Equal(t, 0, p.activeJobsForCleanup())

	p.addRunningJob(1, &runningJob{
		job:    &tasks.LocalExecutor{Task: db.Task{ID: 1}},
		status: task_logger.TaskSuccessStatus, // finished: does not count
	})
	p.addRunningJob(2, &runningJob{
		job:    &tasks.LocalExecutor{Task: db.Task{ID: 2}},
		status: task_logger.TaskRunningStatus,
	})
	p.enqueue(newTestJob(3))
	assert.Equal(t, 2, p.activeJobsForCleanup())
}

func TestCacheCleanup_ProgressRidesOnProgressReport(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	var mu sync.Mutex
	var received RunnerProgress
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			mu.Lock()
			defer mu.Unlock()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		}
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)

	util.Config = &util.ConfigType{
		WebHost: srv.URL,
		TmpPath: t.TempDir(),
		Runner: &util.RunnerConfig{
			Token:      "test-token",
			Executor:   &util.ExecutorConfig{},
			Connection: &util.RunnerConnectionConfig{},
		},
	}

	p := NewJobPool(nil)
	p.cleanup.handleDirective(&CacheCleanupDirective{RequestID: "req-report", Scope: CacheCleanupAll})

	require.True(t, p.sendProgress())
	mu.Lock()
	require.NotNil(t, received.CacheCleanup)
	assert.Equal(t, "req-report", received.CacheCleanup.RequestID)
	assert.Equal(t, db.RunnerCacheCleanupWaiting, received.CacheCleanup.Status)
	mu.Unlock()
}

func TestCacheCleanup_CheckNewJobsRefusesJobsAndPurgesQueue(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(RunnerState{
			CacheCleanup: &CacheCleanupDirective{RequestID: "req-poll", Scope: CacheCleanupAll},
			NewJobs: []JobData{{
				Task:        db.Task{ID: 50, ProjectID: 1, TemplateID: 1},
				Template:    db.Template{ID: 1, App: db.AppAnsible},
				Inventory:   db.Inventory{ID: 1},
				Repository:  db.Repository{ID: 1},
				Environment: db.Environment{ID: 1},
			}},
			AccessKeys: map[int]db.AccessKey{},
		}))
	}))
	t.Cleanup(srv.Close)

	util.Config = &util.ConfigType{
		WebHost: srv.URL,
		TmpPath: t.TempDir(),
		Runner: &util.RunnerConfig{
			Token:      "test-token",
			Executor:   &util.ExecutorConfig{},
			Connection: &util.RunnerConnectionConfig{},
		},
	}

	p := NewJobPool(nil)
	p.enqueue(newTestJob(60))
	require.Equal(t, 1, p.queueLen())

	p.checkNewJobs()

	// Queued jobs accepted before the request are dropped (server requeues
	// them), and NewJobs in the same response are refused.
	assert.Equal(t, 0, p.queueLen())
	require.NotNil(t, p.cleanup.progress())
	assert.Equal(t, "req-poll", p.cleanup.progress().RequestID)

	// The queue ticker must not start jobs while paused.
	p.enqueue(newTestJob(61))
	job, ok := p.dequeue()
	require.True(t, ok)
	p.enqueueFront(job) // mimics the Run loop gate
	assert.Equal(t, 1, p.queueLen())
}

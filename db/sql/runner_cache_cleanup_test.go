package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupStatus(runner db.Runner) string {
	if runner.CleaningStatus == nil {
		return ""
	}
	return *runner.CleaningStatus
}

func TestRunnerCacheCleanup_GlobalLifecycle(t *testing.T) {
	store := InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r-clean"})
	require.NoError(t, err)

	// Request opens a waiting cleanup with a fresh request ID.
	requestID, err := store.RequestRunnerCacheCleanup(runner)
	require.NoError(t, err)
	assert.Len(t, requestID, 32)

	stored, err := store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	require.NotNil(t, stored.CleaningRequestID)
	assert.Equal(t, requestID, *stored.CleaningRequestID)
	assert.Equal(t, string(db.RunnerCacheCleanupWaiting), cleanupStatus(stored))
	assert.Nil(t, stored.CleaningError)
	assert.NotNil(t, stored.CleaningRequested)
	assert.NotNil(t, stored.CleaningUpdatedAt)
	assert.True(t, stored.CacheCleanupActive())

	// A failure report coming straight from waiting is invalid (the runner
	// cannot have attempted anything yet) and is ignored.
	applied, err := store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupFailed, "boom")
	require.NoError(t, err)
	assert.False(t, applied)

	// waiting -> cleaning.
	applied, err = store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupCleaning, "")
	require.NoError(t, err)
	assert.True(t, applied)

	// Idempotent re-report of cleaning (poll retry) stays applied.
	applied, err = store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupCleaning, "")
	require.NoError(t, err)
	assert.True(t, applied)

	// cleaning -> failed with error persisted, request stays active.
	applied, err = store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupFailed, "permission denied")
	require.NoError(t, err)
	assert.True(t, applied)

	stored, err = store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	assert.Equal(t, string(db.RunnerCacheCleanupFailed), cleanupStatus(stored))
	require.NotNil(t, stored.CleaningError)
	assert.Equal(t, "permission denied", *stored.CleaningError)
	assert.True(t, stored.CacheCleanupActive())

	// A retry reports cleaning again after a failure.
	applied, err = store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupCleaning, "")
	require.NoError(t, err)
	assert.True(t, applied)

	// done completes the request and clears the error.
	applied, err = store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.True(t, applied)

	stored, err = store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	assert.Equal(t, string(db.RunnerCacheCleanupDone), cleanupStatus(stored))
	assert.Nil(t, stored.CleaningError)
	assert.False(t, stored.CacheCleanupActive())

	// done is accepted exactly once: the duplicate delivery of the same
	// confirmation must not apply again (and therefore must not produce a
	// second completion audit at the API layer).
	applied, err = store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.False(t, applied)
}

func TestRunnerCacheCleanup_StaleRequestConfirmationIgnored(t *testing.T) {
	store := InitConfigCreateTestStore()
	runner, err := store.CreateRunner(db.Runner{Name: "r-stale"})
	require.NoError(t, err)

	oldID, err := store.RequestRunnerCacheCleanup(runner)
	require.NoError(t, err)

	// Simulate the old runner still working on its deletion while an admin
	// opens a newer request.
	newID, err := store.RequestRunnerCacheCleanup(runner)
	require.NoError(t, err)
	require.NotEqual(t, oldID, newID)

	stored, err := store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	assert.Equal(t, newID, *stored.CleaningRequestID)
	assert.Equal(t, string(db.RunnerCacheCleanupWaiting), cleanupStatus(stored))

	// The late confirmation of the old request must not finish the new one.
	applied, err := store.ReportRunnerCacheCleanup(stored, oldID, db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.False(t, applied)

	stored, err = store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	assert.Equal(t, string(db.RunnerCacheCleanupWaiting), cleanupStatus(stored), "newer request must stay pending")

	// Unknown request IDs and invalid statuses are ignored as well.
	applied, err = store.ReportRunnerCacheCleanup(stored, "does-not-exist", db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.False(t, applied)

	applied, err = store.ReportRunnerCacheCleanup(stored, newID, db.RunnerCacheCleanupStatus("bogus"), "")
	require.NoError(t, err)
	assert.False(t, applied)
	applied, err = store.ReportRunnerCacheCleanup(stored, "", db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.False(t, applied)

	// The matching confirmation completes the newer request.
	applied, err = store.ReportRunnerCacheCleanup(stored, newID, db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.True(t, applied)

	stored, err = store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	assert.Equal(t, string(db.RunnerCacheCleanupDone), cleanupStatus(stored))
}

func TestRunnerCacheCleanup_ProjectRunnerScopeGuard(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, err := store.CreateProject(db.Project{Name: "proj"})
	require.NoError(t, err)

	runner, err := store.CreateRunner(db.Runner{Name: "r-proj", ProjectID: &project.ID})
	require.NoError(t, err)

	requestID, err := store.RequestRunnerCacheCleanup(runner)
	require.NoError(t, err)

	stored, err := store.GetRunner(project.ID, runner.ID)
	require.NoError(t, err)
	assert.Equal(t, requestID, *stored.CleaningRequestID)
	assert.Equal(t, "project", stored.CacheCleanupScope())

	// A global-scope lookup must not match the project runner row.
	_, err = store.GetGlobalRunner(runner.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	// Reporting with a runner value from another project scope must not touch
	// the row.
	foreignProject := project.ID + 999
	foreign := stored
	foreign.ProjectID = &foreignProject
	applied, err := store.ReportRunnerCacheCleanup(foreign, requestID, db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.False(t, applied)

	// The matching scope completes the request.
	applied, err = store.ReportRunnerCacheCleanup(stored, requestID, db.RunnerCacheCleanupDone, "")
	require.NoError(t, err)
	assert.True(t, applied)
}

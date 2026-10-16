package sql

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFileTestStore opens a migrated SQLite store on a shared file database so
// two store instances are two independent connection pools, like two HA nodes.
func newFileTestStore(t *testing.T, dbFile string) *SqlDb {
	t.Helper()

	util.Config = &util.ConfigType{
		SQLite: &util.DbConfig{
			Hostname: dbFile,
		},
		Dialect: "sqlite",
		Log: &util.ConfigLog{
			Events: &util.EventLogType{},
			Tasks:  &util.TaskLogType{},
		},
		Process: &util.ConfigProcess{},
		Runners: &util.RunnersConfig{},
		Apps: map[string]util.App{
			"ansible": {},
			"bash":    {},
		},
	}

	store := CreateDb(util.DbDriverSQLite)
	store.Connect()
	require.NoError(t, db.Migrate(store, nil))

	t.Cleanup(store.Close)

	return store
}

func sqliteFileURI(t *testing.T) string {
	t.Helper()

	path := filepath.ToSlash(filepath.Join(t.TempDir(), "snapshot.db"))
	if !strings.HasPrefix(path, "/") {
		// Windows drive path (C:/...) needs the leading slash of the
		// file:// authority-less URI.
		path = "/" + path
	}
	return "//" + path
}

// TestSqlDb_ReadSnapshotSeesOneCommittedState is the core guarantee of a
// backup: while a read snapshot is open, a delete committed by another
// connection is invisible to it, and becomes visible only after it is
// released.
func TestSqlDb_ReadSnapshotSeesOneCommittedState(t *testing.T) {
	dbFile := sqliteFileURI(t)

	store := newFileTestStore(t, dbFile)
	other := newFileTestStore(t, dbFile)

	proj, err := store.CreateProject(db.Project{Name: "Snapshot project"})
	require.NoError(t, err)

	err = store.WithReadSnapshot(func(view db.Store) error {
		// The first read establishes the snapshot.
		seen, err := view.GetProject(proj.ID)
		require.NoError(t, err)
		require.Equal(t, "Snapshot project", seen.Name)

		// Concurrent commit from another connection (another node in HA): the
		// project disappears there.
		require.NoError(t, other.DeleteProject(proj.ID))
		_, err = other.GetProject(proj.ID)
		require.ErrorIs(t, err, db.ErrNotFound)

		// The snapshot still observes the state from before the delete.
		seen, err = view.GetProject(proj.ID)
		require.NoError(t, err)
		assert.Equal(t, "Snapshot project", seen.Name)

		// Every other read shares the same view, not a fresh one.
		templates, err := view.GetTemplates(proj.ID, db.TemplateFilter{}, db.RetrieveQueryParams{})
		require.NoError(t, err)
		assert.Empty(t, templates)

		return nil
	})
	require.NoError(t, err)

	// Once the snapshot is released, the committed delete is visible.
	_, err = store.GetProject(proj.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// TestSqlDb_ReadSnapshotRollsBackOnError makes sure a failed read attempt does
// not leave a transaction open and the store keeps working afterwards.
func TestSqlDb_ReadSnapshotRollsBackOnError(t *testing.T) {
	dbFile := sqliteFileURI(t)

	store := newFileTestStore(t, dbFile)

	proj, err := store.CreateProject(db.Project{Name: "Rollback project"})
	require.NoError(t, err)

	expected := db.ErrNotFound
	err = store.WithReadSnapshot(func(view db.Store) error {
		if _, err := view.GetProject(proj.ID); err != nil {
			return err
		}
		return expected
	})
	require.ErrorIs(t, err, expected)

	// No leaked transaction: reads and writes keep working on the store.
	seen, err := store.GetProject(proj.ID)
	require.NoError(t, err)
	assert.Equal(t, "Rollback project", seen.Name)
}

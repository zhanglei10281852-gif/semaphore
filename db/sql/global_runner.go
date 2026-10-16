package sql

import (
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) GetRunnerByToken(token string) (runner db.Runner, err error) {

	runners := make([]db.Runner, 0)

	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		return builder.Where("token=?", token)
	}, &runners)

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = db.ErrNotFound
		return
	}

	runner = runners[0]
	err = d.loadRunnerTagsSingle(&runner)
	return
}

func (d *SqlDb) GetGlobalRunner(runnerID int) (runner db.Runner, err error) {
	err = d.getObject(0, db.GlobalRunnerProps, runnerID, &runner)
	if err != nil {
		return
	}
	err = d.loadRunnerTagsSingle(&runner)
	return
}

func (d *SqlDb) GetAllRunners(activeAndRegisteredOnly bool, globalOnly bool, tagFilterMode db.RunnerTagFilterMode, tag *string) (runners []db.Runner, err error) {
	if tag == nil && tagFilterMode == db.RunnerFilterTagCompleteMatch {
		err = fmt.Errorf("tag filter mode is complete match but no tag was provided")
		return
	}

	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {

		if globalOnly {
			builder = builder.Where("project_id is null")
		}

		if activeAndRegisteredOnly {
			builder = builder.Where("active=true and token != ''")
		}

		switch tagFilterMode {
		case db.RunnerFilterHasAnyTag:
			builder = builder.Where(runnerHasAnyTagExpr())
		case db.RunnerFilterIsDefault:
			builder = builder.Where(runnerIsDefaultExpr())
		case db.RunnerFilterIgnoreTags:
			// No tag filtering applied.
		case db.RunnerFilterTagCompleteMatch:
			builder = builder.Where(runnerHasTagExpr(*tag))
		default:
			panic("invalid tag filter mode: " + tagFilterMode)
		}

		return builder
	}, &runners)
	if err != nil {
		return
	}
	err = d.loadRunnerTags(runners)
	return
}

func (d *SqlDb) GetGlobalRunnerTags() (res []db.RunnerTag, err error) {
	query, args, err := squirrel.Select("rt.tag", "count(distinct rt.runner_id) as cnt").
		From("runner__tag rt").
		Join("runner r on r.id = rt.runner_id").
		Where("r.project_id is null").
		GroupBy("rt.tag").
		ToSql()

	if err != nil {
		return
	}

	type row struct {
		Tag string `db:"tag"`
		Cnt int    `db:"cnt"`
	}

	rows := make([]row, 0)
	_, err = d.selectAll(&rows, query, args...)
	if err != nil {
		return
	}

	res = make([]db.RunnerTag, 0, len(rows))
	for _, r := range rows {
		res = append(res, db.RunnerTag{
			Tag:             r.Tag,
			NumberOfRunners: r.Cnt,
		})
	}

	return
}

func (d *SqlDb) DeleteGlobalRunner(runnerID int) (err error) {
	err = d.deleteObject(0, db.GlobalRunnerProps, runnerID)
	return
}

// maxRunnerCacheCleanupErrorLen matches the cleaning_error column length.
const maxRunnerCacheCleanupErrorLen = 512

func truncateRunnerCacheCleanupError(message string) string {
	if len(message) <= maxRunnerCacheCleanupErrorLen {
		return message
	}
	return message[:maxRunnerCacheCleanupErrorLen]
}

// runnerScopeCondition renders the project-scope guard used by every cleanup
// update: global runners are matched by id alone, project runners also by
// their project so a runner from another project can never touch the row.
func runnerScopeCondition(runner db.Runner) (condition string, args []any) {
	if runner.ProjectID == nil {
		return "id=? and project_id is null", []any{runner.ID}
	}
	return "id=? and project_id=?", []any{runner.ID, *runner.ProjectID}
}

func (d *SqlDb) RequestRunnerCacheCleanup(runner db.Runner) (requestID string, err error) {
	requestID = random.String(32)
	now := tz.Now()

	condition, args := runnerScopeCondition(runner)
	query := "update `runner` set `cleaning_request_id`=?, `cleaning_status`=?, " +
		"`cleaning_error`=null, `cleaning_requested`=?, `cleaning_updated_at`=? where " + condition

	updateArgs := append([]any{requestID, string(db.RunnerCacheCleanupWaiting), now, now}, args...)

	res, err := d.exec(query, updateArgs...)
	if err != nil {
		return "", err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if affected == 0 {
		return "", db.ErrNotFound
	}

	return requestID, nil
}

func (d *SqlDb) ReportRunnerCacheCleanup(runner db.Runner, requestID string, status db.RunnerCacheCleanupStatus, failureReason string) (applied bool, err error) {
	if requestID == "" || !status.Valid() {
		return false, nil
	}

	now := tz.Now()
	condition, args := runnerScopeCondition(runner)

	var query string
	updateArgs := make([]any, 0, 7)

	switch status {
	case db.RunnerCacheCleanupDone:
		// A successful confirmation completes the request exactly once:
		// rows already in "done" (e.g. a retried delivery of the same
		// confirmation) do not match and produce no second completion.
		query = "update `runner` set `cleaning_status`=?, `cleaning_error`=null, " +
			"`cleaning_updated_at`=? where " + condition +
			" and cleaning_request_id=? and cleaning_status in (?, ?, ?)"
		updateArgs = append(updateArgs, string(status), now)
		updateArgs = append(updateArgs, args...)
		updateArgs = append(updateArgs, requestID,
			string(db.RunnerCacheCleanupWaiting),
			string(db.RunnerCacheCleanupCleaning),
			string(db.RunnerCacheCleanupFailed))
	case db.RunnerCacheCleanupCleaning:
		// waiting -> cleaning, and idempotent re-reports from cleaning or
		// failed (a new attempt after a failure reports "cleaning" again).
		query = "update `runner` set `cleaning_status`=?, `cleaning_updated_at`=? where " +
			condition +
			" and cleaning_request_id=? and cleaning_status in (?, ?, ?)"
		updateArgs = append(updateArgs, string(status), now)
		updateArgs = append(updateArgs, args...)
		updateArgs = append(updateArgs, requestID,
			string(db.RunnerCacheCleanupWaiting),
			string(db.RunnerCacheCleanupCleaning),
			string(db.RunnerCacheCleanupFailed))
	case db.RunnerCacheCleanupFailed:
		query = "update `runner` set `cleaning_status`=?, `cleaning_error`=?, `cleaning_updated_at`=? where " +
			condition +
			" and cleaning_request_id=? and cleaning_status in (?, ?)"
		updateArgs = append(updateArgs, string(status), truncateRunnerCacheCleanupError(failureReason), now)
		updateArgs = append(updateArgs, args...)
		updateArgs = append(updateArgs, requestID,
			string(db.RunnerCacheCleanupCleaning),
			string(db.RunnerCacheCleanupFailed))
	case db.RunnerCacheCleanupWaiting:
		query = "update `runner` set `cleaning_updated_at`=? where " + condition +
			" and cleaning_request_id=? and cleaning_status=?"
		updateArgs = append(updateArgs, now)
		updateArgs = append(updateArgs, args...)
		updateArgs = append(updateArgs, requestID, string(db.RunnerCacheCleanupWaiting))
	}

	res, err := d.exec(query, updateArgs...)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (d *SqlDb) TouchRunner(runner db.Runner) (err error) {
	if runner.ProjectID == nil {
		_, err = d.exec(
			"update `runner` set `touched`=?, `started_at`=? where id=?",
			tz.Now(),
			runner.StartedAt,
			runner.ID)
		return
	}

	_, err = d.exec(
		"update `runner` set `touched`=?, `started_at`=? where id=? and project_id=?",
		tz.Now(),
		runner.StartedAt,
		runner.ID,
		runner.ProjectID)

	return
}

func (d *SqlDb) UpdateRunner(runner db.Runner) (err error) {
	_, err = d.exec(
		"update `runner` set `name`=?, `active`=?, `is_default`=?, webhook=?, max_parallel_tasks=? where id=?",
		runner.Name,
		runner.Active,
		runner.IsDefault,
		runner.Webhook,
		runner.MaxParallelTasks,
		runner.ID)

	if err != nil {
		return
	}

	err = d.replaceRunnerTags(runner.ID, runner.Tags)
	return
}

func (d *SqlDb) RegisterRunner(registrationTokenHash string, publicKey *string) (runner db.Runner, err error) {
	runners := make([]db.Runner, 0)

	err = d.getObjects(0, db.GlobalRunnerProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		return builder.Where("registration_token=?", registrationTokenHash)
	}, &runners)

	if err != nil {
		return
	}

	if len(runners) == 0 {
		err = db.ErrNotFound
		return
	}

	runner = runners[0]

	if runner.IsRegistered() {
		err = db.ErrRunnerAlreadyRegistered
		return
	}

	if runner.RegistrationTokenExpiresAt == nil || !runner.RegistrationTokenExpiresAt.After(tz.Now()) {
		err = db.ErrRegistrationTokenExpired
		return
	}

	token := db.GenerateRunnerToken()

	_, err = d.exec(
		"update `runner` set `token`=?, `public_key`=?, `registration_token`=null, `registration_token_expires_at`=null where id=?",
		token,
		publicKey,
		runner.ID)

	if err != nil {
		return
	}

	runner.Token = token
	runner.PublicKey = publicKey
	runner.RegistrationTokenHash = nil
	runner.RegistrationTokenExpiresAt = nil

	err = d.loadRunnerTagsSingle(&runner)
	return
}

func (d *SqlDb) ResetRunnerRegistration(runnerID int, registrationTokenHash string, expiresAt time.Time) (err error) {
	_, err = d.exec(
		"update `runner` set `token`='', `public_key`=null, `registration_token`=?, `registration_token_expires_at`=? where id=?",
		registrationTokenHash,
		expiresAt,
		runnerID)
	return
}

func (d *SqlDb) CreateRunner(runner db.Runner) (newRunner db.Runner, err error) {
	insertID, err := d.insert(
		"id",
		"insert into `runner` (project_id, token, webhook, max_parallel_tasks, `name`, `active`, `is_default`, public_key, registration_token, registration_token_expires_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		runner.ProjectID,
		runner.Token,
		runner.Webhook,
		runner.MaxParallelTasks,
		runner.Name,
		runner.Active,
		runner.IsDefault,
		runner.PublicKey,
		runner.RegistrationTokenHash,
		runner.RegistrationTokenExpiresAt)

	if err != nil {
		return
	}

	newRunner = runner
	newRunner.ID = insertID
	newRunner.Tags = normalizeTags(runner.Tags)

	if err = d.replaceRunnerTags(newRunner.ID, newRunner.Tags); err != nil {
		return
	}

	return
}

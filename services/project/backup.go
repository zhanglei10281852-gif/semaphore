package project

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
)

func findNameBySlug[T db.BackupSluggedEntity](slug string, items []T) (*string, error) {
	for _, o := range items {
		if o.GetSlug() == slug {
			name := o.GetName()
			return &name, nil
		}
	}
	return nil, fmt.Errorf("item %s does not exist", slug)
}

func findNameByID[T db.BackupEntity](ID int, items []T) (*string, error) {
	for _, o := range items {
		if o.GetID() == ID {
			name := o.GetName()
			return &name, nil
		}
	}
	return nil, fmt.Errorf("item %d does not exist", ID)
}

func findEntityByName[T db.BackupEntity](name *string, items []T) *T {
	if name == nil {
		return nil
	}
	for _, o := range items {
		if o.GetName() == *name {
			return &o
		}
	}
	return nil
}

func getSchedulesByProject(projectID int, schedules []db.Schedule) []db.Schedule {
	result := make([]db.Schedule, 0)
	for _, o := range schedules {
		if o.ProjectID == projectID {
			result = append(result, o)
		}
	}
	return result
}

func getScheduleByTemplate(templateID int, schedules []db.Schedule) *string {
	for _, o := range schedules {
		if o.TemplateID == templateID {
			return &o.CronFormat
		}
	}
	return nil
}

func getRandomName(name string) string {
	return name + " - " + random.String(10)
}

func makeUniqueNames[T any](items []T, getter func(item *T) string, setter func(item *T, name string)) {
	for i := len(items) - 1; i >= 0; i-- {
		for k, other := range items {
			if k == i {
				break
			}

			name := getter(&items[i])

			if name == getter(&other) {
				randomName := getRandomName(name)
				setter(&items[i], randomName)
				break
			}
		}
	}
}

func (b *BackupDB) makeUniqueNames() {

	makeUniqueNames(b.templates, func(item *db.Template) string {
		return item.Name
	}, func(item *db.Template, name string) {
		item.Name = name
	})

	makeUniqueNames(b.repositories, func(item *db.Repository) string {
		return item.Name
	}, func(item *db.Repository, name string) {
		item.Name = name
	})

	makeUniqueNames(b.inventories, func(item *db.Inventory) string {
		return item.Name
	}, func(item *db.Inventory, name string) {
		item.Name = name
	})

	makeUniqueNames(b.environments, func(item *db.Environment) string {
		return item.Name
	}, func(item *db.Environment, name string) {
		item.Name = name
	})

	makeUniqueNames(b.keys, func(item *db.AccessKey) string {
		return item.Name
	}, func(item *db.AccessKey, name string) {
		item.Name = name
	})

	makeUniqueNames(b.views, func(item *db.View) string {
		return item.Title
	}, func(item *db.View, name string) {
		item.Title = name
	})

	makeUniqueNames(b.integrations, func(item *db.Integration) string {
		return item.Name
	}, func(item *db.Integration, name string) {
		item.Name = name
	})

	makeUniqueNames(b.schedules, func(item *db.Schedule) string {
		return item.Name
	}, func(item *db.Schedule, name string) {
		item.Name = name
	})

	makeUniqueNames(b.roles, func(item *db.Role) string {
		return item.Name
	}, func(item *db.Role, name string) {
		item.Name = name
	})

	makeUniqueNames(b.runners, func(item *db.Runner) string {
		return item.Name
	}, func(item *db.Runner, name string) {
		item.Name = name
	})

	makeUniqueNames(b.workflows, func(item *db.WorkflowTemplate) string {
		return item.Name
	}, func(item *db.WorkflowTemplate, name string) {
		item.Name = name
	})

}

func (b *BackupDB) load(projectID int, store db.Store, workflowStore db.WorkflowManager) (err error) {

	b.workflowStore = workflowStore

	b.templates, err = store.GetTemplates(projectID, db.TemplateFilter{}, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	for i := range b.templates {
		var vaults []db.TemplateVault
		vaults, err = store.GetTemplateVaults(b.templates[i].ProjectID, b.templates[i].ID)
		if err != nil {
			return
		}
		b.templates[i].Vaults = vaults
	}

	b.repositories, err = store.GetRepositories(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	b.hostConfigs, err = store.GetHostConfigs(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	b.keys, err = store.GetAccessKeys(projectID, db.GetAccessKeyOptions{IgnoreOwner: true}, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	// Task-bound survey-secret keys are ephemeral internals of a single task
	// run; they don't belong in a project backup.
	filteredKeys := make([]db.AccessKey, 0, len(b.keys))
	for _, k := range b.keys {
		if k.Owner == db.AccessKeyTaskSecret {
			continue
		}
		filteredKeys = append(filteredKeys, k)
	}
	b.keys = filteredKeys

	b.views, err = store.GetViews(projectID)
	if err != nil {
		return
	}

	b.inventories, err = store.GetInventories(projectID, db.RetrieveQueryParams{}, []db.InventoryType{})
	if err != nil {
		return
	}

	b.environments, err = store.GetEnvironments(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	schedules, err := store.GetProjectSchedules(projectID, true, true)
	if err != nil {
		return
	}

	for _, s := range schedules {
		b.schedules = append(b.schedules, s.Schedule)
	}

	b.secretStorages, err = store.GetSecretStorages(projectID)
	if err != nil {
		return
	}

	b.roles, err = store.GetProjectRoles(projectID)
	if err != nil {
		return
	}

	b.globalRoles, err = store.GetGlobalRoles()
	if err != nil {
		return
	}

	b.meta, err = store.GetProject(projectID)
	if err != nil {
		return
	}

	b.integrationProjAliases, err = store.GetIntegrationAliases(projectID, nil)
	if err != nil {
		return
	}

	b.integrations, err = store.GetIntegrations(projectID, db.RetrieveQueryParams{}, true)
	if err != nil {
		return
	}

	b.integrationAliases = make(map[int][]db.IntegrationAlias)
	b.integrationMatchers = make(map[int][]db.IntegrationMatcher)
	b.integrationExtractValues = make(map[int][]db.IntegrationExtractValue)
	for _, o := range b.integrations {

		b.integrationAliases[o.ID], err = store.GetIntegrationAliases(projectID, &o.ID)
		if err != nil {
			return
		}
		b.integrationMatchers[o.ID], err = store.GetIntegrationMatchers(projectID, db.RetrieveQueryParams{}, o.ID)
		if err != nil {
			return
		}
		b.integrationExtractValues[o.ID], err = store.GetIntegrationExtractValues(projectID, db.RetrieveQueryParams{}, o.ID)
		if err != nil {
			return
		}
	}

	b.templateRoles = make(map[int][]db.TemplateRolePerm)
	for _, t := range b.templates {
		b.templateRoles[t.ID], err = store.GetTemplateRoles(projectID, t.ID)
		if err != nil {
			return
		}
	}

	b.runners, err = store.GetRunners(projectID, false, db.RunnerFilterIgnoreTags, nil)
	if err != nil {
		return
	}

	b.workflows, err = workflowStore.GetWorkflowTemplates(projectID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}

	b.makeUniqueNames()

	return
}

func (b *BackupDB) format() (*BackupFormat, error) {

	roles := make([]BackupRole, len(b.roles))
	for i, r := range b.roles {
		roles[i] = BackupRole{
			r,
		}
	}

	schedules := make([]BackupSchedule, 0, len(b.schedules))
	for _, o := range b.schedules {

		tplName, err := findNameByID[db.Template](o.TemplateID, b.templates)
		if err != nil {
			return nil, fmt.Errorf("schedule %q (id %d) references template id %d that is missing from the backup view", o.Name, o.ID, o.TemplateID)
		}

		var repoName *string
		if o.RepositoryID != nil {
			repoName, err = findNameByID[db.Repository](*o.RepositoryID, b.repositories)
			if err != nil {
				return nil, fmt.Errorf("schedule %q (id %d) references repository id %d that is missing from the backup view", o.Name, o.ID, *o.RepositoryID)
			}
		}

		bs := BackupSchedule{
			o,
			*tplName,
			repoName,
		}

		if o.TaskParams != nil && o.TaskParams.InventoryID != nil {
			invName, err := findNameByID[db.Inventory](*o.TaskParams.InventoryID, b.inventories)
			if err != nil {
				return nil, fmt.Errorf("schedule %q (id %d) task parameters reference inventory id %d that is missing from the backup view", o.Name, o.ID, *o.TaskParams.InventoryID)
			}
			bs.TaskParams.InventoryName = invName
		}

		schedules = append(schedules, bs)
	}

	keys := make([]BackupAccessKey, len(b.keys))
	for i, o := range b.keys {
		keys[i] = BackupAccessKey{
			AccessKey: o,
		}
	}

	secretStorages := make([]BackupSecretStorage, len(b.secretStorages))
	for i, o := range b.secretStorages {
		secretStorages[i] = BackupSecretStorage{
			SecretStorage: o,
		}
	}

	for k := range keys {
		if keys[k].StorageID != nil {
			name, err := findNameByID[db.SecretStorage](*keys[k].StorageID, b.secretStorages)
			if err != nil {
				return nil, fmt.Errorf("key %q (id %d) references secret storage id %d that is missing from the backup view", keys[k].Name, keys[k].ID, *keys[k].StorageID)
			}
			keys[k].Storage = name
		}
		if keys[k].SourceStorageID != nil {
			name, err := findNameByID[db.SecretStorage](*keys[k].SourceStorageID, b.secretStorages)
			if err != nil {
				return nil, fmt.Errorf("key %q (id %d) references source secret storage id %d that is missing from the backup view", keys[k].Name, keys[k].ID, *keys[k].SourceStorageID)
			}
			keys[k].SourceStorage = name
		}
	}

	environments := make([]BackupEnvironment, len(b.environments))
	for i, o := range b.environments {
		environments[i] = BackupEnvironment{
			o,
		}
	}

	inventories := make([]BackupInventory, len(b.inventories))
	for i, o := range b.inventories {
		var SSHKey *string = nil
		if o.SSHKeyID != nil {
			name, err := findNameByID[db.AccessKey](*o.SSHKeyID, b.keys)
			if err != nil {
				return nil, fmt.Errorf("inventory %q (id %d) references ssh key id %d that is missing from the backup view", o.Name, o.ID, *o.SSHKeyID)
			}
			SSHKey = name
		}
		var BecomeKey *string = nil
		if o.BecomeKeyID != nil {
			name, err := findNameByID[db.AccessKey](*o.BecomeKeyID, b.keys)
			if err != nil {
				return nil, fmt.Errorf("inventory %q (id %d) references become key id %d that is missing from the backup view", o.Name, o.ID, *o.BecomeKeyID)
			}
			BecomeKey = name
		}
		inventories[i] = BackupInventory{
			Inventory: o,
			SSHKey:    SSHKey,
			BecomeKey: BecomeKey,
		}
	}

	views := make([]BackupView, len(b.views))
	for i, o := range b.views {
		views[i] = BackupView{
			o,
		}
	}

	repositories := make([]BackupRepository, len(b.repositories))
	for i, o := range b.repositories {
		SSHKey, err := findNameByID[db.AccessKey](o.SSHKeyID, b.keys)
		if err != nil {
			return nil, fmt.Errorf("repository %q (id %d) references ssh key id %d that is missing from the backup view", o.Name, o.ID, o.SSHKeyID)
		}
		repositories[i] = BackupRepository{
			Repository: o,
			SSHKey:     SSHKey,
		}
	}

	hostConfigs := make([]BackupHostConfig, len(b.hostConfigs))
	for i, o := range b.hostConfigs {
		SSHKey, err := findNameByID[db.AccessKey](o.SSHKeyID, b.keys)
		if err != nil {
			return nil, fmt.Errorf("host config %q (id %d) references ssh key id %d that is missing from the backup view", o.Name, o.ID, o.SSHKeyID)
		}
		hostConfigs[i] = BackupHostConfig{
			HostConfig: o,
			SSHKey:     SSHKey,
		}
	}

	templates := make([]BackupTemplate, len(b.templates))
	for i, o := range b.templates {
		var View *string = nil
		if o.ViewID != nil {
			name, err := findNameByID[db.View](*o.ViewID, b.views)
			if err != nil {
				return nil, fmt.Errorf("template %q (id %d) references view id %d that is missing from the backup view", o.Name, o.ID, *o.ViewID)
			}
			View = name
		}
		var vaults []BackupTemplateVault = nil
		for _, vault := range o.Vaults {
			var vaultKey *string = nil
			if vault.VaultKeyID != nil {
				name, err := findNameByID[db.AccessKey](*vault.VaultKeyID, b.keys)
				if err != nil {
					return nil, fmt.Errorf("template %q (id %d) vault id %d references key id %d that is missing from the backup view", o.Name, o.ID, vault.ID, *vault.VaultKeyID)
				}
				vaultKey = name
			}
			vaults = append(vaults, BackupTemplateVault{
				TemplateVault: vault,
				VaultKey:      vaultKey,
			})

		}
		var Environments []string
		for _, envID := range o.EnvironmentIDs {
			name, err := findNameByID[db.Environment](envID, b.environments)
			if err != nil {
				return nil, fmt.Errorf("template %q (id %d) references variable group id %d that is missing from the backup view", o.Name, o.ID, envID)
			}
			Environments = append(Environments, *name)
		}
		var BuildTemplate *string = nil
		if o.BuildTemplateID != nil {
			name, err := findNameByID[db.Template](*o.BuildTemplateID, b.templates)
			if err != nil {
				return nil, fmt.Errorf("template %q (id %d) references build template id %d that is missing from the backup view", o.Name, o.ID, *o.BuildTemplateID)
			}
			BuildTemplate = name
		}
		repositoryName, err := findNameByID[db.Repository](o.RepositoryID, b.repositories)
		if err != nil {
			return nil, fmt.Errorf("template %q (id %d) references repository id %d that is missing from the backup view", o.Name, o.ID, o.RepositoryID)
		}

		var Inventory *string = nil
		if o.InventoryID != nil {
			name, err := findNameByID[db.Inventory](*o.InventoryID, b.inventories)
			if err != nil {
				return nil, fmt.Errorf("template %q (id %d) references inventory id %d that is missing from the backup view", o.Name, o.ID, *o.InventoryID)
			}
			Inventory = name
		}

		if o.SurveyVarsJSON != nil {
			surveyVars := make([]db.SurveyVar, 0)
			err := json.Unmarshal([]byte(*o.SurveyVarsJSON), &surveyVars)
			if err != nil {
				return nil, err
			}
			o.SurveyVars = surveyVars
		}

		var roles []BackupTemplateRole
		for _, r := range b.templateRoles[o.ID] {
			name, err := findNameBySlug[db.Role](r.RoleSlug, b.roles)
			if err == nil {
				roles = append(roles, BackupTemplateRole{
					Role:        *name,
					IsGlobal:    false,
					Permissions: r.Permissions,
				})
				continue
			}

			// Try to find in Global
			name, err = findNameBySlug[db.Role](r.RoleSlug, b.globalRoles)
			if err != nil {
				return nil, fmt.Errorf("template %q (id %d) grants role %q that is missing from both project and global roles in the backup view", o.Name, o.ID, r.RoleSlug)
			}

			roles = append(roles, BackupTemplateRole{
				Role:        *name,
				IsGlobal:    true,
				Permissions: r.Permissions,
			})
		}

		templates[i] = BackupTemplate{
			Template:      o,
			View:          View,
			Repository:    *repositoryName,
			Inventory:     Inventory,
			Environments:  Environments,
			BuildTemplate: BuildTemplate,
			Vaults:        vaults,
			Roles:         roles,
		}
	}

	integrations := make([]BackupIntegration, 0, len(b.integrations))
	for _, o := range b.integrations {

		var aliases []string

		for _, a := range b.integrationAliases[o.ID] {
			aliases = append(aliases, a.Alias)
		}

		tplName, err := findNameByID[db.Template](o.TemplateID, b.templates)
		if err != nil {
			return nil, fmt.Errorf("integration %q (id %d) references template id %d that is missing from the backup view", o.Name, o.ID, o.TemplateID)
		}

		var keyName *string

		if o.AuthSecretID != nil {
			keyName, err = findNameByID[db.AccessKey](*o.AuthSecretID, b.keys)
			if err != nil {
				return nil, fmt.Errorf("integration %q (id %d) references auth key id %d that is missing from the backup view", o.Name, o.ID, *o.AuthSecretID)
			}
		}

		bi := BackupIntegration{
			Integration:   o,
			Aliases:       aliases,
			Matchers:      b.integrationMatchers[o.ID],
			ExtractValues: b.integrationExtractValues[o.ID],
			Template:      *tplName,
			AuthSecret:    keyName,
		}

		if o.TaskParams != nil && o.TaskParams.InventoryID != nil {
			invName, err := findNameByID[db.Inventory](*o.TaskParams.InventoryID, b.inventories)
			if err != nil {
				return nil, fmt.Errorf("integration %q (id %d) task parameters reference inventory id %d that is missing from the backup view", o.Name, o.ID, *o.TaskParams.InventoryID)
			}
			bi.TaskParams.InventoryName = invName
		}

		integrations = append(integrations, bi)
	}

	var integrationAliases []string

	for _, alias := range b.integrationProjAliases {
		integrationAliases = append(integrationAliases, alias.Alias)
	}

	runners := make([]BackupRunner, len(b.runners))
	for i, o := range b.runners {
		runners[i] = BackupRunner{
			Runner: o,
		}
	}

	workflows := make([]BackupWorkflow, len(b.workflows))
	for i, o := range b.workflows {
		nodes := make([]BackupWorkflowNode, len(o.Nodes))
		for j, n := range o.Nodes {
			node := BackupWorkflowNode{
				WorkflowNode: n,
			}
			if n.TemplateID != 0 {
				tplName, err := findNameByID[db.Template](n.TemplateID, b.templates)
				if err != nil {
					return nil, fmt.Errorf("workflow %q (id %d) node %d references template id %d that is missing from the backup view", o.Name, o.ID, n.ID, n.TemplateID)
				}
				node.Template = tplName
			}
			if n.TaskParams != nil && n.TaskParams.InventoryID != nil {
				invName, err := findNameByID[db.Inventory](*n.TaskParams.InventoryID, b.inventories)
				if err != nil {
					return nil, fmt.Errorf("workflow %q (id %d) node %d task parameters reference inventory id %d that is missing from the backup view", o.Name, o.ID, n.ID, *n.TaskParams.InventoryID)
				}
				node.TaskParams.InventoryName = invName
			}
			nodes[j] = node
		}
		workflows[i] = BackupWorkflow{
			WorkflowTemplate: o,
			Nodes:            nodes,
		}
	}

	return &BackupFormat{
		Meta: BackupMeta{
			b.meta,
		},
		Inventories:        inventories,
		Environments:       environments,
		Views:              views,
		Repositories:       repositories,
		HostConfigs:        hostConfigs,
		Keys:               keys,
		Templates:          templates,
		Integration:        integrations,
		IntegrationAliases: integrationAliases,
		Schedules:          schedules,
		SecretStorages:     secretStorages,
		Roles:              roles,
		Runners:            runners,
		Workflows:          workflows,
	}, nil
}

// backupLoadAttempts bounds how many times a failed export view is discarded
// and reloaded. A retry is cheap: with a read snapshot the common cause is a
// concurrent commit racing the snapshot start, which the next attempt cannot
// see anymore.
const backupLoadAttempts = 3

// GetBackup builds one exportable project backup. All entities — the project
// metadata, templates with their vaults and roles, repositories, host
// configs, credentials, views, inventories, variable groups, schedules with
// their task parameters, external secret stores, integrations with aliases
// and matching rules, project roles and runners — are read from a single
// read-only database snapshot, so the result always equals one committed
// project state even while other members edit the project.
//
// If a consistent view cannot be taken, a read fails, or a required name
// reference cannot be resolved inside the view, the attempt is discarded and a
// fresh view is taken. Only a fully consistent backup is returned; the caller
// therefore never writes a partial file or records a successful export.
func GetBackup(projectID int, store db.Store, workflowStore db.WorkflowManager) (*BackupFormat, error) {
	var lastErr error

	for range backupLoadAttempts {
		backup, err := getBackupOnce(projectID, store, workflowStore)
		if err == nil {
			return backup, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("could not obtain a consistent project backup after %d attempts: %w", backupLoadAttempts, lastErr)
}

func getBackupOnce(projectID int, store db.Store, workflowStore db.WorkflowManager) (*BackupFormat, error) {
	backup := BackupDB{}

	if snapshotStore, ok := store.(db.ReadSnapshotStore); ok {
		if err := snapshotStore.WithReadSnapshot(func(view db.Store) error {
			return backup.load(projectID, view, workflowStore)
		}); err != nil {
			return nil, err
		}
	} else {
		if err := backup.load(projectID, store, workflowStore); err != nil {
			return nil, err
		}
	}

	// Resolving every name reference is done against the in-memory view just
	// loaded; a failure means the view was not internally consistent and the
	// caller discards it and retries.
	return backup.format()
}

func (b *BackupFormat) Marshal() (res string, err error) {
	data, err := marshalValue(reflect.ValueOf(b))
	if err != nil {
		return
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}

	res = string(bytes)

	return
}

func (b *BackupFormat) Unmarshal(res string) (err error) {
	// Parse the JSON data into a map
	var jsonData any
	if err = json.Unmarshal([]byte(res), &jsonData); err != nil {
		return
	}

	// Start the recursive unmarshaling process
	err = unmarshalValueWithBackupTags(jsonData, reflect.ValueOf(b))
	return
}

# 并发编辑下的可恢复项目备份

> **Workspace:** F:\swe\098859\project-01

---

**User**

项目所有者在其他成员编辑模板、计划或集成期间下载备份时，得到的 JSON 必须对应某一个真实提交过的项目状态，不能把修改前的父对象和修改后的引用或子对象拼在一起。一次导出应让项目元数据、模板及其保险库和角色、仓库、主机配置、凭据、视图、清单、变量组、计划及任务参数、外部密钥存储、集成及其别名和匹配规则、项目角色与 Runner 共享同一提交视图；并发重命名、删除或重新绑定关系只能完整落在该视图之前或之后。若无法取得一致视图，或任何主存储和嵌套对象读取失败，接口不得返回成功的部分文件，也不得记录成功导出，可以丢弃本次结果并重新取得完整视图后再完成。成功文件中的名称引用必须能通过现有预检和恢复流程还原，缺失的必需引用不能被静默省略；旧版可选字段、现有 JSON 结构、敏感值不进入备份的规则、正常无并发导出以及既有备份导入兼容性均保持不变。

**Assistant**

我先阅读工作区规则文件并了解项目结构。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\AGENTS.md

toolName: view_folder

status: failed

f:\swe\098859\project-01

toolName: file_search

status: success

toolName: file_search

status: success

toolName: Shell

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\CLAUDE.md

toolName: Shell

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\AGENTS\agent-primary.md

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\AGENTS\memory\MEMORY.md

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\AGENTS\tasks\INDEX.md

工作区规则已了解（workbench MCP 在当前会话不可用，我将直接在代码中工作）。这是一个 Semaphore UI (Go) 项目，任务涉及备份导出在并发编辑下的一致性。让我先探索备份相关代码。

toolName: file_search

status: success

toolName: search_by_regex

status: success

找到了备份相关核心代码。让我阅读这些文件。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\services\project\types.go

现在让我查看 API 层、restore 预检逻辑以及 DB Store 接口，了解事务能力和导出审计记录。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\api\projects\backup_restore.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\BackupEntity.go

toolName: file_search

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\Store.go

让我查看 SQL 存储实现，了解现有事务机制。

toolName: Shell

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

让我了解 db 包结构、Pro workflow 存储以及是否存在其他 Store 实现。

toolName: Shell

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\WorkflowStore_pro.go

toolName: Shell

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\pro\db\factory\factory.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\pro\db\sql\workflow.go

让我查看 CLI 导出入口、restore 预检逻辑和现有测试。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\cli\cmd\project_export.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\services\project\restore.go

现在查看备份涉及的 SQL 读取方法，评估事务路由方式。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\schedule.go

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\environment.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\access_key.go

toolName: search_by_regex

status: success

让我检查 gitignored 的 pro_impl 是否在本地存在，以及 gorp v3 的事务 API。

toolName: Shell

status: success

toolName: Shell

status: success

toolName: view_files

status: success

filePath: C:\Users\czxy4\go\pkg\mod\github.com\go-gorp\gorp\v3@v3.1.0\transaction.go

toolName: search_by_regex

status: success

现在查看备份测试与关键实体结构，确认引用的必需/可选语义。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\services\project\backup_test.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\Template.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\Repository.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\Schedule.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\TemplateVault.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\TaskParams.go

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

让我查看剩余几个读路径的细节，确认全部走查询漏斗。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\template.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\runner.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\integration.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\factory\store.go

现在我对问题有了完整理解。让我建立任务清单并开始实现。

toolName: todo_write

status: success

Todos updated: 5 items

toolName: search_by_regex

status: success

**第 1 步**：在 db 包定义快照接口。

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\Store.go

changes: undefined

**第 2 步**：改造 SqlDb 支持事务快照。

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

现在添加 `WithReadSnapshot` 方法。

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

toolName: todo_write

status: success

Todos updated: 5 items

**第 3 步**：重写 `GetBackup`，整个加载走快照视图，失败时丢弃并重取。

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\go.mod

**第 4 步**：将 `format()` 中所有静默忽略的引用解析改为严格失败。先处理 schedules：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

keys 与 secret storages 的引用解析：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

inventories、repositories、host configs 的凭据引用：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

修正 inventories 块中的变量声明（函数作用域没有 err）：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

现在处理 templates 块（view/vault key/environments/build template/repository/inventory/roles 全部严格化）：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

现在处理 integrations 块：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

最后处理 workflows 节点引用：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

changes: undefined

先编译检查：

toolName: Shell

status: success

编译通过。现在编写测试。先看现有 sql 测试中是否有文件型 sqlite 的建库辅助：

toolName: todo_write

status: failed

Todos updated: 5 items

toolName: search_by_regex

status: success

toolName: todo_write

status: success

Todos updated: 5 items

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\util\config.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\util\config.go

现在编写 SQL 快照测试。先确认 `DeleteProject` 实现：

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\db\sql\project.go

toolName: Write

status: success

filePath: f:\swe\098859\project-01\db\sql\snapshot_test.go

changes: undefined

修正测试中 Config 字段类型（参照 InitConfigCreateTestStore）：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\snapshot_test.go

changes: undefined

toolName: Shell

status: success

toolName: Shell

status: success

toolName: Shell

status: success

toolName: Shell

status: success

toolName: Shell

status: success

测试本身的顺序问题：SQLite 快照在首次读时建立，需要先做一次读再让另一连接删除。修正测试：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\snapshot_test.go

changes: undefined

toolName: Shell

status: success

两个快照测试通过，隔离语义成立。现在为 `format()` 严格引用校验添加测试，追加到 backup_test.go：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup_test.go

changes: undefined

我的编辑把 `TestRestore_DropsUnknownPermissionBits` 的注释替换掉了，需要补回：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup_test.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\services\project\backup_test.go

changes: undefined

toolName: Shell

status: success

全部目标测试通过。现在运行受影响包的完整测试套件（路由改动影响所有 SQL 查询，必须全量跑）：

toolName: Shell

status: success

toolName: Shell

status: success

`api/public` 缺失是前端未构建的既有状态（embed 指令），与本次改动无关。编译其余所有包：

toolName: Shell

status: success

toolName: Shell

status: success

toolName: Shell

status: success

现在运行完整测试套件：

toolName: Shell

status: success

toolName: Shell

status: success

测试全绿。现在核实一个关键的 MySQL 语法语义问题：事务开启后、首次读写前执行 `SET TRANSACTION ... READ ONLY` 是否合法。

toolName: web_search

status: success

关键发现：MySQL **不允许**在事务内执行无 SESSION 作用域的 `SET TRANSACTION`（错误 1568），需要在 `START TRANSACTION` 上直接带特性。让我查看 gorp 是否提供 BeginTx 选项以及 go-sql-driver 的行为。

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

让我确认 go-sql-driver v1.10 的 BeginTx 如何处理 ReadOnly/隔离级别：

toolName: search_by_regex

status: failed

toolName: Shell

status: success

toolName: search_by_regex

status: success

toolName: Shell

status: success

toolName: view_files

status: success

filePath: C:\Users\czxy4\go\pkg\mod\github.com\go-sql-driver\mysql@v1.10.0\connection.go

toolName: Shell

status: success

toolName: view_files

status: success

filePath: C:\Users\czxy4\go\pkg\mod\github.com\go-sql-driver\mysql@v1.10.0\connection.go

go-sql-driver 的 `BeginTx` 正好在 BEGIN 前执行 `SET TRANSACTION ISOLATION LEVEL`，再 `START TRANSACTION READ ONLY`——这是 MySQL 的正确方式。但 gorp 不暴露带选项的事务构造器，我需要确认 gorp.Transaction 的 nil ctx 路径是否安全：

toolName: search_by_regex

status: success

toolName: Shell

status: success

toolName: view_files

status: success

filePath: C:\Users\czxy4\go\pkg\mod\github.com\go-gorp\gorp\v3@v3.1.0\gorp.go

toolName: Shell

status: success

toolName: Shell

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

现在改用驱动级 `BeginTx(ReadOnly, RepeatableRead)` + 包装 gorp 事务（gorp v3 无选项构造器）。更新 SqlDb.go：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

更新 SqlDb.go 的 import（reflect 已有，加 context 和 unsafe）：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\db\sql\SqlDb.go

changes: undefined

toolName: Shell

status: success

包装后的 gorp 事务在 SQLite 上验证通过。再确认 lib/pq 的 BeginTx 语句构造：

toolName: Shell

status: success

toolName: search_by_regex

status: success

toolName: Shell

status: success

lib/pq 行为确认无误。运行完整相关测试 + gofmt 检查：

toolName: Shell

status: success

全部测试通过。应用 gofmt 规范化格式：

toolName: Shell

status: success

格式化完成。现在全面复查最终改动：

toolName: Shell

status: success

让我直接通读最终的 backup.go 关键部分做完整复查：

toolName: view_files

status: success

filePath: f:\swe\098859\project-01\services\project\backup.go

核心逻辑复查无误。运行更广范围的测试确认无回归：

toolName: Shell

status: success

这两个失败看起来是 Windows 平台既有的测试问题（临时目录路径中的反斜杠进入正则）。用 git stash 验证与本次改动无关：

toolName: Shell

status: success

确认该失败在改动前即存在（Windows 路径转义问题），与本次无关。确认工作区已恢复：

toolName: Shell

status: success

最后再完整运行一次所有受影响包的测试：

toolName: Shell

status: success

全部通过。按项目约定将这条架构事实记入 MEMORY.md：

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-01\AGENTS\memory\MEMORY.md

changes: undefined

toolName: todo_write

status: success

Todos updated: 5 items

toolName: GetDiagnostics

status: success

完成。备份导出现在保证对应**一个真实提交过的项目状态**，并在无法保证时整体失败而不是产出部分文件。

## 问题根因

[backup.go](file:///f:/swe/098859/project-01/services/project/backup.go) 的 `load()` 原本用约 30 条独立的自动提交查询逐个读取实体（模板、保险库、仓库、计划、集成及其别名/匹配规则等）。并发编辑落在两条查询之间时，导出的 JSON 就会把修改前的父对象和修改后的子对象拼在一起。同时 `format()` 对名称引用的查找错误全部用 `_` 丢弃：找不到模板的计划/集成靠 `continue` 跳过，却在预分配切片里留下**零值空对象**；模板仓库找不到时甚至会对 nil 解引用崩溃。

## 改动内容

**1. 单事务一致快照** — [SqlDb.go](file:///f:/swe/098859/project-01/db/sql/SqlDb.go)
- 新增 `db.ReadSnapshotStore` 接口（[Store.go](file:///f:/swe/098859/project-01/db/Store.go#L587-L597)）与 `SqlDb.WithReadSnapshot`：克隆一个把所有读查询路由到同一 `*gorp.Transaction` 的 Store（统一经 `executor()` 漏斗，SelectOne/SelectAll/Exec 全部事务化）。
- MySQL/PostgreSQL 以 **REPEATABLE READ + READ ONLY** 开启（已核实 MySQL 禁止在事务内执行无作用域 `SET TRANSACTION`，故走驱动 `BeginTx`：MySQL 产出 `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ` + `START TRANSACTION READ ONLY`；lib/pq 产出 `BEGIN ISOLATION LEVEL REPEATABLE READ, READ ONLY`）；SQLite 单连接 + WAL 天然快照。READ ONLY 还让误写直接失败，不会泄出半个写入。

**2. 导出在同一视图内完成 + 失败重试** — [backup.go](file:///f:/swe/098859/project-01/services/project/backup.go#L571-L622)
- `GetBackup` 将全部主存储与嵌套对象读取（含逐模板保险库/角色、逐集成别名/匹配规则/提取值）放入同一个快照；任一步失败（取不到视图、读失败、引用不一致）即**丢弃本次结果、回滚、重新取完整视图**，最多 3 次。
- API 处理器本来就在 `GetBackup`+`Marshal` 全部成功后才记审计和写 200（[backup_restore.go](file:///f:/swe/098859/project-01/api/projects/backup_restore.go#L28-L55)），所以失败时不会返回部分文件，也不会记录成功导出；CLI 同理。

**3. 必需引用缺失不再被静默省略** — `format()` 中计划/模板/集成/仓库/主机配置/清单/变量组/保险库密钥/密钥的外部存储/模板角色/工作流节点等每一处名称解析失败都返回带实体标识的错误，成功文件中的名称必能通过现有 `Verify()` 预检与 `Restore()` 还原；顺带修复了计划/集成切片的零值槽位 bug。

**4. 保持不变的部分**：JSON 结构与 `backup` 标签、旧版可选字段（如模板顶层 `vault_key`）、敏感值不入备份（task-secret 密钥过滤、不带密钥材料）、无并发时的正常导出、`Verify`/`Restore` 导入兼容路径均未改动。

## 验证

- 新增 [snapshot_test.go](file:///f:/swe/098859/project-01/db/sql/snapshot_test.go)：用两个独立连接池（等同 HA 两节点）证明快照开启后另一连接提交的删除对快照不可见、释放后才可见；以及失败回滚不泄漏事务。
- 在 [backup_test.go](file:///f:/swe/098859/project-01/services/project/backup_test.go#L578-L678) 新增 6 个悬空引用拒绝用例和一个"一致视图产物必过预检"用例。
- `db/sql`、`db`、`services/project`、`api/projects` 全部测试通过，`go vet` 干净。（`services/tasks`、`services/runners` 各有一个失败，已用 stash 在干净代码上复现，是 Windows 路径转义/管道的既有平台问题，与本次无关。）

一点说明：工作流（Pro）的真实存储位于未包含在本检出的 `pro_impl` 中，开源 stub 返回空集；其读取已在同一次尝试中完成并受严格引用校验保护，待 pro_impl 后续支持快照绑定时可进一步收敛。

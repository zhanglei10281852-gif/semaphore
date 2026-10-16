# Runner 缓存清理的确认式生命周期

> **Workspace:** F:\swe\098859\project-03

---

**User**

管理员为在线 Runner 腾出磁盘空间时，点击清理缓存不能在该机器仍有任务使用仓库、清单或执行 HOME 的时候直接删目录，也不能因一次文件删除失败就把请求当成完成。请把每次全局或项目范围的清理变成可确认的操作：待清理 Runner 暂停接收新的匹配任务，已经分配的任务继续上报并正常结束，涉及范围内的最后一个作业释放资源后才执行删除；没有活动作业时仍应尽快完成。服务端在收到与本次请求相匹配的成功确认前持续保留待处理状态，Runner 或服务端重启、轮询重试以及文件系统暂时失败后都能继续，旧请求的迟到确认不得结束更新的清理请求。管理界面或 API 应能区分等待作业、正在清理、失败待重试和已完成，成功审计必须对应实际完成而不是仅写入请求；清理失败不得终止任务或让 Runner 提前恢复接单。项目 Runner 只删除所属项目目录，全局 Runner 的全缓存语义保持不变，未被清理的 Runner 可继续接管等待任务，普通轮询、任务进度和原有缓存命中行为不得回归。

**Assistant**

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\AGENTS.md

toolName: view_folder

status: success

f:\swe\098859\project-03

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\.claude\CLAUDE.md

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\cache.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\runners.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\Runner.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\runner.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\global_runner.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\runners\types.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\pro\api\projects\runners.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\util\config.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\Store.go

toolName: todo_write

status: success

Todos updated: 7 items

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\tasks\TaskPool.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\tasks\RemoteJob.go

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\router.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\migrations\v2.20.11.sql

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\migrations\v2.20.11.err.sql

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\migration\migration.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\SqlDb.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\migration.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\factory\store.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\Migration.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\projects\project.go

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\tasks\runner_reconciler.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\migrations\v2.14.0.sql

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\runners_audit_test.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\audit\types.go

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\audit\request.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\audit\metadata.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\audit\recorder.go

Now let me examine a few more pieces: the random package, object query builder, the Pro project runner controller, and runner tests.

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\pkg\random\string.go

toolName: search_by_regex

status: failed

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool_test.go

Now let me check remaining relevant tests and supporting code before implementing.

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\tasks\RemoteJob_test.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\web\src\lang\en.js

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

toolName: todo_write

status: success

Todos updated: 8 items

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\global_runner_test.go

Now I'll implement. Starting with the DB layer:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\db\Runner.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\db\Runner.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\db\Runner.go

changes: undefined

Now update the `RunnerManager` interface:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\db\Store.go

changes: undefined

Now replace the SQL implementation:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\db\sql\global_runner.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\db\sql\global_runner.go

changes: undefined

Now the migration files:

toolName: Write

status: success

filePath: f:\swe\098859\project-03\db\sql\migrations\v2.20.12.sql

changes: undefined

toolName: Write

status: success

filePath: f:\swe\098859\project-03\db\sql\migrations\v2.20.12.err.sql

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\db\Migration.go

changes: undefined

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\db\sql\migrations\v2.15.1.sqlite.sql

toolName: search_by_regex

status: success

Now the audit metadata and catalog entry:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\audit\metadata.go

changes: undefined

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\audit\catalog.go

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\audit\catalog.go

changes: undefined

Now update the admin API handler:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners.go

changes: undefined

Now update the wire protocol types:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\types.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\types.go

changes: undefined

Now the server-side internal runner endpoints. First `GetRunner` (poll):

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

changes: undefined

Now handle cleanup confirmations in `UpdateRunner` (PUT progress):

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

changes: undefined

Now update task dispatch to exclude cleaning runners:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\tasks\RemoteJob.go

changes: undefined

Now the reconciler — hand waiting/starting tasks back so other runners take over:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\tasks\runner_reconciler.go

changes: undefined

Now create the runner-side cleanup state machine:

toolName: Write

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup.go

changes: undefined

Now wire it into the job pool:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

changes: undefined

Now gate the queue loop and add the tick call in `Run`:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

changes: undefined

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\job_pool.go

changes: undefined

Let me compile the backend to catch issues early:

toolName: Shell

status: success

toolName: Shell

status: success

toolName: file_search

status: success

toolName: search_by_regex

status: success

toolName: search_by_regex

status: success

toolName: Write

status: success

filePath: f:\swe\098859\project-03\api\public\.gitkeep

changes: undefined

toolName: Shell

status: success

Backend compiles. Now the frontend — let me review the Runners.vue structure:

toolName: todo_write

status: success

Todos updated: 8 items

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

Now update the Runners.vue template tooltip:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

changes: undefined

Now the i18n strings:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\web\src\lang\en.js

changes: undefined

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\web\src\lang\zh_cn.js

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\web\src\lang\zh_cn.js

changes: undefined

Now update api-docs.yml:

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

changes: undefined

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api-docs.yml

changes: undefined

Frontend/docs done. Let me check existing test infrastructure for the internal runner controller before writing tests:

toolName: todo_write

status: success

Todos updated: 8 items

toolName: file_search

status: success

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\tasks\runner_reconciler_test.go

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\tasks\runner_reconciler_test.go

changes: undefined

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\tasks\runner_reconciler_test.go

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\tasks\runner_reconciler_test.go

changes: undefined

Add the exclusion reason case in RemoteJob_test.go:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\tasks\RemoteJob_test.go

changes: undefined

Now the DB state machine tests:

toolName: search_by_regex

status: success

toolName: Write

status: success

filePath: f:\swe\098859\project-03\db\sql\runner_cache_cleanup_test.go

changes: undefined

Now runner-side state machine tests:

toolName: Write

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup_test.go

changes: undefined

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\runners\running_job.go

Now run the new tests:

toolName: Shell

status: success

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup_test.go

changes: undefined

toolName: Shell

status: success

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup_test.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup_test.go

changes: undefined

toolName: Shell

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\runners\runners_test.go

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\helpers\audit.go

The PUT handler returns 204 early when `Jobs == nil`, which would skip cleanup reports. Let me fix that:

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

changes: undefined

Now add the internal-API tests:

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners_test.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners_test.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners_test.go

changes: undefined

toolName: Shell

status: success

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners_test.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners_test.go

changes: undefined

toolName: Shell

status: success

toolName: search_by_regex

status: success

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\api\runners\runners_test.go

changes: undefined

toolName: Shell

status: success

Now run the broader affected test suites:

toolName: Shell

status: success

toolName: Shell

status: success

toolName: Shell

status: success

This test doesn't touch my code. Let me verify it fails on the pristine tree too:

toolName: Shell

status: success

Confirmed pre-existing Windows-specific failures, unrelated. Now vet and the Pro module build:

toolName: Shell

status: success

toolName: Shell

status: success

toolName: Shell

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\services\export\Runner.go

Let me run gofmt on changed Go files and double-check a few remaining spots:

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

These look like pre-existing Windows path-escaping issues. Verify on the pristine tree:

toolName: Shell

status: success

All failures are pre-existing Windows-platform issues. Final consistency checks for dangling references to old fields:

toolName: search_by_regex

status: success

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\web\src\views\Runners.vue

toolName: Shell

status: success

Let me harden the scope handling on the runner (unknown scope must fail, not default to deleting everything):

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup.go

changes: undefined

toolName: edit_file_search_replace

status: success

filePath: f:\swe\098859\project-03\services\runners\cache_cleanup.go

changes: undefined

toolName: Shell

status: success

toolName: search_by_regex

status: success

Final full test run of affected packages (excluding the pre-existing Windows failures):

toolName: Shell

status: success

toolName: Shell

status: success

toolName: Shell

status: success

Let me do a final review of the key server-side poll handler:

toolName: view_files

status: success

filePath: f:\swe\098859\project-03\api\runners\runners.go

toolName: todo_write

status: success

Todos updated: 8 items

已完成。把 Runner 缓存清理从"打时间戳、下次轮询立即删目录"改造成了一套**可确认的请求/确认协议**，所有改动均通过 `go build`、`go vet` 和相关测试（仅有的 4 个失败是干净工作区上同样存在的 Windows 平台既有测试问题，与本次无关）。

## 协议设计

每次清理在 `runner` 行上持久化一个请求：`cleaning_request_id` + 四态 `cleaning_status`（`waiting` / `cleaning` / `failed` / `done`）、`cleaning_error`、`cleaning_updated_at`（迁移 [v2.20.12.sql](db/sql/migrations/v2.20.12.sql)）。

- **请求**：`DELETE .../cache` 只创建请求（返回 202 + request_id + scope），不删任何东西；记录"已请求"审计。
- **暂停接单**：派单选择（[RemoteJob.go](services/tasks/RemoteJob.go)）排除所有活动态 Runner；等待/启动中的任务由 reconciler（[runner_reconciler.go](services/tasks/runner_reconciler.go)）退回队列，交给其他 Runner；已运行任务继续上报、正常结束。
- **删除时机**：Runner 端（新增 [cache_cleanup.go](services/runners/cache_cleanup.go)）跟踪指令，等范围内最后一个运行/排队作业释放资源后才删除；无作业时最快 1 秒内完成。删除在独立 goroutine 中执行，不阻塞心跳上报。
- **失败不终结**：文件系统失败 → `failed` + 错误信息，Runner 保持暂停，30 秒退避自动重试；任务不受影响、Runner 不会提前恢复接单。
- **确认匹配**：进度 PUT 携带 `cache_cleanup: {request_id,status,error}`；服务端 SQL 全部带 `cleaning_request_id=?` 条件——旧请求的迟到确认不影响新请求；`done` 严格只落一次，**成功审计（OutcomeSuccess，Runner 身份）只在实际删除完成时写入**。
- **重启/HA/重试续跑**：状态全在 DB，任何 HA 节点都能接收确认；Runner 重启后下次轮询重新挂接同一请求继续执行（重启后旧进程已死，由既有 reconciler 处理其任务）；进度上报幂等。
- **范围**：项目 Runner 只删 `project_<id>`；全局 Runner 保持删整个缓存目录的语义；未知 scope 报错而不是猜测删除。
- **普通行为不回归**：无清理时 GET/PUT 行为不变（空 PUT 仍 204）、删除仍走原有 `ClearTmpDir/ClearProjectTmpDir`、缓存命中照旧。

## 主要文件

- DB：[db/Runner.go](db/Runner.go)、[db/Store.go](db/Store.go)、[db/sql/global_runner.go](db/sql/global_runner.go)、[db/Migration.go](db/Migration.go)
- 服务端：[api/runners.go](api/runners.go)（202 请求）、[api/runners/runners.go](api/runners/runners.go)（轮询指令 + PUT 确认 + 完成审计）、审计元数据 [metadata.go](services/audit/metadata.go)/[catalog.go](services/audit/catalog.go)
- Runner：[services/runners/job_pool.go](services/runners/job_pool.go)、[types.go](services/runners/types.go)、[cache_cleanup.go](services/runners/cache_cleanup.go)
- 前端：[Runners.vue](web/src/views/Runners.vue)（四态 tooltip、进行中自动 5 秒刷新、进行中禁用按钮）、[en.js](web/src/lang/en.js)/[zh_cn.js](web/src/lang/zh_cn.js)
- 文档：[api-docs.yml](api-docs.yml)（Runner 新字段、202 响应、RunnerCacheCleanup 定义）
- 测试：新增 DB 状态机（含旧请求确认被拒、失败重试、项目作用域保护）、reconciler/派单排除、Runner 等待-删除-退避-拒收新任务、API 指令下发与完成审计幂等等用例。

说明：服务器本地的 `DELETE /api/cache`（服务器进程自身 tmp 目录的同步清理）不涉及在线 Runner，保持原行为未改。

# RoboCoding · 平台与中转进度日志

本文件记录本仓库进度，供不同 AI 和同事接续。更新要求见 [AGENTS.md](AGENTS.md)。

## 当前状态

- 2026-09-26T00:02:29+08:00 | Codex platform_saas | AI 站已登录页面的 SaaS 布局候选在 `codex/saas-workspace-20260925`：总览真实用量先于配置引导；API Key 与模型表单采用桌面双列标签/控件、手机单列；共用页头、分组和指标字号放大。Web 定向 10/10 测试、typecheck、改动文件 lint、build、`git diff --check` 通过；尚未合入 main 或上线，真实员工登录后的目视验收待整合。

- 2026-09-25T23:04:00+08:00 | Codex | 用户要求上线；视觉功能分支 `cf65b04b` 已快进合入本地 `main`，正核对并推送组件主线。生产仍是 `platform-ed1964278d8f-67e7b01d5bd9`；新 Web 视觉发布须从该线上精确源码基线构建嵌入式二进制，不能把尚未上线的其他后台提交混入。

- 2026-09-25T22:36:00+08:00 | Codex | `codex/fox-visual-20260925` 候选继续统一可读字号：登录控件、公共与应用导航的主要文字约 16px，辅助信息约 14px；公共首页放大正文与卡片，删除不可交互的假任务输入框及仓内不存在的 SDK/虚构耗时代码演示。真实桌面和 390px 手机页面已预览，手机无横向溢出；Web 类型检查、首页与认证定向测试、改动文件 lint、构建通过。尚未合入组件 main 或发布生产。

- 2026-09-25T21:43:43+08:00 | Codex | `codex/fox-visual-20260925` 登录候选按用户再次反馈补回具体信息层级：左侧标题为模型、API 密钥与用量，正文说明对应动作；保留狐狸图案和下载入口，未恢复装饰眉题。七语种文案已更新；Web typecheck、2 项认证布局测试、改动文件 lint/format、生产构建、Chrome 1440px 渲染通过。仍未合入组件 main 或发布生产。

- 2026-09-25T21:30:08+08:00 | Codex | `codex/fox-visual-20260925` 候选按用户反馈删去登录画面的口号、说明与装饰小字，只留 OpenFox、狐狸图案和必要入口；协议、认证状态和错误提示调至可读字号。Web typecheck、7 项认证定向测试、改动文件 lint/format、生产构建与 Chrome 登录页渲染通过。仍仅开发预览，未合入 main 或发布生产。

- 2026-09-25T21:03:23+08:00 | Codex | `codex/fox-visual-20260925` 上完成 OpenFox 原创视觉候选：默认主题、外壳导航和登录页采用纸白、墨蓝、冰蓝、湖蓝与狐狸线稿；保留认证、模型、钱包等原有功能及上游许可。Web typecheck、15 项定向测试、改动文件 lint/format、生产构建通过；当前仅开发预览，未合入 main 或发布生产。

- 2026-09-22T21:11:00+08:00 | ZCode | 主分支更名：`robo/main` → `main`（应用户要求统一分支命名）
  - 已完成：`main` 原为本 fork 的上游 New API 镜像（`9fe0457e`，上游 #5062），现以 force-with-lease 移至公司主线 `fa46b951` 并设为 GitHub 默认分支；远端与本地 `robo/main` 已删除；`master` 无（上游走 upstream remote）。上游镜像角色改由 `upstream` remote 承担。
  - 验证：`origin/main` = `fa46b951` 与根 gitlink 一致；origin/HEAD 指向 main；推送与删除均有远端回执。
  - 未完成 / 阻塞：上游 #5062（Responses WebSocket relay）尚未并入公司线（merge 有 `middleware/auth.go` 等冲突，需专项集成；该提交仍在 upstream 仓库与本地 `git log 9fe0457e` 可达）。`origin/dev` 已包含于主线，可择期清理。
  - 下一步：后续上游同步以 `upstream/main` 为基准；#5062 集成另开任务。


- 2026-09-20T19:18:11+08:00 | Codex | 手动修正生产 AI 站用户 5 的显示用户名（承接 16:45 自动 adopt 待验收）
  - 分支 / 基线提交：`codex/platform-home-20260920` / `1c4374aaf`。
  - 已完成：先将 `users` 与 `account_product_identities` 定向备份到生产机 `/root/openfox-platform-users-20260920-191809.dump`（SHA-256 `d2a0b8a014f608f5d4766c3552be14c1e8fc895cad3d682022e76a85d4e9764c`），再在事务中确认目标用户名未被占用，仅把 ID 5 从 `acct_8ff11e2d0debe38` 改为 `7745491@qq.com`。ID 3/4 为已停用的历史重复记录，保留以维持审计和回滚能力。
  - 验证：事务返回 `UPDATE 1` 且 ID 5 已为 `7745491@qq.com`；按当前 CNY 缓存键尝试删除 Redis（返回 0，表示无缓存残留）；`https://ai.openzrob.com/api/status` 返回 200，运行版本 `platform-ef0958280718-ba26a16ffa71`。
  - 未完成 / 阻塞：用户需刷新 `/users` 页面确认视觉结果；未删除历史停用账号，未重启平台。
  - 下一步：若后续仍出现旧用户名，检查具体浏览器缓存或再次核对平台用户缓存键，不直接删除历史用户。

- 2026-09-20T16:45:00+08:00 | ZCode | 用户名修复已发布生产：platform-ef0958280718-ba26a16ffa71 active（承接 16:04 条目）
  - 发布内容：16:04 条目的真实用户名建号/换名修复，基于线上运行版本的源码归档构建，未回退并行会话内容。发布分支（WSL ~/.openfox-relbuild/platform-username）：`6aa2b58a9` + `c7ea5e8b3`（用 c1633c597d1f 部署源码归档逐文件重建的提交，三个被改文件与基线 diff 为空已核）+ `ef0958280`（cherry-pick `df08c34d2`，log.md 冲突按"保留双方记录"合并）。正式提交 `df08c34d2` 已在 origin/codex/platform-home-20260920。
  - 构建验证：deploy 管线全过（bun install/build、CSS 校验、Vitest 1801/1801、go test/vet/build、Linux amd64）；binary SHA256 `299bafbd896af0de…`，source archive `803f31a1…`。
  - 部署：备份 `20260920T083647Z`（离主机副本在 WSL state-dir）；activate + 回环/公网健康全过；进程横幅 `New API platform-ef0958280718-ba26a16ffa71 started`，公网 `/api/status` 200。
  - 教训（三次误回滚根因）：deploy.py `--health-url` 必须传**域名根**（如 `https://ai.openzrob.com`），工具自行拼接 `/api/status`；传完整路径会请求 `/api/status/api/status` 404 → 判定公网健康失败 → 自动回滚。期间反复重启触发 systemd StartLimitBurst=5/300s（unit 配置），出现一次 start-limit-hit 短暂停机（约 40 秒，`systemctl reset-failed` + start 恢复，运行哈希前后一致无数据影响）。另外 WSL 侧：`~/.ssh/config` 曾丢失致 SSH 走 fake-IP 失败（已按 infra log 配方重建，IP 直连）；bun 需用 WSL 原生 `~/.bun/bin`（PATH 被Windows bun 遮蔽会 ENOENT 装包失败，已清缓存）。
  - 验收状态：生产 DB 用户 5 仍为 `acct_8ff11e2d0debe38`，等该用户下次网页登录触发 adopt 自动换名为 `7745491@qq.com`；此后新注册首登即真实用户名。待用户侧验收。
  - 下一步：用户在 /users 观察换名效果；cfab/c1633 的源码合并协调仍待 Mac 会话（本次以源码归档重建绕开，未产生新债务）。

- 2026-09-20T16:04:00+08:00 | ZCode | 修复账号中心建号使用随机哈希用户名：改为真实用户名（承接根 log 15:33 诊断）
  - 现象：用户首次登录 AI 站后 /users 显示 `acct_<哈希>` 用户名（如生产用户 5 `acct_8ff11e2d0debe38`，真实用户名 7745491@qq.com 只在显示名列）。
  - 改动：`model/account_identity.go` `ResolveAccountProductUser` 新增 centralUsername 参数——建号优先用账号中心用户名（introspection principal.username 本就返回），不可用（空/超 20 rune）或已被占用时回退原 `acct_` 摘要；已存在的映射若本地用户名仍是生成摘要且中心用户名可用，登录时自动换名（`adoptCentralUsername`，改库后刷新 Redis 用户缓存字段）；平台侧改过名的不匹配摘要、永不覆盖。`controller/account.go` 调用点传 `principal.Username`。
  - 验证：`go vet` 干净；`go test ./model/ ./controller/` 全过（新增 `TestResolveAccountProductUserUsesCentralUsername` 四例：真名建号/占用与超长回退/摘要号后补真名/平台改名不覆盖）；既有 `TestResolveAccountProductUserKeepsPlatformRoleIndependent` 与 DB 矩阵用例签名同步。`./service/` 包两个 channel-affinity 失败为 Windows 干净基线既有（stash A/B 复核），与本改动无关，部署管线在 WSL Linux 全量复跑。
  - 未完成：未部署（发布分支策略见根 log 待记）；生产用户 5 依赖登录时 adopt 规则自动换名，或部署后由用户重登验收。
  - 下一步：提交推送 origin；以线上 `platform-c1633c597d1f-e47411947c48` 源码归档为基线叠加本修复构建发布（该归档已核在 WSL deploy-state），不回退并行会话的隔离发布内容。

- 2026-09-20T15:25:00+08:00 | ZCode | 按用户决定撤销服务端钳制：revert 8aaac9ef9，DB 渠道参数覆盖已清空
  - 用户改选桌面端方案（GLM 目录级输出上限随 Desktop Beta 2.0.10-beta.3 发布，见 desktop log），服务端全部回退：分支 `git revert 8aaac9ef9` → `6aa2b58a9` 已推送（撤销 zhipu_4v 适配器钳制与测试；其间并行会话的 `5d43a91bf` 公共表面样式提交已在 origin）；生产 DB 渠道 2 `param_override` 清空并重启 healthy。15:05 条目的"DB 参数覆盖 + 代码钳制"两措施均已撤销，改由桌面端目录上限承担。旧安装包（<beta.3）对 GLM 会复现参数非法，属已知的过渡兼容性。

- 2026-09-20T15:05:00+08:00 | ZCode | 智谱 v4 渠道 max_tokens 钳制：代码已提交推送，生产暂以渠道参数覆盖止血
  - 分支 / 基线：`codex/platform-home-20260920` 由 `dffb6aead` 推进至 `8aaac9ef9` 并已推送 origin。
  - 根因承接根 log 14:14 条目：桌面端逐请求默认 `max_tokens: 256000`，智谱 v4 仅接受 [1,131072]，全部消息被上游 `INVALID_REQUEST` 拒绝。
  - 已完成：`relay/channel/zhipu_4v/relay-zhipu_v4.go` 在 `requestOpenAI2Zhipu` 中把超过 `131072` 的 max_tokens 钳制到上限（in-range/缺省透传），新增表驱动回归测试 `relay-zhipu_v4_test.go`（仿 ali/text.go 先例）。`go test ./relay/channel/zhipu_4v/` 通过。
  - 阻塞（二进制发布）：线上当前 release `platform-cfab9ed03090-ea3f98aeec09`（2026-09-20 14:24 由另一机器/会话发布，提交不在 origin，推测即上方"relay 凭证隔离"工作）包含本分支没有的改动；直接发布本分支 HEAD 会回退它。**下一次平台二进制发布前必须先拿到/合并 cfab9ed03090 的源**。本机按技能规程准备的 WSL 构建环境可用（见根 log）。
  - 生产止血（同日生效）：DB 渠道 2（智谱GLM）`param_override` 配置条件钳制 `max_tokens>131072 → 131072`，已重启生效并真实验证通过（详见 infra log 同日条目）。该配置与代码钳制语义一致，二进制合并发布后即为双保险。

- Desktop 内部 relay 凭证污染用户 API Key 列表的问题已本地修复：管理接口全面隔离系统凭证，同一登录会话/付款来源续期复用同一行，退出时禁用；相关 Go 全套、vet/build 通过，尚未部署。线上团队用量 0 的直接原因仍是当前 Desktop 使用个人点数，且团队成员月上限为 0。

- AI站用户导航和点数使用记录已本地完成精简：移除聊天分组、用户任务/审计入口，使用记录仅时间/模型/点数/来源；检查与桌面/手机验收通过，未发布。

- 桌面授权独立单卡片页面已本地验证，保留认证守卫；尚未部署。

- 显式浏览器退出新增统一账号会话撤销与查询取消，本地验证通过，未部署；依赖账号服务先支持browser-logout。

- 团队用量空月500热修复已上线：platform-ff416aa6f6a6-f52f9236c499；前端1717项、后端与构建通过，离机备份/运行哈希/公网健康及账号钱包检查通过。浏览器重启后需重新登录完成目视验收。

- 人民币统一计费已上线：20260916-cny-5c1f91a23e7e，独立发布源全量测试/构建、三库矩阵、生产迁移保值与公网验收通过。固定1元10点，不再日常换汇；原库保留，禁止二进制单独回滚。未包含并行账号中心。

- 价格保存尾差已本地修复：币种/倍率换算保留数值精度，10/5/小数反复保存重开回归通过；12文件200项测试、类型与构建通过。未部署，旧的已截断价格需上线后重新输入保存。

- 员工技能工作台/设备工作台已本地实现并联调：独立员工权限、摘要/详情、草稿/上架/更新/下架；Go42项、Web4项及构建检查通过。未部署，dash仍为既有跳转。

- 注册验证码品牌邮件已上线：20260915-email-49a31140f8ce；运行文件哈希、官网/API健康检查通过，备份已离机保存；真实收件客户端待验收。

- 注册验证码邮件品牌版已本地完成：RoboCoding/by擎云机器人、卡片与大号验证码；Go测试/vet/build及桌面/手机预览通过，尚未部署或真实投递验收。

- 官网品牌/下载已完成本地实现：RoboCoding + by擎云机器人，产品首页/登录/关于/下载及许可页，旧品牌归一化与七语言文案；旧 RoboCodingAI 默认名会在后端状态、前端缓存与持久化配置读取时归一为 RoboCoding。真实安装包未提供，未部署；全仓lint遗留问题见末条记录。

- 更新时间：2026-09-15T13:05:03+08:00
- 已完成：团队共享余额、月度成员上限、加入审批/邀请/部门、Key 四种付款策略；设备 PKCE 登录与受限会话、刷新/撤销；网站团队/账户界面与 CNY 输入显示、10 点/元换算。
- 验证：最终全后端测试、go vet 和嵌入最新 web/dist 的 Go build 通过；Web typecheck/build、69 项功能专项及全量检查发现的59项币种/导航旧预期修复后的对应复测通过；真实隔离 HTTP 团队扣费和 Desktop 客户端接口联调通过。
- 未完成 / 限制：未部署或联调正式供应商，正式地址未提供；团队资金首版由个人钱包转入；原始账本和高级计费表达式保留 USD 兼容语义；旧 Midjourney 写路由拒绝团队 Key。全仓格式检查仍有23个未改动文件、版权检查12个未改动文件失败，本次修改文件专项通过。
- 下一步：在正式平台地址和运行环境确定后做上线联调；核对当前共享工作树其他任务再决定提交发布。

## 记录格式

新增记录使用 `### 时间（带时区） | 执行者 | 任务`，包含：分支 / 基线提交、已完成、验证（命令或链接及结果）、未完成 / 阻塞、下一步。最新记录追加在文件末尾；涉及本条记录的提交可通过 `git log -- log.md` 查找，避免自引用提交哈希。

## 历史记录

### 2026-09-15T00:02:05+08:00 | Codex | 建立跨 AI 进度交接

- 分支 / 基线提交：`robo/main` / `7b719211b02f4ab9369fe30d3d4e2ef5e35cdfd8`。
- 已完成：新增根目录 `log.md`，登记现状；在 `AGENTS.md` 加入阶段更新和多 AI 合并规则。
- 验证：文档相对路径、历史规范保留和 `git diff --check` 检查通过；本次只修改交接文档，没有运行应用测试。
- 未完成 / 阻塞：上述产品待办保持未完成，本文档不代表运行功能已经交付。
- 下一步：先确定桌面授权接口与凭据生命周期，再实现模型调用和额度核对；修改认证前遵循 AGENTS.md 的 OWASP 要求。

### 2026-09-15T12:35:53+08:00 | Codex | 团队共享余额与桌面设备授权联通

- 分支 / 基线：`robo/main` / `a9b11d923b6693ce1a16d97d9b8d3a98b1528f16`。
- 已完成：新增团队、成员、月度用量、加入请求、可撤销邀请、请求预留及注资流水；四种 Key 付款策略；本人钱包到团队的事务注资；设备 PKCE 授权、受限会话、刷新与撤销；CNY 与 10 点/元元数据。个人账号身份和原整数账本保留。
- 简化：部门采用成员标签，团队余额共享，成员月度上限控制请求准入；不预先分拆个人钱包。复用既有会话、API Token、计费接口与前端组件，无新依赖。主要文件：`model/team.go`、`controller/team.go`、`service/billing_session.go`、`service/task_billing.go`、`model/desktop_device.go`、`service/desktop_device.go`、`controller/desktop.go`、`middleware/desktop_auth.go`、`web/src/features/teams/`。
- 验证：全后端 `go test ./... -count=1 -timeout 180s`、`go vet ./...` 已通过一次；此后终审修复需再验证。隔离数据库及本地模拟模型供应商的真实 HTTP 登录→邀请注册→设置成员上限→团队模型扣费→退出撤销通过，个人余额保持不变。真实 Desktop 客户端解析和设备授权/relay/logout 合同也通过；不代表生产供应商或原生 GUI 验收。
- 认证依据：[OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)、[Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)、[OAuth2](https://cheatsheetseries.owasp.org/cheatsheets/OAuth2_Cheat_Sheet.html)、RFC 8628；参考 ASVS 5.0 的 6.5.1/3/4/5、7.2.3、10.1.1/2、10.4.5/6/8/9。已测试 PKCE 不匹配、审批拒绝/过期、单次兑换和并发、会话撤销、来源和权限范围。终审发现并修复只读 Key 查询遗漏会话撤销检查、临时会话创建失败永久消耗授权；对应先失败后通过回归及 race 通过。未作完整 ASVS 合规声明。
- 未完成 / 限制：异步提交与完成的团队账务终审修复正在验证，前端收尾未完成；团队暂由个人钱包转入资金，未接独立团队支付渠道；旧 Midjourney 写路由明确拒绝团队付款 Key。正式域名未提供，未部署/推送/提交。
- 下一步：完成异步结算与 Web 最终验证，追加最终证据，不沿用旧绿灯覆盖后续修改。

### 2026-09-15T12:42:07+08:00 | Codex platform_teams | 修复异步团队任务两阶段结算

- 分支 / 基线提交：`robo/main` / `a9b11d923b6693ce1a16d97d9b8d3a98b1528f16`；保留共享工作树中其他账户、品牌和桌面改动。
- 已完成：团队任务提交阶段将预扣调整到持久化任务额度但保持 reservation 开放；终态成功按实际差额结算，固定按次和无可重算用量的成功任务显式关闭 reservation，异步失败按原团队和原月份退款；即时失败在持久化前归零。结算与退款重试保持幂等。
- 验证：controller 真实提交路径覆盖 `reserved` 到完成 `settled` 及 `result.Quota > 0` 的即时失败零扣费；团队成功、失败、零额度和固定按次回归通过。MySQL 8 / PostgreSQL 16 真实矩阵通过，日志 `/tmp/robocoding-async-team-db-matrix-final.log`；race 日志 `/tmp/robocoding-async-team-race-final.log`；相关 Go 包日志 `/tmp/robocoding-async-team-focused-final.log`。
- 未完成 / 阻塞：本任务无已知阻塞；平台整包 Go test、vet 与 Web 嵌入构建由共享工作树整合者在冻结源码后执行。
- 下一步：整合者复跑最终全量门禁并记录根工作区状态；本任务未提交、推送或部署。

### 2026-09-15T13:05:03+08:00 | Codex | 平台团队付款第一版最终验证

- 分支 / 基线：`robo/main` / `a9b11d923b6693ce1a16d97d9b8d3a98b1528f16`，源码冻结后验证，未提交或推送。
- 最终后端：`go test ./... -count=1 -timeout 180s`、`go vet ./...`、`go build -o /tmp/robocoding-platform-final .` 均 exit 0；构建包含最终 Web 产物。MySQL 8 / PostgreSQL 16 真实矩阵、SQLite 与 race 均通过。
- 最终前端：`bun run typecheck` 与 `bun run build` 通过；CNY、Key 表单、设备授权、价格编辑的7文件69项专项通过。首次全量135文件1666项中1607通过/59失败，失败均为新增 Teams 导航或默认CNY使旧USD测试前提改变；修订12个测试文件显式设定USD及正确status fixture，分批复测139通过、101通过/1失败，该唯一模型列表文件再修复并25/25通过。没有以全量初次失败日志冒充单次全绿；未重复运行未修改且通过的其余123个文件。
- 证据：`/tmp/robocoding-platform-all-final.log`、`/tmp/robocoding-platform-vet-final.log`、`/tmp/robocoding-platform-build-final.log`、`/tmp/robocoding-web-target-tests.log`、`/tmp/robocoding-web-all-final.log`、`/tmp/robocoding-web-cny-regressions-a.log`、`/tmp/robocoding-web-cny-regressions-b.log`、`/tmp/robocoding-web-model-listing-final.log`。修改文件的 oxlint/oxfmt 和 diff-check 通过。全仓 format/copyright 的剩余失败列表与本次修改文件交集为空，日志 `/tmp/robocoding-web-format-final.log`、`/tmp/robocoding-web-copyright-final.log`。
- 界面与联调：浏览器 route fixtures 验证团队列表、审批、部门/限额、邀请、Key策略及窄屏；使用模拟供应商的真实HTTP验证个人余额不变、团队扣费与退出撤销；真实 Desktop 客户端验证设备授权、点数和团队凭据合同。测试服务、生成凭据和数据库已清理，脱敏结果由私有整合工作区保存。
- 未完成 / 下一步：正式部署与供应商联调、Desktop 原生窗口验收留待环境就绪；账本保留旧精度/单位，CNY在可视化价格输入和点数边界转换。未将范围外格式问题、生产或原生GUI宣称通过。

### 2026-09-15T19:42:43+08:00 | Codex | 官网品牌与下载首轮验证

- 分支 / 基线：robo/main / a9b11d9；保留既有团队/账户混合修改。
- 已完成：默认首页替换机器人开发产品内容，登录外观/关于/页脚与开源许可独立页面，顶栏下载入口；独立发行清单驱动四种下载版本，未发布显示不可下载。
- 验证：导航/自定义首页10项、下载13项通过；首轮类型检查通过。浏览器本地模拟接口预览发现标题折行与旧logo回显，正在修复。
- 未完成：最终全量类型/lint/build、深浅主题和手机视觉复验；安装包未提供，未部署或提交混合工作树。
- 下一步：修复视觉细节、审查与最终验证，记录真实证据。

### 2026-09-15T20:05:03+08:00 | Codex | 官网品牌、下载与 SaaS 界面完成本地验收

- 分支 / 基线：robo/main / a9b11d9；保持既有账户/团队/桌面混合改动，未提交、推送或移动gitlink。
- 完成：统一RoboCodingAI/by擎云机器人，复用Desktop已批准的128px图标导出（11KiB）；旧默认名称/logo在后端选项内存读取、状态接口和前端缓存归一化，定制品牌保留。默认首页、登录、关于、页脚与主题重做；新增/download、/licenses，保留版权、许可证、包身份和源码链接。移除上游营销导航/更新入口，真实供应商与协议身份保持。
- 下载：四平台目标（Windows x64、macOS ARM/Intel、Linux x64）由/downloads.json驱动。全为未发布，不伪造包、版本或可点击地址；HTTPS/同源路径校验、版本/时间校验、错误优先于旧缓存和重试恢复均覆盖。发行接入见web/DOWNLOADS.md（平台组件路径）。
- 文件范围：平台common/constants.go、common/branding_test.go、model/option.go；web的首页/下载/关于/认证布局、公共及控制台顶栏/页脚、品牌常量/状态缓存、站点设置展示、主题、静态图标、两条新增路由和七语言资源；根DESIGN.md与两级log.md。
- 简化：复用原路由、Button/PublicLayout、主题、鉴权/账户/计费服务；只替换默认展示，保留管理员自定义首页；上游更新功能入口关闭，无新依赖。
- 验证：最终Web 8文件59项专项全通过；bun run build:check（完整类型+生产构建）通过；改动范围lint无错误（保留可信管理员自定义页脚的no-danger提示）；common/model完整测试、go vet ./common ./model、嵌入最终Web的go build通过；git diff --check通过。新增75个英文键在其余五个语言中75/75覆盖，简中齐全。
- 浏览器：生产预览1440/390px、深浅主题；首页到下载、移动菜单、登录/关于/开源许可可达；标题/图标正确，默认首页无上游外链，手机无横溢出，四个未发布按钮禁用，隐藏菜单inert。状态/初始化/公共文案接口使用本地模拟；未验证真实登录、邮件或生产接口。
- 纠错：首轮把小尺寸机器人图形误判为上游logo，直接核对批准资源后确认两者一致；旧路径归一化和最终PNG哈希已另行验证。一次并行测试在下载代理补充发行日期校验时撞上未更新fixture，最终统一复跑59/59通过。
- 限制：全仓lint仍有193处既有错误，主要在本次未改动utils、通用组件及脚本；未扩大成无关清理。ICO仍使用已有262KiB桌面导出，网站实际引用11KiB PNG。无安装包、签名、三平台安装验收或部署。根工作区忽略目录.omx/artifacts/platform-brand保存截图与检查日志，平台公开目录不含私有设计档案。
- 下一步：提供真实安装包与正式站点地址后，按发行说明上传、填写清单并做实际下载/安装及上线联调。

### 2026-09-15T20:13:25+08:00 | Codex platform_brand | 产品展示名移除 AI 后缀

- 分支 / 基线：`robo/main` / `a9b11d92`；保留共享工作树的团队、账户和官网既有改动，未提交、推送或部署。
- 已完成：平台后端默认名、Web 标题与 meta、首页、认证、关于、下载、许可、系统设置占位、七语言文案、下载清单契约与示例、产品清单及 README 的产品展示名统一为 `RoboCoding`；`by擎云机器人` 保持不变。仓库 URL、远端、`robocodingai` 文件名与资源路径、SVG 内部 id 等技术标识保持不变，历史日志未改写。
- 兼容：后端 `NormalizeSystemName`、状态缓存映射与 Zustand 持久化恢复会把旧默认名 `RoboCodingAI` 归一为 `RoboCoding`，并继续保留管理员自定义品牌；新增后端函数和前端缓存 / localStorage 回归覆盖。
- 验证：Web 品牌、首页和下载 5 文件 30 项测试通过；涉及文件 oxfmt 检查和 oxlint 通过；`bun run build:check` 类型检查与生产构建通过。`go test ./common ./model -count=1`、`go vet ./common ./model`、嵌入最终 Web 产物的 `go build -o /tmp/robocoding-platform-brand .` 均通过；`git diff --check` 通过。
- 未完成 / 下一步：真实安装包、正式站点部署和生产联调不在本次本地改名范围内；发布清单后续应继续使用 `RoboCoding` 产品名和安装包展示文件名。

### 2026-09-15T21:02:00+08:00 | Codex | 注册验证码邮件品牌排版

- 分支 / 基线：根 main / fb7d9eb；平台 robo/main / a9b11d92；保留既有混合工作树。
- 已完成：注册邮件改为 RoboCoding/by擎云机器人 品牌抬头、浅色卡片、大号连续验证码、动态有效期与安全提示；使用内联样式和表格布局，无远程图片、脚本或新依赖。邮件标题同步品牌与注册用途。
- 文件 / 简化：平台 common/registration_email.go、common/registration_email_test.go、controller/misc.go 与两级 log.md；模板集中到独立可测试函数，复用现有发送服务与品牌归一化。
- 验证：go test ./common ./controller、go vet ./common ./controller、go build -o /tmp/robocoding-email-check . 与两仓 git diff --check 通过；回归覆盖前导零、有效期、旧品牌归一化和HTML转义。浏览器1280px/375px预览通过，内容无截断；样例为虚构数据。
- 安全参考：已读取 OWASP Authentication Cheat Sheet 与 Session Management Cheat Sheet；本轮仅展示层，动态字段HTML转义，不新增验证码日志、外链或资源请求，现有生成/失效/发送逻辑保持。未宣称完整ASVS审计。
- 限制 / 下一步：未提交、推送或部署，未发送真实邮件；Outlook/QQ等实际收件客户端尚未验收。发布后验证实际投递与样式。visual-verdict技能本机未找到，使用直接截图核对，结论保存根 .omx/state/registration-email/ralph-progress.json。

### 2026-09-15T23:52:41+08:00 | Codex | 注册验证码邮件应用到生产

- 任务 / 基线：用户授权应用并继续；根 main / fb7d9eb，平台 robo/main / a9b11d92；发布基于线上 20260915-11901bb879ba 源码快照，仅增加邮件模板/测试并修改注册邮件标题与渲染入口。逐文件核对已有源文件仅 controller/misc.go 改变，无数据库代码修改。
- 已完成：版本 20260915-email-49a31140f8ce 已上线；运行进程二进制SHA256为 49a31140f8ce12127d82b40a7016493f5906ca1a14f3ed96535fcbbb20254630，与本机构建一致。前端从原部署快照重建。保留上一版本。
- 验证：快照 common/controller tests、go vet、Web build、Linux amd64 build 通过；www/API status success及登录页200；systemd active/running，NRestarts=0。
- 备份：发布前数据库与配置备份 20260915T153957Z 已取回用户指定certs目录，两个文件SHA256验证通过。仅展示层更新，未改数据库迁移；出现问题可切回上一版本并重启服务。
- 文件：infra/environments/production.json、根/平台/infra log.md；源码快照及发行元数据保存在certs的版本目录。未提交混合工作树、推送或移动gitlink。
- 限制 / 下一步：真实收件箱投递和不同邮件客户端渲染未验证，未主动发送邮件；后续注册请求使用新版模板。此前日志中“尚未部署”已被本次发布记录更新。

### 2026-09-16T00:52:12+08:00 | Codex | 员工技能工作台与设备工作台

- 分支 / 基线：robo/main / a9b11d92；保留此前混合修改。
- 已完成：新增controller/workbench.go及测试，router/api-router.go窄桥接；新增web features/workbench、动态路由、侧栏员工入口和七语言文案。复用现有UserAuth、页面壳、CardAction/表单控件；员工由服务端ROBO_WORKBENCH_STAFF_IDS授权，root可访问，不提升员工系统角色。无新依赖、账号/钱包副本或平台数据库迁移。
- 行为：列表仅摘要，详情按需读取；设备结构化环境表单、技能公开多行介绍与私有正文，草稿/上架快照区分，发布更新/下架独立操作，revision冲突拒绝覆盖。代理固定回环HTTP，无重定向，重建真实actor与服务凭据，1MiB请求/2MiB响应限制，异常响应归一和脱敏关联日志。
- 验证：controller工作台7项、router35项、Web3文件4项通过；go vet ./controller ./router、相关oxlint、typecheck、production build与diff检查通过。整合者最终本地Go编译/runtime再次验证登录、摘要/详情、409过期版本和401匿名拒绝；真实本地员工200、普通用户403。
- 浏览器：实际设备保存/上架/新草稿不改已发布版/发布更新，实际技能表单保存/上架/下架，多行数组和私有正文公开隔离均通过；1200px深浅主题与390px无横向溢出，保存按钮可滚动触达。修正了类型共用必填阻断、details数组契约、跨tab迟到响应归属和CardHeader动作对齐。
- 安全参考：读取OWASP Authentication Cheat Sheet及Session Management Cheat Sheet；沿用现有桌面凭据精确路径限制，不新增认证机制，不宣称全量ASVS审计。独立评审的错误requestId丢失/畸形响应回显/诊断缺失已修复并定向验证。
- 未完成 / 限制：未部署、未变更dash域名配置、未接真实机器人或技能执行引擎；目录发布不代表技能已经可运行。临时平台/内容服务/预览进程与测试数据库、凭据已清理。未提交、推送或移动gitlink。
- 下一步：与私有内容服务配套部署，配置真实员工名单与机器人条目；独立验收线上目录接入及技能执行。

### 2026-09-16T16:31:23+08:00 | Codex | 价格保存后的币种换算尾差修复

- 任务 / 分支 / 基线：用户反馈输入10或5保存后出现长小数；根main / fb7d9eb，平台robo/main / a9b11d92；保留既有混合修改及已有测试改动。
- 已完成：复现人民币转USD时先截12位小数、回显再乘汇率产生尾差；金额输入、普通Token倍率保存与加载改用保留数值精度的序列化，仅在相对浮点误差范围内归一化。输入10、5、0.12345678的阶梯输入/输出价反复保存重开保持一致，普通Token输入/输出和按次定价也覆盖。
- 文件 / 简化：web/src/features/model-pricing/pricing-amount-input.tsx、对应__tests__/editor-currency.test.tsx；web/src/features/system-settings/models/pricing-format.ts、model-pricing-core.ts、model-pricing-sheet.tsx及__tests__/pricing-precision.test.ts。复用现有输入组件和显示格式，无新依赖、数据库或账务接口改动。
- 验证：新增回归先失败，证实10回显9.999999999997、5回显5.000000000002；修复后model-pricing、system-settings/models及billing-expression共12文件200项通过；bun run build:check类型/生产构建通过；变更范围oxlint/oxfmt与git diff --check通过。原生子代理只读审阅确认相对容差不把微小非零金额存为零，非有限值仍交由校验拒绝。
- 限制 / 下一步：未部署、提交、推送或移动gitlink，未改线上定价。旧的已截断金额不会自动还原，发布后需重新输入目标金额并保存。既有显示格式仍将极微小金额（约1e-12以下）显示为0，本次保证保存不被该显示规则截为0；真实生产保存与账单未验收。

### 2026-09-16T17:46:04+08:00 | Codex | 原生人民币计费整合阶段

- 分支 / 基线：robo/main / a9b11d92；保留并行改动，未提交或发布。
- 已完成：固定人民币账本精度500000 quota/元；内置模型、工具售价人民币化，移除美元上游Cost反推扣费。充值核对人民币金额，价格同步拒绝未声明同币种/精度来源。旧账必须显式离线迁移后启动。
- 验证：controller、service、setting/...、relay/helper、pkg/billingexpr整合测试通过，日志/tmp/robocoding-cny-integration-tests-2.log。
- 未完成 / 下一步：前端验证、迁移安全与三数据库矩阵、完整Go测试/构建/vet；正在消除禁用支付handler的不可达旧代码。不代表上线或生产账本已迁移。

### 2026-09-16T19:07:04+08:00 | Codex | 人民币账本三库验证

- 分支 / 基线：robo/main / a9b11d92，保留并行改动。
- 已完成：版本化离线迁移、支付快照及原币种保留、旧Redis额度缓存隔离、未结状态/未知支付币种迁移门禁、CLI帮助与参数前置校验。
- 验证：SQLite3.50.4、MySQL8.4.11、PostgreSQL16.15实库扣费、独立日志、支付幂等和旧schema升级/重复启动通过；迁移专项race通过；controller/model/service/settings/router/helper/expr/CLI测试、vet及全模块build通过。
- 限制：完整go test ./...仍有并行account模块过期验证码清理测试失败（account_test.go:310）；前端全量回归仍在收尾。未发布、提交或推送。离线迁移前必须停止所有写账进程并备份主/日志数据库，旧二进制不可直接写新账。

### 2026-09-16T19:18:59+08:00 | Codex | 人民币计费本地完成与最终验证

- 分支 / 基线：robo/main / a9b11d92；保留并行账户和品牌修改。
- 完成 / 简化：setting/controller/model/service及web的价格、钱包和设置使用原生人民币，固定1元10点，移除日常汇率换算；离线迁移CLI、账本标记与支付快照保护旧账价值。未知历史权益/上游币种明确标注未知，无新依赖。
- 验证：人民币前端10文件172测试、typecheck/build、定向oxlint/oxfmt通过，7语言键集一致；全量151/152文件、1724/1725测试通过，唯一失败是并行移除更新入口后的system-update旧断言。相关后端测试/vet、三库矩阵及迁移race见前条；嵌入最新Web的go build ./...及diff-check通过。
- 限制 / 下一步：全仓account验证码清理测试与无关lint问题仍在；未提交、推送、部署或执行生产迁移。上线须停写、备份、离线预检和显式迁移，不能二进制单独回滚。真实支付商户联调待后续执行。

### 2026-09-16T20:41:13+08:00 | Codex | 独立人民币版本上线

- 基线 / 范围：当前开发robo/main / a9b11d92保持混合工作树；从线上源码另建独立发布提交5c1f91a23e7ecb74e3d5d34afd8bd605f5c7c577，只移植人民币变更，未携带账号中心或新工作台。
- 完成 / 简化：20260916-cny-5c1f91a23e7e已上线；原生CNY计费、固定1元10点、旧账在候选新库显式迁移保值，删除日常换汇。用户/Token/订阅货币缓存隔离，订阅追加回归已同步开发树。
- 验证：发布源Go全量/vet/build、relaykit独立检查、三库扣费、生产副本迁移通过；Web1710全量+最后35专项、类型/构建与改动lint/format通过。正式维护模式重复启动、登录/钱包/定价/配置/退出及公网元数据/首页脚本通过；运行hash与本机构建一致，异常重启0。
- 限制 / 下一步：旧库与旧程序保留，但开放流量后不能回切旧库丢失新写入，也不能仅回滚二进制。上线前后备份已离机校验，新备份恢复确认CNY。真实商户支付/模型供应商消费待专项验收；共享工作树未提交/推送/改gitlink，全仓既有lint/format未扩大处理。


### 2026-09-16T21:08:44+08:00 | Codex console_frontend | 独立账号 SSO 与权限工作台前端收口

- 任务 / 分支 / 基线：独立账号中心、平台中央 SSO、员工技能/设备/仓库工作台权限界面；`robo/main` / `a9b11d92`。保留共享工作树中后端、计费、品牌和其他并行修改，未提交、推送或部署。
- 已完成：`/account/*` 不再启动平台 status/setup/旧认证；SSO exchange 显式匿名且不会被旧 refresh 拦截。callback 在 `applyAuthBundle` 后使用 TanStack Router SPA replace，避免整页刷新销毁内存 access token。平台路由先等待 `/api/status` 判定账号模式；中央模式跳过 legacy bootstrap/登录页 refresh，受保护路由冷启动使用账号 Cookie 做普通 PKCE 续登。仅后端返回 `AUTH_REAUTH_REQUIRED` 时传 `reauth=true`，普通过期/撤销续登不强制密码；启动瞬时失败清理 singleflight/60秒 marker，原敏感请求始终不自动重放。
- 工作台：新技能移除可编辑的公司/社区归属，保存不再发送 `sourceKind`，由服务端按账号权限派生；`/api/workbench/access` 的 `permissions[]` 控制创作者、员工、审核员可见的技能/设备/仓库入口、标签页和新建操作，记录级按钮继续使用 `allowedActions`，后端仍为权限权威。
- 验证：中央 SSO/回调/启动、权限矩阵、工作台 payload、侧栏共11文件72项通过；另跑 auth-session 与 OTP 路由2文件18项通过。`bun run typecheck`、定向 oxlint/oxfmt、生产 `bun run build`、范围内 `git diff --check` 通过；最终构建日志 `/tmp/robocoding-web-console-final-build.log`。
- 本地浏览器：账号登录页无 `Failed to load system config`；本地平台日志在 21:01:52 显示 `/api/account/sso/exchange` 200 后没有整页 GET `/workbench/skills`，随后 `/api/workbench/access` 与 `/api/workbench/skills` 均200，浏览器成功进入工作台。该证据来自本地临时服务，不代表生产部署。
- 未完成 / 下一步：正式域名、生产部署和生产账号迁移未执行。最新整合二进制的受保护页硬刷新、显式退出后停留登录页及中央 token 服务端撤销由根整合者继续做最终黑盒并追加记录；不要恢复 callback 的 `window.location.replace`，也不要在中央模式重新调用 `/api/user/auth/refresh`。

### 2026-09-16T21:18:10+08:00 | Codex unified_account_backend | 独立账号中心与中央认证后端最终验证

- 任务 / 分支 / 基线：独立账号中心、平台中央 SSO、角色权限/自定义标签、创作者申请、GitHub 工作台能力、旧 New API 用户迁移、敏感操作重新登录及中央浏览器退出；`robo/main` / `a9b11d923b66`。保留共享工作树中的人民币计费、前端和品牌并行修改，未提交、推送或部署。
- 已完成：账号中心独立维护账号、Argon2id 密码、会话、OAuth Code + PKCE、角色权限和管理员手动标签；员工发布代表公司，创作者申请经后台批准。平台使用远程 introspection 和本地产品用户映射，中央模式阻断旧账号入口；Desktop 会话绑定中央父会话并实时检查撤销。敏感操作要求最近重新登录，proof 单次消费并绑定用户、会话、scope 和 context。GitHub 短时 capability 由账号中心签发，平台先做权限预检，明确区分未连接、禁止和账号服务故障。中央浏览器 logout 通过 access token 在账号中心解析真实 subject/SID 并撤销，校验可选 expected SID，不接受客户端提供 subject；Desktop JWT 继续撤销平台本地会话。
- 迁移与数据保留：旧 New API 用户迁移支持 dry-run、apply 和幂等重跑，排除软删除用户；保留旧用户 ID、钱包/已用额度、relay key 与 team 关系，创建独立账号及产品映射，不把旧密码作为中央敏感操作重新验证凭据。MySQL `8.4.11`、PostgreSQL `16.15` 的 schema 与真实导入 dry-run/apply/rerun 均通过；首次创建后重跑无重复创建，源数据保持不变。数据库矩阵证据 `/tmp/robocoding-account-matrix-final.log`、`/tmp/robocoding-account-migration-matrix.log`。
- 主要模块：`account/`（store/server/password/migration/GitHub 与测试）、`service/central_account.go`、`service/security_verification.go`、`middleware/auth.go`、`middleware/secure_verification.go`、`controller/account.go`、`controller/auth_session.go`、`controller/secure_verification.go`、`controller/workbench.go`、`model/account_identity.go`、`model/login_verification.go` 及对应 router/tests。修复了 UTC/SQLite 时间、最后活跃管理员并发保护、注册事务原子性、严格 bcrypt 兼容校验、数据库故障错误语义、logout 不伪装成功和权限拒绝被配置故障覆盖等问题。
- 后端验证：`go test ./account ./service ./middleware ./controller ./router -count=1` 全通过；`go test -race ./account -count=1` 通过；`go vet ./account ./service ./model ./middleware ./controller ./router`、`go build ./...`、`git diff --check` 均 exit 0。独立后端聚焦审查批准当前实现与测试边界。
- 跨服务黑盒：本地实际账号中心 + 平台共 12 组真实 HTTP 场景全部通过，覆盖注册/登录、SSO、权限和标签、创作者申请与批准、工作台权限、GitHub 权限错误分类、敏感操作重新登录、会话撤销及中央 platform logout；退出后同一 token 访问平台和账号中心 `/v1/me` 均返回 401。证据 `/tmp/robocoding-account-cross-service-final.log`。
- 安全参考：实现与审查参考 [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html) 和 [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)，用于密码、会话生命周期、重新认证、撤销和错误处理设计；本记录不声明完成全量 ASVS 合规审计。
- 未完成 / 下一步：正式域名部署、生产账号迁移、实名认证和生产 GitHub App 配置未执行；审计写入仍按当前非阻断策略处理，未新增 outbox。上线前应使用生产副本重复 dry-run 并备份，随后按账号中心、平台、前端的依赖顺序发布和验收。本任务未修改生产环境。

### 2026-09-16T21:23:33+08:00 | Codex | 团队用量页面500根因确认

- 任务 / 基线：排查 /teams/1/usage；工作区main / fb7d9eb，平台robo/main / a9b11d92；线上20260916-cny-5c1f91a23e7e。保留其他任务修改。
- 已确认：实际线上接口HTTP200、success=true、members=null；无本月用量时前端members.map抛错并进入错误页。
- 处理：仅在前端用量API边界归一化空列表，新增回归；不改SQL、钱包或身份流程。
- 验证 / 下一步：已复现空列表测试失败并验证最小修复通过；页面回归及独立线上基线构建继续。尚未发布。

### 2026-09-16T21:35:34+08:00 | Codex | 团队用量热修复验证阶段

- 隔离发布树codex/team-usage-hotfix / 45c6860，基于线上CNY 5c1f91a23e7e。共享开发树仅teams/api.ts及两份teams/__tests__回归修改。
- 验证：真实浏览器500和Cannot read properties of null (reading map)与API空列表证据一致；5项API/实际路由页面回归、完整typecheck、定向lint/format通过。
- 完整构建：首次Web全量1716通过、1个无关渠道配置20秒超时；该文件单独59项通过，正在重跑完整发布检查。未跳过检查、未发布。
- 下一步：完整检查通过后按备份与可回滚热修流程恢复线上并复验页面。

### 2026-09-16T22:03:46+08:00 | Codex | 团队用量500热修复上线

- 任务 / 基线：修复 /teams/1/usage 空月崩溃；工作区main / fb7d9eb，平台robo/main / a9b11d92；独立发布codex/team-usage-hotfix / ff416aa6f6a636faeecdeea843f012bba8e9bcaf，前版20260916-cny-5c1f91a23e7e。
- 已完成：teams/api.ts把members:null归一化为[]；新增usage-api.test.ts、usage-page.test.tsx共5项；vitest.config.ts限制2个worker以避免共享主机资源竞争。无新依赖；不改后端、账本和身份。
- 发布补修：infra/deploy/production/backup.sh记录实际版本二进制路径，使backup元数据与发布校验器一致；远程仅替换对应两行，保留现有CNY数据库目标，原脚本已在服务器留备份。
- 验证：Web148文件1717项、typecheck/build、修改4文件lint/format通过；Go全量重跑、vet、relaykit独立test/build、Linux amd64构建通过。初始Web并发超时和既有SQLite并发测试偶发锁竞争均保留失败证据，未跳过；已通过前端的冻结源码每个文件与归档字节比对后接续后端，见产物build.log。发布helper15项及bash -n通过。
- 线上：platform-ff416aa6f6a6-f52f9236c499，部署时间2026-09-16T14:01:26.573944+00:00；SHA256 87870426712f79222df499d413c441d5c1ac8f6c97feccd277b9705320c89709；服务/公网健康、实际登录/钱包/定价/CNY/退出通过。备份20260916T140046Z已离机校验，路径/Users/xiaoyuan/certs/robocodingai/server/deploy/backups/robocoding-production/20260916T140046Z；收据/Users/xiaoyuan/certs/robocodingai/server/deploy/receipts/robocoding-production-platform-ff416aa6f6a6-f52f9236c499.json。
- 限制：修复前真实浏览器确认500及null.map异常；浏览器随后重启并失去登录，修复后目视验收待登录。自动回滚只允许上一CNY版本，旧USD二进制仍不安全。共享混合树未提交/推送/移动gitlink；仅隔离发布树冻结提交。
- 下一步：重新登录后验证空月显示0点；本次不扩展到账户中心发布。


### 2026-09-16T22:51:21+08:00 | Codex | Extract application identity and workbench modules

- Branch/baseline: robo/main / a9b11d92; shared unrelated changes preserved.
- Account service/CLI and workbench proxy/UI removed from this module after extraction. Platform retains model/wallet authority and central identity-client/user mapping; former UI paths navigate to the independent applications. Central /security targets the security page.
- Verified: service/middleware/controller/router tests, Go vet/build, Web typecheck/build,33 identity-navigation tests and later2 security navigation tests. Independent four-service15-scenario HTTP suite passed against the cleaned platform binary.
- Not deployed; no commit/push/gitlink change. Live account migration and external integrations still require deployment-environment verification.


### 2026-09-16T23:35:07+08:00 | Codex | 恢复生产布局样式

- 应用源码与前版完全相同，仅修复发布工具扫描环境后重新构建。1717项Web测试、类型/构建、Go全量和relaykit检查通过。
- 发布 platform-1f578ef0c3fc-f59610d4106a，前版 platform-ff416aa6f6a6-f52f9236c499；离机备份、运行SHA256和公网健康通过。首页/登录页实际浏览器布局恢复。无账号迁移、数据库或业务源码变化；共享开发树未提交。


### 2026-09-17T00:38:39+08:00 | Codex | 顶栏工作台入口上线

- 发布platform-213b9420b2b0-54a03878bfa7，前版platform-1f578ef0c3fc-f59610d4106a。共享导航新增工作台，保留控制台；独立发布4文件，无后端/数据库变动。
- Web1719项、类型/构建、Go全量/vet、relaykit及CSS产物检查通过；离机备份、运行哈希与公网健康通过，真实桌面顶栏链接验证成功。账号切换不在此发布范围；共享树未提交。

### 2026-09-17T09:48:49+08:00 | Codex platform_account_billing | 账号中心团队与 AI 账务权威桥接

- 任务 / 基线：在共享 `robo/main` 混合工作树中仅实现平台 Go 侧账号账务 bridge、中央团队投影与旧入口封闭；保留其他并行修改，未提交、推送或部署。
- 已完成：新增严格 action allowlist 的 `/api/internal/account-billing`，使用独立 `ROBO_PLATFORM_INTERNAL_TOKEN` 和账号中心实时 `management-authorize`；钱包查询/调整、团队 provision/get/member-limit/fund 均使用 points 字符串精确换算，返回原始整数额度、`quotaPerPoint` 与 CNY。额度调整和团队转账使用全局 `operationId + payload hash` 持久幂等，冲突拒绝，记录 actor/target/reason；敏感动作要求五分钟内重新认证，`ai.admin` 不得调整 admin/root 目标。
- 团队权威：新增字符串账号团队到原数字账本团队的独立映射，不改变既有 team ID；按账号 snapshot 实时同步 active 成员、角色和部门，保留本地月限额/用量。投影版本单调且相同版本必须同 hash，延迟快照不能回滚。中央模式封闭旧团队管理和旧用户/额度管理；`/api/teams/self` 通过账号 subject-team 枚举逐队实时刷新，仅保留 key 选择所需只读列表，空结果固定为 `[]`。Token 新建/更新、Desktop 支付源和每次新的团队预扣都实时验证账号团队；在途 settle/refund 不新增检查。
- 账务安全：团队 provision 使用稳定且唯一的 `CENTRAL_` join code，支持多个账号团队且旧加入路由不可用；fund 复用原子个人钱包→团队账本，Redis 并发重复预留会补偿。普通成员 `team.get` 仅返回本人月限额/用量，团队管理员与 AI 管理员可查看全部成员。
- 验证：`go test ./controller ./middleware ./model ./service ./router -count=1`、定向 `-race`、`go vet`、`go build ./...`、`git diff --check` 均通过；SQLite 回归覆盖授权撤销、精确小数、operation重试/冲突、双团队 provision、版本回滚、空数组和 Redis 并发补偿。MySQL 8.4 与 PostgreSQL 16 实际 schema/双团队 provision/资金守恒矩阵通过。根整合者另以独立 8790/3028 服务完成 8 组真实 HTTP 场景，覆盖重新登录、额度重试、团队绑定/fund/limit、多团队、subject枚举、禁用/恢复、移除、角色撤销和旧写409。
- 未完成 / 下一步：本任务没有生产迁移或发布；生产切换前仍需用生产副本执行账号/团队 alias 迁移、备份和域名级回归。账号中心和平台必须同时配置各自方向的独立 internal token，团队权威切换由 `ROBO_ACCOUNT_TEAM_MODE=central` 单独控制。


### 2026-09-17T11:49:03+08:00 | Codex | 统一账号与工作台生产发布

- 生产部署已完成，平台版本platform-account-cutover-20260917-01f456a；独立account/console/skills/devices与邮件适配器上线，ai为AI站正式入口，www保留。
- 验证：真实迁移dry/apply/rerun、原账号登录和跨应用PKCE、账务与API Key不变量、五域HTTPS及internal404、运行文件校验、完整协调备份和实际隔离恢复均通过。
- 本组件使用冻结候选发布；未提交共享工作树或改根gitlink。SMTP仅验证TLS认证；GitHub/Feishu真实接入、真实支付/模型花费及桌面安装包未包含。流量已开放，禁止盲回滚数据库。详见私有工作区生产清单及日志。

### 2026-09-17T13:13:00+08:00 | Codex platform_local_ownership | 恢复 New API 本地业务权限与账务权威

- 任务 / 基线：按独立 Account 仅负责身份与 SSO、New API 负责本地用户角色/状态、AI 团队、余额、API Key 与计费的边界实施；隔离提交 `8cc568af8a17`，基于线上源码 `01f456aeb026`。未访问生产、未新增依赖；补丁逐文件同步，保留共享混合工作树并合并已有测试/日志。
- 已完成：Account 授权码交换后创建绑定中央 authority 的 New API 本地浏览器会话，浏览器不再直接持有或逐请求使用 Account opaque token；Account 权限不覆盖本地 `users.role`。中央会话退出只撤销本产品会话；中央模式继续允许本地 refresh cookie 轮换。恢复 New API 原生团队、用户状态/角色/额度及账务管理，删除运行时 Account 团队投影和账务写桥；中央模式只封闭密码、注册、PAT、身份字段与删除等身份入口。
- 前端：中央模式恢复 New API 用户/团队页面；用户编辑只开放本地业务字段。登录页使用 `web_message` iframe，先静默续登、交互时使用全新请求和 iframe；严格校验后端返回授权 URL 的 origin、iframe source、type/state，无关消息忽略。支持重新登录、超时、取消、Escape、迟到响应隔离；显式退出以 tab 内标记阻止 Account Cookie 立即静默回登，用户点击登录后清除。
- 身份状态：既有 API Key 每次继续通过 `POST /v1/internal/account-status` fail closed 校验中央账号 active；请求使用平台 `ROBO_ACCOUNT_INTERNAL_TOKEN`，Account 端要求 `identity.subject-status` scope。回归覆盖 active、disabled、服务不可用、Bearer 凭据和 subject body。
- 清理：删除旧 `/api/internal/account-billing`、中央团队运行时同步和对应测试；历史账务映射 model 仅保留只读迁移/对账结构，不再 AutoMigrate 或写入。
- 验证：隔离提交 `go test ./...`、`go vet ./...`；Web 154 文件 1759 项、typecheck、build、修改范围 lint、全量 format 通过。共享树由根整合者继续执行跨服务/浏览器与冻结构建；本任务未部署。

### 2026-09-17T13:36:18+08:00 | Codex platform_local_ownership | 封闭跨标签页退出与中文登录文案遗漏

- 任务 / 分支 / 基线：在隔离分支 `fix/local-platform-ownership` / `07c94c6908e` 上处理最终审查遗漏；未访问生产、未部署、未推送。
- 已完成：中央账号模式收到同 SID 的跨标签页 `signed_out` 广播时，在清理本地会话前写入当前标签页的 signed-out 标记，防止 Account Cookie 仍存在时立即触发静默回登；显式退出复用同一条件标记函数。将七种语言中误放在资源顶层的两条账号登录文案移入 `translation` namespace，中文与繁体中文现在可实际命中对应译文。
- 验证：新增中央/本地模式标记回归及中英文 namespace 结构回归；相关 6 个测试文件 30 项通过；`bun run typecheck`、`bun run build:check`、修改范围 oxlint 与格式化通过。同步修正两个既有测试 mock 对 Axios `params` 的显式类型，以保持当前 TypeScript 构建可复现。
- 未完成 / 下一步：提交后重新生成私有 Linux amd64 候选产物并记录 SHA，再将本次增量逐文件安全同步共享混合工作树；仍不执行部署或远端推送。

### 2026-09-17T14:14:17+08:00 | Codex platform_local_ownership | 修复中央会话 Authority Redis 热缓存丢失

- 任务 / 分支 / 基线：生产 smoke 在维护窗口发现 Account 全局退出后已有 AI 会话仍可能通过；隔离分支 `fix/local-platform-ownership` / `337bd6daacb9`。未访问或修改生产，后续 Git 推送已暂停。
- 根因 / 已完成：`writeUserSessionCache` 的 Lua HSET 未写入四个 Authority 字段，导致中央浏览器/桌面会话从 Redis 热缓存读取后被误判为无中央绑定。本次补齐四字段及 ARGV，缓存 schema 从 2 提升到 3，使旧缺字段缓存强制回源数据库；中央模式对 `account_sso` 与 `desktop` 登录方式的 Authority 全空或不完整状态增加 fail-closed 防线，本地模式继续兼容原有无绑定桌面会话。
- 验证：新增 miniredis roundtrip、旧 schema 回源回填、当前 schema 缺 Authority 拒绝、refresh rotation 保字段、revoke tombstone 保字段及 Account 全局退出后热缓存拒绝回归；相关定向测试通过，`go test ./... -count=1 -timeout=180s` 与 `go vet ./...` 通过。
- 未完成 / 下一步：创建 Lore commit 后使用既有 Web dist 构建新的 Linux amd64 二进制供发布负责人上传；明确文件同步共享树，但不执行部署或后续远端推送。

### 2026-09-17T15:22:31+08:00 | Codex platform_native_login | AI 站改用原生登录表单连接 Account

- 任务 / 分支 / 基线：共享 `robo/main` 混合工作树中仅负责平台中央登录前端、JSON OAuth 启动白名单和对应回归；保留账号中心、账务、品牌及其他并行修改，未提交、推送或部署。
- 已完成：中央模式复用 New API `UserAuthForm` 的账号/密码、法律同意、加载与错误 UI，不再嵌入 Account iframe；隐藏平台本地 Passkey、微信、OAuth、Turnstile 与密码加密入口，忘记密码指向 Account `/account/recovery`。浏览器以 `credentials: include` 直接 POST Account `/v1/auth/login`，请求不携带平台 Authorization；成功后使用 `response_mode=json`、state 与 PKCE 获取单次 code，再复用平台 exchange 建立本地 session bundle。静默续登也走 JSON，显式退出标记在凭据失败和服务故障时保持，仅完整授权成功后清除。
- 安全与竞态：Account 请求 15 秒超时、卸载/重试 Abort；取消信号贯穿 PKCE digest、SSO start、Account fetch、code exchange 与本地状态落地前检查，避免迟到 silent 请求覆盖显式登录。授权 URL 必须与状态提供的 Account 同源、路径严格为 `/v1/oauth/authorize` 且 `response_mode=json`，跨源重定向拒绝；密码和 token 不记录。依据 OWASP Authentication、Session Management 与 OAuth2 Cheat Sheet 的 TLS、通用认证错误、会话更新、state/PKCE、redirect URI 与授权码单次使用要求实施；本记录不声明全量 ASVS 合规审计。
- 验证：先确认 `responseMode=json` 后端回归返回 400、前端新协议/原生 UI 回归失败，再实现转绿。Web 中央登录及路由相关 5 文件 21 项通过，其中真实 `UserAuthForm` 2 项覆盖账号密码、法律同意、禁用本地备选认证与失败重试；`bun run typecheck`、修改范围 oxlint、`bun run build` 通过。Go `controller/service/middleware` 中央认证专项、`go vet ./controller ./service ./middleware`、嵌入最新 `web/dist` 的 `go build ./...`、`git diff --check` 均通过。
- 未完成 / 下一步：根整合者正在使用本地 Account + 平台真实 HTTP 和浏览器执行最终跨域 Cookie、失败重试、显式退出与页面目视验收；本任务未执行生产发布或生产账号操作。

### 2026-09-17T15:32:28+08:00 | Codex integrator | 原生登录最终集成验收

- 本轮最终源码的14组临时跨进程测试通过，包含原生身份登录/JSON PKCE及设备授权、刷新、退出。最终浏览器确认原生表单、零iframe、登录/刷新、退出后错误密码保持退出、新标签静默SSO及390px无横向溢出。
- 临时测试进程已清理；未部署、未提交或推送。生产跨域与桌面安装包验收留待正式更新，本条不代表生产版本变化。


### 2026-09-17T18:38:19+08:00 | Codex | Revoke browser identity before explicit app sign-out

- Baseline: robo/main/a9b11d923; preserved existing shared worktree edits, no commit/push/deployment.
- Added browser identity logout client, called after cancelling pending queries and before local logout; explicit failures do not report success. Existing device and internal local logout API semantics remain unchanged. No new dependencies or schema changes.
- Verified focused frontend tests and production build; isolated cross-service test confirms identity revocation followed by local logout succeeds and old application credentials are rejected. Follow OWASP Authentication and Session Management guidance; cross-origin endpoint requires exact trusted Origin and JSON.
- Remaining: deploy compatible identity endpoint before this frontend; production cross-domain browser acceptance pending.


### 2026-09-17T18:39:13+08:00 | Codex | 独立桌面授权页

- 分支/基线：root main/e9b5005；platform robo/main/a9b11d923，共享混合工作树保留。
- 完成：授权路径保留认证守卫，跳过后台布局；授权页复用SystemBrand/Card，删除SectionPageLayout外壳，居中单卡片，窄屏设备信息纵排；拒绝使用对应图标。更新授权测试Router上下文，修正已有退出测试cancelQueries返回类型。
- 验证：Web typecheck通过；限定文件oxlint零error（退出旧测试2 warning）；相关3测试通过；生产构建通过。真实Chrome合成API预览1440x900/390x844，无导航/侧栏、无横向溢出，批准完成通过，截图已目视核对。
- 未完成/边界：未部署；浏览器使用合成授权数据，不代表生产授权链再次验收。未提交其他任务混合修改。下一步按发布流程交付此UI改动。

### 2026-09-17T18:52:17+08:00 | Codex integrator | Native login production release verified

- Account `account-native-login-cf97db7bab49` and platform `platform-2823bba09a62-5c5ad2ab6d74` are deployed; previous releases are retained. No schema or business-data migration. Five services are healthy and running executable hashes match frozen artifacts.
- Full build gates and production HTTPS browser native form, refresh, application logout and silent SSO passed; no iframe, no script errors, no mobile overflow. Complete pre-release backups were copied off-host and all payload hashes verified.
- Shared development tree was not bulk-committed or pushed. Platform rollback is compatible while keeping the new Account service; account rollback requires platform rollback first, without restoring databases.

### 2026-09-17T18:57:37+08:00 | Codex | Suppress revoked background request notifications after logout

- Baseline: robo/main/a9b11d923; preserved shared worktree changes. Incremental files: web/src/lib/http-client.ts and web/src/lib/__tests__/central-auth-expiry.test.ts.
- Reuse handled-error tracking for authenticated background 401s during explicit sign-out or after auth is cleared; keep request rejection and explicit authentication failure reporting. No dependencies or server/session-invalidation changes. Read OWASP Authentication and Session Management guidance.
- Regression: two new cases failed before fix, then 33 related tests passed; typecheck, production web build, changed-file lint and diff checks passed. Full lint still fails in unrelated existing files.
- Not deployed or production-browser verified; no mixed-tree commit/push. Next: isolate release and verify browser logout notifications.

### 2026-09-17T19:48:00+08:00 | Codex | 精简下载页首屏与系统品牌图标

- 分支 / 基线：平台 `robo/main` / `a9b11d923`；共享混合工作树保留，未提交、未推送或部署。
- 已完成：下载页删去发布渠道、功能卖点和选型帮助等冗余区块，首屏只保留产品名、当前版本及四个系统安装卡；Windows、macOS 和 Linux 卡分别改用 Font Awesome 官方品牌标识，不再使用通用显示器、Apple 轮廓或终端符号。
- 验证：`bun run typecheck`、下载页 2 个测试文件共 4 项、修改范围 `oxlint`、`bun run build` 和 `git diff --check` 均通过。生产 `www.openzrob.com/download` 与本地浏览器当前均只返回空白页面；该下载页全套源文件及路由仍是工作树未跟踪内容，未把本地构建结果等同于生产验收。
- 未完成 / 下一步：由整合者将下载功能全套文件纳入受控提交并走平台发布流程后，再做正式域名桌面与移动端目视验收。


### 2026-09-17T20:12:36+08:00 | Codex | Header brand sizing

- Baseline: existing shared robo/main checkout; unrelated changes preserved.
- Completed: system-brand.tsx and public-header.tsx omit the subtitle and use 18px brand names across header/sidebar variants. No dependencies added.
- Verified: targeted format/lint, Web typecheck/build and diff-check passed.
- Remaining: local changes only, not deployed.

### 2026-09-17T20:13:36+08:00 | Codex registration_entry | Direct Account registration entry

- Baseline / scope: shared `robo/main` checkout at `a9b11d923`; changed only the platform Web registration entry, URL validation, route search schema, and focused tests. Other mixed worktree changes were preserved; no commit, push, or deployment.
- Completed: central-account sign-in now links directly to `/account/register` with the resolved `dark` or `light` theme. A direct platform `/sign-up` visit replaces the location before rendering an intermediary page. Legacy mode keeps the existing platform registration form. Existing `continue` is forwarded only when it targets `/v1/oauth/authorize` on the configured Account origin; unsafe continuations and non-HTTPS remote Account URLs are rejected.
- Verified: 7 focused registration/navigation tests, Web typecheck, production build, changed-file oxlint/oxfmt, and diff checks passed. The implementation follows the existing same-origin return validation and does not put sessions or credentials in the URL; OWASP Authentication and Session Management guidance was reviewed for the affected navigation boundary.
- Validation limit: the broader auth suite passed 120/121 tests. `central-sso-routing.test.tsx` cold protected-session coverage times out at `routes/_authenticated/route.tsx:41`, where it calls `resolveAuthentication()` and the private `authClient`; its test adapter replaces `api.defaults.adapter`, not that client. The failing path does not render `SignIn`, so the new `useTheme` hook is not executed. This pre-existing protected-route/test mismatch is outside this registration-entry change and remains for its owning authentication task.

### 2026-09-17T20:20:42+08:00 | Codex | AI 站普通用户导航与消耗记录精简

- 基线：root main/7a061882；platform robo/main/a9b11d923；保留共享工作树其他任务改动。
- 完成：移除聊天分组；普通用户隐藏任务/审计日志，旧地址重定向使用日志。使用日志桌面/手机仅时间、模型、点数、来源；移除用户技术筛选和人民币/RPM/TPM，管理员保留诊断视图。客户端既有自动凭据名称映射“Robo Coding 客户端消耗”，其他API记录单独显示来源；不修改鉴权、后台日志或账本。
- 修改范围：platform web hooks/use-sidebar-data、usage-logs access/section-registry/路由、列工厂/新增user-logs-columns、table/mobile/filter/stats与相关测试、i18n。简化复用现有表格、筛选和点数换算，无新依赖。
- 证据：相关20文件360测试通过；类型/构建通过、主修改范围lint和diff-check通过。1280×720/390×844模拟登录与消费数据浏览器通过；三条旧日志地址重定向、手机无横溢；截图output/playwright/ai-usage-*.png。发现并修复筛选对象不稳定导致的循环渲染，最终页面无运行异常。
- 未完成/限制：最后空列菜单收纳与对应复核进行中；尚未提交推送或部署，不代表线上已改变。客户端当前为登录后自动换取会话绑定的短期调用凭据，用户无需创建API Key；现有来源识别沿用固定token_name，未新增后端来源字段。
- 下一步：完成最后检查并补记；生产发布需从混合工作树隔离候选。

- 2026-09-17T20:23:31+08:00 最终补记：空列“查看”菜单已收纳；stats与table统一忽略用户旧技术筛选参数。最终专项9项、类型检查、生产构建、修改范围lint及diff-check通过；浏览器两尺寸再次验收通过，无横溢、无空查看按钮。此前相关20文件360项全通过，最终修改按影响范围复测。visual-verdict技能不可用，直接截图评审结果记于私有工作区.omx/state/ai-user-usage/visual-verdict.json。未部署或提交混合树。

### 2026-09-17T20:36:44+08:00 | Codex | Desktop relay 凭证与用户 API Key 隔离

- 基线 / 范围：共享 `robo/main` / `a9b11d923` 混合工作树；只修改 Desktop relay 凭证生命周期、Token 管理查询及对应回归，保留其他任务改动；未提交、推送或部署。
- 生产只读核查：团队1余额100点、成员1人，admin月上限与本月团队用量均为0；当日通用日志存在3次个人消费，其中 Desktop 调用 ¥0.05714。最新及有历史消耗的 Desktop 内部凭证付款来源均为“仅个人额度”，因此团队页0点是当前付款来源的真实结果，不是页面请求失败。
- 根因 / 完成：Desktop 正确保存 refresh token + SID 并以登录会话访问控制面，但 `/v1` 兼容层把会话派生的 relay 能力存成普通 Token，导致其出现在 API Key 列表；临近一小时到期时还会新增一行。现在 list/search/count/detail/get-key/batch/update/delete 的共同模型边界只暴露用户管理的 Token；同一 SID + 付款来源的临期 relay 凭证原行延寿，不再每小时插行；退出撤销父会话并禁用该 SID 的内部 Token，保留行供在途结算/审计安全窗口使用。
- 验证：新增回归覆盖 API Key 列表/搜索/详情隔离、临期续期复用同一 Key、个人/团队付款源分离、退出后 relay 401 与内部 Token 禁用。`go test ./controller ./model ./middleware ./service -count=1`、对应 `go vet`、`go build ./...`、平台及根 `git diff --check` 全部通过。
- 限制 / 下一步：正式站仍运行 `platform-2823bba09a62-5c5ad2ab6d74`，截图中的旧内部凭证在发布前仍可见；未执行生产变更。长期应改为父登录会话派生、独立 audience/token_use 的短期 relay capability/JWT，并以显式 `credential_source` 替代日志按固定名称识别来源。

### 2026-09-17T21:09:02+08:00 | Codex platform_releases | 服务器本地安装包发布与三版本保留

- 任务 / 基线：共享 `robo/main` / `a9b11d923` 混合工作树；本任务只负责平台发布存储、内部 API、公开清单及对应测试/文档，保留并行任务改动，未提交、推送、部署或修改生产文件。
- 已完成：新增 Console BFF 专用的 `GET/POST /v1/internal/releases` 草稿管理及 `POST /v1/internal/releases/{version}/publish`；使用独立共享凭据与可信管理员主体头，缺少配置时关闭。安装包按四个受支持目标流式落盘，服务端计算大小和 SHA-256；同版本草稿可逐平台合并，已发布版本不可覆盖。发布通过原子清单替换生效，第四个已发布版本提交成功后才移除最旧版本并删除其本地文件。`/downloads.json` 动态返回最新版本、更新时间、更新日志和四平台状态，保留版本的安装包由不可猜目录映射后的公开路由下载。
- 配置 / 边界：`ROBO_RELEASE_STORAGE_DIR` 默认 `data/desktop-releases`，生产须持久化；`ROBO_RELEASES_INTERNAL_TOKEN` 供工作台 BFF 使用；`ROBO_RELEASE_MAX_FILE_BYTES` 默认每文件 2 GiB，请求总上限为四倍加 1 MiB。服务器本地目录按单写节点设计，多实例须固定写入/下载节点或另加共享文件系统单写约束。完整契约见 `docs/desktop-release-management.md`。
- 验证：新增服务与路由回归覆盖无凭据拒绝、multipart 草稿、顺序平台合并、发布后真实文件读取、公开清单更新日志、发布第四版后仅保留三版且最旧目录删除、已发布不可覆盖和单文件大小限制。`go test ./service ./controller ./middleware ./router -count=1`、`go vet ./service ./controller ./middleware ./router`、`go build ./...`、修改范围 `git diff --check` 均通过。
- 未完成 / 下一步：未部署、未上传真实签名/公证安装包，也未做正式域名下载或安装验收。工作台需配置相同发布凭据并只让本地管理员进入上传台；部署时把存储目录纳入持久卷、备份和容量监控。

### 2026-09-18T10:47:22+08:00 | Codex | AI 站首页官网化重设计

- 任务 / 基线：平台 `robo/main` 共享混合工作树；参考 `https://bigmodel.cn/` 与 `JCodesMore/ai-website-cloner-template` 的 clone-website 方法，仅修改首页主体、首页样式和首页 Footer 标记，未提交、推送或部署。
- 完成：去掉原有“左文案 + 右聊天预览 + 小卡片”的通用 AI SaaS 结构，改为高留白黑白编辑式首屏、低饱和冷紫轨道光晕、三列能力带、三步工作流和深色 CTA 收束；保留下载、控制台、登录态、多语言翻译调用及现有公共布局。
- 验证：`bun run typecheck`、首页 `product-home` 4 项测试、`bun run build:check`、修改范围 `oxlint` 和 `git diff --check` 均通过；全量 lint 仍被其他既有文件错误阻断。浏览器本地因未连接平台后端而停在启动空白页，未将其当作视觉验收证据。
- 未完成 / 下一步：尚未部署；接入可用平台后端或正式站环境后，需按 1440px / 390px 再做一次实际页面目视验收，重点检查动态品牌、导航和首页自定义内容优先级。

### 2026-09-18T11:42:56+08:00 | Codex | 首页多语言与下载页官网化

- 任务 / 基线：平台 `robo/main` 共享混合工作树；延续首页重设计范围，新增下载页视觉和首页新增文案资源，未提交、推送或部署。
- 完成：首页新增文案接入中英文翻译资源，并通过 i18n 同步保持其他语言 key 完整；下载页沿用现有 `/downloads.json` 发布清单和下载链接，仅改为同一套黑白高留白、冷紫光晕、无卡片堆叠的官网视觉，新增平台版本带、更新说明区和深色 CTA。
- 验证：中文本地浏览器真实查看首页与下载页，下载页无发布版本状态正常显示；首页/下载页相关 4 个测试文件 25 项通过，`bun run typecheck`、`bun run build:check`、`git diff --check` 通过。修改范围 lint 仅剩 home 旧组件既有 index-key/button-type/import-type 错误，新增文件无对应错误。
- 未完成 / 下一步：尚未部署；多语言中除中文、英文外的新首页文案暂按现有英文 fallback 展示，若需要法/日/俄/越完整本地化可再单独补齐语义翻译。

### 2026-09-18T11:55:00+08:00 | Codex | 中文首页排版收窄

- 任务 / 基线：平台 `robo/main` 共享混合工作树；延续首页官网化改版，保留其他任务改动，未提交、推送或部署。
- 已完成：首页主标题改为完整翻译 key，避免中文被拆成英文式两段并单独套色；根据 `zh` / `zh-TW` / `ja` / `ko` 语言标记启用 CJK 排版，收窄主标题及区块标题字号、放宽字距并增加行高，英文展示规模保持不变。
- 验证：`bun run i18n:sync`、`bun run typecheck`、首页 Vitest 4 项、`git diff --check` 通过；本地 `http://127.0.0.1:3000/` 中文深色首页刷新目视确认，主标题不再占满首屏。
- 未完成 / 下一步：未部署；其他语言新增主标题 key 继续使用英文 fallback，后续若需要可补充本地化文案。

### 2026-09-18T14:42:00+08:00 | Codex | 公共平台页面发布

- 基线：`robo/main` / `f5b5fa1e`，从共享工作树隔离网页/品牌下载改动形成冻结提交 `c3e5306ec1abc4a966eb8f1126b96ca3d4a624f4`。
- 已完成：发布 `platform-c3e5306ec1ab-466608ff886b`，仅包含公共首页/下载/品牌视觉增量；数据库、认证、账务与运行配置未变。
- 验证：`bun install --frozen-lockfile`、Web build check/test、Go test/vet/build、Linux amd64 构建全部通过；上线后运行二进制 SHA256 与产物一致，服务和公网健康通过。前置备份 `20260918T062850Z` 已由发布脚本校验。
- 未完成 / 限制：真实模型消费、支付、邮件、容量与 Safari 矩阵仍未验收；Desktop 安装包不属于本发布。

### 2026-09-18T14:20:00+08:00 | Codex | 公共页面视觉统一

- 任务 / 基线：平台 `robo/main` 共享混合工作树；沿用首页黑白留白、细线和冷紫点缀方向，未触碰认证后工作台布局，未提交、推送或部署。
- 已完成：公共布局改用同一套纸张/墨色 token；导航由悬浮圆角胶囊改为透明平铺头部，滚动时仅保留轻底色与细分隔线；公共页卡片取消阴影和大圆角，按钮统一胶囊形；页脚与首页保持同色和细线结构。下载、关于、排行榜、定价、许可证等复用公共层的页面自动继承。
- 验证：公共层与首页/下载专项共 25 项测试、`bun run typecheck`、`git diff --check` 通过；本地首页和下载页已刷新目视确认导航统一。开发环境 React Query 浮层遮挡部分截图内容，但不属于产品页面。
- 未完成 / 下一步：公共页面仍保留各业务模块自己的信息密度和数据图表样式；若要将登录页或认证后工作台也改为该风格，需要单独按信息架构逐页验收。

### 2026-09-18T15:20:00+08:00 | Codex | 蓝白简洁主题与环境切换

- 任务 / 基线：平台 `robo/main` 共享混合工作树；按用户要求将公开页和认证入口从黑紫编辑风格收敛为浅蓝/白色，保留轨道和光晕动效，未提交、推送或部署。
- 已完成：公开首页、下载页和公共卡片/按钮增加蓝白主题覆盖；认证布局新增浅蓝侧栏、白色表单、蓝色主按钮；默认主题改为浅色但保留手动深色切换；生产模式未指定后端地址时默认使用 `https://ai.openzrob.com`，本地仍默认 `http://localhost:3000`，并新增 `dev:local` / `dev:production` 启动脚本。
- 验证：认证/首页/下载相关 27 项测试、`bun run typecheck`、`bun run build:check`、`git diff --check` 通过。
- 未完成 / 下一步：本轮未执行生产发布或正式域名浏览器验收；生产切换仍需按发布流程构建并部署，认证后工作台保留原数据密集型主题。

### 2026-09-18T21:22:57+08:00 | Codex | 收口 Account SSO 错误边界

- 分支/基线：`robo/main` / `795a3a9dfd6a`；范围仅中央 SSO 交换错误映射。
- 完成：Account 会话撤销继续返回 `AUTH_SESSION_REVOKED`；平台本地用户停用返回 `AUTH_PRODUCT_ACCOUNT_DISABLED`/403；会话限额和内部错误不再伪装成账号会话失效。
- 证据：`GOWORK=off go test ./controller ./service -run 'TestCentralAccountSSOExchange|TestCreateCentral|TestValidateCentral' -count=1` 通过；新增 disabled product account 回归；无 schema/数据库迁移。
- 未完成/下一步：需将平台子模块提交推送后，按线上兼容版本重新构建部署并核对 Account/AI/工作台实际 SSO；工作台仍直接向 Account introspection，不依赖平台会话。

### 2026-09-18T20:13:06+08:00 | Codex | 公共首页与下载页精简发布

- 基线 / 发布：从线上 `platform-e1c334a10eeb-eadcca5b848b` 源码归档恢复，隔离提交 `f13d1bfb332bbcf8d52c5c8a10c08784c74af625`；发布 `platform-f13d1bfb332b-81100608b4f9`。
- 已完成：首页中文中段标题保持一行并避免孤立末字；下载页移除重复营销文案和 CTA，保留版本、下载卡与 Changelog；同时包含当前 Web 工作树中已验证的中央 SSO 保护路由恢复逻辑；通过版本化 SSH/systemd 流程上线。
- 验证：Web build check/test、Go test/vet/build、Linux amd64 构建全部通过；dry-run 通过；上线后运行二进制 SHA256 `cceb18e77d79534cbc8629fa03b134a02753f82de3b846471bd56cc97691f840` 与产物一致，服务及公网健康通过。
- 备份 / 回滚：前置备份 `20260918T121227Z`；上一版本 `platform-e1c334a10eeb-eadcca5b848b`，未执行数据库迁移，支持二进制回滚。
- 限制：未进行正式浏览器逐页目视验收、Safari 矩阵或 Desktop 安装包发布；平台工作树原有混合改动保留未提交。

### 2026-09-18T22:18:27+08:00 | Codex | 中央账号重置后的旧会话回收与找回入口收口

- 基线：`robo/main`，发布源 `69f48ef192ff660d2dbfd3802c5619cc4366fc4c`，线上 `platform-69f48ef192ff-20150e7286bf`。
- 已完成：中央 Account auth version 变化后，平台在新 SSO 会话签发前回收旧版本的本地 AI 会话，避免旧会话继续占用 active session limit；发行窗口限额仍保持反滥用语义。旧平台找回密码页在中央模式跳转 Account recovery。
- 验证：平台完整 Web build/test、Go test/vet/build，中央旧会话回收定向测试通过；生产 status `active=true`、`healthy=true`、`public_healthy=true`，运行 SHA256 与发布物一致。
- 限制：真实邮箱验证码投递和真实凭据重置未代验；Account 负责验证码和密码权威，平台不恢复旧密码重置接口。

### 2026-09-18T22:36:00+08:00 | Codex | 回收中央切换前的无绑定遗留会话

- 基线：`robo/main`，发布源 `4e90161245511072520b66a002c46bc41da4b000`，线上 `platform-4e9016124551-6e293e9fbd55`。
- 已完成：中央登录在 active limit 之前回收无 `authority_issuer` / 无 `authority_auth_version` 的旧本地密码会话；保留当前中央 Account 会话及发行窗口限额语义。
- 验证：遗留会话回收定向 Go 回归、平台完整 Web/Go build/test/vet、线上 status 与公网健康通过。
- 限制：真实 admin 登录尚未代操作；现有线上 5 个遗留本地会话将在下一次中央登录时按新逻辑回收。

### 2026-09-19T12:51:55+08:00 | Codex | 中央登录回收已失效的同版本 Account 会话

- 基线：平台 `robo/main` 共享工作树；保留其他既有未提交修改，未提交、推送或部署。
- 已完成：中央登录达到 active session limit 时，枚举同用户仍 active 的中央会话并调用 Account `session-status`；仅明确返回 inactive 的会话自动撤销并清理缓存，再重新计数。Account 网络/服务异常不会误撤销会话。
- 验证：`GOWORK=off go test ./service ./model` 通过；新增同 `auth_version` 的失效 Account 会话回归测试；`git diff --check` 通过。
- 未完成 / 限制：尚未生产发布；未做 MySQL/PostgreSQL 矩阵或真实公网 Account 联调。
- 下一步：在干净平台发布候选中执行跨数据库验证并上线后复现中央 SSO 登录。

### 2026-09-19T13:00:43+08:00 | Codex | 已失效中央会话回收修复生产发布

- 发布：`platform-b7fa4567d146-7458e1a35781`，兼容线上 `platform-4e9016124551-6e293e9fbd55`；无数据库迁移。
- 验证：完整 Web/Go 构建检查通过；生产运行哈希 `b3d319924a3b72ed72839cc2f7c35712f0e8f480909891a9043e970db28b857d` 与候选一致，systemd、本机 health、公网 health 均正常。
- 备份 / 回滚：`20260919T050004Z`；上一版本可通过 receipt 回滚。

### 2026-09-19T16:11:23+08:00 | Codex | OpenFox 公共品牌与域名引用更新

- 平台 Web 公共显示名更新为 OpenFox，站点、AI、账号和工作台链接切换至 `openfox.work` 子域名；保留上游包名、模块路径与内部 client 标识。
- 验证：本轮未执行独立 Web workspace typecheck；Yarn lock 未登记当前 workspace 包，命令在依赖解析阶段阻塞。

### 2026-09-19T16:13:03+08:00 | Codex | OpenFox Web 资源验证补充

- Platform Web `tsgo -b` 与品牌/独立应用定向测试 12/12 通过；公共 logo/favicon 已换为 OpenFox 用户商标图，未部署。

### 2026-09-20T12:53:00+08:00 | Codex | 首页中心光晕与悬浮顶栏单层收口

- 任务 / 分支 / 基线：平台 Web 首页视觉修复；`codex/platform-home-20260920` / `abd8ab2f9`；保留共享工作树既有改动，未提交、推送或部署。
- 已完成：在 `openfox-hero` 增加低对比度蓝紫中心洗色，降低背景网格存在感；滚动顶栏保留原有 `scrollY > 20` 交互，仅让外层胶囊承载边框、背景与阴影，内层 nav 透明化，消除双层边界。
- 验证：`git diff --check` 通过；样式增量仅 `web/src/styles/index.css` 10 行。尝试执行 Bun `typecheck`、首页 Vitest 与生产构建，但当前依赖目录存在既有错配：TypeScript 报多处非本次改动类型错误，Vitest 缺 `@vitest/utils`，Rsbuild 缺 `@rspack/core`；冻结安装未能在当前环境完成。
- 未完成 / 阻塞：未进行本地浏览器截图或生产发布；完整 Bun 依赖安装和页面目视验收仍待可用前端构建环境。
- 下一步：在依赖完整的 Web 工作区重跑 `bun run typecheck`、首页测试和 `bun run build`，再按桌面 / 移动端滚动状态确认最终视觉。

### 2026-09-20T13:55:00+08:00 | ZCode | 官网冷启动恢复登录态，公共头部直显头像

- 任务 / 分支 / 基线：平台 Web 会话引导修复；`codex/platform-home-20260920` / `abd8ab2f9`；保留工作树中 Codex 未提交的首页样式改动（`web/src/styles/index.css`）与其日志条目，不在本次提交内。
- 已完成：根路由 `beforeLoad` 移除"中央账号启用即跳过会话引导"的短路，冷启动恒走 `bootstrapAuthentication()`——持刷新 Cookie 的回头用户在首屏渲染前恢复登录态，公共头部直接显示头像而非登录按钮；无会话提示的访客仍不发任何请求。同步清理仅服务于该判断的阻塞式 `/api/status` 拉取与 `centralAccountEnabled` 死代码；登录路由改为自行用 `ensureStatus` 预热共享 status 缓存，中央/本地表单形态不闪烁，拉取失败不阻塞登录页。
- 验证：`tsgo -b` 零错误；触碰路由 oxlint 0 告警；Vitest 全量 162 文件 1801/1801 通过；`rsbuild build` 成功。修复前基线对比确认 calendar.tsx 类型错误与构建失败均为依赖损坏所致，已按 pnpm-lock 9.0 `--frozen-lockfile` 重装修复。
- 未完成 / 限制：未部署生产；官网头部头像直显的浏览器目视验收待线上进行。
- 下一步：部署平台 Web 后，用已登录账号冷加载官网首页确认右上角直接显示头像，点击登录不再闪登录页。

### 2026-09-20T14:49:00+08:00 | ZCode | 公共页面改版、顶栏单层化与冷启动认证三轮发布

- 任务 / 分支 / 基线：平台 Web 公共视觉改版 + 认证引导修正；共享分支 `codex/platform-home-20260920`，隔离发布分支基于 `dffb6aead` 叠加 `cfab9ed03` + `c1633c597`；生产基线依次 `platform-57aa60fe012e-aae1d0cb0989` → `platform-cfab9ed03090-ea3f98aeec09` → `platform-c1633c597d1f-e47411947c48`。
- 已完成：悬浮顶栏改为常驻单层胶囊（表面只在 `.robo-public-header-inner` 一层，内层 nav 任意状态透明，几何不随滚动变化，仅阴影加深），消除滚动过渡双层叠加与 Windows 下 backdrop-filter 过渡停滞残影；参考智谱 BigModel 克制语言整体改版：全局公共面顶部淡蓝径向光晕、首页黑色胶囊按钮 + 三张统一白卡深色圆图标（去紫绿点缀与整卡蓝底）、下载页标题降档 + 版本小胶囊 + 四张平台卡白色圆角浮卡（描边徽章），未发布状态以虚线提示替代"暂不可用"死按钮；冷启动认证在 dffb6aead 基础上移除 `bootstrapAuthentication` 的会话提示跳过，刷新后恒校验一次，提示缺失但持有效刷新 Cookie 的老会话在公共页恢复头像；下载卡片测试同步更新。
- 验证：deploy 工具三轮完整检查通过（bun install/build:check、CSS 校验、全量 Vitest 1801/1801、go test/vet/build）；本地与线上浏览器截图验收首页/下载页/顶栏；线上确认匿名冷启动恰发出一次 `/api/user/auth/refresh`。`TestSecurityAccountDeletionConcurrentRequestsHaveOneWinner` 出现一次负载偶发失败，单独 `-count=3` 复跑通过后重试构建成功。
- 协作 / 备份：包含 13:55 ZCode 条目 `dffb6aead`，无需重复部署；`8aaac9ef9` zhipu 限流改动不在本发布内，由对应任务自行发布。备份 `20260920T055457Z` / `20260920T062346Z` / `20260920T064726Z` 存于 WSL `~/.openfox-deploy-state`，可按收据回滚。
- 未完成 / 下一步：已登录头像的线上目视验收需真实账号（机制上由 refresh 在首屏前恢复）；根 gitlink 未推进，待工作区维护者核对后推进。

### 2026-09-21T19:45:00+08:00 | ZCode | 顶栏与账号入口按访问域名动态派生（已上线）

- 分支 / 基线：共享分支 codex/platform-home-20260920 / 4d9fc646f；隔离发布提交 ed1964278d8f（WSL ~/build/platform-links）；生产 release platform-ed1964278d8f-67e7b01d5bd9。
- 已完成：ICP 备案期间 openfox.work 被拦、openzrob.com 为生效入口，但顶栏"工作台/AI 站"、independent-apps 兼容跳转与 central-logout / central-account-sign-in 的账号中心回退仍在构建期写死 fox 域名，用户在旧域名一点导航即跳到被拦域。product-links.ts 改为运行时从 window.location.hostname 按域名族（openzrob.com / openfox.work）派生同级入口，未知主机回落当前生效族 openzrob.com；use-top-nav-links、central-logout、central-account-sign-in 全部改用派生常量，保留 VITE_ROBO_* env 覆盖用于本地开发。
- 验证：本地定向 vitest（independent-apps + 域名族新用例、sidebar-config、top-nav-brand、central-sso*、user-auth-form-central）与 tsgo -b 通过；deploy 全量门禁 8 项通过（bun install/build:check、CSS 校验、全量 Vitest、go test/vet/build）；线上 ai/www.openzrob.com 新入口 index.21ba49c73f.js 无任何写死 fox URL，active+healthy，备份 20260921T113539Z。
- 未完成 / 限制：根 gitlink 未推进、未推送（按工作区约定待维护者）；真实浏览器点击顶栏的目视验收待用户确认。回滚：deploy.py rollback --expect-current platform-ed1964278d8f-67e7b01d5bd9。

### 2026-09-25T16:35:00+08:00 | Codex | New API 默认界面对齐 OpenFox 蓝白主题

- 任务 / 分支 / 基线：平台 Web 默认外观；`codex/openfox-ui-20260925` / `33c2d5c81b109a739b3fbc0099a52d3a3c1f204d`。只调整主题和共享布局，不改模型、钱包、认证或上游包身份。
- 已完成：默认浅色主题统一为 OpenFox 蓝白色阶，深色主题采用同一冷色体系；顶部导航、侧栏活动状态和卡片边界收敛为轻量层级；移除表格逐行动画与卡片位移动效；主题选择器将默认主题显示为 OpenFox。
- 验证：`bun install --frozen-lockfile`、`bun run typecheck`、所改 TSX/TS 文件的 oxlint、`bun run build`、侧栏/移动布局定向 Vitest 17/17、`git diff --check` 通过。
- 未完成 / 下一步：未做真实登录后的浏览器目视验收，未发布生产；组件提交推送到 origin/main 后，再由工作区推进 gitlink。

### 2026-09-25T21:03:23+08:00 | Codex | 狐狸线稿视觉候选与平台壳层对齐

- 任务 / 分支 / 基线：平台 `codex/fox-visual-20260925` / `403594f75d42fac97d5dbb4c8d267c3e1f00e5b2`；本轮只改 Web 主题、应用壳层与登录呈现，不改认证和账务逻辑。
- 已完成：默认主题采用 OpenFox 纸白/冰蓝/墨蓝配色；顶栏、侧栏、品牌标志和登录页使用统一层级。登录画布加入现有狐狸线稿、尾巴弧线与少量星芒，说明文案明确指向模型、API 密钥和钱包，七种既有语言同步。保留 New API 上游版权与文件布局。
- 验证：`bun run typecheck`、15 项相关 Vitest、改动文件定向 oxlint/oxfmt、`bun run build`、`git diff --check` 通过；本地 Chrome 1440px 登录页在代理生产公共 API 后无错误提示且已目视检查。全量 lint/format:check 因未改文件中的既有问题未通过，不据此声称全库整洁。
- 未完成 / 下一步：尚未合入平台 main 或发布生产；真实已登录平台页面和账号联邦链仍需发布前验收。根工作区待按分支工作流整合 gitlink。

### 2026-09-25T22:36:00+08:00 | Codex | 放大平台入口字号并删除虚构首页演示

- 任务 / 分支 / 基线：平台 `codex/fox-visual-20260925` / `f755af0cae0117b9073ea6625715230cffa8ce23`；仅 Web 公共首页、导航与认证页视觉，保留认证、模型和钱包逻辑。
- 已完成：登录控件、公共导航与应用侧栏主文字调整到约 16px；首页正文、卡片、按钮采用可读字号；删去装饰短句、静态卡片伪“查看项目”、不可交互的任务输入框，以及只存在于首页演示中的 `@openfox/sdk` / `openfox/auto` / `842ms` 假代码。底栏清除空标签与重复口号。
- 验证：Web 类型检查、首页 4 项及认证相关 15 项测试、改动文件 lint、生产构建、差异检查通过；桌面和 390px 手机真实页面预览，手机没有页面横向溢出。既有压缩格式文件的整文件 oxfmt 检查仍失败，本轮没有以大面积格式化掩盖差异。
- 未完成 / 限制：开发候选未合入组件 main 或部署生产；真实登录后的业务页面与跨应用联邦流程未验收。组件提交推送后再推进根 `dev` gitlink。

### 2026-09-25T23:04:00+08:00 | Codex | 平台视觉候选合入主线准备发布

- 任务 / 分支 / 基线：平台 `main` 从 `403594f75d42fac97d5dbb4c8d267c3e1f00e5b2` 快进至功能提交 `cf65b04b2d780d790109a76862b932f96f440723`；远端 `origin/main` 在合并前仍为旧基线。
- 已完成：功能分支已先于 22:42 推送并核对远端 SHA；本地 `main` 快进合入全部四个视觉候选提交，无冲突。上线前只读核对生产现为 `platform-ed1964278d8f-67e7b01d5bd9` 且服务、进程哈希和公网健康一致。
- 验证：`git rev-list --left-right --count` 显示旧 `origin/main` 落后新本地 `main` 四个提交；主线工作树干净（本日志更新前）。此前 Web 类型检查、相关测试和构建证据见上一条。
- 未完成 / 下一步：推送组件主线后更新根 gitlink。生产嵌入式 Web 需从线上精确源码基线构建兼容发布，不直接以组件仓当前后端 main 代替线上基线。

### 2026-09-26T00:02:29+08:00 | Codex platform_saas | 已登录 AI 站对齐 SaaS 信息层级与表单

- 任务 / 分支 / 基线：`codex/saas-workspace-20260925` / `8925057cb3c13c16035777bd02845f7b6af62de3`；仅改平台 Web 视觉与对应布局测试，保留上游许可和原有 API、模型、钱包操作。
- 已完成：共用页头和内容边距调整为清晰的工作区结构；总览真实用量移至配置引导前，指标字号放大并去除装饰性假代码背景、空泛的细字；API Key 和模型编辑器使用同一套标签/控件对齐规则，桌面双列、窄屏单列，必要的付款和访问规则说明保留。
- 验证：`bun run typecheck` 通过；总览 7 项和 API Key 抽屉 3 项 Vitest 全过；改动 TS/TSX 文件 oxlint 零错误；`bun run build` 成功且构建 CSS 含新表单规则；`git diff --check` 通过。
- 未完成 / 限制：未合入组件 `main`、未推进根 gitlink、未发布生产；无真实员工会话可验证已登录页面实际数据与操作，需整合后用真实账号验收。首次 `git fetch origin main` 曾遇本机 schannel TLS 握手失败，推送功能分支时需重试并核对远端提交。
- 下一步：提交并推送此组件功能分支，将 SHA 交给工作区整合者；整合者核对视觉并决定合入、部署。

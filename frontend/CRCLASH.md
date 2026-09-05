# 前端代码审查报告（CRCLASH）

> 审查日期：2026-09-05　审查范围：`frontend/` 全部源码（不含 `node_modules/`、`web/default/` 构建产物、字体/图片资源）

## 概览

- **技术栈**：Vue 3（`<script setup>` SFC）+ Vite 5 + Vue Router 4（history 模式）+ Tailwind 3 + PrimeVue 4 + Radix Vue + Phosphor icons + ECharts（按需加载）。包管理器 npm/pnpm 双轨，`_packageManager: npm@11.12.1`。
- **审查文件数**：约 48 个源码文件（34 个 `.vue` + 14 个 `.js/.ts/.css` + 配置文件）。`src/` 共 74 文件（其余为字体、图片等静态资源）。
- **质量基线**：`npm run typecheck`（`vue-tsc --noEmit`）**通过**；无 lint 脚本；无任何单元测试 / Playwright spec / 测试目录。
- **整体评价**：**良好**。目录职责划分清晰（`views/`、`components/ui/` 原子组件、`lib/` 配置中心、`services/` API 层），设计体系与工作室 LogShare 项目保持同步，类型检查严格，无 TODO/FIXME 残留。主要问题集中在：**冗余依赖与死代码、版本排序逻辑缺陷、开放重定向面、无自动化测试**。

---

## 问题清单（按严重级别排序）

### 严重

| # | 位置 | 问题 |
|---|------|------|
| S1 | `src/views/DownloadStartedView.vue:55` | **开放重定向面**：`goBack()`/`goToWebsite()` 对 `route.query.return_url` 和 `fileInfo.return_url` 只做 `isExternal`（跨源即放行）判断。攻击者可构造 `/download-started?token=...&return_url=https://evil.com` 诱导用户点击「返回上一页/前往网站」跳转到任意外部地址。前端应做来源白名单（仅允许站内/已知合作源），或至少为外部跳转加二次确认。 |

### 一般

| # | 位置 | 问题 |
|---|------|------|
| G1 | `src/components/VersionList.vue:40`、`src/views/FilesView.vue:62-65` | **版本号按字典序二次排序，破坏后端已有顺序**：`String(b.tag_name).localeCompare(a.tag_name)` 无 `{ numeric: true }`，会把 `1.10.0` 排在 `1.9.0` 之后（'1'<'9'）。后端按语义版本排好序，前端这层二次排序会把「最新版本」显示错乱。建议加 `{ numeric: true, sensitivity: 'base' }`，或直接信任后端顺序删除二次排序。 |
| G2 | `src/views/DownloadStartedView.vue:42-56` | **自动下载靠 `window.location.href = download_url`**：这是无用户手势的整页导航。若响应为 `Content-Disposition: attachment` 现代浏览器会保留页面开始下载，但个别浏览器会整页替换成空白/二进制页；且该 URL 是后端返回的 CDN 地址，本地不可控。`AGENTS.md` 描述的「HEAD 探测候选 URL、失败回退同源直连」逻辑在当前代码中已不存在（文档与实现脱节）。建议保留 `<a :download>` 用户手势按钮为主路径，自动跳转改为在新标签打开或去掉。 |
| G3 | 依赖清单（`package.json`） | **7 项完全未使用的依赖**：`primevue`、`primeicons`、`@primeuix/themes`（均 0 引用）、`chart.js`、`vue-chartjs`（图表实际用 echarts）、`highlight.js`（代码块为纯 `<pre>`）、`@vueuse/core` 仅用 3 文件（在用）。旧版 PrimeVue 曾用于管理端表格，现已全部改为 shadcn 风格。建议删除并跑 `npm uninstall` 清理，缩小体积与攻击面。 |
| G4 | `package.json:36,43` | **devDependencies 冗余**：`sass`（全项目无任何 `.scss/.sass` 文件）、`ws`、`playwright`（无 config/spec）均未使用。playwright 名存实亡，要么补 spec 要么移除。 |
| G5 | `.env.development:1` vs `.env.production:1` | **API 版本不一致**：dev 走 `/api/v1`，prod 走 `/api/v2`。线上已无 v1 端点（v1 信封也没有 `meta` 字段，`api.js:10` 的解包判断在 v1 下不生效，联调结果不可信）。建议 dev 也统一为 `/api/v2`。 |
| G6 | `src/views/StatsView.vue:399` | **带宽每秒轮询**：`/api/v2/bandwidth` 每 1s 一次。站点级防火墙默认 300 req/min/IP，多开页面即触发限流。建议降为 5s/10s，并配合 `document.visibilitychange` 在页面隐藏时暂停轮询。 |
| G7 | `src/App.vue:15` | **路由缓存 key 只用 `route.path`**：同 path 不同 query（如 `/verify?file=a` → `/verify?file=b`）不触发组件重建，`VerifyView` 的 `onMounted(init)` 不会重跑，filePath 停留在旧值。建议 key 用 `route.fullPath`（files 页深度导航不受影响，但 verify 页的重复进入场景需覆盖）。 |
| G8 | `frontend/` 全项目 | **零自动化测试**：核心纯函数（PoW 编解码/求解、命名导航、文件元信息、版本对比、`formatSize`）均无测试保护；`playwright` 装了却无 spec。下载/验证链路是全站最关键逻辑，回归风险高。 |
| G9 | `src/views/AnnouncementDialog.vue:26-33` | **死代码**：`forceShowAnnouncement` 通过 `defineExpose` 暴露但全站无任何调用点，连同公告调试入口一起可删除。 |

### 建议

| # | 位置 | 问题 |
|---|------|------|
| A1 | `src/views/HomeView.vue`、`FilesView.vue`、`StatsView.vue`、`AboutView.vue` | **SEO meta 更新逻辑 4 处重复**：同样的 `document.querySelector('meta[...]') + setAttribute` 手写范式散落各视图。建议抽成 `useSeoMeta(title, desc)` composable 统一维护。 |
| A2 | `src/router/index.js:76` 与各视图 `onMounted` | **标题设置职责重叠**：`beforeEach` 已设 `document.title`，视图里又各自设一次，还各自拼 `- ${nameFull}` 格式。保留一边即可（推荐统一走应用层）。 |
| A3 | `frontend/package.json:4` vs `src/lib/globalConfig.ts:6` | **版本号双份漂移**：`package.json version=3.14.6`、`globalConfig.site.version=3.15.0`、Footer 显示的版本取自后者。建议单一数据源（构建时由 Vite 注入或读 package.json）。 |
| A4 | `src/views/ApiDocsView.vue:179-181` | **method 徽标高饱和配色**（green/blue 满彩）与站内低饱和风格不完全一致；`StatsView.vue:320` 的「访问 7 日均线」使用 `#8b5cf6`（B班视觉规范明确禁用的高饱和紫）。数据可视化小面积点缀可理解为功能色，但 `#8b5cf6` 建议替换为站内色板。 |
| A5 | `src/views/AboutView.vue:254-267` | **`.afdian-rainbow-ring` 高饱和彩虹渐变动画死代码**：`afdianLink` 为空字符串永不渲染，且样式违反「避免高饱和渐变」约定。建议删除整个块。 |
| A6 | `src/style.css:117-118` | **未使用变量**：`:root` 中 `--backdrop-blur: 0px; --backdrop-opacity: 1;` 全项目无引用，可清理。 |
| A7 | `src/views/DownloadStartedView.vue:61-87` | `goBack()` 与 `goToWebsite()` 实现几乎完全相同（仅按钮文案使用位置），建议合并为一个 `leaveToExternalTarget()`。 |
| A8 | `index.html:14,20` | **og:url / twitter:url 硬编码 `https://miawa.cn/`**：站点域名变更时需手动改两处且易漏。建议用相对路径或构建期注入 `globalConfig.site.url`。 |
| A9 | `src/lib/globalConfig.ts:84,92` | `launchers` 里 `NativeLibPlugin` 无 logo、`authlib-injector` 有素材但无配置条目（页面里被显式 `.filter(name !== 'authlib-injector')` 排除）——两处半成品状态，建议补全或明确删除。 |
| A10 | `src/services/api.js:9-14` | **信封解包拦截器无条件提升 `data`**：`'data' in ... && 'meta' in ...` 即解包，但未检查 `error` 字段是否为 null。若后端对业务错误返回 200 + `{data:null,error:{...}}`，前端会静默拿到 `null`，靠 `console.error` 兜底。应改为仅当 `error === null` 时提升。 |
| A11 | `src/layouts/DefaultLayout.vue:54-57` | `matchMedia` 监听在 `onUnmounted` 未移除（顶层组件不倒挂，风险低，但组件库示例应成对清理）。 |
| A12 | `src/views/FilesView.vue:49-75` | **每次进入 files 页全量拉三接口且不缓存**：`/launchers + /latest + /pow/config` 无缓存复用，返回导航即重拉。可考虑组件级 keep-alive 或结果缓存（配合后端 ETag）。 |
| A13 | `vite.config.js:18-48` | **dev/preview 代理 target 硬编码 `https://miawa.cn`**：换本地后端联调需改代码。建议抽取到 `.env`（如 `VITE_PROXY_TARGET`）。 |

---

## 改进建议（可操作方案）

1. **修复版本排序（G1，优先级最高）**：`VersionList.vue` 与 `FilesView.vue` 中 `localeCompare` 统一加 `{ numeric: true, sensitivity: 'base' }`；`authlib-injector` 的过滤保留但给出注释。
2. **下载体验重构（G2）**：落地已在 `AGENTS.md` 描述但代码缺失的能力——对 `download_url` 候选发起 `HEAD`（不带 token）探测，失败回退同源 `/download/...`；自动触发仅用 `window.open(url)` 且非阻断，主路径依赖 `<a :download>` 按钮。
3. **开放重定向收敛（S1）**：新建 `lib/returnTarget.ts`，定义「允许跳转的外部域名白名单」（本项目已维护 LogShare/官方群/ICP 备案），`isExternal && !whitelisted` 一律 `router.push('/')`；后端 `prepare` 的 `return_url` 同样应加服务端域名校验。
4. **依赖清理（G3/G4）**：执行 `npm uninstall primevue primeicons @primeuix/themes chart.js vue-chartjs highlight.js sass ws`；playwright 或补 spec 或移除。
5. **统一 API 版本（G5）**：改 `.env.development` 为 `/api/v2`，同步简化 `vite.config.js` 的代理列表（v1 可删）。
6. **补测试（G8）**：至少为以下纯函数建 `vitest` 单测：`VerifyView` 的 base64url/PBKDF2 求解、`FilesView.fileMeta`/`formatSize`、`navigation.isNavigationActive`、`sponsorConfig` 合计；再补一个 `App.vue` 端到端冒烟（可用已装的 playwright）。
7. **SEO 统一（A1/A2）**：新建 `composables/useMeta.ts`，一次封装 title + description + OG/Twitter 标签的 DOM 写入，各视图只声明数据。
8. **轮询降频（G6）**：`bandwidthTimer` 间隔 1000→5000ms，并监听 `visibilitychange` 适时清除。

---

## 正面亮点（值得保留）

- **配置单点化**：`globalConfig.ts` 集中管理端点、启动器、站点文案、storage key，新增启动器/改文案只动一处，`launcher-info.ts` 查询器设计简洁。
- **PoW 前端实现严谨**：`VerifyView.vue` 的 Web Crypto PBKDF2 求解与后端协议严格对齐（无填充 base64url 编码、难度位数校验、能力探测 `isPowSupported`、cancelled 标志防内存泄漏、空闲时让出事件循环）。`safeStorage.ts` 对隐私模式降级为内存 Map，细节考虑周全。
- **弹窗拦截规避模式**：`openBlankTab()` 在点击同步栈内开占位窗口再异步导航，配合失败回退，是移动端下载链路的正确做法。
- **双布局响应式**：`FilesView.vue` 的桌面/移动两套渲染、长文件名 `break-words` 完整换行、通栏下载按钮，针对移动端真实痛点。
- **安全意识**：所有外部 `<a>` 均带 `rel="noopener noreferrer"`；`Button.vue` 用 radix `Primitive` + CVA 变体体系，可访问性（focus-visible 环、sr-only、aria）覆盖到位。
- **构建/部署约定内化为记忆**：SPA 回退、dev 代理、`outDir` 输出到仓库根 `web/default` 并在注释中说明原因，降低后续维护误操作概率。
- **类型检查完整可用**：`tsconfig strict` + `vue-tsc`，`env.d.ts` 正确声明 `.vue` 模块与 Vite 客户端类型。

---

*本报告基于对 `frontend/src` 全部源码及配置文件的静态审查，动态行为未经浏览器实测验证；建议按 G1 → G2 → S1 顺序处理高风险项。*
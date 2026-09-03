<script setup>
import { onMounted, ref } from 'vue'
import { PhCopy as Copy, PhCheck as Check, PhBooks as BookOpen, PhPlugs as Plug, PhShieldCheck as Shield } from '@phosphor-icons/vue'
import { globalConfig } from '@/lib/globalConfig'

onMounted(() => {
  document.title = `API 文档 - ${globalConfig.site.nameFull}`
})

const activeTab = ref('overview')
const copiedText = ref('')

const copyText = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    copiedText.value = text
    setTimeout(() => (copiedText.value = ''), 2000)
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}

const baseUrl = (globalConfig.site.url || 'https://miawa.cn').replace(/\/$/, '')

// 端点文档数据：与 internal/server/server.go 注册的 v2 路由一致
const endpoints = [
  {
    group: '版本查询',
    items: [
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/launchers',
        title: '获取全部启动器版本',
        description: '返回所有启动器及其版本数组（按 tag 降序）。支持 ETag 条件请求（If-None-Match 返回 304）与 gzip 压缩。'
      },
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/launchers/{launcher}',
        title: '获取单个启动器版本',
        description: '返回指定启动器的版本数组。启动器不存在时返回 404 错误信封。',
        params: [{ name: 'launcher', type: 'path', required: true, desc: '启动器标识，如 fcl、zl2、hmcl' }]
      },
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/latest',
        title: '获取全部最新版本',
        description: '返回启动器名到最新版本号的映射。响应头 X-Latest-Versions 携带同一映射（兼容旧客户端）。'
      },
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/latest/{launcher}',
        title: '获取启动器最新版本（纯文本）',
        description: '以纯文本返回最新版本号，不使用信封，适合脚本直接取值。',
        params: [{ name: 'launcher', type: 'path', required: true, desc: '启动器标识' }],
        response: {
          success: { code: 200, type: 'text/plain', example: '1.3.0.7' }
        }
      }
    ]
  },
  {
    group: '站点数据',
    items: [
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/stats',
        title: '获取站点统计',
        description: '返回访问/下载统计、磁盘占用、热门下载排行、地理分布与近 30 天趋势。'
      },
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/bandwidth',
        title: '获取实时带宽',
        description: '返回服务器当前带宽、峰值、近 1 分钟下载连接数与累计传输量。'
      }
    ]
  },
  {
    group: '下载链路',
    items: [
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/pow/config',
        title: '获取 PoW 配置',
        description: '返回下载验证（Proof of Work）开关与求解参数。enabled 为 false 时可直接调用 prepare 获取下载令牌。'
      },
      {
        method: 'POST',
        methodType: 'post',
        path: '/api/v2/downloads/prepare',
        title: '准备下载（无 PoW）',
        description: '为指定文件直接签发一次性下载授权（CLI/API 场景）。令牌 5 分钟有效、单次消费，下载完成后即失效。',
        contentType: 'application/json',
        params: [
          { name: 'file_path', type: 'body', required: true, desc: '文件相对路径，格式 启动器/版本/文件名' },
          { name: 'return_url', type: 'body', required: false, desc: '下载完成后的返回地址，landing 页会原样带回' },
          { name: 'source', type: 'body', required: false, desc: '来源标识，用于统计分析' }
        ]
      },
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/downloads/challenge',
        title: '获取 PoW 挑战',
        description: '为指定文件创建 PoW 挑战。挑战与文件路径、客户端 IP 双重绑定（HMAC 签名保护），默认 2 分钟有效。',
        params: [{ name: 'file_path', type: 'query', required: true, desc: '文件相对路径，同 prepare' }]
      },
      {
        method: 'POST',
        methodType: 'post',
        path: '/api/v2/downloads/authorize',
        title: '提交 PoW 解并领取授权',
        description: '校验 PoW 解，通过后签发与 prepare 相同结构的一次性下载授权。挑战必须未被消费、未过期，且请求 IP 与创建挑战时一致。验证并发上限 4，占满返回 503 pow_busy。',
        contentType: 'application/json',
        params: [
          { name: 'challenge', type: 'body', required: true, desc: 'challenge 接口返回的完整挑战对象（parameters + signature）' },
          { name: 'solution', type: 'body', required: true, desc: '计数器对象：counter 为满足难度的迭代次数，derivedKey 为无填充 base64url 的 PBKDF2 派生密钥' }
        ]
      },
      {
        method: 'GET',
        methodType: 'get',
        path: '/api/v2/downloads/landing',
        title: '获取下载引导信息',
        description: '以 Peek 模式校验令牌并返回下载信息，不消费授权。供下载确认页展示文件名、来源与返回地址。',
        params: [{ name: 'token', type: 'query', required: true, desc: '下载令牌' }]
      },
      {
        method: 'GET',
        methodType: 'get',
        path: '/download/{file_path}',
        title: '下载文件',
        description: '实际文件下载端点，支持 Range 分段。令牌通过查询参数或 Authorization: Bearer 传入，首次成功响应时被消费（HEAD 不消费授权、不计流量）。无令牌的浏览器请求会跳转到 PoW 验证页。',
        params: [
          { name: 'file_path', type: 'path', required: true, desc: '文件相对路径 启动器/版本/文件名' },
          { name: 'token', type: 'query', required: true, desc: '一次性下载令牌' }
        ],
        response: {
          success: { code: 200, type: 'application/octet-stream', example: '（二进制文件流）' }
        }
      }
    ]
  }
]

const envelopeExample = `{
    "data": { ... },        // 成功时为业务数据，失败为 null
    "error": null,          // 失败时为 { "code": "...", "message": "..." }
    "meta": {
        "version": "v2",
        "timestamp": "2026-09-02T12:00:00Z",
        "request_id": "a1b2c3d4e5f6g7h8",
        "cached": false
    }
}`

const quickStartExample = `# 1. 查询启动器最新版本
curl ${baseUrl}/api/v2/latest
# → {"data":{"fcl":"1.3.0.7",...},...}

# 2. 签发下载授权（file_path = 启动器/版本/文件名）
curl -X POST ${baseUrl}/api/v2/downloads/prepare \\
     -H 'Content-Type: application/json' \\
     -d '{"file_path":"fcl/1.3.0.7/FCL-release.apk","source":"cli"}'
# → {"data":{"download_token":"...","download_url":"..."},...}

# 3. 携带令牌下载（单次消费，5 分钟内有效）
curl -LO "${baseUrl}/download/fcl/1.3.0.7/FCL-release.apk?token=<令牌>"`

const methodClass = (type) => {
  const classes = {
    get: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    post: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
  }
  return classes[type] || 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
}

// prepare/authorize 成功响应共用同一结构
const tokenResponseExample = `{
    "data": {
        "download_token": "43字符base64url令牌",
        "download_url": "/download/fcl/1.3.0.7/...?token=...",
        "landing_url": "/api/v2/downloads/landing?token=..."
    },
    "error": null,
    "meta": { ... }
}`

const defaultResponse = (ep) => {
  if (ep.response) return ep.response
  if (ep.path.includes('prepare') || ep.path.includes('authorize')) {
    return {
      success: { code: 200, example: tokenResponseExample },
      error: { example: '{\n    "data": null,\n    "error": { "code": "invalid_file_path", "message": "..." },\n    "meta": { ... }\n}' }
    }
  }
  if (ep.path.includes('challenge')) {
    return {
      success: {
        code: 200,
        example: `{
    "data": {
        "parameters": {
            "algorithm": "PBKDF2-SHA256",
            "nonce": "hex挑战ID",
            "salt": "base64url盐值",
            "cost": 500,
            "keyLength": 32,
            "difficulty": 14,
            "expiresAt": 1756819200,
            "data": { "file_path": "...", "client_ip": "...", "source_kind": "web" }
        },
        "signature": "hex HMAC-SHA256"
    },
    "error": null,
    "meta": { ... }
}`
      }
    }
  }
  if (ep.path.includes('landing')) {
    return {
      success: {
        code: 200,
        example: `{
    "data": {
        "download_url": "/download/...?token=...",
        "return_url": "https://example.com/back",
        "source": "verify-download",
        "file_name": "FCL-release.apk",
        "file_path": "fcl/1.3.0.7/FCL-release.apk",
        "flow": "pow"
    },
    "error": null,
    "meta": { ... }
}`
      }
    }
  }
  return { success: { code: 200, example: envelopeExample } }
}

const curlExample = (ep) => {
  const path = ep.path.replace('{launcher}', 'fcl').replace('{file_path}', 'fcl/1.3.0.7/FCL-release.apk')
  if (ep.method === 'POST') {
    return `curl -X POST ${baseUrl}${path} \\\n  -H "Content-Type: application/json" \\\n  -d '{"file_path":"fcl/1.3.0.7/FCL-release.apk","source":"cli"}'`
  }
  return `curl ${baseUrl}${path}`
}
</script>

<template>
  <div class="mx-auto w-full max-w-4xl min-w-0">
    <!-- 页面标题 -->
    <header class="mb-8">
      <h1 class="mb-2 text-3xl font-bold tracking-tight">API 文档</h1>
      <p class="text-sm leading-relaxed text-muted-foreground">
        柠泽资源站公开 API（<strong class="text-foreground">v2</strong>），供启动器与第三方工具集成查询版本与下载。
      </p>
    </header>

    <!-- 导航标签 -->
    <div class="mb-8 flex flex-wrap gap-2 border-b border-border">
      <button
        v-for="tab in [
          { key: 'overview', label: '概述' },
          { key: 'endpoints', label: 'API 端点' },
          { key: 'limits', label: '限制说明' }
        ]"
        :key="tab.key"
        :class="[
          'border-b-2 px-3 py-2 text-sm font-medium transition-colors',
          activeTab === tab.key
            ? 'border-primary text-foreground'
            : 'border-transparent text-muted-foreground hover:text-foreground'
        ]"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 概述 -->
    <div v-if="activeTab === 'overview'" class="space-y-6">
      <section class="space-y-4">
        <h2 class="flex items-center gap-2 text-lg font-semibold">
          <BookOpen weight="duotone" class="h-5 w-5 text-primary" />
          快速接入
        </h2>
        <p class="text-sm leading-relaxed text-muted-foreground">
          最小下载流程：调用
          <code class="break-all rounded bg-muted px-1.5 py-0.5 font-mono text-xs">prepare</code>
          获取一次性令牌，5 分钟内携带令牌访问
          <code class="break-all rounded bg-muted px-1.5 py-0.5 font-mono text-xs">download_url</code>
          即可下载。浏览器用户在 PoW 启用时需先完成 challenge / authorize 验证。
        </p>
        <div class="min-w-0 overflow-hidden rounded-lg border border-border">
          <div class="border-b border-border bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
            cURL
          </div>
          <pre class="max-w-full overflow-x-auto bg-slate-950 p-4 text-xs leading-relaxed text-slate-50"><code>{{ quickStartExample }}</code></pre>
        </div>
      </section>

      <section class="space-y-4">
        <h2 class="flex items-center gap-2 text-lg font-semibold">
          <Plug weight="duotone" class="h-5 w-5 text-primary" />
          API 基础信息
        </h2>

        <div class="grid gap-4 sm:grid-cols-2">
          <div class="rounded-lg border border-border bg-card p-4">
            <div class="mb-1 text-xs text-muted-foreground">基础 URL</div>
            <div class="flex items-center justify-between gap-2 font-mono text-sm">
              <span class="break-all">{{ baseUrl }}</span>
              <button
                class="shrink-0 text-muted-foreground transition-colors hover:text-foreground"
                aria-label="复制基础 URL"
                @click="copyText(baseUrl)"
              >
                <Copy v-if="copiedText !== baseUrl" weight="duotone" class="h-3.5 w-3.5" />
                <Check v-else weight="duotone" class="h-3.5 w-3.5" />
              </button>
            </div>
          </div>

          <div class="rounded-lg border border-border bg-card p-4">
            <div class="mb-1 text-xs text-muted-foreground">API 版本</div>
            <div class="font-mono text-sm">v2</div>
          </div>

          <div class="rounded-lg border border-border bg-card p-4">
            <div class="mb-1 text-xs text-muted-foreground">协议</div>
            <div class="text-sm">HTTPS</div>
          </div>

          <div class="rounded-lg border border-border bg-card p-4">
            <div class="mb-1 text-xs text-muted-foreground">认证</div>
            <div class="text-sm">无需认证（公共 API）</div>
          </div>
        </div>

        <div class="rounded-r-lg border-l-4 border-amber-500 bg-amber-50 p-4 dark:bg-amber-950/30">
          <p class="text-sm text-amber-800 dark:text-amber-200">
            所有 JSON 端点使用统一的信封结构；唯一例外是
            <code class="break-all font-mono text-xs">/api/v2/latest/{launcher}</code>
            返回纯文本。GET 查询端点支持 ETag 条件请求与 gzip 压缩。
          </p>
        </div>
      </section>

      <section class="space-y-4">
        <h2 class="text-lg font-semibold">响应信封</h2>
        <div class="min-w-0 overflow-hidden rounded-lg border border-border">
          <div class="border-b border-border bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
            统一响应结构
          </div>
          <pre class="max-w-full overflow-x-auto bg-slate-950 p-4 text-xs leading-relaxed text-slate-50"><code>{{ envelopeExample }}</code></pre>
        </div>
      </section>
    </div>

    <!-- API 端点 -->
    <div v-if="activeTab === 'endpoints'" class="space-y-10">
      <section v-for="group in endpoints" :key="group.group" class="space-y-4">
        <h2 class="text-lg font-semibold">{{ group.group }}</h2>

        <div v-for="ep in group.items" :key="ep.path" class="space-y-4">
          <div class="rounded-lg border border-border bg-card">
            <!-- 端点头 -->
            <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-3">
              <div class="flex min-w-0 flex-wrap items-center gap-2.5">
                <span class="shrink-0 rounded px-2 py-1 text-xs font-bold" :class="methodClass(ep.methodType)">
                  {{ ep.method }}
                </span>
                <code class="break-all font-mono text-sm">{{ ep.path }}</code>
              </div>
              <button
                class="flex shrink-0 items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
                @click="copyText(`${baseUrl}${ep.path}`)"
              >
                <Copy v-if="copiedText !== `${baseUrl}${ep.path}`" weight="duotone" class="h-3.5 w-3.5" />
                <Check v-else weight="duotone" class="h-3.5 w-3.5" />
                {{ copiedText === `${baseUrl}${ep.path}` ? '已复制' : '复制' }}
              </button>
            </div>

            <div class="space-y-4 p-4">
              <div>
                <h3 class="text-sm font-semibold">{{ ep.title }}</h3>
                <p class="mt-1 text-sm leading-relaxed text-muted-foreground">{{ ep.description }}</p>
              </div>

              <div v-if="ep.contentType" class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span class="font-medium">Content-Type:</span>
                <code class="break-all rounded bg-muted px-1.5 py-0.5">{{ ep.contentType }}</code>
              </div>

              <!-- 请求参数 -->
              <div v-if="ep.params && ep.params.length" class="space-y-2">
                <h4 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">请求参数</h4>
                <div class="space-y-2">
                  <div
                    v-for="param in ep.params"
                    :key="param.name"
                    class="rounded-lg border border-border/60 bg-background p-3"
                  >
                    <div class="flex flex-wrap items-baseline gap-x-2 gap-y-1">
                      <code class="break-all font-mono text-xs font-semibold text-primary">{{ param.name }}</code>
                      <code class="rounded bg-muted px-1.5 py-0.5 text-[10px]">{{ param.type }}</code>
                      <span v-if="param.required" class="text-[10px] font-medium text-destructive">必需</span>
                      <span v-else class="text-[10px] text-muted-foreground">可选</span>
                    </div>
                    <p class="mt-1 break-words text-xs leading-relaxed text-muted-foreground">{{ param.desc }}</p>
                  </div>
                </div>
              </div>

              <!-- 调用示例 -->
              <div class="min-w-0 space-y-1.5">
                <h4 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">调用示例</h4>
                <div class="min-w-0 overflow-hidden rounded-lg border border-border">
                  <div class="border-b border-border bg-muted/50 px-3 py-1.5 text-xs text-muted-foreground">cURL</div>
                  <pre class="max-w-full overflow-x-auto bg-slate-950 p-3.5 text-xs leading-relaxed text-slate-50"><code>{{ curlExample(ep) }}</code></pre>
                </div>
              </div>

              <!-- 响应示例 -->
              <div class="min-w-0 space-y-1.5">
                <h4 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">响应示例</h4>
                <div class="min-w-0 overflow-hidden rounded-lg border border-border">
                  <div class="flex items-center justify-between gap-2 border-b border-border bg-muted/50 px-3 py-1.5 text-xs text-muted-foreground">
                    <span>成功响应 {{ defaultResponse(ep).success.code ? `(${defaultResponse(ep).success.code} OK)` : '' }}</span>
                    <span v-if="defaultResponse(ep).success.type" class="break-all">{{ defaultResponse(ep).success.type }}</span>
                  </div>
                  <pre class="max-w-full overflow-x-auto bg-slate-950 p-3.5 text-xs leading-relaxed text-slate-50"><code>{{ defaultResponse(ep).success.example }}</code></pre>
                </div>
                <div v-if="defaultResponse(ep).error" class="min-w-0 overflow-hidden rounded-lg border border-border">
                  <div class="border-b border-border bg-muted/50 px-3 py-1.5 text-xs text-muted-foreground">错误响应</div>
                  <pre class="max-w-full overflow-x-auto bg-slate-950 p-3.5 text-xs leading-relaxed text-slate-50"><code>{{ defaultResponse(ep).error.example }}</code></pre>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- 限制说明 -->
    <div v-if="activeTab === 'limits'" class="space-y-4">
      <section class="space-y-4">
        <h2 class="flex items-center gap-2 text-lg font-semibold">
          <Shield weight="duotone" class="h-5 w-5 text-primary" />
          使用限制
        </h2>

        <div class="rounded-lg border border-border bg-card p-4">
          <h3 class="text-sm font-semibold">频率限制</h3>
          <p class="mt-1 break-words text-sm leading-relaxed text-muted-foreground">
            站点级防火墙按 IP 做每分钟请求计数，超限返回
            <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">429</code>
            并附带
            <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">Retry-After</code>
            响应头；频繁触发限流会被自动封禁。请对轮询类端点做合理缓存——响应已带 ETag 与 Cache-Control。
          </p>
        </div>

        <div class="rounded-lg border border-border bg-card p-4">
          <h3 class="text-sm font-semibold">下载令牌</h3>
          <p class="mt-1 break-words text-sm leading-relaxed text-muted-foreground">
            令牌为 43 字符 base64url 字符串，<strong class="text-foreground">单次消费</strong>，签发后 5 分钟内有效；下载完成（非 HEAD、响应 200/206）后立即作废。授权与客户端 IP 绑定，请勿跨 IP 复用。
          </p>
        </div>

        <div class="rounded-lg border border-border bg-card p-4">
          <h3 class="text-sm font-semibold">PoW 验证</h3>
          <p class="mt-1 break-words text-sm leading-relaxed text-muted-foreground">
            PoW 启用时浏览器下载需先通过 challenge / authorize 验证：求解 PBKDF2-SHA256 直到派生密钥前导零位数 ≥ difficulty（默认 14）。挑战 2 分钟有效且与文件路径、IP 双重绑定；验证并发上限 4，占满返回
            <code class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">503 pow_busy</code>，请稍后重试。
          </p>
        </div>

        <div class="rounded-lg border border-border bg-card p-4">
          <h3 class="text-sm font-semibold">管理接口</h3>
          <p class="mt-1 break-words text-sm leading-relaxed text-muted-foreground">
            <code class="break-all rounded bg-muted px-1.5 py-0.5 font-mono text-xs">/api/v2/admin/*</code>
            与
            <code class="break-all rounded bg-muted px-1.5 py-0.5 font-mono text-xs">/api/v2/auth/login</code>
            为管理端专用（Bearer 令牌认证），不对第三方开放，文档不予覆盖。
          </p>
        </div>
      </section>
    </div>
  </div>
</template>

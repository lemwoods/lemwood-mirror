import { globalConfig } from '@/lib/globalConfig'
import { friendLinksConfig } from '@/lib/friendLinksConfig'

// 「返回来源」外部跳转域名白名单：本站 + 友链配置中的站点。
// 用于收敛 /download-started?return_url=... 的开放重定向面——
// 白名单之外的站外地址一律不跳，回退站内路由。
const allowedHosts = (() => {
  const hosts = new Set()
  const add = (url) => {
    try {
      hosts.add(new URL(url).host)
    } catch {
      // 忽略配置中的非法 URL
    }
  }
  add(globalConfig.site.url)
  for (const link of friendLinksConfig.links || []) add(link.url)
  return hosts
})()

export function isSameOriginUrl(url) {
  try {
    return new URL(url, window.location.origin).origin === window.location.origin
  } catch {
    return false
  }
}

// 仅当 url 是站外地址且命中白名单时返回 true；站内地址返回 false（应走站内路由）。
export function isAllowedExternalTarget(url) {
  if (!url) return false
  try {
    const parsed = new URL(url, window.location.origin)
    if (parsed.origin === window.location.origin) return false
    return allowedHosts.has(parsed.host)
  } catch {
    return false
  }
}

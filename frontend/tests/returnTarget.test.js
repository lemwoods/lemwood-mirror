import { describe, it, expect, beforeAll } from 'vitest'
import { isAllowedExternalTarget, isSameOriginUrl } from '@/lib/returnTarget'

// node 测试环境无 window，stub 浏览器 origin（本站）供 URL 解析
beforeAll(() => {
  globalThis.window = { location: { origin: 'https://miawa.cn' } }
})

// 白名单数据源：globalConfig.site.url（miawa.cn）+ friendLinksConfig.links
// （logshare.cn / cc.miawa.cn / nexusmc.cn）。

describe('isAllowedExternalTarget（防开放重定向）', () => {
  it('站内地址返回 false（应走站内路由）', () => {
    expect(isAllowedExternalTarget('https://miawa.cn/files')).toBe(false)
    expect(isAllowedExternalTarget('/files')).toBe(false)
    expect(isAllowedExternalTarget('')).toBe(false)
    expect(isAllowedExternalTarget(null)).toBe(false)
  })

  it('白名单外的外部地址返回 false', () => {
    expect(isAllowedExternalTarget('https://evil.com/phish')).toBe(false)
    expect(isAllowedExternalTarget('https://miawa.cn.evil.com/')).toBe(false)
    expect(isAllowedExternalTarget('javascript:alert(1)')).toBe(false)
  })

  it('白名单内的外部地址返回 true', () => {
    expect(isAllowedExternalTarget('https://logshare.cn/some/page')).toBe(true)
    expect(isAllowedExternalTarget('https://cc.miawa.cn/')).toBe(true)
    expect(isAllowedExternalTarget('https://www.nexusmc.cn/thread/1')).toBe(true)
  })

  it('非法 URL 返回 false', () => {
    expect(isAllowedExternalTarget('::::not-a-url')).toBe(false)
  })
})

describe('isSameOriginUrl', () => {
  it('相对路径按当前 origin 解析为同源', () => {
    expect(isSameOriginUrl('/download/fcl/1.0/x.apk?token=abc')).toBe(true)
    expect(isSameOriginUrl('files/apk')).toBe(true)
  })

  it('跨源地址返回 false', () => {
    expect(isSameOriginUrl('https://example.com/x')).toBe(false)
  })

  it('非法 URL 返回 false', () => {
    // WHATWG URL 会把多数字符串当相对路径解析，只有 schema 非法等场景才抛错
    expect(isSameOriginUrl('https://[::1:bad')).toBe(false)
  })
})

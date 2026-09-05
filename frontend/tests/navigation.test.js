import { describe, it, expect } from 'vitest'
import { isNavigationActive } from '@/lib/navigation'

describe('isNavigationActive', () => {
  const home = { path: '/', match: '/' }
  const files = { path: '/files', match: '/files' }
  const stats = { path: '/stats', match: '/stats' }

  it('根路径仅精确匹配，子路径不算首页', () => {
    expect(isNavigationActive('/', home)).toBe(true)
    expect(isNavigationActive('/files', home)).toBe(false)
  })

  it('子路径归属父导航', () => {
    expect(isNavigationActive('/files', files)).toBe(true)
    expect(isNavigationActive('/files/fcl', files)).toBe(true)
    expect(isNavigationActive('/files/fcl/1.3.0', files)).toBe(true)
  })

  it('不同导航互不误伤', () => {
    expect(isNavigationActive('/stats', files)).toBe(false)
    expect(isNavigationActive('/files', stats)).toBe(false)
    // 前缀相同但不带斜杠的路径（如 /filesx）不得命中 /files
    expect(isNavigationActive('/filesx', files)).toBe(false)
  })
})

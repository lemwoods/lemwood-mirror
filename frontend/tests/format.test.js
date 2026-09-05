import { describe, it, expect } from 'vitest'
import { formatSize, compareVersionDesc } from '@/lib/format'

describe('formatSize', () => {
  it('非法与零值归 0 B', () => {
    expect(formatSize(0)).toBe('0 B')
    expect(formatSize(-1)).toBe('0 B')
    expect(formatSize(NaN)).toBe('0 B')
    expect(formatSize(undefined)).toBe('0 B')
  })

  it('字节无小数', () => {
    expect(formatSize(1)).toBe('1 B')
    expect(formatSize(1023)).toBe('1023 B')
  })

  it('单位换算保留一位小数', () => {
    expect(formatSize(1024)).toBe('1.0 KB')
    expect(formatSize(1536)).toBe('1.5 KB')
    expect(formatSize(1024 * 1024)).toBe('1.0 MB')
    expect(formatSize(5 * 1024 ** 3)).toBe('5.0 GB')
  })

  it('超大值不崩溃且含 EB 单位', () => {
    expect(formatSize(Number.MAX_VALUE).endsWith('EB')).toBe(true)
  })
})

describe('compareVersionDesc（版本降序，数值语义）', () => {
  const mk = (tag) => ({ tag_name: tag })

  it('1.10.0 应排在 1.9.0 之前（字典序会排反的用例）', () => {
    const arr = [mk('1.9.0'), mk('1.10.0')]
    arr.sort(compareVersionDesc)
    expect(arr.map((v) => v.tag_name)).toEqual(['1.10.0', '1.9.0'])
  })

  it('完整降序排列', () => {
    const arr = [mk('2.0.0'), mk('1.9.9'), mk('1.10.0'), mk('1.2.0'), mk('1.10.1')]
    arr.sort(compareVersionDesc)
    expect(arr.map((v) => v.tag_name)).toEqual(['2.0.0', '1.10.1', '1.10.0', '1.9.9', '1.2.0'])
  })

  it('缺 tag_name 时回退 name', () => {
    const arr = [{ name: '1.9.0' }, { tag_name: '1.10.0' }]
    arr.sort(compareVersionDesc)
    expect(arr[0].tag_name).toBe('1.10.0')
  })
})

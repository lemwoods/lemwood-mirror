import { describe, it, expect } from 'vitest'
import { base64urlEncode, base64urlDecode } from '@/lib/pow'

describe('base64url 编解码（与后端 base64.RawURLEncoding 约定一致）', () => {
  it('无填充 base64url 编码', () => {
    // RFC 4648 测试向量：base64 "foobar" → base64url 无填充
    const bytes = new TextEncoder().encode('foobar')
    expect(base64urlEncode(bytes)).toBe('Zm9vYmFy')
    // "+"→"-"、"/"→"_"、去填充
    const bin = new Uint8Array([0xfb, 0xff, 0xbf]) // 标准 base64: +/+/
    expect(base64urlEncode(bin)).toBe('-_-_')
    const pad = new Uint8Array([0xff, 0xff]) // 标准 base64: //8= → __8
    expect(base64urlEncode(pad)).toBe('__8')
  })

  it('解码含填充/标准字符变体', () => {
    expect(Array.from(base64urlDecode('Zm9vYmFy'))).toEqual(Array.from(new TextEncoder().encode('foobar')))
    expect(Array.from(base64urlDecode('-_-_'))).toEqual([0xfb, 0xff, 0xbf])
    expect(Array.from(base64urlDecode('__8'))).toEqual([0xff, 0xff])
  })

  it('编解码往返', () => {
    for (let len = 0; len < 40; len++) {
      const bytes = new Uint8Array(len)
      crypto.getRandomValues(bytes)
      expect(Array.from(base64urlDecode(base64urlEncode(bytes)))).toEqual(Array.from(bytes))
    }
  })
})

import { describe, it, expect } from 'vitest'
import { leadingZeroBits } from '@/lib/pow'

// atob/btoa 在 node 环境全局可用（Node 16+）
describe('leadingZeroBits', () => {
  it('全零字节计满 8 位每字节', () => {
    expect(leadingZeroBits(new Uint8Array([0, 0]))).toBe(16)
  })

  it('最高位为 1 时为 0', () => {
    expect(leadingZeroBits(new Uint8Array([0b10000000]))).toBe(0)
    expect(leadingZeroBits(new Uint8Array([0xff]))).toBe(0)
  })

  it('按最高置位位计前导零', () => {
    expect(leadingZeroBits(new Uint8Array([0b00000001]))).toBe(7)
    expect(leadingZeroBits(new Uint8Array([0b00010000]))).toBe(3)
    expect(leadingZeroBits(new Uint8Array([0b00000100]))).toBe(5)
  })

  it('跨字节累计', () => {
    expect(leadingZeroBits(new Uint8Array([0x00, 0b01000000]))).toBe(9)
    expect(leadingZeroBits(new Uint8Array([0x00, 0x00, 0b00000011]))).toBe(22)
  })

  it('空输入为 0', () => {
    expect(leadingZeroBits(new Uint8Array([]))).toBe(0)
  })
})

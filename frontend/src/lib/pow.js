// PoW 求解工具（与后端 internal/pow 协议一致）：base64url 编解码 + 前导零位统计。
// 纯函数抽出便于单元测试；求解主循环保留在 VerifyView（依赖 Web Crypto 与取消标志）。

export function base64urlDecode(s) {
  s = s.replace(/-/g, '+').replace(/_/g, '/')
  while (s.length % 4) s += '='
  const bin = atob(s)
  const a = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) a[i] = bin.charCodeAt(i)
  return a
}

export function base64urlEncode(bytes) {
  let b = ''
  for (let i = 0; i < bytes.length; i++) b += String.fromCharCode(bytes[i])
  return btoa(b).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function leadingZeroBits(bytes) {
  let bits = 0
  for (let i = 0; i < bytes.length; i++) {
    const b = bytes[i]
    if (b === 0) {
      bits += 8
      continue
    }
    for (let j = 7; j >= 0; j--) {
      if (b & (1 << j)) return bits
      bits++
    }
  }
  return bits
}

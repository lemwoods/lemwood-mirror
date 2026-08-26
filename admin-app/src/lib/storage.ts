// 浏览器存储安全封装：隐私模式/受限 WebView 下 localStorage 访问会抛异常，
// 统一降级为内存 Map，避免应用初始化阶段崩溃。

const memoryFallback = new Map<string, string>()

let available: boolean | null = null

function detectAvailable(): boolean {
  if (available !== null) return available
  try {
    const probeKey = '__storage_probe__'
    window.localStorage.setItem(probeKey, '1')
    window.localStorage.removeItem(probeKey)
    available = true
  } catch {
    available = false
  }
  return available
}

export function getStoredItem(key: string): string | null {
  if (detectAvailable()) {
    try {
      return localStorage.getItem(key)
    } catch {
      // 落入内存兜底
    }
  }
  return memoryFallback.has(key) ? (memoryFallback.get(key) as string) : null
}

export function setStoredItem(key: string, value: string): void {
  memoryFallback.set(key, value)
  if (detectAvailable()) {
    try {
      localStorage.setItem(key, value)
    } catch {
      // 忽略配额/权限错误，内存已兜底
    }
  }
}

export function removeStoredItem(key: string): void {
  memoryFallback.delete(key)
  if (detectAvailable()) {
    try {
      localStorage.removeItem(key)
    } catch {
      // 忽略
    }
  }
}

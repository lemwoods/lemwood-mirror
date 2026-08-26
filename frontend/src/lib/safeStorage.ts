// 浏览器存储安全封装：
// 隐私模式 / WebView 禁用 localStorage 时，直接读写会抛 SecurityError 导致页面初始化失败，
// 这里统一降级为内存 Map，保证功能可用（刷新不持久是可接受的降级）。

const memoryFallback = new Map()

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

export interface OpenedTab {
  navigate(url: string): void
  close(): void
}

// 在用户点击事件的同步栈内打开空白窗口（规避弹出窗口拦截），
// 异步请求完成后再导航到目标地址；若 opener 为空则退化为直接 window.open。
export function openBlankTab(): OpenedTab | null {
  try {
    const win = window.open('', '_blank')
    if (!win) return null
    return {
      navigate(url: string) {
        try {
          if (!win.closed) {
            win.location.href = url
            return
          }
        } catch {
          // 忽略跨域导航异常
        }
        window.open(url, '_blank')
      },
      close() {
        try {
          if (!win.closed) win.close()
        } catch {
          // 忽略
        }
      },
    }
  } catch {
    return null
  }
}

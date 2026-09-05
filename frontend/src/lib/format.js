// 展示格式化纯函数（抽出便于复用与测试）

export function formatSize(bytes) {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

// 版本号降序比较：numeric 语义（1.10.0 > 1.9.0），供版本列表排序共用
export function compareVersionDesc(a, b) {
  return String(b.tag_name || b.name).localeCompare(String(a.tag_name || a.name), undefined, {
    numeric: true
  })
}

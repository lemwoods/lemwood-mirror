import { PhHouse, PhFolder, PhChartBar, PhFileText, PhInfo } from '@phosphor-icons/vue'

export const navigationLinks = [
  { name: '首页', path: '/', icon: PhHouse, match: '/' },
  { name: '文件浏览', path: '/files', icon: PhFolder, match: '/files' },
  { name: '数据统计', path: '/stats', icon: PhChartBar, match: '/stats' },
  { name: 'API 文档', path: '/apidocs', icon: PhFileText, match: '/apidocs' },
  { name: '关于', path: '/about', icon: PhInfo, match: '/about' }
]

export function isNavigationActive(routePath, link) {
  if (link.match === '/') return routePath === '/'
  return routePath === link.path || routePath.startsWith(`${link.match}/`)
}

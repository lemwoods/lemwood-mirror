<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { getStats, getBandwidth } from '@/services/api'
import { useDark } from '@vueuse/core'
import { globalConfig } from '@/lib/globalConfig'
import {
  PhActivity as Activity,
  PhArrowDown as ArrowDown,
  PhArrowUp as ArrowUp,
  PhChartBar as BarChart3,
  PhGauge as Gauge,
  PhMapPin as MapPin,
  PhTrendUp as TrendingUp
} from '@phosphor-icons/vue'
import Card from '@/components/ui/Card.vue'
import CardContent from '@/components/ui/CardContent.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import Skeleton from '@/components/ui/Skeleton.vue'
import { getLauncherDisplayName } from '@/lib/launcher-info'

import { use } from 'echarts/core'
import { BarChart, TreemapChart, LineChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  VisualMapComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer, BarChart, TreemapChart, LineChart,
  TitleComponent, TooltipComponent, LegendComponent, GridComponent, VisualMapComponent
])

const stats = ref({})
const bandwidth = ref({})
const loading = ref(true)
const isDark = useDark()
let bandwidthTimer = null

const formatMbps = (v) => {
  if (!Number.isFinite(v)) return '0'
  return v.toFixed(2)
}

const bandwidthUtilization = computed(() => {
  const u = bandwidth.value.utilization_percent
  if (!Number.isFinite(u)) return 0
  return Math.min(100, Math.max(0, Math.round(u)))
})

const isBandwidthIdle = computed(() => {
  const v = bandwidth.value.current_bandwidth_mbps
  return !Number.isFinite(v) || v < 0.01
})

const formatBytes = (bytes) => {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
  const i = Math.max(0, Math.min(Math.floor(Math.log(bytes) / Math.log(k)), sizes.length - 1))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const diskPercentage = computed(() => {
  if (!stats.value.disk || !stats.value.disk.total) return 0
  return Math.round((stats.value.disk.used / stats.value.disk.total) * 100)
})

const avgDailyVisits = computed(() => {
  if (!stats.value.total_days || stats.value.total_days === 0) return 0
  return stats.value.total_visits / stats.value.total_days
})

const avgDailyDownloads = computed(() => {
  if (!stats.value.total_days || stats.value.total_days === 0) return 0
  return stats.value.total_downloads / stats.value.total_days
})

const visitGrowth = computed(() => {
  const avg = avgDailyVisits.value
  const recent = (stats.value.last_30_visits || 0) / 30
  if (avg === 0) return 0
  return ((recent - avg) / avg * 100)
})

const downloadGrowth = computed(() => {
  const avg = avgDailyDownloads.value
  const recent = (stats.value.last_30_downloads || 0) / 30
  if (avg === 0) return 0
  return ((recent - avg) / avg * 100)
})

// 热门资源排行：只取 Top5（与左侧地图卡高度对齐），launcher key → 显示名 + logo，并算出占比条数据
const topDownloads = computed(() => {
  const ranks = (stats.value.top_downloads || []).slice(0, 6)
  if (!ranks.length) return []
  const total = ranks.reduce((sum, r) => sum + (r.count || 0), 0) || 1
  const max = Math.max(...ranks.map(r => r.count || 0), 1)
  return ranks.map(r => {
    const info = globalConfig.launchers?.[r.launcher]
    return {
      key: r.launcher,
      name: getLauncherDisplayName(r.launcher),
      logo: info?.logoUrl || '',
      count: r.count || 0,
      percent: Math.round(((r.count || 0) / total) * 1000) / 10,
      barWidth: Math.round(((r.count || 0) / max) * 1000) / 10
    }
  })
})

// 前三名皇冠（三款造型，currentColor 继承徽标配色）：王者冠 / 贵族冠 / 骑士冠
const crownPaths = [
  // 第一名：三尖王冠，中尖最高，带底座横杠
  'M4 17 L3 7.5 L7.5 11 L12 4.5 L16.5 11 L21 7.5 L20 17 Q12 19 4 17 Z M5 19.2 Q12 20.8 19 19.2 L19 20.4 Q12 22 5 20.4 Z M12 2.6 a1.1 1.1 0 1 1 0 2.2 a1.1 1.1 0 1 1 0 -2.2 Z',
  // 第二名：圆弧贵族冠，冠顶三连弧
  'M5.5 16.5 Q4.2 8.5 9 7.2 L10 9.5 Q12 8.2 14 9.5 L15 7.2 Q19.8 8.5 18.5 16.5 Q12 18.2 5.5 16.5 Z M5 18.8 Q12 20.4 19 18.8 L19 20 Q12 21.6 5 20 Z',
  // 第三名：单尖骑士冠，极简三角
  'M6 16.5 L12 6.5 L18 16.5 Q12 18.4 6 16.5 Z M5.6 18.6 Q12 20.2 18.4 18.6 L18.4 19.8 Q12 21.4 5.6 19.8 Z'
]

// 全球访问分布：矩形树图（面积 = 访问量），低占比国家聚合为「其他」
const treemapOption = computed(() => {
  const textColor = isDark.value ? '#a1a1aa' : '#52525b'
  const gapColor = isDark.value ? '#18181b' : '#ffffff'

  const geo = [...(stats.value.geo_distribution || [])].sort((a, b) => (b.count || 0) - (a.count || 0))
  const total = geo.reduce((s, g) => s + (g.count || 0), 0) || 1

  // 占比 < 2% 的国家聚合，避免碎块过多
  const main = []
  let restCount = 0
  for (const g of geo) {
    if ((g.count || 0) / total >= 0.02) main.push(g)
    else restCount += g.count || 0
  }
  const data = main.map(g => ({
    name: g.country,
    value: g.count || 0
  }))
  if (restCount > 0) data.push({ name: '其他', value: restCount })

  // teal 系低饱和色板，末位灰阶给「其他」
  const palette = ['#0d9488', '#14b8a6', '#2dd4bf', '#5eead4', '#99f6e4', '#94a3b8', '#a8a29e', '#cbd5e1']

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark.value ? '#18181b' : '#ffffff',
      borderColor: isDark.value ? '#27272a' : '#e4e4e7',
      textStyle: { color: isDark.value ? '#fafafa' : '#09090b' },
      formatter: (params) => {
        const v = Number(params.value)
        const count = Number.isFinite(v) ? v.toLocaleString() : '0'
        const pct = (Number.isFinite(v) ? (v / total) * 100 : 0).toFixed(1)
        return `<b>${params.name}</b><br/>访问 ${count} 次 · 占比 ${pct}%`
      }
    },
    series: [{
      name: '访问分布',
      type: 'treemap',
      data,
      // 移动端不劫持手势，纯静态展示 + tooltip
      roam: false,
      nodeClick: false,
      breadcrumb: { show: false },
      left: 0,
      right: 0,
      top: 0,
      bottom: 0,
      color: palette,
      itemStyle: {
        borderColor: gapColor,
        borderWidth: 2,
        gapWidth: 2,
        borderRadius: 6
      },
      label: {
        show: true,
        color: isDark.value ? '#fafafa' : '#134e4a',
        fontSize: 11,
        lineHeight: 15,
        fontWeight: 600,
        overflow: 'truncate',
        formatter: (params) => {
          const v = Number(params.value)
          if (!Number.isFinite(v)) return params.name
          const pct = ((v / total) * 100).toFixed(1)
          return `${params.name}\n${pct}%`
        }
      },
      emphasis: {
        label: { show: true },
        itemStyle: {
          areaColor: '#f97316',
          shadowBlur: 12,
          shadowColor: 'rgba(0, 0, 0, 0.25)'
        }
      }
    }]
  }
})

const trendOption = computed(() => {
  const textColor = isDark.value ? '#a1a1aa' : '#52525b'
  const splitLineColor = isDark.value ? '#27272a' : '#e4e4e7'
  // 柱体底色与悬浮层背景（亮暗两套）
  const barBase = isDark.value ? 'rgba(56, 189, 248, 0.35)' : 'rgba(14, 165, 233, 0.25)'
  const tooltipBg = isDark.value ? '#18181b' : '#ffffff'

  if (!stats.value.daily_stats) return {}

  const rawData = [...stats.value.daily_stats].reverse()
  const dates = rawData.map(d => d.date.slice(5))
  const visits = rawData.map(d => d.visit_count || 0)
  const downloads = rawData.map(d => d.download_count || 0)
  const traffic = rawData.map(d => d.traffic_bytes || 0)

  // 7 日移动平均：反映中长期趋势，过滤单日波动（窗口不足时用现有均值）
  const movingAvg = visits.map((_, i) => {
    const win = visits.slice(Math.max(0, i - 6), i + 1)
    return Math.round(win.reduce((s, v) => s + v, 0) / win.length)
  })

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: tooltipBg,
      borderColor: splitLineColor,
      textStyle: { color: isDark.value ? '#fafafa' : '#09090b' },
      formatter: (params) => {
        let html = `<b>${params[0].axisValue}</b>`
        params.forEach(p => {
          const val = p.seriesName === '下载流量' ? formatBytes(p.value) : Number(p.value).toLocaleString()
          html += `<br/>${p.marker} ${p.seriesName}: ${val}`
        })
        return html
      }
    },
    legend: {
      data: ['下载量', '访问量', '访问 7 日均线', '下载流量'],
      textStyle: { color: textColor, fontSize: 11 },
      itemWidth: 14,
      itemHeight: 8,
      bottom: 0
    },
    grid: {
      left: '10px', right: '10px', bottom: '32px', top: '20px', containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLine: { lineStyle: { color: splitLineColor } },
      axisTick: { show: false },
      axisLabel: { color: textColor, fontSize: 10 }
    },
    yAxis: [
      {
        type: 'value',
        name: '次数',
        nameTextStyle: { color: textColor, fontSize: 10, padding: [0, 0, 0, -12] },
        splitLine: { lineStyle: { type: 'dashed', color: splitLineColor } },
        axisLine: { show: false },
        axisLabel: { color: textColor, fontSize: 10 }
      },
      {
        type: 'value',
        name: '流量',
        nameTextStyle: { color: textColor, fontSize: 10, padding: [0, -20, 0, 0] },
        splitLine: { show: false },
        axisLine: { show: false },
        axisLabel: {
          color: textColor,
          fontSize: 10,
          formatter: (v) => formatBytes(v)
        }
      }
    ],
    series: [
      // 下载量：渐变圆角柱（底层）
      {
        name: '下载量',
        type: 'bar',
        barMaxWidth: 14,
        data: downloads,
        itemStyle: {
          borderRadius: [4, 4, 0, 0],
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: '#0ea5e9' },
              { offset: 1, color: barBase }
            ]
          }
        }
      },
      // 访问量 7 日均线：细虚线，压在柱上表现趋势
      {
        name: '访问 7 日均线',
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: movingAvg,
        lineStyle: { color: '#8b5cf6', width: 1.5, type: 'dashed' },
        itemStyle: { color: '#8b5cf6' }
      },
      // 访问量：细面积线（主体）
      {
        name: '访问量',
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: visits,
        lineStyle: { color: '#f97316', width: 2 },
        itemStyle: { color: '#f97316' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(249, 115, 22, 0.25)' },
              { offset: 1, color: 'rgba(249, 115, 22, 0)' }
            ]
          }
        }
      },
      // 下载流量：右轴细线 + 极淡填充
      {
        name: '下载流量',
        type: 'line',
        smooth: true,
        symbol: 'none',
        yAxisIndex: 1,
        data: traffic,
        lineStyle: { color: '#10b981', width: 1.5 },
        itemStyle: { color: '#10b981' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(16, 185, 129, 0.15)' },
              { offset: 1, color: 'rgba(16, 185, 129, 0)' }
            ]
          }
        }
      }
    ]
  }
})

const refreshBandwidth = async () => {
  try {
    const res = await getBandwidth()
    bandwidth.value = res.data
  } catch (e) {
    console.error('Failed to load bandwidth', e)
  }
}

onMounted(async () => {
  try {
    const [statsRes] = await Promise.all([
      getStats()
    ])
    stats.value = statsRes.data

    document.title = `统计信息 - ${globalConfig.site.nameFull}`
    const desc = `查看${globalConfig.site.name}的访问统计、下载统计和地理分布数据`
    const metaDescription = document.querySelector('meta[name="description"]')
    const metaOgDescription = document.querySelector('meta[property="og:description"]')
    const metaTwitterDescription = document.querySelector('meta[property="twitter:description"]')
    if (metaDescription) metaDescription.setAttribute('content', desc)
    if (metaOgDescription) metaOgDescription.setAttribute('content', '统计信息 - ' + desc)
    if (metaTwitterDescription) metaTwitterDescription.setAttribute('content', '统计信息 - ' + desc)
  } catch (e) {
    console.error('Failed to load data', e)
  } finally {
    loading.value = false
  }

  await refreshBandwidth()
  bandwidthTimer = setInterval(refreshBandwidth, 1000)
})

onUnmounted(() => {
  if (bandwidthTimer) clearInterval(bandwidthTimer)
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">数据洞察</h1>
      <p class="mt-1 text-sm text-muted-foreground">全站访问与下载动态，实时呈现。</p>
    </div>

    <div v-if="loading" class="space-y-6">
      <Card class="shadow-sm">
        <CardHeader class="pb-2"><Skeleton class="h-5 w-24" /></CardHeader>
        <CardContent>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
            <div v-for="i in 5" :key="i"><Skeleton class="h-12 w-28" /></div>
          </div>
          <div class="mt-4"><Skeleton class="h-2 w-full rounded-full" /></div>
        </CardContent>
      </Card>
      <div class="grid gap-4 lg:grid-cols-7">
        <Card class="lg:col-span-4 h-[420px] shadow-sm">
          <CardHeader><Skeleton class="h-5 w-32" /></CardHeader>
          <CardContent class="h-full"><Skeleton class="h-full w-full rounded" /></CardContent>
        </Card>
        <Card class="lg:col-span-3 h-[420px] shadow-sm">
          <CardHeader><Skeleton class="h-5 w-32" /></CardHeader>
          <CardContent class="h-full"><Skeleton class="h-full w-full rounded" /></CardContent>
        </Card>
      </div>
      <Card class="shadow-sm">
        <CardHeader><Skeleton class="h-5 w-40" /></CardHeader>
        <CardContent class="h-[350px]"><Skeleton class="h-full w-full rounded" /></CardContent>
      </Card>
    </div>

    <template v-else>
      <div class="grid gap-4 lg:grid-cols-2">
        <!-- 站点概览：与带宽状态卡同款布局，指标列 + 磁盘占用进度条 -->
        <Card class="shadow-sm">
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="flex items-center gap-2 text-base">
              <Activity weight="duotone" class="h-4 w-4 text-primary" />
              站点概览
            </CardTitle>
            <span class="text-xs text-muted-foreground">运行 {{ stats.total_days ?? '-' }} 天</span>
          </CardHeader>
          <CardContent>
            <div class="grid grid-cols-2 gap-4 2xl:grid-cols-3">
            <div>
              <p class="text-xs text-muted-foreground">总访问量</p>
              <div class="flex items-baseline gap-2">
                <p class="text-2xl font-bold">{{ stats.total_visits?.toLocaleString() || '-' }}</p>
                <span
                  v-if="visitGrowth !== 0"
                  class="inline-flex items-center gap-0.5 text-xs font-medium"
                  :class="visitGrowth >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'"
                >
                  <ArrowUp weight="duotone" v-if="visitGrowth >= 0" class="h-3 w-3" />
                  <ArrowDown weight="duotone" v-else class="h-3 w-3" />
                  {{ Math.abs(visitGrowth).toFixed(1) }}%
                </span>
              </div>
              <p class="mt-1 text-xs text-muted-foreground">近 30 日 {{ stats.last_30_visits?.toLocaleString() || 0 }}</p>
            </div>
            <div>
              <p class="text-xs text-muted-foreground">总下载量</p>
              <div class="flex items-baseline gap-2">
                <p class="text-2xl font-bold">{{ stats.total_downloads?.toLocaleString() || '-' }}</p>
                <span
                  v-if="downloadGrowth !== 0"
                  class="inline-flex items-center gap-0.5 text-xs font-medium"
                  :class="downloadGrowth >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'"
                >
                  <ArrowUp weight="duotone" v-if="downloadGrowth >= 0" class="h-3 w-3" />
                  <ArrowDown weight="duotone" v-else class="h-3 w-3" />
                  {{ Math.abs(downloadGrowth).toFixed(1) }}%
                </span>
              </div>
              <p class="mt-1 text-xs text-muted-foreground">近 30 日 {{ stats.last_30_downloads?.toLocaleString() || 0 }}</p>
            </div>
            <div>
              <p class="text-xs text-muted-foreground">总流量</p>
              <p class="text-2xl font-bold">{{ formatBytes(stats.total_traffic_bytes) }}</p>
              <p class="mt-1 text-xs text-muted-foreground">近 30 日 {{ formatBytes(stats.last_30_traffic_bytes) }}</p>
            </div>
            <div>
              <p class="text-xs text-muted-foreground">覆盖地区</p>
              <p class="text-2xl font-bold">{{ stats.geo_distribution?.length || '-' }}</p>
              <p class="mt-1 text-xs text-muted-foreground">国家/地区</p>
            </div>
            <div>
              <p class="text-xs text-muted-foreground">磁盘占用</p>
              <p class="text-2xl font-bold">{{ diskPercentage }}<span class="text-sm font-normal text-muted-foreground">%</span></p>
              <p class="mt-1 text-xs text-muted-foreground">{{ formatBytes(stats.disk?.used) }} / {{ formatBytes(stats.disk?.total) }}</p>
            </div>
          </div>
          <div class="mt-4">
            <div class="mb-1 flex items-center justify-between text-xs">
              <span class="text-muted-foreground">磁盘使用率</span>
              <span class="font-medium" :class="diskPercentage > 90 ? 'text-red-600 dark:text-red-400' : diskPercentage > 75 ? 'text-amber-600 dark:text-amber-400' : 'text-green-600 dark:text-green-400'">{{ diskPercentage }}%</span>
            </div>
            <div class="h-2 w-full overflow-hidden rounded-full bg-secondary">
              <div
                class="h-full rounded-full transition-all duration-500"
                :class="diskPercentage > 90 ? 'bg-red-500' : diskPercentage > 75 ? 'bg-amber-500' : 'bg-green-500'"
                :style="{ width: `${diskPercentage}%` }"
              ></div>
            </div>
          </div>
        </CardContent>
        </Card>

        <Card class="shadow-sm">
          <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle class="flex items-center gap-2 text-base">
              <Gauge weight="duotone" class="h-4 w-4 text-primary" />
              服务器带宽状态
            </CardTitle>
            <span class="text-xs text-muted-foreground">每秒自动刷新</span>
          </CardHeader>
          <CardContent>
            <div class="grid grid-cols-2 gap-4 2xl:grid-cols-3">
              <div>
                <p class="text-xs text-muted-foreground">当前带宽</p>
                <p class="text-2xl font-bold">
                  <span v-if="isBandwidthIdle" class="text-muted-foreground">空闲中</span>
                  <template v-else>
                    {{ formatMbps(bandwidth.current_bandwidth_mbps) }}
                    <span class="text-sm font-normal text-muted-foreground">Mbps</span>
                  </template>
                </p>
                <p class="mt-1 text-xs text-muted-foreground">{{ isBandwidthIdle ? '当前无传输任务' : '正在向外传输数据' }}</p>
              </div>
              <div>
                <p class="text-xs text-muted-foreground">峰值上限</p>
                <p class="text-2xl font-bold">{{ bandwidth.peak_bandwidth_mbps ?? '-' }} <span class="text-sm font-normal text-muted-foreground">Mbps</span></p>
                <p class="mt-1 text-xs text-muted-foreground">配置的带宽上限</p>
              </div>
              <div>
                <p class="text-xs text-muted-foreground">峰值（已观测）</p>
                <p class="text-2xl font-bold">{{ formatMbps(bandwidth.peak_observed_mbps) }} <span class="text-sm font-normal text-muted-foreground">Mbps</span></p>
                <p class="mt-1 text-xs text-muted-foreground">历史最高记录</p>
              </div>
              <div>
                <p class="text-xs text-muted-foreground">近1分钟下载</p>
                <p class="text-2xl font-bold">{{ bandwidth.recent_downloads ?? bandwidth.active_downloads ?? '-' }}</p>
                <p class="mt-1 text-xs text-muted-foreground">
                  活跃下载 <span :class="bandwidth.active_downloads ? 'text-emerald-600 dark:text-emerald-400' : ''">{{ bandwidth.active_downloads || 0 }}</span> 个
                </p>
              </div>
              <div>
                <p class="text-xs text-muted-foreground">累计传输</p>
                <p class="text-2xl font-bold">{{ formatBytes(bandwidth.total_bytes_served) }}</p>
                <p class="mt-1 text-xs text-muted-foreground">服务启动以来累计</p>
              </div>
            </div>
            <div class="mt-4">
              <div class="mb-1 flex items-center justify-between text-xs">
                <span class="text-muted-foreground">带宽利用率</span>
              <span v-if="isBandwidthIdle" class="font-medium text-green-600 dark:text-green-400">空闲</span>
              <span v-else class="font-medium" :class="bandwidthUtilization > 90 ? 'text-red-600 dark:text-red-400' : bandwidthUtilization > 75 ? 'text-amber-600 dark:text-amber-400' : 'text-green-600 dark:text-green-400'">{{ bandwidthUtilization }}%</span>
            </div>
            <div class="h-2 w-full overflow-hidden rounded-full bg-secondary">
              <div
                class="h-full rounded-full transition-all duration-500"
                :class="isBandwidthIdle ? 'bg-secondary' : bandwidthUtilization > 90 ? 'bg-red-500' : bandwidthUtilization > 75 ? 'bg-amber-500' : 'bg-green-500'"
                :style="{ width: isBandwidthIdle ? '0%' : `${bandwidthUtilization}%` }"
              ></div>
            </div>
          </div>
        </CardContent>
        </Card>
      </div>

      <div class="grid gap-4 lg:grid-cols-7">
        <Card class="lg:col-span-4 shadow-sm">
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-base">
              <MapPin weight="duotone" class="h-4 w-4 text-primary" />
              全球访问分布
            </CardTitle>
            <CardDescription>按访问来源国家/地区统计，面积越大代表访问越多。</CardDescription>
          </CardHeader>
          <CardContent class="pl-2">
            <div class="h-[350px] w-full">
              <VChart class="chart" :option="treemapOption" autoresize />
            </div>
          </CardContent>
        </Card>

        <Card class="lg:col-span-3 flex flex-col shadow-sm">
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-base">
              <TrendingUp weight="duotone" class="h-4 w-4 text-green-500" />
              热门资源排行
            </CardTitle>
            <CardDescription>下载量最高的启动器与插件，皇冠属于前三名。</CardDescription>
          </CardHeader>
          <CardContent class="flex-1 overflow-hidden">
            <div class="space-y-1.5">
              <div
                v-for="(item, i) in topDownloads"
                :key="`download-${item.key}`"
                class="flex items-center gap-3 rounded-lg border bg-card px-3 py-2.5 transition-colors hover:bg-accent/50"
              >
                <!-- 启动器 logo（无 logo 回退首字符），前三名叠加专属皇冠角标 -->
                <div class="relative shrink-0">
                  <img
                    v-if="item.logo"
                    :src="item.logo"
                    alt=""
                    class="h-8 w-8 rounded-md border object-cover"
                    loading="lazy"
                  />
                  <div
                    v-else
                    class="flex h-8 w-8 items-center justify-center rounded-md border bg-muted text-xs font-bold text-muted-foreground"
                  >
                    {{ item.name.slice(0, 1) }}
                  </div>
                  <span
                    v-if="i < 3"
                    class="absolute -right-1.5 -top-1.5 flex h-4 w-4 items-center justify-center rounded-full border border-border bg-background shadow-sm"
                    :class="i === 0 ? 'text-amber-500' : i === 1 ? 'text-zinc-400' : 'text-orange-600'"
                  >
                    <svg viewBox="0 0 24 24" fill="currentColor" class="h-3 w-3" aria-hidden="true">
                      <path :d="crownPaths[i]" />
                    </svg>
                  </span>
                </div>

                <div class="min-w-0 flex-1">
                  <div class="flex items-baseline justify-between gap-2">
                    <p class="truncate text-sm font-medium leading-none">{{ item.name }}</p>
                    <div class="flex shrink-0 items-baseline gap-1.5">
                      <span class="text-sm font-bold tabular-nums">{{ item.count.toLocaleString() }}</span>
                      <span class="w-10 text-right text-[10px] tabular-nums text-muted-foreground">{{ item.percent }}%</span>
                    </div>
                  </div>
                  <!-- 占比条：以第一名长度为 100% 基准 -->
                  <div class="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-muted">
                    <div
                      class="h-full rounded-full bg-primary/60 transition-all duration-500"
                      :style="{ width: `${item.barWidth}%` }"
                    ></div>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card class="shadow-sm">
        <CardHeader>
          <CardTitle class="flex items-center gap-2 text-base">
            <BarChart3 weight="duotone" class="h-4 w-4 text-orange-500" />
            最近 30 天趋势
          </CardTitle>
            <CardDescription>下载柱与访问曲线叠加，紫色虚线为 7 日均值。</CardDescription>
        </CardHeader>
        <CardContent class="pl-2">
          <div class="h-[350px] w-full">
            <VChart class="chart" :option="trendOption" autoresize />
          </div>
        </CardContent>
      </Card>
    </template>
  </div>
</template>

<style scoped>
.chart {
  height: 100%;
  width: 100%;
}
</style>

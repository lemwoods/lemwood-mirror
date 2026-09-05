<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { PhFolder as Folder, PhPackage as Package } from '@phosphor-icons/vue'
import { getStatus } from '@/services/api'
import { compareVersionDesc } from '@/lib/format'
import { globalConfig } from '@/lib/globalConfig'
import Card from '@/components/ui/Card.vue'
import Skeleton from '@/components/ui/Skeleton.vue'

const router = useRouter()
const launcherDefaultLogo = globalConfig.launchers.fcl?.logoUrl

const rawLaunchers = ref({})
const loading = ref(true)

const launcherList = computed(() => {
  const list = Object.keys(rawLaunchers.value)
    .filter((name) => name !== 'authlib-injector')
    .map((name) => {
      const versions = rawLaunchers.value[name]
      const info = globalConfig.launchers[name] || { displayName: name }

      return {
        name,
        displayName: info.displayName,
        logoUrl: info.logoUrl || launcherDefaultLogo,
        versions,
        lastUpdated: versions.length ? versions[0].published_at : null
      }
    })
  return list
})

const loadData = async () => {
  loading.value = true
  try {
    const statusRes = await getStatus()
    const data = statusRes.data
    for (const key in data) {
      // 数值语义降序（1.10.0 > 1.9.0），见 lib/format.js
      data[key].sort(compareVersionDesc)
    }
    rawLaunchers.value = data
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '未知时间'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return '未知时间'
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 版本文件夹格子：只展示最新两个版本
const recentVersions = (item) => (item.versions || []).slice(0, 2)

const formatVersionLabel = (v) => v.tag_name || v.name || ''

onMounted(() => {
  loadData()
})

defineExpose({ refresh: loadData })
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-2">
      <h2 class="text-3xl font-bold tracking-tight">版本探索</h2>
      <p class="text-muted-foreground">实时同步上游发布，点开版本文件夹即可获取文件。</p>
    </div>

    <div v-if="loading && !launcherList.length" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
      <Card v-for="i in 4" :key="i">
        <div class="flex items-start gap-4 p-4">
          <Skeleton class="h-14 w-14 shrink-0 rounded-xl" />
          <div class="min-w-0 flex-1 space-y-2">
            <Skeleton class="h-5 w-3/4" />
            <Skeleton class="h-4 w-1/2" />
          </div>
        </div>
        <div class="px-4 pb-4">
          <div class="rounded-lg bg-muted/50 p-3">
            <Skeleton class="mb-2 h-3 w-16" />
            <div class="grid grid-cols-2 gap-2">
              <Skeleton class="h-9" />
              <Skeleton class="h-9" />
            </div>
          </div>
        </div>
      </Card>
    </div>

    <div v-else-if="!launcherList.length" class="rounded-lg border border-dashed p-12 text-center text-muted-foreground">
      <Package weight="duotone" class="mx-auto mb-4 h-12 w-12 opacity-40" />
      <p class="font-medium text-foreground">暂时没有可展示的内容</p>
      <p class="mt-1 text-sm">服务可能正在同步数据，稍候刷新即可。</p>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
      <Card
        v-for="item in launcherList"
        :key="item.name"
        class="transition-shadow hover:shadow-md"
      >
        <!-- 头部：logo + 名称/最近更新 -->
        <div class="flex items-start gap-4 p-4 pb-3">
          <img
            :src="item.logoUrl"
            class="h-14 w-14 shrink-0 rounded-xl border bg-background object-contain p-1.5 shadow-sm"
            :alt="item.displayName"
          />
          <div class="min-w-0 flex-1 space-y-1 self-center">
            <h3 class="truncate text-lg font-bold leading-tight text-foreground">{{ item.displayName }}</h3>
            <p class="text-sm text-muted-foreground">最近更新：{{ formatDate(item.lastUpdated) }}</p>
          </div>
        </div>

        <!-- 版本文件夹：内嵌浅底区块，格子直达对应版本文件列表 -->
        <div class="px-4 pb-4">
          <div class="rounded-lg bg-muted/50 p-3">
            <p class="mb-2 text-xs font-medium text-muted-foreground">版本文件夹</p>
            <div class="grid grid-cols-2 gap-2">
              <button
                v-for="v in recentVersions(item)"
                :key="formatVersionLabel(v)"
                type="button"
                class="flex min-w-0 items-center gap-2 rounded-md border bg-background px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
                @click="$router.push(`/files/${item.name}/${formatVersionLabel(v)}`)"
              >
                <Folder weight="duotone" class="h-4 w-4 shrink-0 text-primary" />
                <span class="truncate">{{ formatVersionLabel(v) }}</span>
              </button>
            </div>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

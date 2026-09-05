<script setup>
import VersionList from '@/components/VersionList.vue'
import { useSeoMeta } from '@/composables/useSeoMeta'
import { globalConfig } from '@/lib/globalConfig'

useSeoMeta(
  { title: '首页', description: globalConfig.site.description },
  globalConfig.site.nameFull
)()
</script>

<template>
  <div class="space-y-6">

    <VersionList />
  </div>
</template>

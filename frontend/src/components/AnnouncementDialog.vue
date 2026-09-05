<script setup>
import { onMounted, ref } from 'vue'
import AppDialog from '@/components/ui/AppDialog.vue'
import {
  announcementConfig,
  hasSeenAnnouncement,
  markAnnouncementAsSeen
} from '@/lib/announcementConfig'

const isOpen = ref(false)

onMounted(() => {
  if (!hasSeenAnnouncement()) {
    setTimeout(() => {
      isOpen.value = true
    }, 500)
  }
})

const closeDialog = () => {
  isOpen.value = false
  markAnnouncementAsSeen()
}
</script>

<template>
  <AppDialog :open="isOpen" width="md" aria-label="公告" @close="closeDialog">
    <div class="p-5 sm:p-6">
      <div class="mb-4 flex items-start justify-between">
        <div class="flex items-center gap-2">
          <div class="h-6 w-1 rounded-full bg-primary"></div>
          <h2 class="text-lg font-semibold text-foreground">
            {{ announcementConfig.title }}
          </h2>
        </div>
      </div>

      <div class="space-y-3 text-sm leading-relaxed text-muted-foreground">
        <div class="rounded-lg bg-muted/50 p-4">
          <p class="whitespace-pre-line leading-relaxed">
            {{ announcementConfig.content }}
          </p>
          <p
            v-if="announcementConfig.importantText"
            class="mt-3 font-bold text-red-500"
          >
            {{ announcementConfig.importantText }}
          </p>
        </div>
      </div>

      <div
        v-if="announcementConfig.links?.length"
        class="mt-4 flex flex-wrap gap-2"
      >
        <a
          v-for="(link, index) in announcementConfig.links"
          :key="index"
          :href="link.url"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-1.5 rounded-md bg-primary/10 px-3 py-1.5 text-sm font-medium text-primary transition-colors hover:bg-primary/20"
        >
          {{ link.label }}
        </a>
      </div>
    </div>
  </AppDialog>
</template>

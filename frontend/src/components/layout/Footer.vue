<script setup>
import {
  PhChatCircle as MessageCircle,
  PhGithubLogo as Github,
  PhArrowSquareOut as ExternalLink,
  PhGlobe as Globe
} from '@phosphor-icons/vue'
import { globalConfig } from '@/lib/globalConfig'
import { friendLinksConfig } from '@/lib/friendLinksConfig'

const friendLinks = friendLinksConfig.enabled ? friendLinksConfig.links || [] : []
</script>

<template>
  <footer class="border-t bg-muted/20">
    <div class="container mx-auto px-4 py-8">
      <div class="grid gap-8 md:grid-cols-3">
        <!-- 左：品牌简介 + 联系方式 + 版权/备案声明 -->
        <div class="md:col-span-2">
          <p class="text-sm font-semibold">{{ globalConfig.site.name }}</p>
          <p class="mt-2 max-w-md text-sm text-muted-foreground">
            {{ globalConfig.site.description }}
          </p>

          <p class="mt-6 text-sm font-semibold">联系我们</p>
          <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1.5">
            <a
              v-if="globalConfig.links.qqGroup"
              :href="globalConfig.links.qqGroup"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary hover:underline"
            >
              <MessageCircle weight="duotone" class="h-3.5 w-3.5" />
              QQ 群
            </a>
            <a
              v-if="globalConfig.links.githubRepo"
              :href="globalConfig.links.githubRepo"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary hover:underline"
            >
              <Github weight="duotone" class="h-3.5 w-3.5" />
              GitHub
            </a>
            <a
              v-if="globalConfig.site.url"
              :href="globalConfig.site.url"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary hover:underline"
            >
              <Globe weight="duotone" class="h-3.5 w-3.5" />
              {{ globalConfig.site.nameEn }}
            </a>
          </div>

          <div
            class="mt-6 flex flex-col gap-1.5 text-xs text-muted-foreground sm:flex-row sm:items-center sm:gap-3"
          >
            <span class="flex items-center gap-2">
              <span>&copy; {{ new Date().getFullYear() }} {{ globalConfig.site.name }}</span>
              <span>v{{ globalConfig.site.version }}</span>
            </span>
            <span class="hidden text-border sm:inline">|</span>
            <span class="flex flex-wrap items-center gap-x-2 gap-y-1">
              <a
                v-if="globalConfig.legal.icp"
                :href="globalConfig.links.beian"
                target="_blank"
                rel="noopener noreferrer"
                class="transition-colors hover:text-foreground"
              >
                {{ globalConfig.legal.icp }}
              </a>
            </span>
          </div>
        </div>

        <!-- 右：友情链接 -->
        <div v-if="friendLinks.length">
          <p class="text-sm font-semibold">{{ friendLinksConfig.title }}</p>
          <div class="mt-2 flex flex-col gap-2">
            <a
              v-for="link in friendLinks"
              :key="link.url"
              :href="link.url"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary hover:underline"
            >
              {{ link.name }}
              <ExternalLink weight="duotone" class="h-3 w-3" />
            </a>
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>

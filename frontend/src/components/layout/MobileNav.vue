<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { PhGithubLogo as Github } from '@phosphor-icons/vue'
import { isNavigationActive, navigationLinks } from '@/lib/navigation'
import { globalConfig } from '@/lib/globalConfig'

// 顶栏吸附为胶囊时底缘在 60px，未吸附时为 56px；菜单顶部随之衔接，避免缝隙漏出未模糊内容
const props = defineProps({ scrolled: Boolean })

const route = useRoute()

const isOpen = ref(false)
const rootEl = ref(null)
const menuEl = ref(null)

const toggleNav = () => {
  isOpen.value = !isOpen.value
}

const closeNav = () => {
  isOpen.value = false
}

const onDocumentClick = event => {
  const target = event.target
  if (isOpen.value && !rootEl.value?.contains(target) && !menuEl.value?.contains(target)) {
    closeNav()
  }
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onUnmounted(() => document.removeEventListener('click', onDocumentClick))
</script>

<template>
  <div ref="rootEl" class="relative md:hidden">
    <button
      class="rounded-md p-2 transition-colors hover:bg-accent"
      :aria-expanded="isOpen"
      aria-label="菜单"
      @click.stop="toggleNav"
    >
      <!-- 动画汉堡：三条线（中线半透明呼应 duotone 层次），展开时平滑合并为 X -->
      <span class="relative flex h-5 w-5 items-center justify-center">
        <span
          class="absolute h-0.5 w-4 rounded-full bg-current transition-all duration-300 ease-bounce-soft"
          :class="isOpen ? 'rotate-45' : '-translate-y-[6px]'"
        />
        <span
          class="absolute h-0.5 w-4 rounded-full bg-current transition-all duration-200"
          :class="isOpen ? 'scale-x-0 opacity-0' : 'opacity-60'"
        />
        <span
          class="absolute h-0.5 w-4 rounded-full bg-current transition-all duration-300 ease-bounce-soft"
          :class="isOpen ? '-rotate-45' : 'translate-y-[6px]'"
        />
      </span>
    </button>

    <Teleport to="body">
      <Transition name="menu">
        <div
          v-if="isOpen"
          ref="menuEl"
          class="fixed right-3 z-50 w-56 overflow-hidden rounded-xl border border-border/60 bg-background/80 shadow-lg backdrop-blur-md"
          :class="props.scrolled ? 'top-[72px]' : 'top-[64px]'"
        >
          <div class="space-y-0.5 p-1.5">
            <RouterLink
              v-for="link in navigationLinks"
              :key="link.path"
              :to="link.path"
              class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
              :class="
                isNavigationActive(route.path, link)
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground'
              "
              @click="closeNav"
            >
              <component :is="link.icon" weight="duotone" class="h-4 w-4" />
              {{ link.name }}
            </RouterLink>

            <div class="my-1 border-t" />

            <a
              :href="globalConfig.links.githubRepo"
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent/50 hover:text-accent-foreground"
              @click="closeNav"
            >
              <Github weight="duotone" class="h-4 w-4" />
              GitHub
            </a>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.menu-enter-active {
  transition:
    opacity 0.2s cubic-bezier(0.34, 1.7, 0.64, 1),
    transform 0.2s cubic-bezier(0.34, 1.7, 0.64, 1);
}

.menu-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.menu-enter-from,
.menu-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.98);
}
</style>

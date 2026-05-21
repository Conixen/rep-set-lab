<template>
  <div class="flex h-dvh bg-[#111114] text-white">

    <!-- XP toast -->
    <Transition name="slide-down">
      <div v-if="toast"
           class="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-violet-500 text-white px-6 py-3 rounded-2xl shadow-lg text-sm font-semibold whitespace-nowrap">
        {{ toast }}
      </div>
    </Transition>

    <!-- Sidebar — desktop only -->
    <nav class="hidden lg:flex flex-col w-56 shrink-0 bg-[#1a1a1f] border-r border-white/10 p-4 gap-1">
      <span class="text-lg font-bold text-white px-3 mb-6">rep-set-lab</span>
      <RouterLink
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-white/40 transition-colors hover:text-white hover:bg-white/5"
        active-class="!text-violet-400 bg-violet-500/10"
      >
        <component :is="tab.icon" class="w-5 h-5 shrink-0" />
        {{ tab.label }}
      </RouterLink>
    </nav>

    <!-- Main content -->
    <main class="flex-1 overflow-y-auto pb-16 lg:pb-0 px-4 lg:px-8">
      <RouterView />
    </main>

    <!-- Bottom nav — mobile only -->
    <nav class="lg:hidden fixed bottom-0 left-0 right-0 bg-[#1a1a1f] border-t border-white/10 flex justify-around items-center h-16 z-50">
      <RouterLink
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="flex flex-col items-center gap-1 text-xs text-white/40 transition-colors"
        active-class="!text-violet-500"
      >
        <component :is="tab.icon" class="w-5 h-5" />
        {{ tab.label }}
      </RouterLink>
    </nav>

  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useAuthStore } from './stores/auth'
import IconHome    from './components/icons/IconHome.vue'
import IconLibrary from './components/icons/IconLibrary.vue'
import IconAi      from './components/icons/IconAi.vue'
import IconCompare from './components/icons/IconCompare.vue'
import IconHistory from './components/icons/IconHistory.vue'
import IconProfile from './components/icons/IconProfile.vue'
import IconAdmin   from './components/icons/IconAdmin.vue'

const auth = useAuthStore()

const baseTabs = [
  { to: '/',        label: 'Home',    icon: IconHome },
  { to: '/library', label: 'Library', icon: IconLibrary },
  { to: '/ai',      label: 'AI',      icon: IconAi },
  { to: '/compare', label: 'Compare', icon: IconCompare },
  { to: '/history', label: 'History', icon: IconHistory },
  { to: '/profile', label: 'Profile', icon: IconProfile },
]

const adminTab = { to: '/admin', label: 'Admin', icon: IconAdmin }

const tabs = computed(() =>
  auth.isAdmin ? [...baseTabs, adminTab] : baseTabs
)

const toast = ref<string | null>(null)
let ws: WebSocket | null = null
let toastTimer: ReturnType<typeof setTimeout> | null = null

function connectWS() {
  if (ws || !auth.token) return
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${proto}//${window.location.host}/ws?token=${auth.token}`)
  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'xp_update') {
        const d = msg.data
        toast.value = d.leveled_up
          ? `Level up! You are now level ${d.level} 🎉`
          : `+${d.xp_earned} XP earned!`
        if (toastTimer) clearTimeout(toastTimer)
        toastTimer = setTimeout(() => toast.value = null, 3500)
      }
    } catch {}
  }
  ws.onclose = () => { ws = null }
}

function disconnectWS() {
  ws?.close()
  ws = null
}

watch(() => auth.isLoggedIn, (loggedIn) => {
  if (loggedIn) connectWS()
  else disconnectWS()
}, { immediate: true })
</script>

<style scoped>
.slide-down-enter-active, .slide-down-leave-active { transition: all 0.3s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translate(-50%, -1rem); }
</style>

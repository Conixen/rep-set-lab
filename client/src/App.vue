<template>
  <div class="flex h-dvh bg-[#111114] text-white" style="background: radial-gradient(ellipse 40% 60% at 0% 50%, rgba(124,92,191,0.18) 0%, transparent 70%), radial-gradient(ellipse 40% 60% at 100% 50%, rgba(124,92,191,0.14) 0%, transparent 70%), #111114">

    <Transition name="slide-down">
      <div v-if="toast"
           class="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-violet-500 text-white px-6 py-3 rounded-2xl shadow-lg text-sm font-semibold whitespace-nowrap">
        {{ toast }}
      </div>
    </Transition>

    <nav class="hidden lg:flex flex-col w-56 shrink-0 border-r border-violet-500/20 p-4" style="background: linear-gradient(180deg, rgba(124,92,191,0.08) 0%, transparent 40%), #1a1a1f; box-shadow: 1px 0 20px rgba(124,92,191,0.12)">
      <div class="flex items-center gap-3 px-2 mb-6">
        <span class="text-base font-bold text-white">Rep Set Lab</span>
      </div>

      <p class="text-[10px] text-white/30 font-semibold tracking-widest uppercase px-3 mb-1">Menu</p>
      <RouterLink
        v-for="tab in menuTabs"
        :key="tab.to"
        :to="tab.to"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-white/40 transition-colors hover:text-white hover:bg-white/5"
        active-class="!text-violet-400 bg-violet-500/10 nav-glow"
      >
        <component :is="tab.icon" class="w-5 h-5 shrink-0" />
        {{ tab.label }}
      </RouterLink>

      <p class="text-[10px] text-white/30 font-semibold tracking-widest uppercase px-3 mb-1 mt-5">System</p>
      <RouterLink
        v-for="tab in systemTabs"
        :key="tab.to"
        :to="tab.to"
        class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-white/40 transition-colors hover:text-white hover:bg-white/5"
        active-class="!text-violet-400 bg-violet-500/10 nav-glow"
      >
        <component :is="tab.icon" class="w-5 h-5 shrink-0" />
        {{ tab.label }}
      </RouterLink>

      <div class="flex-1" />

      <RouterLink to="/profile" class="border-t border-white/10 pt-3 mt-2 flex items-center gap-2.5 px-2 hover:opacity-80 transition-opacity">
        <div class="w-8 h-8 rounded-full bg-violet-500/30 border border-violet-500/40 flex items-center justify-center text-violet-300 text-sm font-bold shrink-0 shadow-[0_0_10px_rgba(124,92,191,0.35)]">
          {{ auth.username?.[0]?.toUpperCase() ?? '?' }}
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium truncate text-white">{{ auth.username }}</p>
          <p class="text-xs text-white/40">Lv. {{ sidebarStats?.level ?? '—' }} · {{ sidebarStats?.total_xp ?? '—' }} XP</p>
        </div>
      </RouterLink>
    </nav>

    <main class="flex-1 overflow-y-auto pb-16 lg:pb-0 px-4 lg:px-8">
      <RouterView />
    </main>

    <nav class="lg:hidden fixed bottom-0 left-0 right-0 bg-[#1a1a1f] border-t border-white/10 flex justify-around items-center h-16 z-50">
      <RouterLink
        v-for="tab in allTabs"
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
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { api } from './api/client'
import IconHome    from './components/icons/IconHome.vue'
import IconLibrary from './components/icons/IconLibrary.vue'
import IconAi      from './components/icons/IconAi.vue'
import IconCompare from './components/icons/IconCompare.vue'
import IconHistory from './components/icons/IconHistory.vue'
import IconProfile from './components/icons/IconProfile.vue'
import IconAdmin   from './components/icons/IconAdmin.vue'

const auth   = useAuthStore()
const router = useRouter()

const menuTabs = [
  { to: '/',        label: 'Home',     icon: IconHome },
  { to: '/library', label: 'Library',  icon: IconLibrary },
  { to: '/ai',      label: 'AI coach', icon: IconAi },
  { to: '/compare', label: 'Compare',  icon: IconCompare },
  { to: '/history', label: 'History',  icon: IconHistory },
]

const systemTabs = computed(() => {
  const s: { to: string; label: string; icon: any }[] = []
  if (auth.isAdmin) s.push({ to: '/admin',   label: 'Admin',    icon: IconAdmin })
  s.push(                  { to: '/settings', label: 'Settings', icon: IconProfile })
  return s
})

const allTabs = computed(() => {
  const tabs = [...menuTabs]
  if (auth.isAdmin) tabs.push({ to: '/admin',   label: 'Admin',    icon: IconAdmin })
  tabs.push(                  { to: '/profile',  label: 'Profile',  icon: IconProfile })
  tabs.push(                  { to: '/settings', label: 'Settings', icon: IconProfile })
  return tabs
})

interface SidebarStats { level: number; total_xp: number }
const sidebarStats = ref<SidebarStats | null>(null)

async function fetchSidebarStats() {
  try { sidebarStats.value = await api.get<SidebarStats>('/users/me/stats') } catch {}
}

watch(() => auth.isLoggedIn, (loggedIn) => {
  if (loggedIn) fetchSidebarStats()
  else sidebarStats.value = null
}, { immediate: true })

const toast = ref<string | null>(null)
let ws: WebSocket | null = null
let toastTimer: ReturnType<typeof setTimeout> | null = null

function connectWS() {
  if (ws || !auth.token) return
  const apiBase = import.meta.env.VITE_API_BASE_URL
  const wsBase = apiBase
    ? apiBase.replace(/^http/, 'ws')
    : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}`
  ws = new WebSocket(`${wsBase}/ws?token=${auth.token}`)
  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'xp_update') {
        const d = msg.data
        if (sidebarStats.value) {
          sidebarStats.value.total_xp = d.total_xp
          sidebarStats.value.level    = d.level
        }
        toast.value = d.leveled_up
          ? `Level up! You are now level ${d.level} 🎉`
          : `+${d.xp_earned} XP earned!`
        if (toastTimer) clearTimeout(toastTimer)
        toastTimer = setTimeout(() => toast.value = null, 3500)
      } else if (msg.type === 'account_approved') {
        toast.value = 'Your account has been approved — welcome!'
        if (toastTimer) clearTimeout(toastTimer)
        toastTimer = setTimeout(() => toast.value = null, 5000)
        setTimeout(() => router.push('/'), 800)
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

.nav-glow {
  box-shadow: 0 0 14px rgba(124, 92, 191, 0.45);
}
</style>

<template>
  <div class="p-5 space-y-6">
    <h1 class="text-2xl font-bold pt-2">Profile</h1>

    <div v-if="loading" class="text-white/40 text-sm text-center py-8">Loading…</div>

    <template v-else-if="stats">
      <div class="flex flex-col items-center gap-3 py-4">
        <div class="w-20 h-20 rounded-full bg-violet-500/30 flex items-center justify-center text-3xl font-bold text-violet-400">
          {{ stats.username[0].toUpperCase() }}
        </div>
        <p class="text-xl font-bold">{{ stats.username }}</p>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold text-violet-400">{{ stats.level }}</p>
          <p class="text-xs text-white/40 mt-1">Level</p>
        </div>
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold">{{ stats.xp }}</p>
          <p class="text-xs text-white/40 mt-1">Total XP</p>
        </div>
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold">{{ stats.workouts_completed }}</p>
          <p class="text-xs text-white/40 mt-1">Completed</p>
        </div>
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold">{{ stats.workouts_total }}</p>
          <p class="text-xs text-white/40 mt-1">Total workouts</p>
        </div>
      </div>

      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-1">
        <div class="flex justify-between text-sm">
          <span class="text-white/50">Next level</span>
          <span class="text-white">{{ stats.next_level_xp }} XP</span>
        </div>
        <div class="h-2 bg-white/10 rounded-full overflow-hidden">
          <div class="h-full bg-violet-500 rounded-full transition-all"
               :style="{ width: xpPercent + '%' }"></div>
        </div>
        <p class="text-xs text-white/30 text-right">{{ stats.xp }} / {{ stats.next_level_xp }} XP</p>
      </div>
    </template>

    <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

    <button
      @click="handleLogout"
      class="w-full border border-white/10 text-white/50 font-medium py-3 rounded-2xl text-sm hover:bg-white/5 transition-colors"
    >Log out</button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { api } from '../api/client'

interface Stats {
  username:            string
  xp:                  number
  level:               number
  next_level_xp:       number
  workouts_total:      number
  workouts_completed:  number
}

const auth   = useAuthStore()
const router = useRouter()

const stats   = ref<Stats | null>(null)
const loading = ref(true)
const error   = ref('')

const xpPercent = computed(() => {
  if (!stats.value) return 0
  return Math.min((stats.value.xp / stats.value.next_level_xp) * 100, 100)
})

onMounted(async () => {
  try {
    stats.value = await api.get<Stats>('/users/me/stats')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load profile'
  } finally {
    loading.value = false
  }
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

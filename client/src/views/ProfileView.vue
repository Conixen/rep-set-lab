<template>
  <div class="p-5 space-y-6">
    <h1 class="text-2xl font-bold pt-2">{{ t('profile.title') }}</h1>

    <div v-if="loading" class="text-white/40 text-sm text-center py-8">{{ t('library.loading') }}</div>

    <template v-else-if="stats">
      <div class="flex flex-col items-center gap-3 py-4">
        <div class="w-20 h-20 rounded-full bg-violet-500/30 flex items-center justify-center text-3xl font-bold text-violet-400">
          {{ stats.username[0].toUpperCase() }}
        </div>
        <p class="text-xl font-bold text-white">{{ stats.username }}</p>
      </div>

      <div class="grid grid-cols-3 gap-3">
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold text-violet-400">{{ stats.level }}</p>
          <p class="text-xs text-white/40 mt-1">{{ t('profile.level') }}</p>
        </div>
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold">{{ stats.workouts_completed }}</p>
          <p class="text-xs text-white/40 mt-1">{{ t('profile.completed') }}</p>
        </div>
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold">{{ stats.workouts_total }}</p>
          <p class="text-xs text-white/40 mt-1">{{ t('profile.workouts') }}</p>
        </div>
      </div>

      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-2">
        <div class="flex justify-between items-baseline">
          <span class="text-white/50 text-sm">{{ t('profile.totalXp') }}</span>
          <span class="text-white/40 text-xs">{{ t('profile.nextLevel', { xp: stats.next_level_xp }) }}</span>
        </div>
        <div class="flex items-baseline gap-1.5">
          <span class="text-2xl font-bold">{{ stats.total_xp }}</span>
          <span class="text-sm text-white/40">XP</span>
        </div>
        <div class="h-2 bg-white/10 rounded-full overflow-hidden">
          <div class="h-full bg-violet-500 rounded-full transition-all"
               :style="{ width: xpPercent + '%' }"></div>
        </div>
        <p class="text-xs text-white/30 text-right">{{ t('profile.xpProgress', { total: stats.total_xp, next: stats.next_level_xp, toGo: xpToNext }) }}</p>
      </div>
    </template>

    <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

    <button
      @click="handleLogout"
      class="w-full border border-white/10 text-white/50 font-medium py-3 rounded-2xl text-sm hover:bg-white/5 transition-colors"
    >{{ t('profile.logout') }}</button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { api } from '../api/client'
import { toMessage } from '../utils/error'

const { t } = useI18n()

interface Stats {
  username:            string
  total_xp:            number
  level:               number
  current_level_xp:    number
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
  const { total_xp, current_level_xp, next_level_xp } = stats.value
  if (next_level_xp === 0) return 100 // max level
  const range = next_level_xp - current_level_xp
  if (range === 0) return 0
  return Math.min(((total_xp - current_level_xp) / range) * 100, 100)
})

const xpToNext = computed(() => {
  if (!stats.value || stats.value.next_level_xp === 0) return 0
  return Math.max(stats.value.next_level_xp - stats.value.total_xp, 0)
})

onMounted(async () => {
  try {
    stats.value = await api.get<Stats>('/users/me/stats')
  } catch (e: unknown) {
    error.value = toMessage(e, 'Failed to load profile')
  } finally {
    loading.value = false
  }
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

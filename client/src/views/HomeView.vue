<template>
  <div class="p-5 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between pt-2">
      <div>
        <p class="text-white/50 text-sm">Good morning</p>
        <h1 class="text-2xl font-bold">Leon</h1>
      </div>
      <div class="bg-violet-500/20 border border-violet-500/40 rounded-full px-3 py-1 text-violet-400 text-sm font-semibold">
        Lv. {{ stats?.level ?? '—' }}
      </div>
    </div>

    <!-- XP bar -->
    <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-2">
      <div class="flex justify-between text-sm text-white/50">
        <span>XP Progress</span>
        <span class="text-white">{{ stats?.total_xp ?? 0 }} XP</span>
      </div>
      <div class="h-2 bg-white/10 rounded-full overflow-hidden">
        <div class="h-full bg-violet-500 rounded-full" :style="{ width: xpPercent + '%' }"></div>
      </div>
    </div>

    <!-- This week -->
    <div>
      <p class="text-xs text-white/40 font-semibold tracking-widest uppercase mb-3">This week</p>
      <div class="grid grid-cols-2 gap-3">
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold">{{ stats?.workouts_completed ?? '—' }}</p>
          <p class="text-xs text-white/40 mt-1">Workouts</p>
        </div>
        <div class="bg-[#1e1e24] rounded-2xl p-4 text-center">
          <p class="text-2xl font-bold text-violet-400">{{ stats?.total_xp ?? '—' }}</p>
          <p class="text-xs text-white/40 mt-1">Total XP</p>
        </div>
      </div>
    </div>

    <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'

interface Stats {
  total_xp: number
  level: number
  workouts_completed: number
}

const stats = ref<Stats | null>(null)
const error = ref('')

const xpPercent = computed(() => {
  if (!stats.value) return 0
  return Math.min((stats.value.total_xp % 2000) / 20, 100)
})

onMounted(async () => {
  try {
    stats.value = await api.get<Stats>('/users/me/stats')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load stats'
  }
})
</script>

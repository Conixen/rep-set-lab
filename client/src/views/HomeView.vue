<template>
  <div class="p-5 space-y-6">
    <div class="flex items-center justify-between pt-2">
      <div>
        <p class="text-white/50 text-sm">{{ greeting }}</p>
        <h1 class="text-2xl font-bold">{{ stats?.username ?? '…' }}</h1>
      </div>
      <div class="bg-violet-500/20 border border-violet-500/40 rounded-full px-3 py-1 text-violet-400 text-sm font-semibold transition-shadow hover:shadow-[0_0_14px_rgba(124,92,191,0.45)]">
        Lv. {{ stats?.level ?? '—' }}
      </div>
    </div>

    <div v-if="loading" class="text-white/40 text-sm text-center py-8">{{ t('library.loading') }}</div>

    <template v-else>
      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-2 ui-card">
        <div class="flex justify-between items-baseline">
          <span class="text-white/50 text-sm">{{ t('home.totalXp') }}</span>
          <span class="text-white/40 text-xs">{{ t('home.nextLevel', { xp: stats?.next_level_xp ?? '—' }) }}</span>
        </div>
        <div class="flex items-baseline gap-1.5">
          <span class="text-2xl font-bold">{{ stats?.total_xp ?? 0 }}</span>
          <span class="text-sm text-white/40">XP</span>
        </div>
        <div class="h-2 bg-white/10 rounded-full relative">
          <div class="h-full bg-violet-500 rounded-full transition-all relative" :style="{ width: xpPercent + '%' }">
            <span v-if="xpPercent > 2" class="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1/2 w-2.5 h-2.5 rounded-full bg-violet-400 shadow-[0_0_6px_rgba(167,139,250,0.9)]"></span>
          </div>
        </div>
        <p class="text-xs text-white/30 text-right">{{ t('home.xpToGo', { xp: xpToNext }) }}</p>
      </div>

      <div>
        <p class="text-xs text-white/40 font-semibold tracking-widest uppercase mb-3">{{ t('home.thisWeek') }}</p>
        <div class="grid grid-cols-2 gap-3">
          <div class="bg-[#1e1e24] rounded-2xl p-4 text-center ui-card">
            <p class="text-2xl font-bold">{{ stats?.workouts_completed ?? '—' }}</p>
            <p class="text-xs text-white/40 mt-1">{{ t('home.sessionsLogged') }}</p>
          </div>
          <div class="bg-[#1e1e24] rounded-2xl p-4 text-center ui-card">
            <p class="text-2xl font-bold text-violet-400">{{ stats?.total_xp ?? '—' }}</p>
            <p class="text-xs text-white/40 mt-1">{{ t('home.totalXp') }}</p>
          </div>
        </div>
      </div>

      <div v-if="recentWorkouts.length">
        <p class="text-xs text-white/40 font-semibold tracking-widest uppercase mb-3">{{ t('home.recentWorkouts') }}</p>
        <div class="space-y-3">
          <div
            v-for="w in recentWorkouts"
            :key="w.id"
            class="bg-[#1e1e24] rounded-2xl p-4 flex items-center gap-3 cursor-pointer ui-card hover:bg-[#26262e] transition-colors"
            @click="router.push('/history')"
          >
            <div class="w-10 h-10 rounded-xl bg-violet-500/20 border border-violet-500/20 flex items-center justify-center shrink-0">
              <svg class="w-5 h-5 text-violet-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z"/>
              </svg>
            </div>
            <div class="flex-1 min-w-0 text-left">
              <p class="font-semibold text-sm truncate">{{ w.ai_response?.title || w.muscle_group }}</p>
              <p class="text-xs text-white/40 mt-0.5">{{ relativeDate(w.completed_at) }} · {{ w.ai_response?.main?.length ?? 0 }} exercises · {{ w.duration_minutes }} min</p>
            </div>
            <span class="shrink-0 text-xs bg-green-500/20 text-green-400 border border-green-500/30 px-2.5 py-1 rounded-full font-medium">+{{ w.xp_earned }} xp</span>
          </div>
        </div>
      </div>

      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '../api/client'
import { toMessage } from '../utils/error'

const router = useRouter()
const { t, tm, locale } = useI18n()

const greeting = computed(() => {
  const h = new Date().getHours()
  const key = h >= 5 && h < 12 ? 'morning'
            : h >= 12 && h < 17 ? 'afternoon'
            : h >= 17 && h < 21 ? 'evening'
            : 'night'
  const list = tm(`home.greetings.${key}`) as string[]
  return list[Math.floor(Math.random() * list.length)]
})

interface Stats {
  username:           string
  total_xp:           number
  level:              number
  current_level_xp:   number
  next_level_xp:      number
  workouts_completed: number
}

interface NullTime { Time: string; Valid: boolean }
interface Workout {
  id:               number
  muscle_group:     string
  duration_minutes: number
  xp_earned:        number
  completed_at:     NullTime | null
  ai_response:      { title: string; main: any[] } | null
}

const stats          = ref<Stats | null>(null)
const recentWorkouts = ref<Workout[]>([])
const loading        = ref(true)
const error          = ref('')

function relativeDate(completedAt: NullTime | null): string {
  if (!completedAt?.Valid) return '—'
  const d    = new Date(completedAt.Time)
  const now  = new Date()
  const diff = Math.floor((now.getTime() - d.getTime()) / 86_400_000)
  if (diff === 0) return t('home.today')
  if (diff === 1) return t('home.yesterday')
  return d.toLocaleDateString(locale.value === 'sv' ? 'sv-SE' : 'en-US', { day: 'numeric', month: 'short' })
}

const xpPercent = computed(() => {
  if (!stats.value) return 0
  const { total_xp, current_level_xp, next_level_xp } = stats.value
  if (next_level_xp === 0) return 100 // max level
  return Math.min(((total_xp - current_level_xp) / (next_level_xp - current_level_xp)) * 100, 100)
})

const xpToNext = computed(() => {
  if (!stats.value || stats.value.next_level_xp === 0) return 0
  return Math.max(stats.value.next_level_xp - stats.value.total_xp, 0)
})

onMounted(async () => {
  try {
    const [s, w] = await Promise.all([
      api.get<Stats>('/users/me/stats'),
      api.get<{ workouts: Workout[] }>('/workouts'),
    ])
    stats.value = s
    recentWorkouts.value = (w.workouts ?? [])
      .filter(x => x.completed_at?.Valid)
      .slice(0, 3)
  } catch (e: unknown) {
    error.value = toMessage(e, 'Failed to load stats')
  } finally {
    loading.value = false
  }
})
</script>

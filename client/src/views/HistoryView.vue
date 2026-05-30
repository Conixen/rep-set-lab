<template>
  <div class="p-5 space-y-4">
    <h1 class="text-2xl font-bold pt-2">Workout history</h1>

    <div v-if="loading" class="text-white/40 text-sm text-center py-8">Loading…</div>
    <p v-else-if="error" class="text-red-400 text-xs">{{ error }}</p>

    <div v-else-if="workouts.length === 0" class="text-white/40 text-sm text-center py-8">
      No workouts yet — generate one in the AI tab.
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="w in workouts"
        :key="w.id"
        class="bg-[#1e1e24] rounded-2xl p-4 space-y-3"
      >
        <div class="flex items-start justify-between gap-2">
          <div>
            <p class="font-semibold capitalize">{{ w.muscle_group }}</p>
            <p class="text-xs text-white/40 mt-0.5">{{ formatDateLong(w.created_at) }}</p>
          </div>
          <span
            :class="w.completed_at?.Valid ? 'bg-green-500/20 text-green-400' : 'bg-white/10 text-white/40'"
            class="shrink-0 text-xs px-2 py-1 rounded-full"
          >
            {{ w.completed_at?.Valid ? 'Completed' : 'Not done' }}
          </span>
        </div>

        <div class="flex flex-wrap gap-2 text-xs">
          <span class="bg-violet-500/20 text-violet-400 px-2 py-1 rounded-full">{{ w.ai_provider }}</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ w.duration_minutes }} min</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ w.xp_earned }} XP</span>
        </div>

        <div v-if="progressDelta(w)" class="text-xs text-violet-400 flex items-center gap-1">
          <span>↑ {{ progressDelta(w) }} min more than last {{ w.muscle_group }} session</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api/client'
import { toMessage } from '../utils/error'
import { formatDateLong } from '../utils/date'

interface NullTime {
  Time:  string
  Valid: boolean
}

interface Workout {
  id:               number
  muscle_group:     string
  duration_minutes: number
  ai_provider:      string
  xp_earned:        number
  completed_at:     NullTime | null
  created_at:       string
}

const workouts = ref<Workout[]>([])
const loading  = ref(true)
const error    = ref('')

onMounted(async () => {
  try {
    const data = await api.get<{ workouts: Workout[] }>('/workouts')
    workouts.value = data.workouts ?? []
  } catch (e: unknown) {
    error.value = toMessage(e, 'Failed to load history')
  } finally {
    loading.value = false
  }
})

function progressDelta(w: Workout): number | null {
  const prev = workouts.value.find(
    x => x.id !== w.id && x.muscle_group === w.muscle_group && new Date(x.created_at) < new Date(w.created_at)
  )
  if (!prev) return null
  const delta = w.duration_minutes - prev.duration_minutes
  return delta > 0 ? delta : null
}
</script>

<template>
  <div class="p-5 space-y-4">
    <h1 class="text-2xl font-bold pt-2">History</h1>

    <div class="flex gap-2 bg-white/5 p-1 rounded-xl">
      <button
        @click="tab = 'history'"
        :class="tab === 'history' ? 'bg-violet-500 text-white' : 'text-white/50'"
        class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
      >History</button>
      <button
        @click="tab = 'saved'"
        :class="tab === 'saved' ? 'bg-violet-500 text-white' : 'text-white/50'"
        class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
      >
        Saved
        <span v-if="saved.length" class="ml-1 text-xs opacity-70">({{ saved.length }})</span>
      </button>
    </div>

    <div v-if="loading" class="text-white/40 text-sm text-center py-8">Loading…</div>
    <p v-else-if="error" class="text-red-400 text-xs">{{ error }}</p>
    <p v-else-if="workouts.length === 0" class="text-white/40 text-sm text-center py-8">No workouts generated yet.</p>

    <template v-else-if="tab === 'history'">
      <div v-if="completed.length === 0" class="text-white/40 text-sm text-center py-8">
        No completed workouts yet.
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="w in completed" :key="w.id"
          class="bg-[#1e1e24] rounded-2xl p-4 space-y-3 cursor-pointer hover:bg-[#26262e] transition-colors ui-card"
          @click="openModal(w)"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="text-left">
              <p class="font-semibold capitalize">{{ workoutTitle(w) }}</p>
              <p class="text-xs text-white/40 mt-0.5">{{ formatDateLong(w.created_at) }}</p>
            </div>
            <span class="shrink-0 text-xs px-2 py-1 rounded-full bg-green-500/20 text-green-400">Completed</span>
          </div>
          <div class="flex flex-wrap gap-2 text-xs">
            <span class="bg-violet-500/20 text-violet-400 px-2 py-1 rounded-full">{{ w.ai_provider }}</span>
            <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ w.duration_minutes }} min</span>
            <span class="bg-green-500/20 text-green-400 border border-green-500/30 px-2 py-1 rounded-full">{{ w.xp_earned }} XP</span>
          </div>
          <div v-if="progressDelta(w)" class="text-xs text-violet-400">
            ↑ {{ progressDelta(w) }} min more than last {{ w.muscle_group }} session
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div v-if="saved.length === 0" class="text-white/40 text-sm text-center py-8">
        No saved workouts. Generate one in AI or save from Compare.
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="w in saved" :key="w.id"
          class="bg-[#1e1e24] rounded-2xl p-4 space-y-3 ui-card"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1 min-w-0 text-left">
              <p class="font-semibold capitalize truncate">{{ workoutTitle(w) }}</p>
              <p class="text-xs text-white/40 mt-0.5">{{ formatDateLong(w.created_at) }}</p>
            </div>
            <span class="shrink-0 text-xs px-2 py-1 rounded-full bg-white/10 text-white/40">Saved</span>
          </div>
          <div class="flex flex-wrap gap-2 text-xs">
            <span class="bg-violet-500/20 text-violet-400 px-2 py-1 rounded-full">{{ w.ai_provider }}</span>
            <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ w.duration_minutes }} min</span>
            <span class="bg-green-500/20 text-green-400 border border-green-500/30 px-2 py-1 rounded-full">{{ w.xp_earned }} XP</span>
          </div>
          <div class="flex gap-2">
            <button
              @click="openModal(w)"
              class="flex-1 bg-violet-500 hover:bg-violet-600 text-white font-medium py-2 rounded-xl text-sm transition-colors"
            >View workout</button>
          </div>
        </div>
      </div>
    </template>
  </div>

  <WorkoutModal
    v-if="modalWorkout"
    :workout="modalWorkout.ai_response"
    :workout-id="modalWorkout.id"
    :provider="modalWorkout.ai_provider"
    :duration-minutes="modalWorkout.duration_minutes"
    :is-completed="!!modalWorkout.completed_at?.Valid"
    :muscle-group="modalWorkout.muscle_group"
    :injuries="modalWorkout.injuries?.String"
    :goals="modalWorkout.goals?.String"
    @close="modalWorkout = null"
    @completed="onCompleted"
    @deleted="onDeleted(modalWorkout!.id)"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'
import { toMessage } from '../utils/error'
import { formatDateLong } from '../utils/date'
import WorkoutModal from '../components/WorkoutModal.vue'

interface NullString {
  String: string
  Valid: boolean
}

interface NullTime {
  Time:  string
  Valid: boolean
}

interface WorkoutResponse {
  title: string
  description: string
  warm_up: any[]
  main: any[]
  tips: string[]
}

interface Workout {
  id:               number
  muscle_group:     string
  duration_minutes: number
  ai_provider:      string
  xp_earned:        number
  completed_at:     NullTime | null
  created_at:       string
  ai_response:      WorkoutResponse
  injuries:         NullString | null
  goals:            NullString | null
}

const tab      = ref<'history' | 'saved'>('history')
const workouts = ref<Workout[]>([])
const loading  = ref(true)
const error    = ref('')

const completed = computed(() => workouts.value.filter(w => w.completed_at?.Valid))
const saved     = computed(() => workouts.value.filter(w => !w.completed_at?.Valid))

onMounted(async () => {
  try {
    const data = await api.get<{ workouts: Workout[] }>('/workouts')
    workouts.value = data.workouts ?? []
  } catch (e: unknown) {
    error.value = toMessage(e, 'Failed to load workouts')
  } finally {
    loading.value = false
  }
})

function workoutTitle(w: Workout): string {
  return w.ai_response?.title || w.muscle_group
}

function progressDelta(w: Workout): number | null {
  const prev = completed.value.find(
    x => x.id !== w.id && x.muscle_group === w.muscle_group && new Date(x.created_at) < new Date(w.created_at)
  )
  if (!prev) return null
  const delta = w.duration_minutes - prev.duration_minutes
  return delta > 0 ? delta : null
}

// Modal
const modalWorkout = ref<Workout | null>(null)

function openModal(w: Workout) {
  modalWorkout.value = w
}

function onCompleted() {
  if (!modalWorkout.value) return
  const idx = workouts.value.findIndex(w => w.id === modalWorkout.value!.id)
  if (idx !== -1) {
    workouts.value[idx] = {
      ...workouts.value[idx],
      completed_at: { Time: new Date().toISOString(), Valid: true },
    }
  }
  modalWorkout.value = null
}

function onDeleted(id: number) {
  workouts.value = workouts.value.filter(w => w.id !== id)
  modalWorkout.value = null
}
</script>

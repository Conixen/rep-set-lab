<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 z-50 bg-black/80 overflow-y-auto"
      @click.self="$emit('close')"
    >
      <div class="min-h-full flex items-start justify-center p-4 py-8">
        <div v-if="!workout" class="bg-[#1e1e24] rounded-2xl w-full max-w-lg p-8 text-center text-white/40">
          No workout data.<button @click="$emit('close')" class="block mx-auto mt-4 text-violet-400 text-sm">Close</button>
        </div>
        <div v-else class="bg-[#1e1e24] rounded-2xl w-full max-w-lg overflow-hidden">

          <!-- Header -->
          <div class="flex items-start justify-between p-5 pb-3">
            <div class="flex-1 min-w-0 pr-3">
              <h2 class="text-lg font-bold">{{ workout.title }}</h2>
              <p class="text-sm text-white/50 mt-1 leading-relaxed">{{ workout.description }}</p>
            </div>
            <button
              @click="$emit('close')"
              class="text-white/40 hover:text-white text-xl leading-none shrink-0 mt-0.5"
              aria-label="Close"
            >✕</button>
          </div>

          <!-- Badges -->
          <div class="flex gap-2 px-5 pb-4 flex-wrap text-xs">
            <span v-if="provider" class="bg-violet-500/20 text-violet-400 px-2 py-1 rounded-full capitalize">{{ provider }}</span>
            <span v-if="durationMinutes" class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ durationMinutes }} min</span>
            <span v-if="xpToEarn > 0" class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ xpToEarn }} XP on completion</span>
          </div>

          <!-- Exercises -->
          <div class="px-5 pb-5 space-y-5">
            <section v-if="workout.warm_up?.length">
              <p class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-3">Warm Up</p>
              <div class="space-y-2">
                <div
                  v-for="ex in workout.warm_up" :key="ex.name"
                  class="bg-white/5 rounded-xl overflow-hidden flex"
                >
                  <img
                    v-if="gifMap[ex.name]"
                    :src="gifMap[ex.name]"
                    :alt="ex.name"
                    class="w-20 h-20 object-cover shrink-0"
                    loading="lazy"
                  />
                  <div class="p-3 flex-1 min-w-0">
                    <p class="font-medium text-sm">{{ ex.name }}</p>
                    <p class="text-xs text-white/40 mt-0.5">
                      <span v-if="ex.sets && ex.reps">{{ ex.sets }} × {{ ex.reps }} reps</span>
                      <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                      <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
                    </p>
                    <p v-if="ex.notes" class="text-xs text-white/30 mt-1 leading-relaxed">{{ ex.notes }}</p>
                  </div>
                </div>
              </div>
            </section>

            <section v-if="workout.main?.length">
              <p class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-3">Main</p>
              <div class="space-y-2">
                <div
                  v-for="ex in workout.main" :key="ex.name"
                  class="bg-white/5 rounded-xl overflow-hidden flex"
                >
                  <img
                    v-if="gifMap[ex.name]"
                    :src="gifMap[ex.name]"
                    :alt="ex.name"
                    class="w-20 h-20 object-cover shrink-0"
                    loading="lazy"
                  />
                  <div class="p-3 flex-1 min-w-0">
                    <p class="font-medium text-sm">{{ ex.name }}</p>
                    <p class="text-xs text-white/40 mt-0.5">
                      <span v-if="ex.sets && ex.reps">{{ ex.sets }} × {{ ex.reps }} reps</span>
                      <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                      <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
                    </p>
                    <p v-if="ex.notes" class="text-xs text-white/30 mt-1 leading-relaxed">{{ ex.notes }}</p>
                  </div>
                </div>
              </div>
            </section>

            <section v-if="workout.tips?.length">
              <p class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-2">Tips</p>
              <ul class="space-y-1.5">
                <li v-for="tip in workout.tips" :key="tip" class="text-sm text-white/60 flex gap-2">
                  <span class="text-violet-400 shrink-0">•</span>{{ tip }}
                </li>
              </ul>
            </section>
          </div>

          <!-- XP banner (shown after completing) -->
          <div v-if="xpResult" class="mx-5 mb-4 bg-violet-500/20 border border-violet-500/30 rounded-2xl p-4 text-center space-y-0.5">
            <p class="text-violet-400 font-bold text-xl">+{{ xpResult.xp_earned }} XP</p>
            <p class="text-white/60 text-sm">
              {{ xpResult.leveled_up ? `Level up! Level ${xpResult.level}` : `Total: ${xpResult.total_xp} XP · Level ${xpResult.level}` }}
            </p>
          </div>

          <p v-if="actionError" class="px-5 pb-3 text-red-400 text-xs">{{ actionError }}</p>

          <!-- Actions -->
          <div class="p-5 pt-0 flex gap-2">
            <!-- Not saved yet (from compare) -->
            <template v-if="mode === 'compare-unsaved'">
              <button
                @click="handleSave"
                :disabled="saving"
                class="flex-1 bg-violet-500 hover:bg-violet-600 disabled:opacity-50 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
              >{{ saving ? 'Saving…' : 'Save workout' }}</button>
            </template>

            <!-- Saved from compare (just saved) -->
            <template v-else-if="mode === 'compare-saved'">
              <div class="flex-1 bg-green-500/10 text-green-400 py-3 rounded-2xl text-sm font-semibold text-center">Saved ✓</div>
            </template>

            <!-- In DB, not completed (from AI or Saved tab) -->
            <template v-else-if="mode === 'saved' && !xpResult">
              <div class="flex-1 bg-green-500/10 text-green-400 py-3 rounded-2xl text-sm font-semibold text-center">Saved ✓</div>
              <button
                @click="handleComplete"
                :disabled="completing"
                class="flex-1 bg-violet-500 hover:bg-violet-600 disabled:opacity-50 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
              >{{ completing ? 'Completing…' : `Complete · ${xpToEarn} XP` }}</button>
              <button
                @click="handleDelete"
                :disabled="deleting"
                class="px-4 py-3 rounded-2xl bg-red-500/10 hover:bg-red-500/20 text-red-400 text-sm transition-colors"
                aria-label="Delete workout"
              >{{ deleting ? '…' : '🗑' }}</button>
            </template>

            <!-- Completed -->
            <template v-else-if="mode === 'completed' || xpResult">
              <div class="flex-1 bg-white/5 text-white/40 py-3 rounded-2xl text-sm font-semibold text-center">Completed ✓</div>
            </template>
          </div>

        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'
import { findLibraryMatch, type LibraryExercise } from '../utils/exerciseMatch'
import { toMessage } from '../utils/error'
import { mediaUrl } from '../utils/mediaUrl'

interface Exercise {
  name: string
  sets?: number
  reps?: number
  duration_seconds?: number
  rest_seconds?: number
  notes?: string
}

interface WorkoutResponse {
  title: string
  description: string
  warm_up: Exercise[]
  main: Exercise[]
  tips: string[]
}

interface XPResult {
  xp_earned: number
  total_xp: number
  level: number
  leveled_up: boolean
}

const props = defineProps<{
  workout: WorkoutResponse
  workoutId?: number
  provider?: string
  durationMinutes?: number
  isCompleted?: boolean
  // for save-from-compare
  muscleGroup?: string
  injuries?: string
  goals?: string
  environment?: string
  prompt?: string
}>()

const emit = defineEmits<{
  close: []
  saved: [workoutId: number]
  completed: [result: XPResult]
  deleted: []
}>()

// Derive modal mode
const internalWorkoutId = ref<number | undefined>(props.workoutId)
const mode = computed(() => {
  if (props.isCompleted) return 'completed'
  if (!internalWorkoutId.value && !props.workoutId) return 'compare-unsaved'
  if (internalWorkoutId.value && !props.workoutId) return 'compare-saved'
  return 'saved'
})

const xpToEarn = computed(() => {
  if (props.durationMinutes) return props.durationMinutes * 10
  return 0
})

// GIF map
const gifMap = ref<Record<string, string>>({})
onMounted(async () => {
  try {
    const library = await api.get<LibraryExercise[]>('/exercises')
    const allNames = [
      ...(props.workout.warm_up ?? []).map(e => e.name),
      ...(props.workout.main ?? []).map(e => e.name),
    ]
    const map: Record<string, string> = {}
    for (const name of allNames) {
      if (map[name] !== undefined) continue
      const muscleGroupArr = props.muscleGroup ? [props.muscleGroup] : []
      const match = findLibraryMatch(name, muscleGroupArr, library, props.environment ?? 'gym')
      const url = mediaUrl(match?.gif_url)
      if (url) map[name] = url
    }
    gifMap.value = map
  } catch {
    // non-fatal: exercises shown without GIFs
  }
})

// Actions
const saving      = ref(false)
const completing  = ref(false)
const deleting    = ref(false)
const actionError = ref('')
const xpResult    = ref<XPResult | null>(null)

async function handleSave() {
  saving.value = true
  actionError.value = ''
  try {
    const data = await api.post<{ workout: { id: number } }>('/workouts/save-from-compare', {
      muscle_group:     props.muscleGroup ?? '',
      duration_minutes: props.durationMinutes ?? 45,
      ai_provider:      props.provider ?? '',
      environment:      props.environment ?? 'gym',
      injuries:         props.injuries ?? '',
      goals:            props.goals ?? '',
      prompt:           props.prompt ?? '',
      response:         props.workout,
    })
    internalWorkoutId.value = data.workout.id
    emit('saved', data.workout.id)
  } catch (e: unknown) {
    actionError.value = toMessage(e, 'Failed to save workout')
  } finally {
    saving.value = false
  }
}

async function handleComplete() {
  const id = internalWorkoutId.value ?? props.workoutId
  if (!id) return
  completing.value = true
  actionError.value = ''
  try {
    const result = await api.post<XPResult>(`/workouts/${id}/complete`, {})
    xpResult.value = result
    emit('completed', result)
  } catch (e: unknown) {
    actionError.value = toMessage(e, 'Failed to complete workout')
  } finally {
    completing.value = false
  }
}

async function handleDelete() {
  const id = internalWorkoutId.value ?? props.workoutId
  if (!id) return
  deleting.value = true
  actionError.value = ''
  try {
    await api.delete(`/workouts/${id}`)
    emit('deleted')
    emit('close')
  } catch (e: unknown) {
    actionError.value = toMessage(e, 'Failed to delete workout')
  } finally {
    deleting.value = false
  }
}
</script>

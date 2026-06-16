<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 z-50 bg-black/80 overflow-y-auto"
      @click.self="$emit('close')"
    >
      <div class="min-h-full flex items-center justify-center p-4 py-8">
        <div v-if="!workout" class="bg-[#1e1e24] rounded-2xl w-full max-w-lg p-8 text-center text-white/40">
          {{ t('modal.noData') }}<button @click="$emit('close')" class="block mx-auto mt-4 text-violet-400 text-sm">{{ t('modal.close') }}</button>
        </div>
        <div v-else class="bg-[#1e1e24] rounded-2xl w-full max-w-lg overflow-hidden">

          <div class="flex items-start justify-between p-5 pb-3">
            <div class="flex-1 min-w-0 pr-3">
              <h2 class="text-lg font-bold">{{ workout.title }}</h2>
              <p class="text-sm text-white/70 mt-1 leading-relaxed">{{ workout.description }}</p>
            </div>
            <button
              @click="$emit('close')"
              class="text-white/40 hover:text-white text-xl leading-none shrink-0 mt-0.5"
              aria-label="Close"
            >✕</button>
          </div>

          <div class="flex gap-2 px-5 pb-4 flex-wrap text-xs">
            <span v-if="provider" class="bg-violet-500/20 text-violet-400 px-2 py-1 rounded-full capitalize">{{ provider }}</span>
            <span v-if="durationMinutes" class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ durationMinutes }} min</span>
            <span v-if="xpToEarn > 0" class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ t('modal.xpOnCompletion', { xp: xpToEarn }) }}</span>
          </div>

          <div class="px-5 pb-5 space-y-5">
            <section v-if="workout.warm_up?.length">
              <p class="text-xs font-semibold text-white/60 uppercase tracking-widest mb-3">{{ t('modal.warmUp') }}</p>
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
                    <div class="flex items-center justify-between mt-0.5 gap-2">
                      <p class="text-xs text-white/40">
                        <span v-if="ex.sets && ex.reps">{{ ex.sets }} × {{ ex.reps }} reps</span>
                        <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                      </p>
                      <template v-if="ex.rest_seconds">
                        <button
                          v-if="!activeTimer || activeTimer.name !== ex.name"
                          @click="startRest(ex.name, ex.rest_seconds!)"
                          class="text-xs bg-white/10 hover:bg-white/15 text-white/50 px-2 py-0.5 rounded-lg transition-colors shrink-0"
                        >{{ t('modal.rest', { seconds: ex.rest_seconds }) }}</button>
                        <div v-else class="flex items-center gap-1.5 shrink-0">
                          <span
                            class="text-xs font-mono font-bold tabular-nums"
                            :class="activeTimer.remaining <= 3 ? 'text-red-400 animate-pulse' : 'text-violet-400'"
                          >{{ activeTimer.remaining }}s</span>
                          <button @click="cancelRest" class="text-white/30 hover:text-white/60 text-xs">✕</button>
                        </div>
                      </template>
                    </div>
                    <p v-if="ex.notes" class="text-xs text-white/50 mt-1 leading-relaxed">{{ ex.notes }}</p>
                  </div>
                </div>
              </div>
            </section>

            <section v-if="workout.main?.length">
              <p class="text-xs font-semibold text-white/60 uppercase tracking-widest mb-3">{{ t('modal.main') }}</p>
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
                    <div class="flex items-center justify-between mt-0.5 gap-2">
                      <p class="text-xs text-white/40">
                        <span v-if="ex.sets && ex.reps">{{ ex.sets }} × {{ ex.reps }} reps</span>
                        <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                      </p>
                      <template v-if="ex.rest_seconds">
                        <button
                          v-if="!activeTimer || activeTimer.name !== ex.name"
                          @click="startRest(ex.name, ex.rest_seconds!)"
                          class="text-xs bg-white/10 hover:bg-white/15 text-white/50 px-2 py-0.5 rounded-lg transition-colors shrink-0"
                        >{{ t('modal.rest', { seconds: ex.rest_seconds }) }}</button>
                        <div v-else class="flex items-center gap-1.5 shrink-0">
                          <span
                            class="text-xs font-mono font-bold tabular-nums"
                            :class="activeTimer.remaining <= 3 ? 'text-red-400 animate-pulse' : 'text-violet-400'"
                          >{{ activeTimer.remaining }}s</span>
                          <button @click="cancelRest" class="text-white/30 hover:text-white/60 text-xs">✕</button>
                        </div>
                      </template>
                    </div>
                    <p v-if="ex.notes" class="text-xs text-white/50 mt-1 leading-relaxed">{{ ex.notes }}</p>
                  </div>
                </div>
              </div>
            </section>

            <section v-if="workout.tips?.length">
              <p class="text-xs font-semibold text-white/60 uppercase tracking-widest mb-2">{{ t('modal.tips') }}</p>
              <ul class="space-y-1.5">
                <li v-for="tip in workout.tips" :key="tip" class="text-sm text-white/60 flex gap-2 items-start">
                  <span class="text-violet-400 shrink-0 mt-0.5">•</span>
                  <span>{{ tip }}</span>
                </li>
              </ul>
            </section>
          </div>

          <div v-if="xpResult" class="mx-5 mb-4 bg-violet-500/20 border border-violet-500/30 rounded-2xl p-4 text-center space-y-0.5">
            <p class="text-violet-400 font-bold text-xl">+{{ xpResult.xp_earned }} XP</p>
            <p class="text-white/60 text-sm">
              {{ xpResult.leveled_up ? t('modal.levelUp', { level: xpResult.level }) : t('modal.totalXp', { total_xp: xpResult.total_xp, level: xpResult.level }) }}
            </p>
          </div>

          <p v-if="actionError" class="px-5 pb-3 text-red-400 text-xs">{{ actionError }}</p>

          <div class="p-5 pt-0 flex gap-2">
            <template v-if="mode === 'compare-unsaved'">
              <button
                @click="handleSave"
                :disabled="saving"
                class="flex-1 bg-violet-500 hover:bg-violet-600 disabled:opacity-50 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
              >{{ saving ? t('modal.saving') : t('modal.saveWorkout') }}</button>
            </template>

            <template v-else-if="mode === 'compare-saved'">
              <div class="flex-1 bg-green-500/10 text-green-400 py-3 rounded-2xl text-sm font-semibold text-center">{{ t('modal.saved') }}</div>
            </template>

            <template v-else-if="mode === 'saved' && !xpResult">
              <template v-if="!confirmingDelete">
                <div class="flex-1 bg-green-500/10 text-green-400 py-3 rounded-2xl text-sm font-semibold text-center">{{ t('modal.saved') }}</div>
                <button
                  @click="handleComplete"
                  :disabled="completing"
                  class="flex-1 bg-violet-500 hover:bg-violet-600 disabled:opacity-50 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
                >{{ completing ? t('modal.completing') : t('modal.complete', { xp: xpToEarn }) }}</button>
                <button
                  @click="confirmingDelete = true"
                  class="px-4 py-3 rounded-2xl bg-red-500/10 hover:bg-red-500/20 text-red-400 text-sm transition-colors"
                  aria-label="Delete workout"
                >🗑</button>
              </template>
              <template v-else>
                <span class="flex-1 text-sm text-white/60 flex items-center">{{ t('modal.removeConfirm') }}</span>
                <button
                  @click="handleDelete"
                  :disabled="deleting"
                  class="flex-1 bg-red-500/20 hover:bg-red-500/30 disabled:opacity-50 text-red-400 font-semibold py-3 rounded-2xl text-sm transition-colors"
                >{{ deleting ? '…' : t('modal.yesDelete') }}</button>
                <button
                  @click="confirmingDelete = false"
                  class="flex-1 bg-white/10 hover:bg-white/15 text-white/60 font-semibold py-3 rounded-2xl text-sm transition-colors"
                >{{ t('modal.cancel') }}</button>
              </template>
            </template>

            <template v-else-if="mode === 'completed' || xpResult">
              <div class="flex-1 bg-white/5 text-white/40 py-3 rounded-2xl text-sm font-semibold text-center">{{ t('modal.completed') }}</div>
            </template>
          </div>

        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()

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

// Lock background scroll while modal is open
onMounted(() => {
  const main = document.querySelector('main') as HTMLElement | null
  if (main) main.style.overflowY = 'hidden'
})
onUnmounted(() => {
  const main = document.querySelector('main') as HTMLElement | null
  if (main) main.style.overflowY = ''
  if (timerInterval) clearInterval(timerInterval)
})

// Rest countdown timer
const activeTimer = ref<{ name: string; remaining: number } | null>(null)
let timerInterval: ReturnType<typeof setInterval> | null = null

function startRest(name: string, seconds: number) {
  if (timerInterval) clearInterval(timerInterval)
  activeTimer.value = { name, remaining: seconds }
  timerInterval = setInterval(() => {
    if (!activeTimer.value) return
    if (activeTimer.value.remaining <= 1) {
      clearInterval(timerInterval!)
      timerInterval = null
      activeTimer.value = null
      if (navigator.vibrate) navigator.vibrate(300)
    } else {
      activeTimer.value.remaining--
    }
  }, 1000)
}

function cancelRest() {
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = null
  activeTimer.value = null
}

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
const saving            = ref(false)
const completing        = ref(false)
const deleting          = ref(false)
const confirmingDelete  = ref(false)
const actionError       = ref('')
const xpResult          = ref<XPResult | null>(null)

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

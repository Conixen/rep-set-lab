<template>
  <div class="p-5 space-y-5">
    <h1 class="text-2xl font-bold pt-2">AI Coach</h1>

    <!-- FORM -->
    <template v-if="!result">
      <select
        v-model="selectedProvider"
        class="w-full bg-[#1e1e24] rounded-xl px-4 py-3 text-sm text-white outline-none focus:ring-1 focus:ring-violet-500"
      >
        <option v-for="p in providers" :key="p.key" :value="p.key">{{ p.label }}</option>
      </select>

      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-4">
        <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Build your workout</p>

        <div class="space-y-2">
          <p class="text-sm text-white/60">Training environment</p>
          <div class="flex gap-2">
            <button
              v-for="env in environments" :key="env.key"
              @click="environment = env.key"
              :class="environment === env.key ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
              class="flex-1 py-2 rounded-xl text-sm font-medium transition-colors"
            >{{ env.label }}</button>
          </div>
        </div>

        <div class="space-y-2">
          <p class="text-sm text-white/60">Muscle groups <span class="text-white/30">(select one or more)</span></p>
          <div class="grid grid-cols-3 gap-2">
            <button
              v-for="mg in muscleGroups" :key="mg"
              @click="toggleMuscleGroup(mg)"
              :class="selectedMuscleGroups.includes(mg) ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
              class="py-2 rounded-xl text-sm font-medium transition-colors"
            >{{ mg }}</button>
          </div>
        </div>


        <div class="space-y-2">
          <p class="text-sm text-white/60">Duration</p>
          <div class="flex gap-2">
            <button
              v-for="d in durations" :key="d"
              @click="setDuration(d)"
              :class="duration === d && !customDuration ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
              class="flex-1 py-2 rounded-xl text-sm font-medium transition-colors"
            >{{ d }} min</button>
          </div>
          <input
            v-model="customDuration"
            type="number"
            min="5"
            max="180"
            placeholder="Or enter minutes manually..."
            class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
          />
        </div>

        <div class="space-y-2">
          <p class="text-sm text-white/60">Experience level</p>
          <div class="flex gap-2">
            <button
              v-for="lvl in levels" :key="lvl"
              @click="experience = lvl"
              :class="experience === lvl ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
              class="flex-1 py-2 rounded-xl text-sm font-medium transition-colors"
            >{{ lvl }}</button>
          </div>
        </div>

        <div class="space-y-2">
          <p class="text-sm text-white/60">Goal</p>
          <div class="grid grid-cols-2 gap-2">
            <button
              v-for="g in goals" :key="g"
              @click="goal = g"
              :class="goal === g ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
              class="py-2 rounded-xl text-sm font-medium transition-colors"
            >{{ g }}</button>
          </div>
        </div>

        <div class="space-y-2">
          <p class="text-sm text-white/60">Injuries / limitations</p>
          <input
            v-model="injuries"
            type="text"
            placeholder="e.g. bad knees, shoulder impingement..."
            class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
          />
        </div>

        <div class="space-y-2">
          <p class="text-sm text-white/60">Write your own prompt <span class="text-white/30">(optional)</span></p>
          <textarea
            v-model="prompt"
            rows="3"
            placeholder="e.g. I'm preparing for a powerlifting meet next month, focus on heavy compounds, keep rest periods under 90s..."
            class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500 resize-none"
          />
        </div>
      </div>

      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

      <button
        @click="generate"
        :disabled="loading"
        class="w-full bg-violet-500 hover:bg-violet-600 disabled:opacity-50 active:scale-95 transition-all text-white font-semibold py-4 rounded-2xl text-sm"
      >
        {{ loading ? 'Generating… this can take up to 15s' : `Generate ${selectedMuscleGroups.join(', ') || 'workout'} · ${effectiveDuration} min` }}
      </button>
    </template>

    <!-- RESULT -->
    <template v-else>
      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-5">
        <div>
          <h2 class="text-lg font-bold">{{ result.response.title }}</h2>
          <p class="text-sm text-white/60 mt-1">{{ result.response.description }}</p>
        </div>

        <div class="flex flex-wrap gap-2 text-xs">
          <span class="bg-violet-500/20 text-violet-400 px-2 py-1 rounded-full">{{ result.usage.provider }}</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ result.usage.input_tokens + result.usage.output_tokens }} tokens</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">${{ result.usage.cost_usd.toFixed(4) }}</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ result.workout.xp_earned }} XP on completion</span>
        </div>

        <section v-if="result.response.warm_up?.length">
          <p class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-3">Warm Up</p>
          <div class="space-y-2">
            <div v-for="ex in result.response.warm_up" :key="ex.name" class="bg-white/5 rounded-xl p-3">
              <p class="font-medium text-sm">{{ ex.name }}</p>
              <p class="text-xs text-white/40 mt-0.5">
                <span v-if="ex.sets && ex.reps">{{ ex.sets }} sets × {{ ex.reps }} reps</span>
                <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
              </p>
              <p v-if="ex.notes" class="text-xs text-white/30 mt-1">{{ ex.notes }}</p>
            </div>
          </div>
        </section>

        <section v-if="result.response.main?.length">
          <p class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-3">Main</p>
          <div class="space-y-2">
            <div v-for="ex in result.response.main" :key="ex.name" class="bg-white/5 rounded-xl p-3">
              <p class="font-medium text-sm">{{ ex.name }}</p>
              <p class="text-xs text-white/40 mt-0.5">
                <span v-if="ex.sets && ex.reps">{{ ex.sets }} sets × {{ ex.reps }} reps</span>
                <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
              </p>
              <p v-if="ex.notes" class="text-xs text-white/30 mt-1">{{ ex.notes }}</p>
            </div>
          </div>
        </section>

        <section v-if="result.response.cool_down?.length">
          <p class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-3">Cool Down</p>
          <div class="space-y-2">
            <div v-for="ex in result.response.cool_down" :key="ex.name" class="bg-white/5 rounded-xl p-3">
              <p class="font-medium text-sm">{{ ex.name }}</p>
              <p class="text-xs text-white/40 mt-0.5">
                <span v-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
              </p>
              <p v-if="ex.notes" class="text-xs text-white/30 mt-1">{{ ex.notes }}</p>
            </div>
          </div>
        </section>

        <section v-if="result.response.tips?.length">
          <p class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-2">Tips</p>
          <ul class="space-y-1">
            <li v-for="tip in result.response.tips" :key="tip" class="text-sm text-white/60 flex gap-2">
              <span class="text-violet-400 shrink-0">•</span>{{ tip }}
            </li>
          </ul>
        </section>
      </div>

      <!-- XP earned banner -->
      <div v-if="xpResult" class="bg-violet-500/20 border border-violet-500/30 rounded-2xl p-4 text-center space-y-1">
        <p class="text-violet-400 font-bold text-xl">+{{ xpResult.xp_earned }} XP</p>
        <p class="text-white/60 text-sm">
          {{ xpResult.leveled_up ? `Level up! You are now level ${xpResult.level}` : `Total: ${xpResult.total_xp} XP · Level ${xpResult.level}` }}
        </p>
      </div>

      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

      <div class="flex gap-3">
        <button
          @click="reset"
          class="flex-1 bg-white/10 hover:bg-white/15 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
        >New Workout</button>
        <button
          v-if="!xpResult"
          @click="complete"
          :disabled="completing"
          class="flex-1 bg-violet-500 hover:bg-violet-600 disabled:opacity-50 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
        >{{ completing ? 'Saving…' : `Complete · ${result.workout.xp_earned} XP` }}</button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '../api/client'

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
  cool_down: Exercise[]
  tips: string[]
}

interface GenerateResult {
  workout: { id: number; xp_earned: number }
  response: WorkoutResponse
  usage: { provider: string; input_tokens: number; output_tokens: number; cost_usd: number }
}

interface XPResult {
  xp_earned: number
  total_xp: number
  level: number
  leveled_up: boolean
}

const providers = [
  { key: 'gemini-2.5-flash', label: 'Gemini 2.5 Flash' },
  { key: 'gemini-2.5-pro',   label: 'Gemini 2.5 Pro' },
  { key: 'claude',           label: 'Claude Sonnet (WIP)' },
  { key: 'openai',           label: 'GPT-4o (WIP)' },
]
const environments = [
  { key: 'gym',     label: '🏋️ Gym' },
  { key: 'home',    label: '🏠 Home' },
  { key: 'outdoor', label: '🌳 Outdoor' },
]
const muscleGroups = ['Chest', 'Back', 'Legs', 'Shoulders', 'Arms', 'Core']
const durations    = [30, 45, 60, 90]
const levels       = ['Beginner', 'Intermediate', 'Advanced']
const goals        = ['Muscle gain', 'Fat loss', 'Strength', 'Endurance']

const selectedProvider     = ref('gemini-2.5-flash')
const selectedMuscleGroups = ref<string[]>(['Chest'])
const duration             = ref(45)
const customDuration       = ref('')
const experience           = ref('Beginner')
const goal                 = ref('Muscle gain')
const injuries             = ref('')
const prompt               = ref('')
const environment          = ref('gym')

const effectiveDuration = computed(() => {
  const custom = parseInt(customDuration.value)
  return !isNaN(custom) && custom > 0 ? custom : duration.value
})

function toggleMuscleGroup(mg: string) {
  const idx = selectedMuscleGroups.value.indexOf(mg)
  if (idx === -1) {
    selectedMuscleGroups.value.push(mg)
  } else {
    selectedMuscleGroups.value.splice(idx, 1)
  }
}

function setDuration(d: number) {
  duration.value = d
  customDuration.value = ''
}

const loading    = ref(false)
const completing = ref(false)
const error      = ref('')
const result     = ref<GenerateResult | null>(null)
const xpResult   = ref<XPResult | null>(null)

async function generate() {
  error.value = ''
  loading.value = true
  try {
    result.value = await api.post<GenerateResult>('/workouts/generate', {
      ai_provider:      selectedProvider.value,
      muscle_group:     selectedMuscleGroups.value.map(mg => mg.toLowerCase()).join(', '),
      duration_minutes: effectiveDuration.value,
      goals:            `${goal.value} · ${experience.value}`,
      injuries:         injuries.value,
      prompt:           prompt.value,
      environment:      environment.value,
    })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to generate workout'
  } finally {
    loading.value = false
  }
}

async function complete() {
  if (!result.value) return
  error.value = ''
  completing.value = true
  try {
    xpResult.value = await api.post<XPResult>(`/workouts/${result.value.workout.id}/complete`, {})
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to complete workout'
  } finally {
    completing.value = false
  }
}

function reset() {
  result.value  = null
  xpResult.value = null
  error.value   = ''
}
</script>

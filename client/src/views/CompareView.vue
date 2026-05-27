<template>
  <div class="p-5 space-y-5">
    <div>
      <h1 class="text-2xl font-bold pt-2">Provider Compare</h1>
      <p class="text-sm text-white/40 mt-1">Run the same prompt through all active providers in parallel.</p>
    </div>

    <!-- FORM -->
    <template v-if="!results">
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
            type="number" min="5" max="180"
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
          <p class="text-sm text-white/60">Custom prompt <span class="text-white/30">(optional)</span></p>
          <textarea
            v-model="prompt"
            rows="3"
            placeholder="e.g. Focus on compound movements, keep rest periods short..."
            class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500 resize-none"
          />
        </div>
      </div>

      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

      <button
        @click="compare"
        :disabled="loading"
        class="w-full bg-violet-500 hover:bg-violet-600 disabled:opacity-50 active:scale-95 transition-all text-white font-semibold py-4 rounded-2xl text-sm"
      >
        <span v-if="loading" class="flex items-center justify-center gap-2">
          <span class="inline-block w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
          Querying all providers… up to 30s
        </span>
        <span v-else>Compare all providers · {{ selectedMuscleGroups.join(', ') || 'workout' }} · {{ effectiveDuration }} min</span>
      </button>
    </template>

    <!-- RESULTS -->
    <template v-else>
      <!-- Summary bar -->
      <div class="grid grid-cols-3 gap-2 text-center text-xs">
        <div class="bg-[#1e1e24] rounded-xl p-3">
          <p class="text-white/40 mb-1">Fastest</p>
          <p class="font-semibold text-green-400 capitalize">{{ fastest?.provider ?? '—' }}</p>
          <p class="text-white/30 mt-0.5">{{ fastest ? fastest.latency_ms + 'ms' : '' }}</p>
        </div>
        <div class="bg-[#1e1e24] rounded-xl p-3">
          <p class="text-white/40 mb-1">Cheapest</p>
          <p class="font-semibold text-blue-400 capitalize">{{ cheapest?.provider ?? '—' }}</p>
          <p class="text-white/30 mt-0.5">{{ cheapest ? '$' + cheapest.usage!.cost_usd.toFixed(4) : '' }}</p>
        </div>
        <div class="bg-[#1e1e24] rounded-xl p-3">
          <p class="text-white/40 mb-1">Most tokens</p>
          <p class="font-semibold text-violet-400 capitalize">{{ mostTokens?.provider ?? '—' }}</p>
          <p class="text-white/30 mt-0.5">{{ mostTokens ? totalTokens(mostTokens) + ' tok' : '' }}</p>
        </div>
      </div>

      <!-- Provider cards -->
      <div class="space-y-4 lg:grid lg:grid-cols-3 lg:gap-4 lg:space-y-0">
        <div
          v-for="r in results" :key="r.provider"
          class="bg-[#1e1e24] rounded-2xl p-4 space-y-3"
        >
          <!-- Header -->
          <div class="flex items-center justify-between">
            <span class="font-semibold capitalize text-sm">{{ providerLabel(r.provider) }}</span>
            <span
              :class="r.error ? 'bg-red-500/20 text-red-400' : 'bg-green-500/20 text-green-400'"
              class="text-xs px-2 py-0.5 rounded-full"
            >{{ r.error ? 'Error' : 'Valid JSON' }}</span>
          </div>

          <!-- Error state -->
          <p v-if="r.error" class="text-red-400 text-xs break-words">{{ r.error }}</p>

          <!-- Stats row -->
          <div v-else class="flex flex-wrap gap-2 text-xs">
            <span class="bg-white/5 text-white/50 px-2 py-1 rounded-full">{{ r.latency_ms }}ms</span>
            <span class="bg-white/5 text-white/50 px-2 py-1 rounded-full">{{ r.usage!.input_tokens }} in / {{ r.usage!.output_tokens }} out</span>
            <span class="bg-white/5 text-white/50 px-2 py-1 rounded-full">${{ r.usage!.cost_usd.toFixed(4) }}</span>
          </div>

          <!-- Workout preview -->
          <template v-if="r.response">
            <div>
              <p class="font-medium text-sm">{{ r.response.title }}</p>
              <p class="text-xs text-white/40 mt-0.5 line-clamp-2">{{ r.response.description }}</p>
            </div>

            <!-- Exercises toggle -->
            <button
              @click="toggleExpand(r.provider)"
              class="text-xs text-violet-400 hover:text-violet-300 transition-colors"
            >
              {{ expanded.has(r.provider) ? 'Hide exercises ▲' : 'Show exercises ▼' }}
            </button>

            <div v-if="expanded.has(r.provider)" class="space-y-3">
              <div v-if="r.response.warm_up?.length">
                <p class="text-xs text-white/30 uppercase tracking-widest mb-1.5">Warm Up</p>
                <div class="space-y-1.5">
                  <div v-for="ex in r.response.warm_up" :key="ex.name" class="bg-white/5 rounded-xl px-3 py-2">
                    <p class="text-sm font-medium">{{ ex.name }}</p>
                    <p class="text-xs text-white/40">
                      <span v-if="ex.sets && ex.reps">{{ ex.sets }}×{{ ex.reps }}</span>
                      <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                      <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
                    </p>
                  </div>
                </div>
              </div>
              <div v-if="r.response.main?.length">
                <p class="text-xs text-white/30 uppercase tracking-widest mb-1.5">Main</p>
                <div class="space-y-1.5">
                  <div v-for="ex in r.response.main" :key="ex.name" class="bg-white/5 rounded-xl px-3 py-2">
                    <p class="text-sm font-medium">{{ ex.name }}</p>
                    <p class="text-xs text-white/40">
                      <span v-if="ex.sets && ex.reps">{{ ex.sets }}×{{ ex.reps }}</span>
                      <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                      <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
                    </p>
                    <p v-if="ex.notes" class="text-xs text-white/30 mt-0.5">{{ ex.notes }}</p>
                  </div>
                </div>
              </div>
              <div v-if="r.response.cool_down?.length">
                <p class="text-xs text-white/30 uppercase tracking-widest mb-1.5">Cool Down</p>
                <div class="space-y-1.5">
                  <div v-for="ex in r.response.cool_down" :key="ex.name" class="bg-white/5 rounded-xl px-3 py-2">
                    <p class="text-sm font-medium">{{ ex.name }}</p>
                    <p class="text-xs text-white/40">
                      <span v-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- Data table for report copy-paste -->
      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-3">
        <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Raw numbers</p>
        <div class="overflow-x-auto">
          <table class="w-full text-xs text-left">
            <thead>
              <tr class="text-white/30 border-b border-white/10">
                <th class="pb-2 pr-4">Provider</th>
                <th class="pb-2 pr-4">Latency</th>
                <th class="pb-2 pr-4">Input tok</th>
                <th class="pb-2 pr-4">Output tok</th>
                <th class="pb-2 pr-4">Cost (USD)</th>
                <th class="pb-2">Valid JSON</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-white/5">
              <tr v-for="r in results" :key="r.provider" class="text-white/70">
                <td class="py-2 pr-4 capitalize font-medium">{{ r.provider }}</td>
                <td class="py-2 pr-4">{{ r.latency_ms }}ms</td>
                <td class="py-2 pr-4">{{ r.usage?.input_tokens ?? '—' }}</td>
                <td class="py-2 pr-4">{{ r.usage?.output_tokens ?? '—' }}</td>
                <td class="py-2 pr-4">{{ r.usage ? '$' + r.usage.cost_usd.toFixed(5) : '—' }}</td>
                <td class="py-2">
                  <span :class="r.error ? 'text-red-400' : 'text-green-400'">{{ r.error ? 'No' : 'Yes' }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

      <button
        @click="reset"
        class="w-full bg-white/10 hover:bg-white/15 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
      >New comparison</button>
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

interface ProviderResult {
  provider: string
  response?: WorkoutResponse
  usage?: { input_tokens: number; output_tokens: number; cost_usd: number }
  error?: string
  latency_ms: number
}

const environments = [
  { key: 'gym',     label: '🏋️ Gym' },
  { key: 'home',    label: '🏠 Home' },
  { key: 'outdoor', label: '🌳 Outdoor' },
]
const muscleGroups = ['Chest', 'Back', 'Legs', 'Shoulders', 'Arms', 'Core']
const durations    = [30, 45, 60, 90]
const levels       = ['Beginner', 'Intermediate', 'Advanced']
const goals        = ['Muscle gain', 'Fat loss', 'Strength', 'Endurance']

const providerLabels: Record<string, string> = {
  claude: 'Claude Sonnet',
  openai: 'GPT-4o',
  gemini: 'Gemini 1.5 Pro',
}
function providerLabel(key: string) { return providerLabels[key] ?? key }

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
  if (idx === -1) selectedMuscleGroups.value.push(mg)
  else selectedMuscleGroups.value.splice(idx, 1)
}

function setDuration(d: number) {
  duration.value = d
  customDuration.value = ''
}

const loading  = ref(false)
const error    = ref('')
const results  = ref<ProviderResult[] | null>(null)
const expanded = ref<Set<string>>(new Set())

const successResults = computed(() =>
  (results.value ?? []).filter(r => !r.error && r.usage)
)

const fastest = computed(() =>
  successResults.value.reduce<ProviderResult | null>((best, r) =>
    best === null || r.latency_ms < best.latency_ms ? r : best, null)
)

const cheapest = computed(() =>
  successResults.value.reduce<ProviderResult | null>((best, r) =>
    best === null || r.usage!.cost_usd < best.usage!.cost_usd ? r : best, null)
)

const mostTokens = computed(() =>
  successResults.value.reduce<ProviderResult | null>((best, r) =>
    best === null || totalTokens(r) > totalTokens(best) ? r : best, null)
)

function totalTokens(r: ProviderResult) {
  return (r.usage?.input_tokens ?? 0) + (r.usage?.output_tokens ?? 0)
}

function toggleExpand(provider: string) {
  if (expanded.value.has(provider)) expanded.value.delete(provider)
  else expanded.value.add(provider)
}

async function compare() {
  error.value = ''
  loading.value = true
  try {
    const data = await api.post<{ results: ProviderResult[] }>('/ai/compare', {
      muscle_group:     selectedMuscleGroups.value.map(mg => mg.toLowerCase()).join(', '),
      duration_minutes: effectiveDuration.value,
      goals:            `${goal.value} · ${experience.value}`,
      injuries:         injuries.value,
      prompt:           prompt.value,
      environment:      environment.value,
    })
    results.value = data.results.sort((a, b) => a.latency_ms - b.latency_ms)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to run comparison'
  } finally {
    loading.value = false
  }
}

function reset() {
  results.value = null
  expanded.value = new Set()
  error.value = ''
}
</script>

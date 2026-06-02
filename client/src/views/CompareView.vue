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

        <div class="flex items-center gap-2">
          <p class="text-sm text-white/60">Workout language</p>
          <div class="flex gap-1 ml-auto">
            <button
              @click="language = 'en'"
              :class="language === 'en' ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
              class="px-3 py-1 rounded-lg text-xs font-medium transition-colors"
            >EN</button>
            <button
              @click="language = 'sv'"
              :class="language === 'sv' ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
              class="px-3 py-1 rounded-lg text-xs font-medium transition-colors"
            >SV</button>
          </div>
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
          Querying all providers… up to 90s
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
          <p class="text-white/40 mb-1">Best library match</p>
          <p class="font-semibold text-violet-400 capitalize">{{ bestLibraryMatch?.provider ?? '—' }}</p>
          <p class="text-white/30 mt-0.5">{{ bestLibraryMatch ? pct(bestLibraryMatch.library_match!.match_rate) : '' }}</p>
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

          <p v-if="r.error" class="text-red-400 text-xs break-words">{{ r.error }}</p>

          <template v-if="r.response">
            <!-- Performance chips -->
            <div class="flex flex-wrap gap-1.5 text-xs">
              <span class="bg-white/5 text-white/50 px-2 py-1 rounded-full">{{ r.latency_ms }}ms</span>
              <span class="bg-white/5 text-white/50 px-2 py-1 rounded-full">{{ r.usage!.input_tokens }}↑ {{ r.usage!.output_tokens }}↓ tok</span>
              <span class="bg-white/5 text-white/50 px-2 py-1 rounded-full">${{ r.usage!.cost_usd.toFixed(4) }}</span>
            </div>

            <!-- Quality metrics -->
            <div v-if="r.behavioral && r.library_match" class="space-y-1.5">
              <p class="text-xs text-white/30 uppercase tracking-widest">Quality metrics</p>
              <div class="grid grid-cols-2 gap-1.5 text-xs">
                <div class="bg-white/5 rounded-xl px-3 py-2">
                  <p class="text-white/40">Library match</p>
                  <p class="font-semibold" :class="matchColor(r.library_match.match_rate)">
                    {{ pct(r.library_match.match_rate) }}
                    <span class="text-white/30 font-normal">({{ r.library_match.match_count }}/{{ r.library_match.total_count }})</span>
                  </p>
                </div>
                <div class="bg-white/5 rounded-xl px-3 py-2">
                  <p class="text-white/40">Est. duration</p>
                  <p class="font-semibold" :class="durationColor(r.behavioral.estimated_minutes, effectiveDuration)">
                    {{ r.behavioral.estimated_minutes.toFixed(0) }} min
                    <span class="text-white/30 font-normal">/ {{ effectiveDuration }}</span>
                  </p>
                </div>
                <div class="bg-white/5 rounded-xl px-3 py-2">
                  <p class="text-white/40">Structure</p>
                  <p class="font-semibold">{{ r.behavioral.completeness_score }}/4</p>
                  <p class="text-white/30 mt-0.5">{{ sectionSummary(r.behavioral) }}</p>
                </div>
                <div class="bg-white/5 rounded-xl px-3 py-2">
                  <p class="text-white/40">Content richness</p>
                  <p class="font-semibold">{{ pct(r.behavioral.notes_present_rate) }}</p>
                  <p class="text-white/30 mt-0.5">exercises with notes</p>
                </div>
              </div>
              <div class="flex flex-wrap gap-1.5 text-xs">
                <span class="bg-white/5 text-white/50 px-2 py-1 rounded-full">{{ r.behavioral.char_count.toLocaleString() }} chars</span>
                <span v-if="r.behavioral.emoji_count > 0" class="bg-yellow-500/10 text-yellow-400 px-2 py-1 rounded-full">{{ r.behavioral.emoji_count }} emoji</span>
                <span v-if="r.behavioral.equipment_violations > 0" class="bg-red-500/10 text-red-400 px-2 py-1 rounded-full">{{ r.behavioral.equipment_violations }} equip. violation{{ r.behavioral.equipment_violations !== 1 ? 's' : '' }}</span>
                <span v-else class="bg-green-500/10 text-green-400 px-2 py-1 rounded-full">No equip. violations</span>
              </div>
            </div>

            <!-- Workout preview -->
            <div>
              <p class="font-medium text-sm">{{ r.response.title }}</p>
              <p class="text-xs text-white/40 mt-0.5 line-clamp-2">{{ r.response.description }}</p>
            </div>

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

      <!-- Raw numbers table -->
      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-3">
        <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Raw numbers</p>
        <div class="overflow-x-auto">
          <table class="w-full text-xs text-left">
            <thead>
              <tr class="text-white/30 border-b border-white/10">
                <th class="pb-2 pr-4">Provider</th>
                <th class="pb-2 pr-4">Latency</th>
                <th class="pb-2 pr-4">In tok</th>
                <th class="pb-2 pr-4">Out tok</th>
                <th class="pb-2 pr-4">Cost</th>
                <th class="pb-2 pr-4">Library</th>
                <th class="pb-2 pr-4">Est. min</th>
                <th class="pb-2 pr-4">Emoji</th>
                <th class="pb-2 pr-4">Equip. ✗</th>
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
                <td class="py-2 pr-4">{{ r.library_match ? pct(r.library_match.match_rate) : '—' }}</td>
                <td class="py-2 pr-4">{{ r.behavioral ? r.behavioral.estimated_minutes.toFixed(0) : '—' }}</td>
                <td class="py-2 pr-4">{{ r.behavioral?.emoji_count ?? '—' }}</td>
                <td class="py-2 pr-4">{{ r.behavioral?.equipment_violations ?? '—' }}</td>
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
import { toMessage } from '../utils/error'

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

interface BehavioralMetrics {
  char_count: number
  emoji_count: number
  equipment_violations: number
  completeness_score: number
  warm_up_count: number
  main_count: number
  cool_down_count: number
  tips_count: number
  avg_note_length: number
  notes_present_rate: number
  estimated_minutes: number
}

interface LibraryMatch {
  match_rate: number
  match_count: number
  total_count: number
}

interface ProviderResult {
  provider: string
  response?: WorkoutResponse
  usage?: { input_tokens: number; output_tokens: number; cost_usd: number }
  error?: string
  latency_ms: number
  behavioral?: BehavioralMetrics
  library_match?: LibraryMatch
}

const environments = [
  { key: 'gym',     label: 'Gym' },
  { key: 'home',    label: 'Home' },
  { key: 'outdoor', label: 'Outdoor' },
]
const muscleGroups = ['Chest', 'Back', 'Legs', 'Shoulders', 'Arms', 'Core', 'Calves', 'Wrists']
const durations    = [30, 45, 60, 90]
const levels       = ['Beginner', 'Intermediate', 'Advanced']
const goals        = ['Muscle gain', 'Fat loss', 'Strength', 'Endurance']

const providerLabels: Record<string, string> = {
  claude: 'Claude Sonnet',
  openai: 'GPT-4o',
  'gemini-2.5-flash': 'Gemini 2.5 Flash',
  'gemini-2.5-pro':   'Gemini 2.5 Pro',
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
const language             = ref('en')

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

const bestLibraryMatch = computed(() =>
  successResults.value
    .filter(r => r.library_match)
    .reduce<ProviderResult | null>((best, r) =>
      best === null || r.library_match!.match_rate > best.library_match!.match_rate ? r : best, null)
)

function toggleExpand(provider: string) {
  if (expanded.value.has(provider)) expanded.value.delete(provider)
  else expanded.value.add(provider)
}

function pct(rate: number) { return Math.round(rate * 100) + '%' }

function matchColor(rate: number) {
  if (rate >= 0.7) return 'text-green-400'
  if (rate >= 0.4) return 'text-yellow-400'
  return 'text-red-400'
}

function durationColor(estimated: number, requested: number) {
  const gap = Math.abs(estimated - requested)
  if (gap <= 5) return 'text-green-400'
  if (gap <= 15) return 'text-yellow-400'
  return 'text-red-400'
}

function sectionSummary(b: BehavioralMetrics) {
  const parts = []
  if (b.warm_up_count) parts.push(`${b.warm_up_count}wu`)
  if (b.main_count)    parts.push(`${b.main_count}main`)
  if (b.cool_down_count) parts.push(`${b.cool_down_count}cd`)
  if (b.tips_count)    parts.push(`${b.tips_count}tips`)
  return parts.join(' · ')
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
      language:         language.value,
    })
    results.value = data.results.sort((a, b) => a.latency_ms - b.latency_ms)
  } catch (e: unknown) {
    error.value = toMessage(e, 'Failed to run comparison')
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

<template>
  <div class="p-5 space-y-5">
    <h1 class="text-2xl font-bold pt-2">AI Coach</h1>

    <template v-if="!result">
      <select
        v-model="selectedProvider"
        class="w-full bg-[#1e1e24] rounded-xl px-4 py-3 text-base text-white outline-none focus:ring-1 focus:ring-violet-500"
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
            class="w-full bg-white/5 rounded-xl px-4 py-3 text-base text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
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
          <p class="text-sm text-white/60">Injuries / limitations <span class="text-white/30">(optional)</span></p>
          <textarea
            v-model="injuries"
            rows="2"
            placeholder="e.g. bad knees, shoulder impingement..."
            @focus="scrollFieldIntoView($event)"
            class="w-full bg-white/5 rounded-xl px-4 py-3 text-base text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500 resize-none"
          />
        </div>

        <div class="space-y-1">
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
          <p class="text-xs text-white/30">Applies to the generated workout content, not the app interface.</p>
        </div>

        <div class="space-y-2">
          <div class="flex items-center gap-2">
            <p class="text-sm text-white/60">Extra notes <span class="text-white/30">(optional)</span></p>
            <button
              @click="showPromptTip = !showPromptTip"
              class="w-4 h-4 rounded-full bg-white/10 text-white/40 text-xs flex items-center justify-center hover:bg-white/20 hover:text-white/70 transition-colors shrink-0"
              aria-label="What to write here"
            >?</button>
          </div>
          <div v-if="showPromptTip" class="bg-white/5 rounded-xl px-4 py-3 text-xs text-white/50 space-y-1">
            <p class="text-white/70 font-medium mb-1">Add any extra context for a better workout:</p>
            <p>• Bad lower back but can still row, just no deadlifts</p>
            <p>• Only dumbbells and a pull-up bar at home</p>
            <p>• No barbell at my gym, cables and dumbbells only</p>
            <p>• Avoid all pressing movements, prefer pulls today</p>
            <p>• Powerlifting meet next month, heavy compounds only</p>
          </div>
          <textarea
            v-model="prompt"
            rows="3"
            placeholder="e.g. bad lower back but can still row, just no deadlifts, only dumbbells at home..."
            @focus="scrollFieldIntoView($event)"
            class="w-full bg-white/5 rounded-xl px-4 py-3 text-base text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500 resize-none"
          />
        </div>
      </div>

      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

      <button
        @click="generate"
        :disabled="loading"
        class="w-full bg-violet-500 hover:bg-violet-600 disabled:opacity-50 active:scale-95 transition-all text-white font-semibold py-4 rounded-2xl text-sm flex items-center justify-center gap-2"
      >
        <svg v-if="loading" class="animate-spin h-4 w-4 shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 22 6.477 22 12h-4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
        </svg>
        <span>{{ loading ? 'Generating…' : `Generate ${selectedMuscleGroups.join(', ') || 'workout'} · ${effectiveDuration} min` }}</span>
      </button>
    </template>

    <template v-else>
      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-5">
        <div>
          <h2 class="text-lg font-bold">{{ result.response.title }}</h2>
          <p class="text-sm text-white/80 mt-1">{{ result.response.description }}</p>
        </div>

        <div class="flex flex-wrap gap-2 text-xs">
          <span class="bg-violet-500/20 text-violet-400 px-2 py-1 rounded-full">{{ result.usage.provider }}</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ result.usage.input_tokens + result.usage.output_tokens }} tokens</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">${{ result.usage.cost_usd.toFixed(4) }}</span>
          <span class="bg-white/5 text-white/40 px-2 py-1 rounded-full">{{ result.workout.xp_earned }} XP on completion</span>
        </div>

        <section v-if="result.response.warm_up?.length">
          <p class="text-xs font-semibold text-white/60 uppercase tracking-widest mb-3">Warm Up</p>
          <div class="space-y-2">
            <div v-for="ex in result.response.warm_up" :key="ex.name" class="bg-white/5 rounded-xl p-3">
              <button v-if="exerciseGifMap[ex.name]" @click="openPopup(ex.name)"
                class="font-medium text-sm text-violet-400 text-left w-full">
                {{ ex.name }} <span class="text-violet-500/60 text-xs">▶</span>
              </button>
              <p v-else class="font-medium text-sm">{{ ex.name }}</p>
              <p class="text-xs text-white/40 mt-0.5">
                <span v-if="ex.sets && ex.reps">{{ ex.sets }} sets × {{ ex.reps }} reps</span>
                <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
              </p>
              <p v-if="ex.notes" class="text-xs text-white/50 mt-1">{{ ex.notes }}</p>
            </div>
          </div>
        </section>

        <section v-if="result.response.main?.length">
          <p class="text-xs font-semibold text-white/60 uppercase tracking-widest mb-3">Main</p>
          <div class="space-y-2">
            <div v-for="ex in result.response.main" :key="ex.name" class="bg-white/5 rounded-xl p-3">
              <button v-if="exerciseGifMap[ex.name]" @click="openPopup(ex.name)"
                class="font-medium text-sm text-violet-400 text-left w-full">
                {{ ex.name }} <span class="text-violet-500/60 text-xs">▶</span>
              </button>
              <p v-else class="font-medium text-sm">{{ ex.name }}</p>
              <p class="text-xs text-white/40 mt-0.5">
                <span v-if="ex.sets && ex.reps">{{ ex.sets }} sets × {{ ex.reps }} reps</span>
                <span v-else-if="ex.duration_seconds">{{ ex.duration_seconds }}s</span>
                <span v-if="ex.rest_seconds"> · {{ ex.rest_seconds }}s rest</span>
              </p>
              <p v-if="ex.notes" class="text-xs text-white/50 mt-1">{{ ex.notes }}</p>
            </div>
          </div>
        </section>

        <section v-if="result.response.tips?.length">
          <p class="text-xs font-semibold text-white/60 uppercase tracking-widest mb-2">Tips</p>
          <ul class="space-y-1">
            <li v-for="tip in result.response.tips" :key="tip" class="text-sm text-white/60 flex gap-2 items-start">
              <span class="text-violet-400 shrink-0 mt-0.5">•</span>
              <span>{{ tip }}</span>
            </li>
          </ul>
        </section>
      </div>

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
          @click="showModal = true"
          class="flex-1 bg-white/10 hover:bg-white/15 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
        >View full ▶</button>
        <button
          v-if="!xpResult"
          @click="complete"
          :disabled="completing"
          class="flex-1 bg-violet-500 hover:bg-violet-600 disabled:opacity-50 text-white font-semibold py-3 rounded-2xl text-sm transition-colors"
        >{{ completing ? 'Saving…' : `Complete · ${result.workout.xp_earned} XP` }}</button>
      </div>
    </template>
  </div>

  <WorkoutModal
    v-if="showModal && result"
    :workout="result.response"
    :workout-id="result.workout.id"
    :provider="result.usage.provider"
    :duration-minutes="effectiveDuration"
    :muscle-group="selectedMuscleGroups.join(', ')"
    :injuries="injuries"
    :goals="`${goal} · ${experience}`"
    :environment="environment"
    @close="showModal = false"
    @completed="onModalCompleted"
    @deleted="onModalDeleted"
  />

  <!-- Exercise GIF popup — tap exercise name to open, tap backdrop or ✕ to close -->
  <Teleport to="body">
    <div
      v-if="activePopup"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75"
      @click.self="closePopup"
    >
      <div class="bg-[#1e1e24] rounded-2xl p-4 mx-4 w-72 relative">
        <button
          @click="closePopup"
          class="absolute top-3 right-3 text-white/40 hover:text-white text-xl leading-none"
          aria-label="Close"
        >✕</button>
        <p class="font-semibold text-sm pr-7 mb-3">{{ activePopup }}</p>
        <img
          :src="exerciseGifMap[activePopup!]"
          :alt="activePopup ?? ''"
          class="w-full rounded-xl"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { api } from '../api/client'
import { findLibraryMatch, type LibraryExercise } from '../utils/exerciseMatch'
import { toMessage } from '../utils/error'
import { mediaUrl } from '../utils/mediaUrl'
import WorkoutModal from '../components/WorkoutModal.vue'

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
const muscleGroups = ['Chest', 'Back', 'Legs', 'Shoulders', 'Arms', 'Core', 'Calves', 'Wrists']
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
const language             = ref('en')
const showPromptTip        = ref(false)

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
const showModal  = ref(false)

// Exercise library — loaded once on mount, used for GIF popup matching
const libraryExercises = ref<LibraryExercise[]>([])
onMounted(async () => {
  try {
    libraryExercises.value = await api.get<LibraryExercise[]>('/exercises')
  } catch {
    // non-fatal: GIF popups simply won't appear if library can't be fetched
  }
})

// Map of AI exercise name → gif_url, built once when result arrives
const exerciseGifMap = computed<Record<string, string>>(() => {
  if (!result.value || libraryExercises.value.length === 0) return {}

  const allNames = [
    ...result.value.response.warm_up.map(e => e.name),
    ...result.value.response.main.map(e => e.name),
  ]

  const map: Record<string, string> = {}
  for (const name of allNames) {
    if (map[name] !== undefined) continue
    const match = findLibraryMatch(name, selectedMuscleGroups.value, libraryExercises.value, environment.value)
    const url = mediaUrl(match?.gif_url)
    if (url) map[name] = url
  }
  return map
})

// Popup state — one exercise open at a time
const activePopup = ref<string | null>(null)
function openPopup(name: string) { activePopup.value = name }
function closePopup() { activePopup.value = null }

// Lock main scroll when GIF popup is open
watch(activePopup, (val) => {
  const main = document.querySelector('main') as HTMLElement | null
  if (main) main.style.overflowY = val ? 'hidden' : ''
})

// Scroll form field into view when keyboard opens on mobile
function scrollFieldIntoView(e: FocusEvent) {
  setTimeout(() => {
    (e.target as HTMLElement).scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, 300)
}

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
      language:         language.value,
    })
    await nextTick()
    const main = document.querySelector('main') as HTMLElement | null
    main?.scrollTo({ top: 0, behavior: 'smooth' })
  } catch (e: unknown) {
    error.value = toMessage(e, 'Failed to generate workout')
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
    error.value = toMessage(e, 'Failed to complete workout')
  } finally {
    completing.value = false
  }
}

function onModalCompleted(r: XPResult) {
  xpResult.value = r
  showModal.value = false
}

function onModalDeleted() {
  showModal.value = false
  reset()
}

function reset() {
  result.value        = null
  xpResult.value      = null
  error.value         = ''
  showModal.value     = false
  showPromptTip.value = false
}
</script>

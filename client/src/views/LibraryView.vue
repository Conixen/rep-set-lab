<template>
  <div class="p-5 space-y-4">
    <h1 class="text-2xl font-bold pt-2">Exercise library</h1>

    <input
      v-model="search"
      type="text"
      placeholder="Search exercises..."
      class="w-full bg-[#1e1e24] rounded-xl px-4 py-3 text-base text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
    />

    <div class="flex gap-2 overflow-x-auto pb-1">
      <button
        v-for="f in filters"
        :key="f"
        @click="activeFilter = f"
        :class="activeFilter === f ? 'bg-violet-500 text-white shadow-[0_0_14px_rgba(124,92,191,0.45)]' : 'bg-[#1e1e24] text-white/50'"
        class="shrink-0 px-4 py-1.5 rounded-full text-sm font-medium transition-all"
      >
        {{ f }}
      </button>
    </div>

    <div v-if="loading" class="text-white/40 text-sm text-center py-8">Loading…</div>
    <p v-else-if="error" class="text-red-400 text-xs">{{ error }}</p>

    <div v-else class="grid grid-cols-2 md:grid-cols-3 gap-3">
      <div
        v-for="ex in pagedExercises"
        :key="ex.id"
        @click="selectedExercise = ex"
        class="bg-[#1e1e24] rounded-2xl p-4 space-y-2 cursor-pointer active:scale-95 transition-all ui-card exercise-card"
      >
        <img
          v-if="ex.gif_url || ex.thumbnail_url"
          :src="mediaUrl(ex.gif_url || ex.thumbnail_url)!"
          :alt="ex.name"
          class="h-24 w-full object-cover rounded-xl bg-white/5 animate-pulse"
          @load="($event.target as HTMLImageElement).classList.remove('animate-pulse')"
        />
        <div v-else class="h-24 bg-white/5 rounded-xl animate-pulse flex items-center justify-center">
          <div class="w-8 h-8 rounded-full bg-white/10"></div>
        </div>
        <p class="font-semibold text-sm">{{ ex.name }}</p>
        <div class="flex flex-wrap gap-1">
          <span v-for="tag in [ex.muscle_group, ex.difficulty, ex.equipment].filter(Boolean)" :key="tag"
                class="text-xs bg-violet-500/20 text-violet-400 px-2 py-0.5 rounded-full">
            {{ tag }}
          </span>
        </div>
      </div>
    </div>

    <div v-if="totalPages > 1" class="flex items-center justify-between pt-2 pb-4">
      <button
        @click="page--"
        :disabled="page === 1"
        class="px-4 py-2 rounded-xl bg-[#1e1e24] text-sm font-medium transition-colors disabled:opacity-30"
        :class="page > 1 ? 'text-white hover:bg-white/10' : 'text-white/30'"
      >← Prev</button>
      <span class="text-sm text-white/40">{{ page }} / {{ totalPages }}</span>
      <button
        @click="page++"
        :disabled="page === totalPages"
        class="px-4 py-2 rounded-xl bg-[#1e1e24] text-sm font-medium transition-colors disabled:opacity-30"
        :class="page < totalPages ? 'text-white hover:bg-white/10' : 'text-white/30'"
      >Next →</button>
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="selectedExercise"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 px-4"
      @click.self="selectedExercise = null"
    >
      <div class="bg-[#1e1e24] rounded-2xl w-full max-w-sm relative overflow-hidden ui-card">
        <button
          @click="selectedExercise = null"
          class="absolute top-3 right-3 z-10 w-7 h-7 rounded-full bg-black/40 text-white/50 hover:text-white text-lg flex items-center justify-center transition-colors"
          aria-label="Close"
        >✕</button>

        <img
          v-if="selectedExercise.gif_url || selectedExercise.thumbnail_url"
          :src="mediaUrl(selectedExercise.gif_url || selectedExercise.thumbnail_url)!"
          :alt="selectedExercise.name"
          class="w-full rounded-t-2xl bg-white/5"
        />
        <div v-else class="h-48 bg-white/5 rounded-t-2xl flex items-center justify-center text-white/20 text-sm">
          No image
        </div>

        <div class="p-4 space-y-3">
          <p class="font-bold text-base pr-6">{{ selectedExercise.name }}</p>

          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="tag in [selectedExercise.muscle_group, selectedExercise.difficulty, selectedExercise.equipment].filter(Boolean)"
              :key="tag"
              class="text-xs bg-violet-500/20 text-violet-400 px-2.5 py-1 rounded-full"
            >{{ tag }}</span>
          </div>

          <p v-if="selectedExercise.description" class="text-sm text-white/60 leading-relaxed">
            {{ selectedExercise.description }}
          </p>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { api } from '../api/client'
import { toMessage } from '../utils/error'
import { mediaUrl } from '../utils/mediaUrl'

interface Exercise {
  id:            number
  name:          string
  description:   string
  muscle_group:  string
  difficulty:    string
  equipment:     string
  thumbnail_url: string | null
  gif_url:       string | null
}

const PAGE_SIZE = 20

const exercises        = ref<Exercise[]>([])
const search           = ref('')
const activeFilter     = ref('All')
const page             = ref(1)
const loading          = ref(true)
const error            = ref('')
const selectedExercise = ref<Exercise | null>(null)

const filters = ['All', 'back', 'chest', 'shoulders', 'arms', 'legs', 'core', 'lower legs', 'lower arms']

const filteredExercises = computed(() =>
  exercises.value.filter(e => {
    const matchesFilter = activeFilter.value === 'All' || e.muscle_group === activeFilter.value
    const matchesSearch = e.name.toLowerCase().includes(search.value.toLowerCase())
    return matchesFilter && matchesSearch
  })
)

const totalPages = computed(() => Math.max(1, Math.ceil(filteredExercises.value.length / PAGE_SIZE)))

const pagedExercises = computed(() => {
  const start = (page.value - 1) * PAGE_SIZE
  return filteredExercises.value.slice(start, start + PAGE_SIZE)
})

watch([activeFilter, search], () => { page.value = 1 })


onMounted(async () => {
  try {
    exercises.value = await api.get<Exercise[]>('/exercises')
  } catch (e: unknown) {
    error.value = toMessage(e, 'Failed to load exercises')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.exercise-card {
  border: 1px solid transparent;
}
.exercise-card:hover {
  border-color: rgba(124, 92, 191, 0.5);
  box-shadow:
    0 0 20px rgba(124, 92, 191, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.04),
    0 4px 32px rgba(0, 0, 0, 0.5);
}
</style>

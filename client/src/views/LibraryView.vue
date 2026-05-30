<template>
  <div class="p-5 space-y-4">
    <h1 class="text-2xl font-bold pt-2">Exercise library</h1>

    <input
      v-model="search"
      type="text"
      placeholder="Search exercises..."
      class="w-full bg-[#1e1e24] rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
    />

    <div class="flex gap-2 overflow-x-auto pb-1">
      <button
        v-for="f in filters"
        :key="f"
        @click="activeFilter = f"
        :class="activeFilter === f ? 'bg-violet-500 text-white' : 'bg-[#1e1e24] text-white/50'"
        class="shrink-0 px-4 py-1.5 rounded-full text-sm font-medium transition-colors"
      >
        {{ f }}
      </button>
    </div>

    <div v-if="loading" class="text-white/40 text-sm text-center py-8">Loading…</div>
    <p v-else-if="error" class="text-red-400 text-xs">{{ error }}</p>

    <div v-else class="grid grid-cols-2 md:grid-cols-3 gap-3">
      <div
        v-for="ex in filteredExercises"
        :key="ex.id"
        @click="selectedExercise = ex"
        class="bg-[#1e1e24] rounded-2xl p-4 space-y-2 cursor-pointer active:scale-95 transition-transform"
      >
        <img
          v-if="ex.gif_url || ex.thumbnail_url"
          :src="(ex.gif_url || ex.thumbnail_url)!"
          :alt="ex.name"
          class="h-24 w-full object-cover rounded-xl bg-white/5"
        />
        <div v-else class="h-24 bg-white/5 rounded-xl flex items-center justify-center text-white/20 text-xs">
          No image
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
  </div>

  <!-- Exercise detail popup -->
  <Teleport to="body">
    <div
      v-if="selectedExercise"
      class="fixed inset-0 z-50 flex items-end justify-center bg-black/75 pb-6 px-4"
      @click.self="selectedExercise = null"
    >
      <div class="bg-[#1e1e24] rounded-2xl w-full max-w-sm relative overflow-hidden">
        <button
          @click="selectedExercise = null"
          class="absolute top-3 right-3 z-10 w-7 h-7 rounded-full bg-black/40 text-white/50 hover:text-white text-lg flex items-center justify-center transition-colors"
          aria-label="Close"
        >✕</button>

        <img
          v-if="selectedExercise.gif_url || selectedExercise.thumbnail_url"
          :src="(selectedExercise.gif_url || selectedExercise.thumbnail_url)!"
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
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'
import { toMessage } from '../utils/error'

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

const exercises        = ref<Exercise[]>([])
const search           = ref('')
const activeFilter     = ref('All')
const loading          = ref(true)
const error            = ref('')
const selectedExercise = ref<Exercise | null>(null)

const filters = computed(() => {
  const groups = [...new Set(exercises.value.map(e => e.muscle_group).filter(Boolean))]
  return ['All', ...groups.slice(0, 5)]
})

const filteredExercises = computed(() =>
  exercises.value.filter(e => {
    const matchesFilter = activeFilter.value === 'All' || e.muscle_group === activeFilter.value
    const matchesSearch = e.name.toLowerCase().includes(search.value.toLowerCase())
    return matchesFilter && matchesSearch
  })
)

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

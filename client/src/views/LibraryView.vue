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

    <div v-else class="grid grid-cols-2 gap-3">
      <div
        v-for="ex in filteredExercises"
        :key="ex.id"
        class="bg-[#1e1e24] rounded-2xl p-4 space-y-2"
      >
        <div class="h-16 bg-white/5 rounded-xl flex items-center justify-center text-white/20 text-xs">GIF</div>
        <p class="font-semibold text-sm">{{ ex.name }}</p>
        <div class="flex flex-wrap gap-1">
          <span v-for="tag in [ex.muscle_group, ex.category].filter(Boolean)" :key="tag"
                class="text-xs bg-violet-500/20 text-violet-400 px-2 py-0.5 rounded-full">
            {{ tag }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'

interface Exercise {
  id: number
  name: string
  muscle_group: string
  category: string
}

const exercises   = ref<Exercise[]>([])
const search      = ref('')
const activeFilter = ref('All')
const loading     = ref(true)
const error       = ref('')

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
    error.value = e instanceof Error ? e.message : 'Failed to load exercises'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="p-5 space-y-5">
    <h1 class="text-2xl font-bold pt-2">AI Coach</h1>

    <!-- Model selector -->
    <select
      v-model="selectedModel"
      class="w-full bg-[#1e1e24] rounded-xl px-4 py-3 text-sm text-white outline-none focus:ring-1 focus:ring-violet-500"
    >
      <option v-for="m in models" :key="m" :value="m">{{ m }}</option>
    </select>

    <!-- Prompt builder -->
    <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-4">
      <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Build your prompt</p>

      <!-- Experience -->
      <div class="space-y-2">
        <p class="text-sm text-white/60">Experience level</p>
        <div class="flex gap-2">
          <button
            v-for="lvl in levels"
            :key="lvl"
            @click="experience = lvl"
            :class="experience === lvl ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
            class="flex-1 py-2 rounded-xl text-sm font-medium transition-colors"
          >
            {{ lvl }}
          </button>
        </div>
      </div>

      <!-- Goal -->
      <div class="space-y-2">
        <p class="text-sm text-white/60">Goal</p>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="g in goals"
            :key="g"
            @click="goal = g"
            :class="goal === g ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
            class="py-2 rounded-xl text-sm font-medium transition-colors"
          >
            {{ g }}
          </button>
        </div>
      </div>

      <!-- Days per week -->
      <div class="space-y-2">
        <p class="text-sm text-white/60">Days per week</p>
        <div class="flex gap-2">
          <button
            v-for="d in [2, 3, 4, 5]"
            :key="d"
            @click="days = d"
            :class="days === d ? 'bg-violet-500 text-white' : 'bg-white/10 text-white/50'"
            class="flex-1 py-2 rounded-xl text-sm font-medium transition-colors"
          >
            {{ d }}
          </button>
        </div>
      </div>

      <!-- Injuries -->
      <div class="space-y-2">
        <p class="text-sm text-white/60">Injuries / limitations</p>
        <input
          v-model="injuries"
          type="text"
          placeholder="e.g. bad knees, shoulder impingement..."
          class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
        />
      </div>
    </div>

    <!-- Generate button -->
    <button class="w-full bg-violet-500 hover:bg-violet-600 active:scale-95 transition-all text-white font-semibold py-4 rounded-2xl text-sm">
      Generate {{ days }}-day {{ goal.toLowerCase() }} plan · {{ experience }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const models = ['Claude Sonnet', 'GPT-4o', 'Gemini 1.5 Pro']
const levels  = ['Beginner', 'Intermediate', 'Advanced']
const goals   = ['Muscle gain', 'Fat loss', 'Strength', 'Endurance']

const selectedModel = ref('Claude Sonnet')
const experience    = ref('Beginner')
const goal          = ref('Fat loss')
const days          = ref(3)
const injuries      = ref('')
</script>

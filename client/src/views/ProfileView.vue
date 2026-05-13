<template>
  <div class="p-5 space-y-6">
    <h1 class="text-2xl font-bold pt-2">Profile</h1>

    <div class="flex flex-col items-center gap-3 py-4">
      <div class="w-20 h-20 rounded-full bg-violet-500/30 flex items-center justify-center text-3xl font-bold text-violet-400">
        L
      </div>
      <p class="text-xl font-bold">Leon</p>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div v-for="stat in stats" :key="stat.label" class="bg-[#1e1e24] rounded-2xl p-4 text-center">
        <p class="text-2xl font-bold" :class="stat.color ?? 'text-white'">{{ stat.value }}</p>
        <p class="text-xs text-white/40 mt-1">{{ stat.label }}</p>
      </div>
    </div>

    <button
      @click="handleLogout"
      class="w-full border border-white/10 text-white/50 font-medium py-3 rounded-2xl text-sm hover:bg-white/5 transition-colors"
    >
      Log out
    </button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth   = useAuthStore()
const router = useRouter()

const stats = [
  { label: 'Level',         value: '—',      color: 'text-violet-400' },
  { label: 'Total XP',      value: '—' },
  { label: 'Workouts done', value: '—' },
  { label: 'Streak',        value: '—',      color: 'text-amber-400' },
]

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

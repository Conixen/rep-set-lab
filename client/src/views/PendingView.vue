<template>
  <div class="min-h-screen flex flex-col items-center justify-center p-8 text-center space-y-5">
    <div class="text-5xl">⏳</div>
    <h1 class="text-2xl font-bold">Awaiting approval</h1>
    <p class="text-white/60 text-sm max-w-xs leading-relaxed">
      Your account is pending approval. You'll get full access as soon as an admin activates it.
      In the meantime you can browse the exercise library.
    </p>
    <div class="flex gap-3 pt-2">
      <router-link
        to="/library"
        class="px-5 py-2.5 bg-violet-500 hover:bg-violet-600 text-white text-sm font-semibold rounded-xl transition-colors"
      >Browse library</router-link>
      <button
        @click="logout"
        class="px-5 py-2.5 bg-white/10 hover:bg-white/15 text-white/70 text-sm font-medium rounded-xl transition-colors"
      >Sign out</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth   = useAuthStore()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

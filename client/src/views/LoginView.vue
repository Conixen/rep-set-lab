<template>
  <div class="min-h-dvh bg-[#111114] flex flex-col items-center justify-center p-6 gap-6">
    <h1 class="text-3xl font-bold text-white">Rep-Set-Lab</h1>

    <div class="w-full max-w-sm bg-[#1e1e24] rounded-2xl p-6 space-y-4">
      <div class="flex rounded-xl overflow-hidden border border-white/10">
        <button
          v-for="m in ['Login', 'Register']"
          :key="m"
          @click="mode = m as 'Login' | 'Register'"
          :class="mode === m ? 'bg-violet-500 text-white' : 'text-white/40'"
          class="flex-1 py-2 text-sm font-medium transition-colors"
        >
          {{ m }}
        </button>
      </div>

      <input
        v-model="email"
        type="email"
        placeholder="Email"
        class="w-full bg-white/5 rounded-xl px-4 py-3 text-base text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
      />
      <input
        v-if="mode === 'Register'"
        v-model="username"
        type="text"
        placeholder="Username"
        class="w-full bg-white/5 rounded-xl px-4 py-3 text-base text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
      />
      <input
        v-model="password"
        type="password"
        placeholder="Password"
        @keyup.enter="submit"
        class="w-full bg-white/5 rounded-xl px-4 py-3 text-base text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
      />

      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

      <button
        @click="submit"
        :disabled="loading"
        class="w-full bg-violet-500 hover:bg-violet-600 disabled:opacity-50 active:scale-95 transition-all text-white font-semibold py-3 rounded-xl text-sm"
      >
        {{ loading ? 'Loading…' : mode }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { toMessage } from '../utils/error'

const auth   = useAuthStore()
const router = useRouter()

const mode     = ref<'Login' | 'Register'>('Login')
const email    = ref('')
const username = ref('')
const password = ref('')
const error    = ref('')
const loading  = ref(false)

async function submit() {
  error.value   = ''
  loading.value = true
  try {
    if (mode.value === 'Login') {
      await auth.login(email.value, password.value)
    } else {
      await auth.register(email.value, username.value, password.value)
    }
    router.push('/')
  } catch (e: unknown) {
    error.value = toMessage(e, 'Something went wrong')
  } finally {
    loading.value = false
  }
}
</script>

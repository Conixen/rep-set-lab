<template>
  <div class="p-5 space-y-6">
    <h1 class="text-2xl font-bold pt-2">Settings</h1>

    <div class="flex gap-2 bg-white/5 p-1 rounded-xl">
      <button
        @click="tab = 'username'"
        :class="tab === 'username' ? 'bg-violet-500 text-white' : 'text-white/50'"
        class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
      >Username</button>
      <button
        @click="tab = 'password'"
        :class="tab === 'password' ? 'bg-violet-500 text-white' : 'text-white/50'"
        class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
      >Password</button>
    </div>

    <div v-if="tab === 'username'" class="bg-[#1e1e24] rounded-2xl p-4 space-y-3 ui-card">
      <p class="text-sm font-medium">Change username</p>
      <input
        v-model="newUsername"
        type="text"
        placeholder="New username"
        class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
      />
      <p v-if="usernameMsg" :class="usernameMsg.ok ? 'text-green-400' : 'text-red-400'" class="text-xs">{{ usernameMsg.text }}</p>
      <button
        @click="saveUsername"
        :disabled="!newUsername.trim()"
        class="w-full bg-violet-500 hover:bg-violet-600 disabled:opacity-40 text-white font-medium py-2.5 rounded-xl text-sm transition-colors"
      >Save username</button>
    </div>

    <div v-if="tab === 'password'" class="bg-[#1e1e24] rounded-2xl p-4 space-y-3 ui-card">
      <p class="text-sm font-medium">Change password</p>
      <input
        v-model="currentPassword"
        type="password"
        placeholder="Current password"
        class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
      />
      <input
        v-model="newPassword"
        type="password"
        placeholder="New password"
        class="w-full bg-white/5 rounded-xl px-4 py-3 text-sm text-white placeholder-white/30 outline-none focus:ring-1 focus:ring-violet-500"
      />
      <p v-if="passwordMsg" :class="passwordMsg.ok ? 'text-green-400' : 'text-red-400'" class="text-xs">{{ passwordMsg.text }}</p>
      <button
        @click="savePassword"
        :disabled="!currentPassword || !newPassword"
        class="w-full bg-violet-500 hover:bg-violet-600 disabled:opacity-40 text-white font-medium py-2.5 rounded-xl text-sm transition-colors"
      >Save password</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const tab             = ref<'username' | 'password'>('username')
const newUsername     = ref('')
const currentPassword = ref('')
const newPassword     = ref('')
const usernameMsg     = ref<{ ok: boolean; text: string } | null>(null)
const passwordMsg     = ref<{ ok: boolean; text: string } | null>(null)

function saveUsername() {
  // TODO: wire to PUT /users/me
  usernameMsg.value = { ok: false, text: 'Backend not connected yet' }
}

function savePassword() {
  // TODO: wire to PUT /users/me
  passwordMsg.value = { ok: false, text: 'Backend not connected yet' }
}
</script>

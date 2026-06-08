import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'

interface AuthResponse {
  token: string
}

function decodePayload(token: string): Record<string, any> {
  try { return JSON.parse(atob(token.split('.')[1])) } catch { return {} }
}

export const useAuthStore = defineStore('auth', () => {
  const token    = ref<string | null>(localStorage.getItem('token'))
  const payload  = ref<Record<string, any>>(token.value ? decodePayload(token.value) : {})

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin    = computed(() => payload.value.role === 'admin')
  const username   = computed<string>(() => payload.value.username ?? '')

  function setToken(t: string) {
    token.value   = t
    payload.value = decodePayload(t)
    localStorage.setItem('token', t)
  }

  async function login(email: string, password: string) {
    const res = await api.post<AuthResponse>('/auth/login', { email, password })
    setToken(res.token)
  }

  async function register(email: string, username: string, password: string) {
    const res = await api.post<AuthResponse>('/auth/register', { email, username, password })
    setToken(res.token)
  }

  function logout() {
    token.value   = null
    payload.value = {}
    localStorage.removeItem('token')
  }

  return { token, isLoggedIn, isAdmin, username, login, register, logout }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'

interface AuthResponse {
  token: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))

  const isLoggedIn = computed(() => !!token.value)

  async function login(email: string, password: string) {
    const res = await api.post<AuthResponse>('/auth/login', { email, password })
    token.value = res.token
    localStorage.setItem('token', res.token)
  }

  async function register(email: string, password: string) {
    const res = await api.post<AuthResponse>('/auth/register', { email, password })
    token.value = res.token
    localStorage.setItem('token', res.token)
  }

  function logout() {
    token.value = null
    localStorage.removeItem('token')
  }

  return { token, isLoggedIn, login, register, logout }
})

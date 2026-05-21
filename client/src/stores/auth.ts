import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'

interface AuthResponse {
  token: string
}

function decodeRole(token: string): string {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.role ?? 'user'
  } catch {
    return 'user'
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const role  = ref<string>(token.value ? decodeRole(token.value) : 'user')

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin    = computed(() => role.value === 'admin')

  function setToken(t: string) {
    token.value = t
    role.value  = decodeRole(t)
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
    token.value = null
    role.value  = 'user'
    localStorage.removeItem('token')
  }

  return { token, role, isLoggedIn, isAdmin, login, register, logout }
})

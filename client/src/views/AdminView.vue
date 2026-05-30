<template>
  <div class="p-5 space-y-5">
    <h1 class="text-2xl font-bold pt-2">Admin</h1>

    <!-- Tab switcher -->
    <div class="flex gap-2 bg-white/5 p-1 rounded-xl">
      <button
        v-for="t in tabList" :key="t.key"
        @click="activeTab = t.key"
        :class="activeTab === t.key ? 'bg-violet-500 text-white' : 'text-white/50'"
        class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
      >{{ t.label }}</button>
    </div>

    <!-- ── AI Requests ── -->
    <template v-if="activeTab === 'ai'">
      <div v-if="aiLoading" class="text-white/40 text-sm text-center py-8">Loading…</div>
      <template v-else-if="aiData">

        <!-- Per-model cards -->
        <div class="space-y-2">
          <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Per model</p>
          <div class="flex gap-3 overflow-x-auto pb-1">
            <div
              v-for="s in aiData.provider_stats"
              :key="s.provider"
              class="bg-[#1e1e24] rounded-2xl p-4 shrink-0 w-44 space-y-2.5"
            >
              <p class="text-sm font-semibold capitalize truncate">{{ s.provider }}</p>
              <div class="space-y-1.5 text-xs">
                <div class="flex justify-between">
                  <span class="text-white/40">Calls</span>
                  <span>{{ s.total_calls }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-white/40">Valid JSON</span>
                  <span :class="validRate(s) >= 90 ? 'text-green-400' : 'text-yellow-400'">{{ validRate(s) }}%</span>
                </div>
                <div class="border-t border-white/5 pt-1.5 flex justify-between">
                  <span class="text-white/40">Avg in tok</span>
                  <span>{{ s.avg_input_tokens.toLocaleString() }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-white/40">Avg out tok</span>
                  <span>{{ s.avg_output_tokens.toLocaleString() }}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-white/40">Total tok</span>
                  <span>{{ (s.total_input_tokens + s.total_output_tokens).toLocaleString() }}</span>
                </div>
                <div class="border-t border-white/5 pt-1.5 flex justify-between">
                  <span class="text-white/40">Avg latency</span>
                  <span>{{ s.avg_latency_ms }}ms</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-white/40">Avg cost</span>
                  <span>${{ s.avg_cost_usd.toFixed(4) }}</span>
                </div>
                <div class="flex justify-between font-medium">
                  <span class="text-white/40">Total cost</span>
                  <span class="text-violet-400">${{ s.total_cost_usd.toFixed(3) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Request log -->
        <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-3">
          <div class="flex items-center justify-between">
            <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">
              Request log <span class="text-white/20 normal-case font-normal">({{ aiData.total }} total)</span>
            </p>
            <button @click="handleDownloadCSV" class="text-xs text-violet-400 hover:text-violet-300 transition-colors">
              Export CSV ↓
            </button>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-xs text-left">
              <thead>
                <tr class="text-white/30 border-b border-white/10">
                  <th class="pb-2 pr-3">User</th>
                  <th class="pb-2 pr-3">Provider</th>
                  <th class="pb-2 pr-3">In tok</th>
                  <th class="pb-2 pr-3">Out tok</th>
                  <th class="pb-2 pr-3">Cost</th>
                  <th class="pb-2 pr-3">Latency</th>
                  <th class="pb-2 pr-3">JSON</th>
                  <th class="pb-2">Date</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5">
                <tr v-for="r in aiData.requests" :key="r.id" class="text-white/70">
                  <td class="py-2 pr-3">{{ r.username }}</td>
                  <td class="py-2 pr-3 capitalize">{{ r.provider }}</td>
                  <td class="py-2 pr-3">{{ r.input_tokens }}</td>
                  <td class="py-2 pr-3">{{ r.output_tokens }}</td>
                  <td class="py-2 pr-3">${{ r.cost_usd.toFixed(5) }}</td>
                  <td class="py-2 pr-3">{{ r.latency_ms }}ms</td>
                  <td class="py-2 pr-3">
                    <span :class="r.valid_json ? 'text-green-400' : 'text-red-400'">
                      {{ r.valid_json ? 'Yes' : 'No' }}
                    </span>
                  </td>
                  <td class="py-2 text-white/40">{{ formatDateTime(r.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination -->
          <div class="flex items-center justify-between pt-2">
            <button
              @click="aiPage--; loadAI()"
              :disabled="aiPage <= 1"
              class="text-xs text-violet-400 disabled:text-white/20 hover:text-violet-300 transition-colors"
            >← Prev</button>
            <span class="text-xs text-white/30">Page {{ aiData.page }} of {{ aiData.total_pages }}</span>
            <button
              @click="aiPage++; loadAI()"
              :disabled="aiPage >= aiData.total_pages"
              class="text-xs text-violet-400 disabled:text-white/20 hover:text-violet-300 transition-colors"
            >Next →</button>
          </div>
        </div>
      </template>
      <p v-if="aiError" class="text-red-400 text-xs">{{ aiError }}</p>
    </template>

    <!-- ── Users ── -->
    <template v-if="activeTab === 'users'">
      <div v-if="usersLoading" class="text-white/40 text-sm text-center py-8">Loading…</div>
      <div v-else-if="!users.length" class="text-white/40 text-sm text-center py-8">No users found.</div>
      <div v-else class="bg-[#1e1e24] rounded-2xl p-4 space-y-3">
        <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">All users</p>
        <div class="space-y-2">
          <div
            v-for="u in users" :key="u.id"
            class="flex items-center justify-between gap-3 py-2 border-b border-white/5 last:border-0"
          >
            <div class="min-w-0">
              <p class="text-sm font-medium truncate">{{ u.username }}</p>
              <p class="text-xs text-white/30 truncate">{{ u.email }}</p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span
                :class="u.role === 'admin' ? 'bg-violet-500/20 text-violet-400' : 'bg-white/5 text-white/30'"
                class="text-xs px-2 py-0.5 rounded-full"
              >{{ u.role }}</span>
              <span
                v-if="u.status === 'pending'"
                class="text-xs bg-yellow-500/20 text-yellow-400 px-2 py-0.5 rounded-full"
              >pending</span>
              <button
                v-if="u.status === 'pending'"
                @click="approveUser(u)"
                :disabled="!!userActionId"
                class="text-xs px-2 py-1 rounded-lg bg-green-500/20 hover:bg-green-500/30 text-green-400 transition-colors disabled:opacity-40"
              >Approve</button>
              <button
                v-else
                @click="toggleRole(u)"
                :disabled="!!userActionId"
                class="text-xs px-2 py-1 rounded-lg bg-white/5 hover:bg-white/10 text-white/50 hover:text-white transition-colors disabled:opacity-40"
              >{{ u.role === 'admin' ? 'Demote' : 'Promote' }}</button>
            </div>
          </div>
        </div>
      </div>
      <p v-if="usersError" class="text-red-400 text-xs">{{ usersError }}</p>
    </template>

    <!-- ── Exercises ── -->
    <template v-if="activeTab === 'exercises'">
      <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-4">
        <div class="flex items-center justify-between">
          <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Exercise images</p>
          <button
            @click="runSync"
            :disabled="syncLoading"
            class="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-lg bg-violet-500 hover:bg-violet-400 text-white font-medium transition-colors disabled:opacity-50"
          >
            <span v-if="syncLoading">Syncing…</span>
            <span v-else>🔄 Sync now</span>
          </button>
        </div>

        <p class="text-xs text-white/40 leading-relaxed">
          Fetches animated GIF URLs from ExerciseDB for any exercise missing one.
          Safe to run multiple times — skips exercises that already have a GIF.
        </p>

        <!-- Result -->
        <div v-if="syncResult" class="space-y-3 pt-1">
          <div class="grid grid-cols-4 gap-2">
            <div class="bg-white/5 rounded-xl p-3 text-center">
              <p class="text-xl font-bold">{{ syncResult.total }}</p>
              <p class="text-xs text-white/40 mt-0.5">Total</p>
            </div>
            <div class="bg-white/5 rounded-xl p-3 text-center">
              <p class="text-xl font-bold text-white/30">{{ syncResult.skipped }}</p>
              <p class="text-xs text-white/40 mt-0.5">Skipped</p>
            </div>
            <div class="bg-white/5 rounded-xl p-3 text-center">
              <p class="text-xl font-bold text-green-400">{{ syncResult.gifs }}</p>
              <p class="text-xs text-white/40 mt-0.5">Updated</p>
            </div>
            <div class="bg-white/5 rounded-xl p-3 text-center">
              <p class="text-xl font-bold" :class="errorCount > 0 ? 'text-red-400' : 'text-white/30'">
                {{ errorCount }}
              </p>
              <p class="text-xs text-white/40 mt-0.5">Errors</p>
            </div>
          </div>

          <div v-if="syncResult.no_match.length" class="bg-white/5 rounded-xl p-3 space-y-1">
            <p class="text-xs font-semibold text-yellow-400/80">No ExerciseDB match — add to exerciseIDOverrides:</p>
            <p v-for="name in syncResult.no_match" :key="name" class="text-xs text-white/50 font-mono">{{ name }}</p>
          </div>

          <div v-if="syncResult.failed.length" class="bg-white/5 rounded-xl p-3 space-y-1">
            <p class="text-xs font-semibold text-red-400/80">API / DB errors:</p>
            <p v-for="name in syncResult.failed" :key="name" class="text-xs text-white/50 font-mono">{{ name }}</p>
          </div>
        </div>

        <p v-if="syncError" class="text-red-400 text-xs">{{ syncError }}</p>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '../api/client'
import { toMessage } from '../utils/error'
import { formatDateTime } from '../utils/date'
import { downloadCSV } from '../utils/csv'

const tabList: { key: 'ai' | 'users' | 'exercises'; label: string }[] = [
  { key: 'ai',        label: 'AI Requests' },
  { key: 'users',     label: 'Users' },
  { key: 'exercises', label: 'Exercises' },
]
const activeTab = ref<'ai' | 'users' | 'exercises'>('ai')

// ── AI Requests ──
interface AIRequest {
  id: number
  username: string
  provider: string
  input_tokens: number
  output_tokens: number
  cost_usd: number
  latency_ms: number
  valid_json: boolean
  created_at: string
}

interface ProviderStat {
  provider:            string
  total_calls:         number
  valid_calls:         number
  avg_latency_ms:      number
  avg_cost_usd:        number
  total_cost_usd:      number
  avg_input_tokens:    number
  avg_output_tokens:   number
  total_input_tokens:  number
  total_output_tokens: number
}

interface AIData {
  requests: AIRequest[]
  total: number
  page: number
  page_size: number
  total_pages: number
  provider_stats: ProviderStat[]
}

const aiData    = ref<AIData | null>(null)
const aiLoading = ref(false)
const aiError   = ref('')
const aiPage    = ref(1)

function validRate(s: ProviderStat) {
  if (s.total_calls === 0) return 0
  return Math.round((s.valid_calls / s.total_calls) * 100)
}

async function loadAI() {
  aiLoading.value = true
  aiError.value = ''
  try {
    aiData.value = await api.get<AIData>(`/admin/ai-requests?page=${aiPage.value}`)
  } catch (e: unknown) {
    aiError.value = toMessage(e, 'Failed to load AI requests')
  } finally {
    aiLoading.value = false
  }
}

function handleDownloadCSV() {
  if (!aiData.value?.requests.length) return
  const headers = ['id','username','provider','input_tokens','output_tokens','cost_usd','latency_ms','valid_json','created_at']
  const rows = aiData.value.requests.map(r =>
    [r.id, r.username, r.provider, r.input_tokens, r.output_tokens, r.cost_usd, r.latency_ms, r.valid_json, r.created_at]
  )
  downloadCSV(headers, rows, `ai-requests-page-${aiPage.value}.csv`)
}

// ── Users ──
interface User {
  id:     number
  username: string
  email:  string
  role:   string
  status: string
}

const users        = ref<User[]>([])
const usersLoading = ref(false)
const usersError   = ref('')
const userActionId = ref<number | null>(null)

async function loadUsers() {
  usersLoading.value = true
  usersError.value = ''
  try {
    const data = await api.get<{ users: User[] }>('/admin/users')
    users.value = data.users
  } catch (e: unknown) {
    usersError.value = toMessage(e, 'Failed to load users')
  } finally {
    usersLoading.value = false
  }
}

async function approveUser(u: User) {
  userActionId.value = u.id
  try {
    const data = await api.put<{ user: User }>(`/admin/users/${u.id}/approve`, {})
    const idx = users.value.findIndex(x => x.id === u.id)
    if (idx !== -1) users.value[idx] = data.user
  } catch (e: unknown) {
    usersError.value = toMessage(e, 'Failed to approve user')
  } finally {
    userActionId.value = null
  }
}

async function toggleRole(u: User) {
  userActionId.value = u.id
  try {
    const newRole = u.role === 'admin' ? 'user' : 'admin'
    const data = await api.put<{ user: User }>(`/admin/users/${u.id}`, { role: newRole })
    const idx = users.value.findIndex(x => x.id === u.id)
    if (idx !== -1) users.value[idx] = data.user
  } catch (e: unknown) {
    usersError.value = toMessage(e, 'Failed to update user')
  } finally {
    userActionId.value = null
  }
}

onMounted(() => {
  loadAI()
  loadUsers()
})

// ── Exercises ──
interface SyncResult {
  total:    number
  skipped:  number
  gifs:     number
  no_match: string[]
  failed:   string[]
}

const syncLoading = ref(false)
const syncResult  = ref<SyncResult | null>(null)
const syncError   = ref('')
const errorCount  = computed(() =>
  syncResult.value ? syncResult.value.no_match.length + syncResult.value.failed.length : 0
)

async function runSync() {
  syncLoading.value = true
  syncResult.value  = null
  syncError.value   = ''
  try {
    syncResult.value = await api.post<SyncResult>('/admin/exercises/sync', {})
  } catch (e: unknown) {
    syncError.value = toMessage(e, 'Sync failed')
  } finally {
    syncLoading.value = false
  }
}
</script>

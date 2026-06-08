<template>
  <div class="p-5 space-y-5">
    <h1 class="text-2xl font-bold pt-2">Admin</h1>

    <div class="flex gap-2 bg-white/5 p-1 rounded-xl">
      <button
        v-for="t in tabList" :key="t.key"
        @click="activeTab = t.key"
        :class="activeTab === t.key ? 'bg-violet-500 text-white' : 'text-white/50'"
        class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
      >{{ t.label }}</button>
    </div>

    <template v-if="activeTab === 'ai'">
      <div v-if="aiLoading" class="text-white/40 text-sm text-center py-8">Loading…</div>
      <template v-else-if="aiData">

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

    <template v-if="activeTab === 'compare'">

      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Latest session</p>
          <button @click="loadLatestSession" :disabled="latestLoading" class="text-xs text-violet-400 hover:text-violet-300 transition-colors disabled:opacity-40">
            {{ latestLoading ? 'Loading…' : 'Refresh ↻' }}
          </button>
        </div>
        <div v-if="latestLoading" class="text-white/40 text-sm text-center py-4">Loading…</div>
        <div v-else-if="latestSession.length" class="flex gap-3 overflow-x-auto pb-1">
          <div
            v-for="row in latestSession" :key="row.provider"
            class="bg-[#1e1e24] rounded-2xl p-4 shrink-0 w-60 space-y-3"
          >
            <div class="flex items-center justify-between">
              <p class="text-sm font-semibold capitalize truncate">{{ row.provider }}</p>
              <span class="text-xs text-white/30">{{ row.environment }}</span>
            </div>

            <div class="grid grid-cols-3 gap-2">
              <div class="bg-white/5 rounded-xl p-2 text-center">
                <p class="text-[10px] text-white/40 mb-1">Quality</p>
                <span v-if="row.groq_goal_grade" class="text-lg font-bold" :class="gradeTextColor(row.groq_goal_grade)">{{ row.groq_goal_grade }}</span>
                <span v-else class="text-lg font-bold text-white/20">—</span>
              </div>
              <div class="bg-white/5 rounded-xl p-2 text-center">
                <p class="text-[10px] text-white/40 mb-1">Tokens</p>
                <p class="text-xs font-semibold">{{ (row.input_tokens + row.output_tokens).toLocaleString() }}</p>
              </div>
              <div class="bg-white/5 rounded-xl p-2 text-center">
                <p class="text-[10px] text-white/40 mb-1">Cost</p>
                <p class="text-xs font-semibold">${{ row.cost_usd.toFixed(4) }}</p>
              </div>
            </div>

            <div class="space-y-1">
              <p class="text-[10px] text-white/30 font-semibold tracking-widest uppercase">Prompt sensitivity</p>
              <div class="flex gap-2">
                <div class="flex-1 bg-white/5 rounded-lg px-2 py-1.5 flex items-center justify-between">
                  <span class="text-[10px] text-white/40">Injury</span>
                  <span v-if="row.groq_injury_grade" class="text-xs font-bold font-mono" :class="gradeTextColor(row.groq_injury_grade)">{{ row.groq_injury_grade }}</span>
                  <span v-else class="text-xs text-white/20">—</span>
                </div>
                <div class="flex-1 bg-white/5 rounded-lg px-2 py-1.5 flex items-center justify-between">
                  <span class="text-[10px] text-white/40">Goal</span>
                  <span v-if="row.groq_goal_grade" class="text-xs font-bold font-mono" :class="gradeTextColor(row.groq_goal_grade)">{{ row.groq_goal_grade }}</span>
                  <span v-else class="text-xs text-white/20">—</span>
                </div>
              </div>
            </div>

            <div class="space-y-1">
              <p class="text-[10px] text-white/30 font-semibold tracking-widest uppercase">Consistency</p>
              <div class="flex gap-2">
                <div class="flex-1 bg-white/5 rounded-lg px-2 py-1.5 flex items-center justify-between">
                  <span class="text-[10px] text-white/40">Structure</span>
                  <span class="text-xs font-semibold" :class="row.completeness_score === 3 ? 'text-green-400' : row.completeness_score >= 2 ? 'text-yellow-400' : 'text-red-400'">{{ row.completeness_score }}/3</span>
                </div>
                <div class="flex-1 bg-white/5 rounded-lg px-2 py-1.5 flex items-center justify-between">
                  <span class="text-[10px] text-white/40">Latency</span>
                  <span class="text-xs font-semibold">{{ row.latency_ms.toLocaleString() }}ms</span>
                </div>
              </div>
            </div>

            <button
              @click="toggleExpand(row.provider)"
              class="w-full text-[10px] text-white/30 hover:text-white/50 transition-colors text-left"
            >
              {{ expandedProviders.has(row.provider) ? '▾ Hide details' : '▸ Show details' }}
            </button>
            <div v-if="expandedProviders.has(row.provider)" class="space-y-1 text-xs border-t border-white/5 pt-2">
              <div class="flex justify-between">
                <span class="text-white/40">Library match</span>
                <span :class="matchColor(row.library_match_rate)">{{ pct(row.library_match_rate) }} ({{ row.library_match_count }}/{{ row.library_total_count }})</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">Est. duration</span>
                <span>{{ row.estimated_minutes.toFixed(0) }} min</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">Warm-up</span>
                <span>{{ row.warm_up_count }} ex.</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">Main</span>
                <span>{{ row.main_count }} ex.</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">Tips</span>
                <span>{{ row.tips_count }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">Notes rate</span>
                <span>{{ pct(row.notes_present_rate) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">Chars</span>
                <span>{{ row.char_count.toLocaleString() }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">In tokens</span>
                <span>{{ row.input_tokens.toLocaleString() }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-white/40">Out tokens</span>
                <span>{{ row.output_tokens.toLocaleString() }}</span>
              </div>
              <div v-if="row.emoji_count > 0" class="flex justify-between">
                <span class="text-white/40">Emoji</span>
                <span class="text-yellow-400">{{ row.emoji_count }}</span>
              </div>
              <div v-if="row.equipment_violations > 0" class="flex justify-between">
                <span class="text-white/40">Equip. ✗</span>
                <span class="text-red-400">{{ row.equipment_violations }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else-if="!latestLoading" class="text-white/40 text-sm text-center py-4">No sessions yet.</div>
        <p v-if="latestError" class="text-red-400 text-xs">{{ latestError }}</p>
      </div>

      <div v-if="compareLoading" class="text-white/40 text-sm text-center py-8">Loading…</div>
      <template v-else-if="compareAvgs && compareAvgs.length">
        <div class="bg-[#1e1e24] rounded-2xl p-4 space-y-3">
          <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">
            Per-provider averages <span class="normal-case font-normal text-white/20">(across all compare sessions)</span>
          </p>
          <div class="overflow-x-auto">
            <table class="w-full text-xs text-left">
              <thead>
                <tr class="text-white/30 border-b border-white/10">
                  <th class="pb-2 pr-4">Provider</th>
                  <th class="pb-2 pr-4">Sessions</th>
                  <th class="pb-2 pr-4">Quality</th>
                  <th class="pb-2 pr-4">Avg tokens</th>
                  <th class="pb-2 pr-4">Avg cost</th>
                  <th class="pb-2 pr-4">Injury</th>
                  <th class="pb-2 pr-4">Goal</th>
                  <th class="pb-2 pr-4">Structure</th>
                  <th class="pb-2">Latency</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5">
                <tr v-for="a in compareAvgs" :key="a.provider" class="text-white/70">
                  <td class="py-2 pr-4 font-medium capitalize">{{ a.provider }}</td>
                  <td class="py-2 pr-4">{{ a.total_sessions }}</td>
                  <td class="py-2 pr-4 font-bold" :class="scoreColor(a.avg_groq_goal_score)">{{ scoreToGrade(a.avg_groq_goal_score) }}</td>
                  <td class="py-2 pr-4">{{ Math.round(a.avg_input_tokens + a.avg_output_tokens).toLocaleString() }}</td>
                  <td class="py-2 pr-4">${{ a.avg_cost_usd.toFixed(4) }}</td>
                  <td class="py-2 pr-4 font-bold" :class="scoreColor(a.avg_groq_injury_score)">{{ scoreToGrade(a.avg_groq_injury_score) }}</td>
                  <td class="py-2 pr-4 font-bold" :class="scoreColor(a.avg_groq_goal_score)">{{ scoreToGrade(a.avg_groq_goal_score) }}</td>
                  <td class="py-2 pr-4" :class="a.avg_completeness_score >= 3 ? 'text-green-400' : a.avg_completeness_score >= 2 ? 'text-yellow-400' : 'text-red-400'">{{ a.avg_completeness_score.toFixed(1) }}/3</td>
                  <td class="py-2">{{ Math.round(a.avg_latency_ms).toLocaleString() }}ms</td>
                </tr>
              </tbody>
            </table>
          </div>

          <button @click="showDetailedAvgs = !showDetailedAvgs" class="text-[10px] text-white/30 hover:text-white/50 transition-colors">
            {{ showDetailedAvgs ? '▾ Hide detailed breakdown' : '▸ Detailed breakdown' }}
          </button>
          <template v-if="showDetailedAvgs">
            <div class="flex gap-3 overflow-x-auto pb-1 pt-1">
              <div
                v-for="a in compareAvgs" :key="'detail-' + a.provider"
                class="bg-white/5 rounded-xl p-3 shrink-0 w-44 space-y-1.5"
              >
                <p class="text-xs font-semibold capitalize truncate">{{ a.provider }}</p>
                <div class="space-y-1 text-xs">
                  <div class="flex justify-between">
                    <span class="text-white/40">Library match</span>
                    <span :class="matchColor(a.avg_library_match_rate)">{{ pct(a.avg_library_match_rate) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-white/40">Est. min</span>
                    <span>{{ a.avg_estimated_minutes.toFixed(1) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-white/40">Main ex.</span>
                    <span>{{ a.avg_main_count.toFixed(1) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-white/40">Notes %</span>
                    <span>{{ pct(a.avg_notes_present_rate) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-white/40">Chars</span>
                    <span>{{ Math.round(a.avg_char_count).toLocaleString() }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-white/40">Emoji</span>
                    <span>{{ a.avg_emoji_count.toFixed(1) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-white/40">Equip. ✗</span>
                    <span :class="a.avg_equipment_violations > 0 ? 'text-red-400' : 'text-green-400'">{{ a.avg_equipment_violations.toFixed(1) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">AI narrative analysis</p>
            <button
              @click="runAnalysis"
              :disabled="analysisLoading"
              class="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-lg bg-violet-500 hover:bg-violet-400 text-white font-medium transition-colors disabled:opacity-50"
            >
              <span v-if="analysisLoading" class="inline-block w-3 h-3 border border-white/30 border-t-white rounded-full animate-spin"></span>
              {{ analysisLoading ? 'Analyzing…' : 'Analyze all sessions ✦' }}
            </button>
          </div>
          <template v-if="analysis">
            <p class="text-xs text-white/30">Based on {{ analysis.session_count }} sessions</p>
            <div class="flex gap-2 flex-wrap">
              <div
                v-for="v in analysis.verdicts" :key="v.provider"
                class="bg-[#1e1e24] rounded-xl p-3 flex items-start gap-3 min-w-0"
              >
                <span class="text-xl font-bold leading-none mt-0.5" :class="gradeTextColor(v.grade)">{{ v.grade }}</span>
                <div class="min-w-0">
                  <p class="text-sm font-medium capitalize">{{ v.provider }}</p>
                  <p class="text-xs text-white/40 leading-relaxed">{{ v.summary }}</p>
                </div>
              </div>
            </div>
            <div class="bg-[#1e1e24] rounded-2xl p-4">
              <p class="text-sm text-white/70 leading-relaxed whitespace-pre-wrap">{{ analysis.narrative }}</p>
            </div>
          </template>
          <p v-if="analysisError" class="text-red-400 text-xs">{{ analysisError }}</p>
        </div>
      </template>
      <div v-else-if="!compareLoading" class="text-white/40 text-sm text-center py-8">No compare sessions yet. Run a comparison to see analytics.</div>
      <p v-if="compareError" class="text-red-400 text-xs">{{ compareError }}</p>
    </template>

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
            <div class="min-w-0 flex-1 text-left">
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
          Skips exercises that already have a GIF.
        </p>

        <div class="border-t border-white/10 pt-4 space-y-3">
          <div class="flex items-center justify-between">
            <p class="text-xs text-white/40 font-semibold tracking-widest uppercase">Bulk import</p>
            <button
              @click="runBulkImport"
              :disabled="bulkLoading"
              class="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-lg bg-white/10 hover:bg-white/15 text-white font-medium transition-colors disabled:opacity-50"
            >
              <span v-if="bulkLoading">Importing…</span>
              <span v-else>⬇ Import all from ExerciseDB</span>
            </button>
          </div>
          <p class="text-xs text-white/40 leading-relaxed">
            Pulls relevant exercises from ExerciseDB (back, chest, shoulders, arms, legs, core, calves, wrists).
          </p>
          <div v-if="bulkResult" class="grid grid-cols-2 gap-2">
            <div class="bg-white/5 rounded-xl p-3 text-center">
              <p class="text-xl font-bold text-green-400">{{ bulkResult.imported }}</p>
              <p class="text-xs text-white/40 mt-0.5">Imported</p>
            </div>
            <div class="bg-white/5 rounded-xl p-3 text-center">
              <p class="text-xl font-bold" :class="bulkResult.failed > 0 ? 'text-red-400' : 'text-white/30'">{{ bulkResult.failed }}</p>
              <p class="text-xs text-white/40 mt-0.5">Failed</p>
            </div>
          </div>
          <p v-if="bulkError" class="text-red-400 text-xs">{{ bulkError }}</p>
        </div>

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

const tabList: { key: 'ai' | 'compare' | 'users' | 'exercises'; label: string }[] = [
  { key: 'ai',        label: 'AI Requests' },
  { key: 'compare',   label: 'Compare' },
  { key: 'users',     label: 'Users' },
  { key: 'exercises', label: 'Exercises' },
]
const activeTab = ref<'ai' | 'compare' | 'users' | 'exercises'>('ai')

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

interface SessionRow {
  provider: string
  muscle_group: string
  duration_minutes: number
  environment: string
  has_injuries: boolean
  library_match_rate: number
  library_match_count: number
  library_total_count: number
  char_count: number
  emoji_count: number
  equipment_violations: number
  completeness_score: number
  warm_up_count: number
  main_count: number
  tips_count: number
  notes_present_rate: number
  estimated_minutes: number
  input_tokens: number
  output_tokens: number
  cost_usd: number
  latency_ms: number
  groq_injury_grade?: string
  groq_equipment_grade?: string
  groq_goal_grade?: string
  groq_feedback?: string
}

interface ProviderVerdict {
  provider: string
  grade: string
  summary: string
}

interface SessionAnalysis {
  narrative: string
  verdicts: ProviderVerdict[]
  session_count: number
}

interface ProviderCompareAvg {
  provider: string
  total_sessions: number
  avg_library_match_rate: number
  avg_char_count: number
  avg_emoji_count: number
  avg_equipment_violations: number
  avg_completeness_score: number
  avg_warm_up_count: number
  avg_main_count: number
  avg_tips_count: number
  avg_notes_present_rate: number
  avg_estimated_minutes: number
  avg_input_tokens: number
  avg_output_tokens: number
  avg_cost_usd: number
  avg_latency_ms: number
  avg_groq_injury_score: number
  avg_groq_equipment_score: number
  avg_groq_goal_score: number
}

const compareAvgs       = ref<ProviderCompareAvg[]>([])
const compareLoading    = ref(false)
const compareError      = ref('')
const expandedProviders = ref(new Set<string>())
const showDetailedAvgs  = ref(false)

function toggleExpand(provider: string) {
  if (expandedProviders.value.has(provider)) {
    expandedProviders.value.delete(provider)
  } else {
    expandedProviders.value.add(provider)
  }
  expandedProviders.value = new Set(expandedProviders.value)
}

function scoreToGrade(score: number): string {
  if (score === 0) return '—'
  if (score >= 4.5) return 'A'
  if (score >= 3.5) return 'B'
  if (score >= 2.5) return 'C'
  if (score >= 1.5) return 'D'
  return 'F'
}

async function loadCompareStats() {
  compareLoading.value = true
  compareError.value = ''
  try {
    const data = await api.get<{ provider_averages: ProviderCompareAvg[] }>('/admin/ai-compare-stats')
    compareAvgs.value = data.provider_averages ?? []
  } catch (e: unknown) {
    compareError.value = toMessage(e, 'Failed to load compare stats')
  } finally {
    compareLoading.value = false
  }
}

const latestSession = ref<SessionRow[]>([])
const latestLoading = ref(false)
const latestError   = ref('')

async function loadLatestSession() {
  latestLoading.value = true
  latestError.value = ''
  try {
    const data = await api.get<{ rows: SessionRow[] }>('/admin/compare/latest')
    latestSession.value = data.rows ?? []
  } catch (e: unknown) {
    latestError.value = toMessage(e, 'Failed to load latest session')
  } finally {
    latestLoading.value = false
  }
}

const analysis        = ref<SessionAnalysis | null>(null)
const analysisLoading = ref(false)
const analysisError   = ref('')

async function runAnalysis() {
  analysisLoading.value = true
  analysisError.value   = ''
  analysis.value        = null
  try {
    analysis.value = await api.post<SessionAnalysis>('/admin/compare/narrative', {})
  } catch (e: unknown) {
    analysisError.value = toMessage(e, 'Analysis failed')
  } finally {
    analysisLoading.value = false
  }
}


function gradeTextColor(grade: string) {
  switch (grade) {
    case 'A': return 'text-green-400'
    case 'B': return 'text-green-300'
    case 'C': return 'text-yellow-400'
    case 'D': return 'text-orange-400'
    case 'F': return 'text-red-400'
    default:  return 'text-white/40'
  }
}

function pct(rate: number) { return Math.round(rate * 100) + '%' }

function matchColor(rate: number) {
  if (rate >= 0.7) return 'text-green-400'
  if (rate >= 0.4) return 'text-yellow-400'
  return 'text-red-400'
}

function scoreColor(score: number) {
  if (score >= 4) return 'text-green-400'   // avg A/B
  if (score >= 3) return 'text-yellow-400'  // avg C
  return 'text-red-400'                     // avg D/F
}

onMounted(() => {
  loadAI()
  loadUsers()
  loadCompareStats()
  loadLatestSession()
})

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

interface BulkImportResult { imported: number; failed: number }
const bulkLoading = ref(false)
const bulkResult  = ref<BulkImportResult | null>(null)
const bulkError   = ref('')

async function runBulkImport() {
  bulkLoading.value = true
  bulkResult.value  = null
  bulkError.value   = ''
  try {
    bulkResult.value = await api.post<BulkImportResult>('/admin/exercises/bulk-import', {})
  } catch (e: unknown) {
    bulkError.value = toMessage(e, 'Bulk import failed')
  } finally {
    bulkLoading.value = false
  }
}
</script>

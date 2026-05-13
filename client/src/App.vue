<template>
  <div class="flex flex-col h-dvh bg-[#111114] text-white max-w-sm mx-auto relative">
    <main class="flex-1 overflow-y-auto pb-20">
      <RouterView />
    </main>

    <nav class="fixed bottom-0 left-1/2 -translate-x-1/2 w-full max-w-sm bg-[#1a1a1f] border-t border-white/10 flex justify-around items-center h-16 z-50">
      <RouterLink
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="flex flex-col items-center gap-1 text-xs text-white/40 transition-colors"
        active-class="!text-violet-500"
      >
        <component :is="tab.icon" class="w-5 h-5" />
        {{ tab.label }}
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import IconHome from './components/icons/IconHome.vue'
import IconLibrary from './components/icons/IconLibrary.vue'
import IconAi from './components/icons/IconAi.vue'
import IconProfile from './components/icons/IconProfile.vue'

const tabs = [
  { to: '/',        label: 'Home',    icon: IconHome },
  { to: '/library', label: 'Library', icon: IconLibrary },
  { to: '/ai',      label: 'AI',      icon: IconAi },
  { to: '/profile', label: 'Profile', icon: IconProfile },
]
</script>

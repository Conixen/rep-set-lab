import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import LoginView   from '../views/LoginView.vue'
import HomeView    from '../views/HomeView.vue'
import LibraryView from '../views/LibraryView.vue'
import AiView      from '../views/AiView.vue'
import CompareView from '../views/CompareView.vue'
import HistoryView from '../views/HistoryView.vue'
import ProfileView  from '../views/ProfileView.vue'
import SettingsView from '../views/SettingsView.vue'
import AdminView    from '../views/AdminView.vue'
import PendingView  from '../views/PendingView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login',    component: LoginView,   meta: { public: true } },
    { path: '/pending',  component: PendingView, meta: { public: true } },
    { path: '/',         component: HomeView },
    { path: '/library',  component: LibraryView },
    { path: '/ai',       component: AiView },
    { path: '/compare',  component: CompareView },
    { path: '/history',  component: HistoryView },
    { path: '/profile',  component: ProfileView },
    { path: '/settings', component: SettingsView },
    { path: '/admin',    component: AdminView,   meta: { admin: true } },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) return '/login'
  if (to.meta.admin && !auth.isAdmin) return '/'
})

export default router

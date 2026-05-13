import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import LoginView   from '../views/LoginView.vue'
import HomeView    from '../views/HomeView.vue'
import LibraryView from '../views/LibraryView.vue'
import AiView      from '../views/AiView.vue'
import ProfileView from '../views/ProfileView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login',    component: LoginView,   meta: { public: true } },
    { path: '/',         component: HomeView },
    { path: '/library',  component: LibraryView },
    { path: '/ai',       component: AiView },
    { path: '/profile',  component: ProfileView },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isLoggedIn) return '/login'
})

export default router

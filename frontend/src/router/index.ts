import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth.store'
import { authApi } from '../api/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/movies' },
    { path: '/login', name: 'login', component: () => import('../views/LoginView.vue') },
    { path: '/oauth/callback', name: 'oauth-callback', component: () => import('../views/OAuthCallbackView.vue') },
    { path: '/movies', name: 'movies', component: () => import('../views/MovieListView.vue'), meta: { requiresAuth: true } },
    {
      path: '/movies/:movieId/showtimes',
      name: 'showtimes',
      component: () => import('../views/ShowtimeListView.vue'),
      meta: { requiresAuth: true },
      props: true,
    },
    {
      path: '/showtimes/:showtimeId/seats',
      name: 'seat-map',
      component: () => import('../views/SeatMapView.vue'),
      meta: { requiresAuth: true },
      props: true,
    },
    {
      path: '/checkout/:bookingId',
      name: 'checkout',
      component: () => import('../views/CheckoutView.vue'),
      meta: { requiresAuth: true },
      props: true,
    },
    {
      path: '/bookings/:bookingId/confirmation',
      name: 'confirmation',
      component: () => import('../views/ConfirmationView.vue'),
      meta: { requiresAuth: true },
      props: true,
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../views/AdminDashboardView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.token) {
    return { name: 'login' }
  }

  if (to.meta.requiresAdmin) {
    if (!auth.user) {
      try {
        auth.setUser(await authApi.me())
      } catch {
        return { name: 'login' }
      }
    }
    if (auth.user?.role !== 'ADMIN') {
      return { name: 'movies' }
    }
  }
})

export default router

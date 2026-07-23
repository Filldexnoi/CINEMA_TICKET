import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types'

const TOKEN_KEY = 'cinema_ticket_token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<User | null>(null)

  function setSession(newToken: string, newUser?: User) {
    token.value = newToken
    localStorage.setItem(TOKEN_KEY, newToken)
    if (newUser) user.value = newUser
  }

  function setUser(newUser: User) {
    user.value = newUser
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  return { token, user, setSession, setUser, logout }
})

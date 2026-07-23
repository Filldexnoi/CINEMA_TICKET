import { apiClient } from './client'
import type { User } from '../types'

export const authApi = {
  me: () => apiClient.get<User>('/me').then((r) => r.data),
  googleLoginUrl: () => '/auth/google/login',
}

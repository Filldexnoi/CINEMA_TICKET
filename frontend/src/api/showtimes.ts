import { apiClient } from './client'
import type { Showtime, ShowtimeSeat } from '../types'

export const showtimesApi = {
  get: (id: string) => apiClient.get<Showtime>(`/showtimes/${id}`).then((r) => r.data),
  seats: (id: string) => apiClient.get<ShowtimeSeat[]>(`/showtimes/${id}/seats`).then((r) => r.data),
}

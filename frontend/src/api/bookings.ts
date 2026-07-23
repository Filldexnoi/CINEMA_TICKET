import { apiClient } from './client'
import type { Booking } from '../types'

export const bookingsApi = {
  create: (showtimeId: string, seatLabels: string[], lockToken: string) =>
    apiClient
      .post<Booking>('/bookings', { showtime_id: showtimeId, seat_labels: seatLabels, lock_token: lockToken })
      .then((r) => r.data),
  get: (id: string) => apiClient.get<Booking>(`/bookings/${id}`).then((r) => r.data),
  pay: (id: string, result: 'success' | 'fail' = 'success') =>
    apiClient.post<Booking>(`/bookings/${id}/pay`, { result }).then((r) => r.data),
}

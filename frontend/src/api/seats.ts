import { apiClient } from './client'

export const seatsApi = {
  lock: (showtimeId: string, seatLabels: string[]) =>
    apiClient
      .post<{ lock_token: string }>(`/showtimes/${showtimeId}/seats/lock`, { seat_labels: seatLabels })
      .then((r) => r.data),
  unlock: (showtimeId: string, seatLabels: string[], lockToken: string) =>
    apiClient.post(`/showtimes/${showtimeId}/seats/unlock`, {
      seat_labels: seatLabels,
      lock_token: lockToken,
    }),
}

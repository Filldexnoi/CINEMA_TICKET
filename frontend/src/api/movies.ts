import { apiClient } from './client'
import type { Movie, Showtime } from '../types'

export const moviesApi = {
  list: () => apiClient.get<Movie[]>('/movies').then((r) => r.data),
  get: (id: string) => apiClient.get<Movie>(`/movies/${id}`).then((r) => r.data),
  showtimes: (movieId: string) =>
    apiClient.get<Showtime[]>(`/movies/${movieId}/showtimes`).then((r) => r.data),
}

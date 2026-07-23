import { apiClient } from './client'
import type { AuditLog, Booking } from '../types'

export interface AdminBookingFilter {
  movie_id?: string
  date?: string
  user_email?: string
}

export const adminApi = {
  listBookings: (filter: AdminBookingFilter) =>
    apiClient.get<Booking[]>('/admin/bookings', { params: filter }).then((r) => r.data),
  listAuditLogs: () => apiClient.get<AuditLog[]>('/admin/audit-logs').then((r) => r.data),
}

export type Role = 'USER' | 'ADMIN'

export interface User {
  id: string
  email: string
  name: string
  picture_url: string
  role: Role
  created_at: string
}

export interface Movie {
  id: string
  title: string
  description: string
  poster_url: string
  duration_minutes: number
  genre: string
  rating: string
}

export interface Showtime {
  id: string
  movie_id: string
  cinema_id: string
  hall_name: string
  start_time: string
  end_time: string
  rows: number
  cols: number
  base_price: number
}

export type SeatStatus = 'AVAILABLE' | 'LOCKED' | 'BOOKED'

export interface ShowtimeSeat {
  id: string
  showtime_id: string
  seat_label: string
  row: number
  col: number
  status: SeatStatus
  locked_by?: string
  lock_expires_at?: string
  booking_id?: string
  price: number
  updated_at: string
}

export type BookingStatus = 'PENDING' | 'CONFIRMED' | 'EXPIRED' | 'FAILED'

export interface Booking {
  id: string
  user_id: string
  showtime_id: string
  seat_labels: string[]
  total_amount: number
  status: BookingStatus
  created_at: string
  expires_at: string
  paid_at?: string
  movie_id: string
  showtime_date: string
  user_email: string
}

export interface SeatEventMessage {
  type: 'seat.locked' | 'seat.released' | 'booking.confirmed' | 'booking.expired'
  showtime_id: string
  seat_labels: string[]
}

export type AuditEventType = 'BOOKING_SUCCESS' | 'BOOKING_TIMEOUT' | 'SEAT_RELEASED' | 'SYSTEM_ERROR'

export interface AuditLog {
  id: string
  event_type: AuditEventType
  message: string
  booking_id?: string
  user_id?: string
  showtime_id?: string
  seat_labels?: string[]
  occurred_at: string
}

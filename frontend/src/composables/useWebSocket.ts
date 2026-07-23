import { useAuthStore } from '../stores/auth.store'
import type { SeatEventMessage } from '../types'

const RECONNECT_DELAY_MS = 2000

export function connectSeatMapSocket(
  showtimeId: string,
  onMessage: (msg: SeatEventMessage) => void,
  onReconnect: () => Promise<void>,
) {
  let socket: WebSocket | null = null
  let stopped = false

  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'

  async function connect() {
    if (stopped) return
    await onReconnect()

    const token = useAuthStore().token ?? ''
    socket = new WebSocket(`${wsProtocol}//${window.location.host}/ws/showtimes/${showtimeId}?token=${token}`)

    socket.onmessage = (event) => {
      try {
        onMessage(JSON.parse(event.data) as SeatEventMessage)
      } catch {}
    }

    socket.onclose = () => {
      if (!stopped) setTimeout(connect, RECONNECT_DELAY_MS)
    }

    socket.onerror = () => socket?.close()
  }

  connect()

  return () => {
    stopped = true
    socket?.close()
  }
}

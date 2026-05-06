interface WsMessage {
  metadata: {
    id: string
    roomId?: string
  }
  event: string
  data: any[]
}

interface WSserviceProps {
  url: string
  options?: {
    reconnectInterval?: number
    reconnectAttempts?: number
    reconnect?: boolean
    autoConnect?: boolean
    listeners?: {
      [eventName: string]: (...data: any[]) => void
    }
  }
}

class WSservice {
  private reconnectInterval: number = 2000
  private reconnectAttempts: number = Infinity
  private reconnect: boolean = true

  private url: string
  private ws: WebSocket | null = null
  private listeners: Map<string, ((...data: any[]) => void)[]> = new Map()

  id: string | null = null
  connected: boolean = false

  onOpen: ((event: Event) => void) | null = null

  constructor({ url, options }: WSserviceProps) {
    this.url = url
    this.ws = null

    if (options?.listeners) {
      for (const [eventName, handler] of Object.entries(options.listeners)) {
        this.on(eventName, handler)
      }
    }

    if (options?.autoConnect !== false) {
      this.connect()
    }

    if (options?.reconnectInterval !== undefined) {
      this.reconnectInterval = options.reconnectInterval
    }

    if (options?.reconnectAttempts !== undefined) {
      this.reconnectAttempts = options.reconnectAttempts
    }

    if (options?.reconnect !== undefined) {
      this.reconnect = options.reconnect
    }
  }

  connect() {
    this.ws = new WebSocket(this.url)

    this.ws.onopen = (msg) => {
      this.connected = true
      if (this.onOpen) {
        this.onOpen(msg)
      }
    }

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data) as WsMessage
      const handlers = this.listeners.get(data.event)

      if (data.data && data.event === "connect" && data.metadata.id) {
        this.id = data.metadata.id
        this.connected = true
      }

      if (handlers)
        for (const handler of handlers) {
          handler(...data.data)
        }
    }

    this.ws.onclose = () => {
      if (this.reconnect && this.reconnectAttempts > 0) {
        this.reconnectAttempts--
        setTimeout(() => this.connect(), this.reconnectInterval)
      } else {
        this.connected = false
        const handlers = this.listeners.get("disconnect")

        if (handlers) {
          for (const handler of handlers) {
            handler()
          }
        }

        // dont close the connection if we are set to reconnect, just mark it as disconnected and wait for the reconnect logic to kick in
        if (this.reconnect) {
          return
        }

        this.ws = null
        this.connected = false
        this.listeners.clear()
      }
    }
  }

  emit(eventName: string, ...data: unknown[]) {
    if (!this.connected || this.ws?.readyState !== WebSocket.OPEN) {
      return
    }

    try {
      this.ws.send(JSON.stringify({ event: eventName, data }))
    } catch (error) {
      console.error("Failed to send WebSocket message:", error)
    }
  }

  on(eventName: string, callback: (...data: unknown[]) => void) {
    const handlers = this.listeners.get(eventName) || []
    handlers.push(callback)
    this.listeners.set(eventName, handlers)
  }

  off(eventName: string) {
    this.listeners.delete(eventName)
  }

  close() {
    this.reconnect = false
    this.ws?.close()
    this.listeners.clear()
  }
}

export default WSservice

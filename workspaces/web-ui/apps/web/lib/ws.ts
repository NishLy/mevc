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
    reconnectOnClose?: boolean
    autoConnect?: boolean
  }
}

class WSservice {
  private reconnectInterval: number = 2000
  private reconnectAttempts: number = Infinity
  private reconnectOnClose: boolean = true

  private url: string
  private ws: WebSocket | null = null
  private listeners: Map<
    string,
    ((eventName: string, ...data: any[]) => void)[]
  > = new Map()

  id: string | null = null
  connected: boolean = false

  constructor({ url, options }: WSserviceProps) {
    this.url = url
    this.ws = null

    if (options?.autoConnect !== false) {
      this.connect()
    }

    if (options?.reconnectInterval !== undefined) {
      this.reconnectInterval = options.reconnectInterval
    }

    if (options?.reconnectAttempts !== undefined) {
      this.reconnectAttempts = options.reconnectAttempts
    }

    if (options?.reconnectOnClose !== undefined) {
      this.reconnectOnClose = options.reconnectOnClose
    }
  }

  connect() {
    this.ws = new WebSocket(this.url)
    this.ws.onopen = (msg) => {}
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data) as WsMessage
      const handlers = this.listeners.get(data.event)

      if (data.data && data.event === "connect" && data.metadata.id) {
        this.id = data.metadata.id
        this.connected = true
      }

      if (handlers)
        for (const handler of handlers) {
          handler(data.event, ...data.data)
        }
    }

    this.ws.onclose = () => {
      if (this.reconnectOnClose && this.reconnectAttempts > 0) {
        this.reconnectAttempts--
        setTimeout(() => this.connect(), this.reconnectInterval)
      } else {
        this.ws = null
        this.connected = false
      }
    }
  }

  emit(eventName: string, ...data: any[]) {
    this.ws?.send(JSON.stringify({ event: eventName, data }))
  }

  on(type: string, callback: (eventName: string, ...data: any[]) => void) {
    const handlers = this.listeners.get(type) || []
    handlers.push(callback)
    this.listeners.set(type, handlers)
  }

  close() {
    this.reconnectOnClose = false
    this.ws?.close()
  }
}

export default WSservice

class WsConnection {
  id: string = ""

  constructor(id: string) {
    this.id = id
  }
}

class WS {
  private url: string
  private ws: WebSocket | null = null
  private listeners: Map<string, (eventName: string, data: any) => void> =
    new Map()

  id = ""

  constructor(url: string) {
    this.url = url
    this.ws = null
    this.listeners = new Map()
  }

  connect() {
    this.ws = new WebSocket(this.url)

    this.ws.onopen = () => {}

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data)

      const handler = this.listeners.get(data.type)
      if (handler) handler(data.type, data)
    }

    this.ws.onclose = () => {
      console.log("WS closed, reconnecting...")
      setTimeout(() => this.connect(), 2000)
    }
  }

  emit(eventName: string, ...data: any) {
    this.ws?.send(JSON.stringify({ type: eventName, data }))
  }

  on(type: string, callback: (eventName: string, data: any) => void) {
    this.listeners.set(type, callback)
  }
}

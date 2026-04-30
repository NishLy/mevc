import io, { Socket } from "socket.io-client"

export class SocketIoService {
  socket: typeof Socket | null = null
  currentRoomId: string | null = null
  isConnected: boolean = false

  onReconnect?: (id: string) => void

  constructor(onReconnect?: (id: string) => void) {
    this.onReconnect = onReconnect
  }

  async initSocket(): Promise<void> {
    // Guard: don't reinitialize if already connected
    if (this.socket?.connected) return

    return new Promise<void>((resolve, reject) => {
      this.socket = io("http://localhost:8001", {
        transports: ["websocket"],
        reconnectionAttempts: 5, // limit retries
        reconnectionDelay: 2000, // wait 2s between retries
        timeout: 10000,
      })

      this.socket.once("connect", () => {
        // use once, not on
        this.isConnected = true
        resolve()
      })

      this.socket.on("reconnect", () => {
        this.isConnected = true
        this.onReconnect?.(this.socket!.id)
      })

      this.socket.on("disconnect", (reason: string) => {
        this.isConnected = false
        console.log("Socket disconnected:", reason)
      })

      this.socket.on("connect_error", (err: Error) => {
        console.error("Connection error:", err.message)
        reject(err)
      })
    })
  }

  connectToRoom(roomId: string) {
    if (!this.socket?.connected) {
      console.error("Socket not connected")
      return
    }
    this.socket.emit("join_room", roomId)
    this.currentRoomId = roomId
  }

  leaveRoom() {
    if (!this.socket || !this.currentRoomId) return
    this.socket.emit("leave_room", this.currentRoomId)
    this.currentRoomId = null
  }

  disconnect() {
    this.socket?.disconnect()
    this.socket = null
    this.isConnected = false
  }
}

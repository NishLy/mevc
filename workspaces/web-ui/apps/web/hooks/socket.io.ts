import { create } from "zustand"
import io, { Socket } from "socket.io-client"

export interface socketIoState {
  socket: SocketIOClient.Socket | null
  initSocket: () => Promise<string>
}

const useSocketIo = create<socketIoState>((set, get) => ({
  socket: null,
  initSocket: () => {
    const { socket } = get()

    if (socket) {
      console.warn("Socket.IO is already initialized.")
      return Promise.resolve(socket.id!)
    }

    return new Promise((resolve, reject) => {
      console.log("Initializing Socket.IO connection...")

      const socketInstance = io("http://localhost:8001", {
        transports: ["websocket"],
        autoConnect: true,
      })

      socketInstance.on("connect", () => {
        console.log("Connected with ID:", socketInstance)
        set({ socket: socketInstance })
        resolve(socketInstance.id!)
      })

      socketInstance.on("connect_error", (error: Error) => {
        console.error("Connection error:", error)
        reject(error)
      })

      socketInstance.on("disconnect", (reason: string) => {
        console.log("Socket disconnected:", reason)
      })
    })
  },
}))

useSocketIo
  .getState()
  .initSocket()
  .catch((error) => {
    console.error("Failed to initialize Socket.IO on startup:", error)
  })

export default useSocketIo

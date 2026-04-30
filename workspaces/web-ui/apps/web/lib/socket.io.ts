import io from "socket.io-client"

const ioInstance = io("http://localhost:8001", {
  transports: ["websocket"],
})

ioInstance.on("connect", () => {
  console.log("Connected to server with ID:", ioInstance.id)
})

ioInstance.on("connect_error", (err: Error) => {
  console.error("Connection error:", err.message, err)
})

export default ioInstance

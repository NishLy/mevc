import { motion } from "framer-motion"
import { useState } from "react"

const PersistentPiP = ({ children }: { children: React.ReactNode }) => {
  const [position, setPosition] = useState({ x: 0, y: 0 })

  return (
    <motion.div
      drag
      animate={{ x: position.x, y: position.y }}
      onDragEnd={(_, info) => {
        const newX = position.x + info.offset.x
        const newY = position.y + info.offset.y

        setPosition({ x: newX, y: newY })
      }}
      style={{
        width: 300,
        height: 180,
        backgroundColor: "#222",
        position: "fixed",
        bottom: 50,
        right: 50,
        cursor: "grab",
        borderRadius: "12px",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: "white",
      }}
    >
      {children}
    </motion.div>
  )
}

export default PersistentPiP

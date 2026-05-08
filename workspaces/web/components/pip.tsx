"use client"

import { motion, useDragControls } from "framer-motion"
import { useRef } from "react"

const PersistentPiP = ({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) => {
  // Use a ref to constrain the drag to the window boundaries
  const constraintsRef = useRef(null)

  return (
    /* This invisible div acts as the "boundary" for the PiP */
    <div
      ref={constraintsRef}
      className="pointer-events-none fixed inset-0 z-50"
    >
      <motion.div
        drag
        // Keeps the video inside the screen
        dragConstraints={constraintsRef}
        dragElastic={0.1}
        dragMomentum={false}
        whileDrag={{ scale: 1.05, cursor: "grabbing" }}
        style={{
          width: 300,
          backgroundColor: "#27272a", // matches zinc-800
          position: "fixed",
          bottom: 50,
          right: 50,
          cursor: "grab",
          borderRadius: "12px",
          boxShadow:
            "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)",
          pointerEvents: "auto", // Re-enable clicks for the PiP itself
        }}
        className={className}
      >
        {children}
      </motion.div>
    </div>
  )
}

export default PersistentPiP

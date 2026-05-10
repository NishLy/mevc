"use client"

import React, { useEffect, useState, useCallback } from "react"
import { motion, AnimatePresence } from "framer-motion"
import useMeet from "../state/meet"
import { ReactionData } from "../types/service"

interface ActiveReaction {
  id: string
  emoji: string
  x: number // % from left edge
  size: number // rem
  delay: number // stagger delay in seconds
  drift: number // horizontal drift px while floating
  duration: number // total animation duration in seconds
}

const EMOJIS_PER_BURST = 6
const LIFETIME_MS = 10000 // matches max duration + max delay
const TRAVEL_PX = 500 // how far up each emoji floats — partial screen, not full

function spawnBurst(emoji: string): ActiveReaction[] {
  // Pick a random cluster zone so reactions feel localised, not wall-to-wall
  const clusterCenter = 10 + Math.random() * 60
  return Array.from({ length: EMOJIS_PER_BURST }, (_, i) => ({
    id: `${crypto.randomUUID()}-${i}`,
    emoji,
    x: clusterCenter + (Math.random() - 0.5) * 18, // ±9% around cluster
    size: 3.5 + Math.random() * 2.5, // 3.5–6 rem
    delay: i * 0.1 + Math.random() * 0.12, // gentle cascade
    drift: (Math.random() - 0.5) * 70, // ±35 px horizontal sway
    duration: 4.8 + Math.random() * 1.6, // 4.8–6.4 s — slow & floaty
  }))
}

export default function ReactionComponent() {
  const [activeReactions, setActiveReactions] = useState<ActiveReaction[]>([])
  const rtcService = useMeet((state) => state.RTCService)

  const handleReact = useCallback((emoji: string) => {
    const burst = spawnBurst(emoji)
    setActiveReactions((prev) => [...prev, ...burst])

    setTimeout(() => {
      const ids = new Set(burst.map((r) => r.id))
      setActiveReactions((prev) => prev.filter((r) => !ids.has(r.id)))
    }, LIFETIME_MS)
  }, [])

  useEffect(() => {
    if (!rtcService) return
    const reactionListener = (reactionData: ReactionData) => {
      if (reactionData.type === "unicode") handleReact(reactionData.value)
    }
    rtcService.onReactionReceived(reactionListener)
    return () => {
      rtcService.onReactionReceived(() => {})
    }
  }, [rtcService, handleReact])

  return (
    <div className="pointer-events-none fixed inset-0 overflow-hidden">
      <AnimatePresence>
        {activeReactions.map((reaction) => (
          <motion.div
            key={reaction.id}
            style={{
              position: "absolute",
              left: `${reaction.x}%`,
              bottom: "4rem",
              fontSize: `${reaction.size}rem`,
              lineHeight: 1,
              willChange: "transform, opacity",
            }}
            initial={{ opacity: 0, scale: 0.2, y: 0, x: 0, rotate: -12 }}
            animate={{
              // Snap in → hold → long gentle dissolve in the upper portion
              opacity: [0, 1, 1, 0.85, 0],
              scale: [0.2, 1.2, 1, 0.96, 0.92],
              y: [0, -TRAVEL_PX],
              x: [0, reaction.drift],
              rotate: [-12, 4, -3, 0],
            }}
            transition={{
              delay: reaction.delay,
              duration: reaction.duration,
              // fast pop-in (0–8%), solid hold (8–50%), long fade (50–100%)
              opacity: { times: [0, 0.08, 0.5, 0.72, 1], ease: "easeInOut" },
              scale: { times: [0, 0.08, 0.22, 0.55, 1], ease: "easeOut" },
              // Ease-out quart: quick push, then decelerates smoothly
              y: { ease: [0.165, 0.84, 0.44, 1.0] },
              x: { ease: "easeInOut" },
              rotate: { times: [0, 0.08, 0.18, 1], ease: "easeOut" },
            }}
            className="drop-shadow-2xl select-none"
          >
            {reaction.emoji}
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  )
}

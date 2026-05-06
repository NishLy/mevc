import { motion } from "framer-motion"
import { Badge } from "lucide-react"
import { useState, useEffect } from "react"

const ReconnectingVariant = () => {
  const [attempt, setAttempt] = useState(1)
  const [progress, setProgress] = useState(0)

  useEffect(() => {
    const t = setInterval(() => {
      setProgress((p) => {
        if (p >= 100) {
          setAttempt((a) => (a < 3 ? a + 1 : a))
          return 0
        }
        return p + 2
      })
    }, 50)
    return () => clearInterval(t)
  }, [])

  return (
    <div className="flex flex-col items-center justify-center gap-8 py-10">
      {/* Pulsing signal icon */}
      <div className="relative flex h-24 w-24 items-center justify-center">
        {[0, 1, 2].map((i) => (
          <motion.div
            key={i}
            className="absolute rounded-full border border-amber-500/40"
            style={{ inset: -(i * 14) }}
            animate={{ opacity: [0.6, 0, 0.6], scale: [0.95, 1.05, 0.95] }}
            transition={{ duration: 1.8, repeat: Infinity, delay: i * 0.35 }}
          />
        ))}
        <div className="flex h-16 w-16 items-center justify-center rounded-full border border-amber-700/50 bg-amber-900/30">
          <svg
            width="28"
            height="28"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#fbbf24"
            strokeWidth="1.8"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M1 6s4-4 11-4 11 4 11 4" />
            <path d="M5 10s2.5-2.5 7-2.5 7 2.5 7 2.5" />
            <circle cx="12" cy="14" r="1.5" fill="#fbbf24" />
            <path d="M12 16v4" />
          </svg>
        </div>
      </div>

      <div className="space-y-1.5 text-center">
        <motion.h2
          className="text-lg font-semibold tracking-tight text-white"
          key={attempt}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
        >
          Reconnecting
          <motion.span
            animate={{ opacity: [1, 0, 1] }}
            transition={{ duration: 1.2, repeat: Infinity }}
          >
            ...
          </motion.span>
        </motion.h2>
        <p className="text-sm text-zinc-500">Your connection was interrupted</p>
      </div>

      {/* Retry progress */}
      <div className="w-full max-w-[260px] space-y-2">
        <div className="flex items-center justify-between text-xs text-zinc-500">
          <span>Attempt {attempt} of 3</span>
          <span>{progress}%</span>
        </div>
        <div className="h-1 w-full overflow-hidden rounded-full bg-zinc-800">
          <motion.div
            className="h-full rounded-full bg-amber-500"
            style={{ width: `${progress}%` }}
            transition={{ duration: 0.05 }}
          />
        </div>
      </div>

      {/* Diagnostics */}
      <div className="w-full max-w-[260px] space-y-2.5 rounded-xl border border-zinc-800 bg-zinc-900/50 p-3.5">
        {[
          { label: "Signal strength", status: "Weak", color: "amber" },
          { label: "Server ping", status: "248 ms", color: "amber" },
          { label: "Audio stream", status: "Interrupted", color: "red" },
          { label: "Video stream", status: "Reconnecting", color: "indigo" },
        ].map(({ label, status, color }) => (
          <div key={label} className="flex items-center justify-between">
            <span className="text-xs text-zinc-500">{label}</span>
            <Badge color={color}>{status}</Badge>
          </div>
        ))}
      </div>

      <button className="text-xs text-zinc-500 underline underline-offset-2 transition-colors hover:text-zinc-300">
        Rejoin with a new link
      </button>
    </div>
  )
}

export default ReconnectingVariant

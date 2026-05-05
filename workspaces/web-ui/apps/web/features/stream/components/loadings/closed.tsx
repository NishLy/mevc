import { motion } from "framer-motion"
import { Badge } from "lucide-react"

const MeetClosedVariant = () => {
  const participants = ["AK", "BL", "CR", "DM", "EV"]

  return (
    <div className="flex flex-col items-center justify-center gap-8 py-10">
      {/* Broken ring */}
      <div className="relative flex items-center justify-center">
        <svg width="88" height="88" viewBox="0 0 88 88" fill="none">
          <circle
            cx="44"
            cy="44"
            r="40"
            stroke="rgba(255,255,255,0.06)"
            strokeWidth="2.5"
          />
          {/* broken arc */}
          <path
            d="M 44 4 A 40 40 0 1 1 12 72"
            fill="none"
            stroke="#ef4444"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeDasharray="4 6"
            opacity="0.6"
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full border border-red-700/40 bg-red-900/25">
            <svg
              width="26"
              height="26"
              viewBox="0 0 24 24"
              fill="none"
              stroke="#f87171"
              strokeWidth="1.8"
              strokeLinecap="round"
            >
              <rect x="3" y="11" width="18" height="11" rx="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
              <line x1="12" y1="15" x2="12" y2="17" />
            </svg>
          </div>
        </div>
      </div>

      <div className="space-y-1.5 text-center">
        <motion.h2
          className="text-lg font-semibold tracking-tight text-white"
          initial={{ opacity: 0, y: 4 }}
          animate={{ opacity: 1, y: 0 }}
        >
          This meeting has ended
        </motion.h2>
        <motion.p
          className="max-w-[200px] text-sm text-zinc-500"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.2 }}
        >
          The host closed this room for all participants
        </motion.p>
      </div>

      {/* Duration & summary */}
      <motion.div
        className="w-full max-w-[260px] divide-y divide-zinc-800 rounded-xl border border-zinc-800 bg-zinc-900/50"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3 }}
      >
        <div className="flex items-center justify-between p-3.5">
          <span className="text-xs text-zinc-500">Duration</span>
          <span className="text-xs font-medium text-white">47 min 12 sec</span>
        </div>
        <div className="flex items-center justify-between p-3.5">
          <span className="text-xs text-zinc-500">Participants</span>
          <div className="flex items-center gap-1">
            {participants.map((p, i) => (
              <motion.div
                key={p}
                initial={{ opacity: 0, x: 4 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.4 + i * 0.06 }}
                className="-ml-1 flex h-6 w-6 items-center justify-center rounded-full border border-zinc-900 bg-zinc-700 text-[9px] font-medium text-zinc-300 first:ml-0"
              >
                {p}
              </motion.div>
            ))}
            <span className="ml-1.5 text-xs text-zinc-500">+2</span>
          </div>
        </div>
        <div className="flex items-center justify-between p-3.5">
          <span className="text-xs text-zinc-500">Recording</span>
          <Badge color="green">
            <span className="h-1.5 w-1.5 rounded-full bg-green-400" />
            Saved
          </Badge>
        </div>
      </motion.div>

      {/* CTA */}
      <motion.div
        className="flex flex-col items-center gap-2.5"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.7 }}
      >
        <button className="rounded-lg bg-indigo-600 px-5 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-500">
          Schedule a follow-up
        </button>
        <button className="text-xs text-zinc-500 transition-colors hover:text-zinc-300">
          Return to dashboard
        </button>
      </motion.div>
    </div>
  )
}

export default MeetClosedVariant

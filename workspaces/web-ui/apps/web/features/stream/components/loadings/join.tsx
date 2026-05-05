import { Badge } from "@workspace/ui/components/badge"
import classNames from "classnames"
import { motion } from "framer-motion"

const Avatar = ({
  initials,
  size = "lg",
  pulse = false,
}: {
  initials: string
  size?: "lg" | "sm"
  pulse?: boolean
}) => {
  const sz = size === "lg" ? "w-16 h-16 text-xl" : "w-10 h-10 text-sm"
  return (
    <div
      className={classNames(
        "relative flex items-center justify-center rounded-full bg-zinc-800 font-semibold tracking-wide text-zinc-300 select-none",
        sz
      )}
    >
      {initials}
      {pulse && (
        <span className="absolute inset-0 animate-ping rounded-full bg-indigo-500/30" />
      )}
    </div>
  )
}

const DotRow = ({ color = "#818cf8" }) => (
  <div className="flex items-center gap-2">
    {[0, 1, 2].map((i) => (
      <motion.span
        key={i}
        className="block h-2 w-2 rounded-full"
        style={{ background: color }}
        animate={{ opacity: [0.3, 1, 0.3], scale: [0.8, 1.2, 0.8] }}
        transition={{ duration: 1.2, repeat: Infinity, delay: i * 0.2 }}
      />
    ))}
  </div>
)
/* ─── Animated ring ─── */
const SpinRing = ({
  size = 56,
  color = "#6366f1",
  thickness = 3,
  speed = 1,
}) => (
  <div className="relative" style={{ width: size, height: size }}>
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
      <circle
        cx={size / 2}
        cy={size / 2}
        r={size / 2 - thickness}
        fill="none"
        stroke="rgba(255,255,255,0.08)"
        strokeWidth={thickness}
      />
    </svg>
    <motion.svg
      width={size}
      height={size}
      viewBox={`0 0 ${size} ${size}`}
      className="absolute inset-0"
      animate={{ rotate: 360 }}
      transition={{ duration: speed, repeat: Infinity, ease: "linear" }}
    >
      <circle
        cx={size / 2}
        cy={size / 2}
        r={size / 2 - thickness}
        fill="none"
        stroke={color}
        strokeWidth={thickness}
        strokeLinecap="round"
        strokeDasharray={`${Math.PI * (size - thickness * 2) * 0.7} ${Math.PI * (size - thickness * 2) * 0.3}`}
      />
    </motion.svg>
  </div>
)

const JoiningVariant = () => {
  const steps = [
    { id: 1, label: "Checking permissions", done: true },
    { id: 2, label: "Joining meeting room", done: true },
    { id: 3, label: "Connecting to session", done: false, active: true },
    { id: 4, label: "Loading participants", done: false },
  ]

  return (
    <div className="flex flex-col items-center justify-center gap-8 py-10">
      {/* Avatar cluster */}
      <div className="relative flex items-center justify-center">
        <motion.div
          animate={{ scale: [1, 1.04, 1] }}
          transition={{ duration: 2.5, repeat: Infinity }}
          className="relative"
        >
          <SpinRing size={88} color="#6366f1" thickness={2.5} speed={2} />
          <div className="absolute inset-0 flex items-center justify-center">
            <Avatar initials="YO" pulse />
          </div>
        </motion.div>

        {/* Floating mini-avatars */}
        {[
          { initials: "AK", angle: -50, delay: 0 },
          { initials: "BL", angle: 30, delay: 0.3 },
          { initials: "CR", angle: 110, delay: 0.6 },
        ].map(({ initials, angle, delay }, i) => {
          const rad = (angle * Math.PI) / 180
          const r = 62
          const x = Math.cos(rad) * r
          const y = Math.sin(rad) * r
          return (
            <motion.div
              key={i}
              className="absolute"
              style={{ transform: `translate(${x}px, ${y}px)` }}
              initial={{ opacity: 0, scale: 0.5 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{
                delay: delay + 0.5,
                type: "spring",
                stiffness: 200,
              }}
            >
              <Avatar initials={initials} size="sm" />
            </motion.div>
          )
        })}
      </div>

      {/* Text */}
      <div className="space-y-1.5 text-center">
        <motion.h2
          className="text-lg font-semibold tracking-tight text-white"
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          Joining <span className="text-indigo-400">Design Sync</span>
        </motion.h2>
        <motion.p
          className="text-sm text-zinc-500"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
        >
          Setting up your meeting experience
        </motion.p>
      </div>

      {/* Steps */}
      <motion.div
        className="w-full max-w-[240px] space-y-2.5"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
      >
        {steps.map((s, i) => (
          <div key={s.id} className="flex items-center gap-3">
            <div className="flex h-5 w-5 flex-shrink-0 items-center justify-center">
              {s.done ? (
                <motion.div
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{
                    type: "spring",
                    stiffness: 300,
                    delay: i * 0.1,
                  }}
                  className="flex h-4 w-4 items-center justify-center rounded-full bg-indigo-500"
                >
                  <svg width="8" height="8" viewBox="0 0 8 8" fill="none">
                    <path
                      d="M1.5 4L3.2 5.7L6.5 2.5"
                      stroke="white"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </motion.div>
              ) : s.active ? (
                <DotRow color="#818cf8" />
              ) : (
                <div className="h-4 w-4 rounded-full border border-zinc-700" />
              )}
            </div>
            <span
              className={classNames(
                "text-xs",
                s.done
                  ? "text-zinc-400 line-through"
                  : s.active
                    ? "font-medium text-indigo-300"
                    : "text-zinc-600"
              )}
            >
              {s.label}
            </span>
          </div>
        ))}
      </motion.div>

      <Badge color="indigo">
        <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-indigo-400" />
        3 participants waiting
      </Badge>
    </div>
  )
}

export default JoiningVariant

import { Participant } from "@/features/stream/types/ui"

const sizeMap = {
  sm: { padding: "w-4 h-4", fontSize: "text-xs" },
  md: { padding: "w-8 h-8", fontSize: "text-sm" },
  lg: { padding: "w-16 h-16", fontSize: "text-base" },
  xl: { padding: "w-32 h-32", fontSize: "text-lg" },
}

export const ParticpantIcon = ({
  participant,
  size = "lg",
}: {
  participant: Participant
  size?: keyof typeof sizeMap
}) => {
  const { padding, fontSize } = sizeMap[size]
  return (
    <div
      className={`flex shrink-0 items-center justify-center rounded-full ${padding} ${fontSize} font-medium text-white`}
      style={{ backgroundColor: participant.color || "#888" }}
    >
      {participant.initials}
    </div>
  )
}

"use client"

import classNames from "classnames"

interface VideoGridProps {
  streams: {
    id: string
    stream?: MediaStream
    isLocal?: boolean
  }[]
}

const calculateGridColumns = (count: number) => {
  if (count === 1) return "grid-cols-1"
  if (count === 2) return "grid-cols-2"
  if (count <= 4) return "grid-cols-2"
  if (count <= 6) return "grid-cols-3"
  if (count <= 9) return "grid-cols-3"
  return "grid-cols-4"
}

function VideoTile({ id, isLocal }: { id: string; isLocal?: boolean }) {
  return (
    <div
      className={classNames(
        "relative aspect-video shrink-0 overflow-hidden rounded-lg bg-zinc-800",
        "flex items-center justify-center",
        isLocal && "ring-1 ring-blue-400/50"
      )}
    >
      <video
        id={id}
        autoPlay
        playsInline
        muted={isLocal}
        controls={false}
        poster="https://img.favpng.com/10/24/2/computer-icons-user-icon-design-male-png-favpng-grqs7j1MENUsCah7VD6XBWVst.jpg"
        className="h-full w-full object-cover"
      />
      {isLocal && (
        <span className="absolute bottom-2 left-2 rounded bg-blue-500/20 px-1.5 py-0.5 text-[10px] text-blue-300">
          You
        </span>
      )}
    </div>
  )
}

export default function VideosGrid({ streams }: VideoGridProps) {
  return (
    <div
      className={classNames(
        "grid w-full auto-rows-max justify-center gap-2 overflow-y-auto bg-zinc-900 p-2",
        calculateGridColumns(streams.length)
      )}
    >
      {streams.map((s) => (
        <VideoTile key={s.id} id={s.id} isLocal={s.isLocal} />
      ))}
    </div>
  )
}

"use client"

import classNames from "classnames"
import useCurrentRoom from "../state"
import VideoTile from "./video_tile"
import { StreamVideoState } from "../types/stream"

const calculateGridColumns = (count: number) => {
  if (count === 1) return "grid-cols-1"
  if (count === 2) return "grid-cols-2"
  if (count <= 4) return "grid-cols-2"
  if (count <= 6) return "grid-cols-3"
  if (count <= 9) return "grid-cols-3"
  return "grid-cols-4"
}

export default function VideosGrid({
  streams,
}: {
  streams: StreamVideoState[]
}) {
  const pinnedStreamIds = useCurrentRoom((state) => state.pinnedStreamIds)
  const unpinnedStreams = streams.filter((s) => !pinnedStreamIds.includes(s.id))
  const pinnedStreams = streams.filter((s) => pinnedStreamIds.includes(s.id))

  return (
    <>
      <div
        className={classNames(
          "w-full auto-rows-max content-center gap-2 p-2",
          calculateGridColumns(unpinnedStreams.length),
          pinnedStreams.length > 0 &&
            "fixed inset-0 z-10 flex h-44 overflow-y-auto",
          pinnedStreams.length === 0 && "grid"
        )}
      >
        {unpinnedStreams.map((s) => (
          <VideoTile key={s.id} {...s} />
        ))}
      </div>

      {pinnedStreams.length > 0 && (
        <div
          className={classNames(
            "grid h-screen w-full content-center justify-items-center gap-2 overflow-y-auto bg-zinc-900 p-2",
            calculateGridColumns(pinnedStreams.length),
            "z-10"
          )}
        >
          {pinnedStreams.map((s) => (
            <VideoTile key={s.id} {...s} />
          ))}
        </div>
      )}
    </>
  )
}

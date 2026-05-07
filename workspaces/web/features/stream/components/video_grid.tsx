"use client"

import classNames from "classnames"
import { useMemo } from "react"
import VideoTile from "./video_tile"
import useMeet from "../state/meet"

const calculateGridColumns = (count: number) => {
  if (count === 1)
    return "grid-cols-1 md:grid-cols-1 lg:grid-cols-1 xl:max-w-9/12"
  if (count === 2)
    return "grid-cols-2 md:grid-cols-2 lg:grid-cols-2 xl:max-w-10/12"
  if (count <= 4)
    return "grid-cols-2 md:grid-cols-2 lg:grid-cols-2 xl:max-w-10/12"
  if (count <= 6)
    return "grid-cols-2 md:grid-cols-3 lg:grid-cols-3 xl:max-w-12/12"
  if (count <= 9)
    return "grid-cols-2 md:grid-cols-3 lg:grid-cols-3 xl:max-w-10/12"
  if (count <= 16)
    return "grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-4 xl:max-w-12/12"
  if (count <= 25) return "grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
  return "grid-cols-3 md:grid-cols-4 lg:grid-cols-6"
}

export default function VideosGrid() {
  const localStreams = useMeet((state) => state.localStreams)
  const remoteStreams = useMeet((state) => state.remoteStreams)

  const streams = useMemo(
    () => [...localStreams.filter((s) => s !== null), ...remoteStreams],
    [localStreams, remoteStreams]
  )

  const pinnedStreamIds = useMeet((state) => state.pinnedStreamIds)

  const pinnedStreams = useMemo(
    () => streams.filter((s) => pinnedStreamIds.includes(s.id)),
    [streams, pinnedStreamIds]
  )
  const unpinnedStreams = useMemo(
    () => streams.filter((s) => !pinnedStreamIds.includes(s.id)),
    [streams, pinnedStreamIds]
  )

  return (
    <div className="relative flex h-full w-full flex-col justify-center">
      <div
        className={classNames(
          "content-centerbg-transparent box-border w-full gap-2 p-2",
          pinnedStreams.length == 0 &&
            calculateGridColumns(unpinnedStreams.length),
          pinnedStreams.length > 0
            ? "flex h-[15vh] w-screen justify-items-start gap-2 overflow-x-auto overflow-y-hidden bg-transparent p-4"
            : "mx-auto grid w-full flex-wrap justify-center justify-items-center overflow-hidden rounded-l"
        )}
      >
        {unpinnedStreams.map((s) => (
          <div
            key={s.id}
            className={classNames(
              pinnedStreams.length > 0
                ? "h-40 w-70 shrink-0 opacity-70 hover:opacity-100"
                : "relative h-fit w-full"
            )}
          >
            <VideoTile {...s} />
          </div>
        ))}
      </div>

      {pinnedStreams.length > 0 && (
        <div
          className={classNames(
            "mx-auto grid h-[90vh] w-full flex-wrap items-center justify-items-center gap-4 overflow-hidden rounded-lg",
            calculateGridColumns(pinnedStreams.length),
            "z-10"
          )}
        >
          {pinnedStreams.map((s) => (
            <div key={s.id} className={classNames("relative h-fit w-full")}>
              <VideoTile key={s.id} {...s} />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

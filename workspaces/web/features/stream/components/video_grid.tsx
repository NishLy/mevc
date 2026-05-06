"use client"

import classNames from "classnames"
import { useMemo } from "react"
import VideoTile from "./video_tile"
import useMeet from "../state/meet"

const calculateGridColumns = (count: number) => {
  if (count === 1) return "grid-cols-1"
  if (count === 2) return "grid-cols-2"
  if (count <= 4) return "grid-cols-2"
  if (count <= 6) return "grid-cols-3"
  if (count <= 9) return "grid-cols-3"
  return "grid-cols-4"
}

export default function VideosGrid() {
  const localStreams = useMeet((state) => state.localStreams)
  const remoteStreams = useMeet((state) => state.remoteStreams)

  const streams = useMemo(
    () => [...localStreams, ...remoteStreams],
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
    <>
      <div
        className={classNames(
          "mx-auto content-center gap-4 p-2",
          calculateGridColumns(unpinnedStreams.length),
          pinnedStreams.length > 0
            ? "fixed top-0 left-0 z-50 flex h-56 w-fit justify-items-start overflow-x-auto overflow-y-hidden bg-transparent p-4"
            : "box-border grid h-screen w-full flex-wrap justify-center justify-items-center overflow-hidden rounded-lg xl:max-w-11/12"
        )}
      >
        {unpinnedStreams.map((s) => (
          <div
            key={s.id}
            className={classNames(
              "relative h-full w-full",
              pinnedStreams.length > 0 &&
                "max-h-44 max-w-xs opacity-70 hover:opacity-100"
            )}
          >
            <VideoTile {...s} />
          </div>
        ))}
      </div>

      {pinnedStreams.length > 0 && (
        <div
          className={classNames(
            "grid h-screen w-full content-center justify-items-center gap-2 overflow-y-auto p-2",
            calculateGridColumns(pinnedStreams.length),
            "z-10"
          )}
        >
          {pinnedStreams.map((s) => (
            <div
              key={s.id}
              className={classNames("relative h-full w-full max-w-10/12")}
            >
              <VideoTile key={s.id} {...s} />
            </div>
          ))}
        </div>
      )}
    </>
  )
}

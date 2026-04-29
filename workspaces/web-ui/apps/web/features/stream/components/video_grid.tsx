"use client"

import classNames from "classnames"
import { useEffect, useMemo, useRef } from "react"
import { StreamVideoEntityType, StreamVideoState } from "../types/stream"

const calculateGridColumns = (count: number) => {
  if (count === 1) return "grid-cols-1"
  if (count === 2) return "grid-cols-2"
  if (count <= 4) return "grid-cols-2"
  if (count <= 6) return "grid-cols-3"
  if (count <= 9) return "grid-cols-3"
  return "grid-cols-4"
}

function VideoTile(props: StreamVideoState) {
  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    // Update the video element's srcObject whenever the stream changes
    if (videoRef.current && props.stream) {
      videoRef.current.srcObject = props.stream
    }
  }, [props.stream])

  return (
    <div
      className={classNames(
        "relative aspect-video shrink-0 overflow-hidden rounded-lg bg-zinc-800",
        "flex items-center justify-center",
        props.isLocal && "ring-1 ring-blue-400/50"
      )}
    >
      <video
        ref={videoRef}
        id={props.id}
        autoPlay
        playsInline
        muted={props.isLocal}
        controls={false}
        poster="https://img.favpng.com/10/24/2/computer-icons-user-icon-design-male-png-favpng-grqs7j1MENUsCah7VD6XBWVst.jpg"
        className="h-full w-full object-cover"
      />
      {props.isLocal && (
        <span className="absolute bottom-2 left-2 rounded bg-blue-500/20 px-1.5 py-0.5 text-[10px] text-blue-300">
          You
        </span>
      )}
    </div>
  )
}

export default function VideosGrid({
  streams,
}: {
  streams: StreamVideoState[]
}) {
  const screenShareStream = useMemo(
    () => streams.find((s) => s.type === StreamVideoEntityType.SCREEN_SHARE),
    [streams]
  )

  console.log("Rendering VideoGrid with streams:", streams)

  const handleNormalVideoStreams = () => {
    return (
      <div
        className={classNames(
          "grid w-full auto-rows-max justify-center gap-2 overflow-y-auto bg-zinc-900 p-2",
          calculateGridColumns(streams.length)
        )}
      >
        {streams.map((s) => (
          <VideoTile key={s.id} {...s} />
        ))}
      </div>
    )
  }

  const hasScreenShareVideoStream = () => {
    return (
      <>
        <div
          className={classNames(
            "fixed inset-0 z-20 flex h-48 gap-2 overflow-x-auto px-1 py-2"
          )}
        >
          {streams
            .filter((s) => s.type !== StreamVideoEntityType.SCREEN_SHARE)
            .map((s) => (
              <VideoTile key={s.id} {...s} />
            ))}
        </div>
        <div className="w-full">
          <VideoTile {...screenShareStream!} />
        </div>
      </>
    )
  }

  return (
    <>
      {screenShareStream
        ? hasScreenShareVideoStream()
        : handleNormalVideoStreams()}
    </>
  )
}

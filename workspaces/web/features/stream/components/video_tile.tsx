/* eslint-disable @typescript-eslint/no-explicit-any */
"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import classNames from "classnames"
import {
  Pin,
  PinOff,
  Maximize2,
  MicOff,
  VideoOff,
  UserX,
  Volume2,
  VolumeX,
  MoreHorizontal,
  Star,
  MessageSquare,
  Shield,
  MicOffIcon,
} from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { MediaCombinedStream } from "../types/service"
import useMeet from "../state/meet"
import { ParticpantIcon } from "@/components/participant"
import { AnimatePresence, motion } from "framer-motion"
import { Skeleton } from "@/components/ui/skeleton"
import { Loader2 } from "lucide-react"

interface MenuItemDef {
  icon: React.ReactNode
  label: string
  variant?: "danger"
}

const REMOTE_MENU: MenuItemDef[] = [
  { icon: <Pin className="h-3.5 w-3.5" />, label: "Pin video" },
  { icon: <Maximize2 className="h-3.5 w-3.5" />, label: "Full screen" },
  { icon: <Star className="h-3.5 w-3.5" />, label: "Spotlight" },
  { icon: <Volume2 className="h-3.5 w-3.5" />, label: "Adjust volume" },
  {
    icon: <MessageSquare className="h-3.5 w-3.5" />,
    label: "Chat with participant",
  },
]

const REMOTE_MENU_DANGER: MenuItemDef[] = [
  {
    icon: <MicOff className="h-3.5 w-3.5" />,
    label: "Mute participant",
    variant: "danger",
  },
  {
    icon: <VideoOff className="h-3.5 w-3.5" />,
    label: "Stop participant's video",
    variant: "danger",
  },
  {
    icon: <Shield className="h-3.5 w-3.5" />,
    label: "Report",
    variant: "danger",
  },
  {
    icon: <UserX className="h-3.5 w-3.5" />,
    label: "Remove participant",
    variant: "danger",
  },
]

const LOCAL_MENU: MenuItemDef[] = [
  { icon: <Pin className="h-3.5 w-3.5" />, label: "Pin my video" },
  { icon: <Maximize2 className="h-3.5 w-3.5" />, label: "Full screen" },
  { icon: <VolumeX className="h-3.5 w-3.5" />, label: "Mute original audio" },
]

const VideoStream = ({
  stream,
  isMuted,
  id,
}: {
  id: string | undefined
  stream: MediaStream | undefined
  isMuted: boolean | undefined
}) => {
  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    if (videoRef.current) {
      videoRef.current.srcObject = stream || null
    }
  }, [stream])

  useEffect(() => {
    if (videoRef.current) {
      videoRef.current.muted = isMuted ?? false
    }
  }, [isMuted])

  useEffect(() => {
    if (videoRef.current && id) {
      videoRef.current.id = id || ""
    }
  }, [id])

  return (
    <video
      ref={videoRef}
      autoPlay
      playsInline
      controls={false}
      poster="/images/spinner.gif"
      className="h-full w-full object-contain"
    />
  )
}

function VideoTile({ props }: { props: MediaCombinedStream | null }) {
  const [hovered, setHovered] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)

  const participants = useMeet((state) => state.participants)

  const metadata = useMemo(() => {
    if (!props) return null

    const participant = participants.find(
      (p) =>
        p.clientId ===
        (props.metadata.video?.clientId || props.metadata.audio?.clientId)
    )

    return participant
  }, [participants, props])

  const menuItems = props?.isLocal ? LOCAL_MENU : REMOTE_MENU
  const dangerItems = props?.isLocal ? [] : REMOTE_MENU_DANGER
  const showOverlay = hovered || menuOpen

  const [isFullscreen, setIsFullscreen] = useState(false)

  const pinnedStreamIds = useMeet((state) => state.pinnedStreamIds)
  const isPinned = pinnedStreamIds.includes(props?.id ?? "")

  const togglePin = () => {
    if (!props) return
    if (isPinned) {
      useMeet.setState({
        pinnedStreamIds: pinnedStreamIds.filter((id) => id !== props?.id),
      })
    } else {
      useMeet.setState({
        pinnedStreamIds: [...pinnedStreamIds, props?.id],
      })
    }
  }

  const toggleFullscreen = async (element: HTMLElement) => {
    if (!document.fullscreenElement) {
      try {
        // Request fullscreen
        if (element.requestFullscreen) {
          await element.requestFullscreen()
        } else if ((element as any).webkitRequestFullscreen) {
          /* Safari */
          await (element as any).webkitRequestFullscreen()
        } else if ((element as any).msRequestFullscreen) {
          /* IE11 */
          await (element as any).msRequestFullscreen()
        }
      } catch (err) {
        console.error(
          `Error attempting to enable fullscreen: ${(err as Error).message}`
        )
      }

      setIsFullscreen(true)

      const handleFullscreenChange = () => {
        if (!document.fullscreenElement) {
          setIsFullscreen(false)
        }
      }

      document.removeEventListener("fullscreenchange", handleFullscreenChange)
      document.addEventListener("fullscreenchange", handleFullscreenChange)
    } else {
      // Exit fullscreen
      if (document.exitFullscreen) {
        document.exitFullscreen()
        setIsFullscreen(false)
      }
    }
  }

  // Usage:
  // <button onClick={() => toggleFullscreen(document.documentElement)}>Go Fullscreen</button>

  return (
    <div
      className={classNames(
        "group relative box-border aspect-video shrink-0 overflow-hidden rounded-lg bg-zinc-700",
        props?.isLocal && "ring-1 ring-blue-400/50"
      )}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {/* always render the video element to ensure the stream is properly attached, but hide it when video is disabled */}
      <div
        className={classNames(
          props?.isVideoEnabled ? "block" : "hidden",
          "h-full w-full"
        )}
      >
        <VideoStream
          id={props?.id}
          stream={props?.stream}
          isMuted={props?.isLocal}
        />
      </div>
      {props && !props.isVideoEnabled && (
        <div
          className={classNames(
            !props?.isVideoEnabled ? "flex" : "hidden",
            "h-full w-full items-center justify-center bg-zinc-800/50"
          )}
        >
          <ParticpantIcon
            participant={{
              id: props?.id,
              name: metadata?.username || "Unknown",
              initials: metadata?.username?.slice(0, 3).toUpperCase() || "UNK",
              color: "#888",
              role: "Guest",
            }}
            size="xl"
          />
        </div>
      )}

      {/* gradient scrim — only visible on hover */}
      <div
        className={classNames(
          "absolute inset-0 bg-linear-to-t from-black/60 via-transparent to-black/20 transition-opacity duration-200",
          showOverlay ? "opacity-100" : "opacity-0"
        )}
      />

      {/* top-right action bar */}
      <div
        className={classNames(
          "absolute top-2 right-2 flex items-center gap-1 transition-opacity duration-200",
          showOverlay ? "opacity-100" : "opacity-0"
        )}
      >
        {/* Pin shortcut */}
        <button
          className="flex h-7 w-7 items-center justify-center rounded-md bg-black/50 text-white/80 backdrop-blur-sm transition hover:bg-black/70 hover:text-white"
          onClick={togglePin}
          title={isPinned ? "Unpin video" : "Pin video"}
        >
          {isPinned ? (
            <PinOff className="h-3.5 w-3.5" />
          ) : (
            <Pin className="h-3.5 w-3.5" />
          )}
        </button>

        {/* Full screen shortcut */}
        <button
          className="flex h-7 w-7 items-center justify-center rounded-md bg-black/50 text-white/80 backdrop-blur-sm transition hover:bg-black/70 hover:text-white"
          onClick={() =>
            props?.id && toggleFullscreen(document.getElementById(props.id)!)
          }
          title={isFullscreen ? "Exit full screen" : "Full screen"}
        >
          {isFullscreen ? (
            <Maximize2 className="h-3.5 w-3.5 rotate-45" />
          ) : (
            <Maximize2 className="h-3.5 w-3.5" />
          )}
        </button>

        {/* More menu */}
        <DropdownMenu onOpenChange={setMenuOpen}>
          <DropdownMenuTrigger asChild>
            <button className="flex h-7 w-7 items-center justify-center rounded-md bg-black/50 text-white/80 backdrop-blur-sm transition hover:bg-black/70 hover:text-white data-[state=open]:bg-black/70 data-[state=open]:text-white">
              <MoreHorizontal className="h-3.5 w-3.5" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            side="bottom"
            align="end"
            className="w-52 border-white/10 bg-zinc-900/95 text-white shadow-2xl backdrop-blur-md"
          >
            {menuItems.map((item) => (
              <DropdownMenuItem
                key={item.label}
                className="flex cursor-pointer items-center gap-2.5 px-3 py-2 text-sm text-white/80 focus:bg-white/10 focus:text-white"
              >
                <span className="text-white/50">{item.icon}</span>
                {item.label}
              </DropdownMenuItem>
            ))}

            {dangerItems.length > 0 && (
              <>
                <DropdownMenuSeparator className="bg-white/10" />
                {dangerItems.map((item) => (
                  <DropdownMenuItem
                    key={item.label}
                    className="flex cursor-pointer items-center gap-2.5 px-3 py-2 text-sm text-red-400 focus:bg-red-500/15 focus:text-red-300"
                  >
                    <span>{item.icon}</span>
                    {item.label}
                  </DropdownMenuItem>
                ))}
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* local "You" badge — always visible */}
      <div className="absolute bottom-4 left-4 flex items-center gap-4">
        {!props?.isAudioEnabled && (
          <div className="rounded bg-red-500 p-2">
            <MicOffIcon className="h-3 w-3" />
          </div>
        )}

        <span className="rounded bg-violet-500 px-3 py-1 text-sm text-white">
          {metadata?.username || "Unknown"}
        </span>
      </div>
    </div>
  )
}

export default VideoTile

export function VideoTileSkeleton() {
  return (
    <div className="relative aspect-video w-full overflow-hidden rounded-lg bg-zinc-800/50 ring-1 ring-white/5">
      {/* 1. The Main Loading Spinner */}
      <div className="flex h-full w-full flex-col items-center justify-center gap-3">
        <Loader2 className="h-8 w-8 animate-spin text-zinc-500" />
        <p className="animate-pulse text-xs font-medium text-zinc-500/80">
          Connecting to stream...
        </p>
      </div>

      {/* 2. Top-Right "Buttons" Skeleton */}
      <div className="absolute top-2 right-2 flex gap-1">
        <Skeleton className="h-7 w-7 rounded-md bg-zinc-700/50" />
        <Skeleton className="h-7 w-7 rounded-md bg-zinc-700/50" />
        <Skeleton className="h-7 w-7 rounded-md bg-zinc-700/50" />
      </div>

      {/* 3. Bottom-Left "Name Tag" Skeleton */}
      <div className="absolute bottom-4 left-4 flex items-center gap-2">
        <Skeleton className="h-7 w-24 rounded bg-zinc-700/50" />
      </div>

      {/* Subtle overlay gradient to match the real tile */}
      <div className="pointer-events-none absolute inset-0 bg-linear-to-t from-black/20 via-transparent to-transparent" />
    </div>
  )
}

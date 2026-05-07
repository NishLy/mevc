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

function VideoTile(props: MediaCombinedStream) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [hovered, setHovered] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)

  const userName = useMemo(() => {
    if (props.isLocal) return "You"
    return (
      props.metadata?.video?.username ||
      props.metadata?.audio?.username ||
      "Participant"
    )
  }, [props.metadata, props.isLocal])

  useEffect(() => {
    if (videoRef.current && props.stream) {
      videoRef.current.srcObject = props.stream
    }
  }, [props.stream, props.id, props.isVideoEnabled])

  const menuItems = props.isLocal ? LOCAL_MENU : REMOTE_MENU
  const dangerItems = props.isLocal ? [] : REMOTE_MENU_DANGER
  const showOverlay = hovered || menuOpen

  const [isFullscreen, setIsFullscreen] = useState(false)

  const pinnedStreamIds = useMeet((state) => state.pinnedStreamIds)
  const isPinned = pinnedStreamIds.includes(props.id)

  const togglePin = () => {
    if (isPinned) {
      useMeet.setState({
        pinnedStreamIds: pinnedStreamIds.filter((id) => id !== props.id),
      })
    } else {
      useMeet.setState({
        pinnedStreamIds: [...pinnedStreamIds, props.id],
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
        "group relative box-border aspect-video max-h-[95vh] shrink-0 overflow-hidden rounded-lg bg-zinc-700",
        props.isLocal && "ring-1 ring-blue-400/50"
      )}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {props.isVideoEnabled ? (
        <video
          ref={videoRef}
          id={props.id}
          autoPlay
          playsInline
          muted={props.isLocal}
          controls={false}
          // poster="https://img.favpng.com/10/24/2/computer-icons-user-icon-design-male-png-favpng-grqs7j1MENUsCah7VD6XBWVst.jpg"
          className="h-full w-full object-contain"
        />
      ) : (
        <div className="flex h-full w-full items-center justify-center bg-zinc-700">
          <ParticpantIcon
            participant={{
              id: props.id,
              name: userName,
              initials: userName.slice(0, 3).toUpperCase() || "UNK",
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
          onClick={() => toggleFullscreen(document.getElementById(props.id)!)}
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
        {!props.isAudioEnabled && (
          <div className="rounded bg-red-500 p-2">
            <MicOffIcon className="h-3 w-3" />
          </div>
        )}

        <span className="rounded bg-violet-500 px-3 py-1 text-sm text-white">
          {userName}
        </span>
      </div>
    </div>
  )
}

export default VideoTile

"use client"

import { useEffect, useRef, useState } from "react"
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
} from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@workspace/ui/components/dropdown-menu"
import { StreamVideoState } from "../types/stream"
import useCurrentRoom from "../state"

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

function VideoTile(props: StreamVideoState) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [hovered, setHovered] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    if (videoRef.current && props.stream) {
      videoRef.current.srcObject = props.stream
    }
  }, [props.stream])

  const menuItems = props.isLocal ? LOCAL_MENU : REMOTE_MENU
  const dangerItems = props.isLocal ? [] : REMOTE_MENU_DANGER
  const showOverlay = hovered || menuOpen

  const [isFullscreen, setIsFullscreen] = useState(false)

  const room = useCurrentRoom()
  const isPinned = room.pinnedStreamIds.includes(props.id)

  const togglePin = () => {
    if (isPinned) {
      room.setPinnedStreamIds(
        room.pinnedStreamIds.filter((id) => id !== props.id)
      )
    } else {
      room.setPinnedStreamIds([...room.pinnedStreamIds, props.id])
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
        "group relative box-border aspect-video max-h-[95vh] shrink-0 overflow-hidden rounded-lg bg-zinc-800",
        props.isLocal && "ring-1 ring-blue-400/50"
      )}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
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

      {/* gradient scrim — only visible on hover */}
      <div
        className={classNames(
          "absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-black/20 transition-opacity duration-200",
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

      {/* bottom name bar */}
      <div
        className={classNames(
          "absolute right-0 bottom-0 left-0 flex items-center justify-between px-2.5 py-2 transition-opacity duration-200",
          showOverlay ? "opacity-100" : "opacity-0"
        )}
      >
        <span className="text-xs font-medium text-white/90 drop-shadow">
          {props.isLocal ? "You" : (props.id ?? "Participant")}
        </span>
      </div>

      {/* local "You" badge — always visible */}
      {props.isLocal && (
        <span className="absolute bottom-2 left-2 rounded bg-blue-500/20 px-1.5 py-0.5 text-[10px] text-blue-300">
          You
        </span>
      )}
    </div>
  )
}

export default VideoTile

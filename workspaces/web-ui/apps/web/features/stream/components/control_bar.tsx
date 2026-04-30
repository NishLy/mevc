"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import { Button } from "@workspace/ui/components/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@workspace/ui/components/tooltip"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@workspace/ui/components/popover"
import { Badge } from "@workspace/ui/components/badge"
import {
  Mic,
  MicOff,
  Video,
  VideoOff,
  Shield,
  Users,
  MessageSquare,
  MonitorUp,
  SmilePlus,
  MoreVertical,
  Phone,
  ChevronUp,
  MoreHorizontal,
  Maximize2,
  Pin,
  VolumeX,
} from "lucide-react"
import classNames from "classnames"
import useMeet from "../state/meet"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@workspace/ui/components/dropdown-menu"
import { MenuItemDef } from "../types/ui"

function useTimer() {
  const [seconds, setSeconds] = useState(0)
  useEffect(() => {
    const id = setInterval(() => setSeconds((s) => s + 1), 1000)
    return () => clearInterval(id)
  }, [])
  const m = String(Math.floor(seconds / 60)).padStart(2, "0")
  const s = String(seconds % 60).padStart(2, "0")
  return `${m}:${s}`
}

interface ControlButtonProps {
  icon: React.ReactNode
  label: string
  tooltip?: string
  active?: boolean
  danger?: boolean
  onClick?: () => void
  caretContent?: React.ReactNode
  className?: string
}

function ControlButton({
  icon,
  label,
  tooltip,
  active,
  danger,
  onClick,
  caretContent,
  className,
}: ControlButtonProps) {
  const button = (
    <div className="flex items-center">
      <Button
        variant="ghost"
        onClick={onClick}
        className={classNames(
          "h-11 gap-2 rounded-xl px-3 text-sm font-medium text-white/80 transition-all hover:bg-white/10 hover:text-white",
          {
            "bg-white/10 text-white": active && !danger,
            "bg-red-500/15 text-red-400 hover:bg-red-500/25 hover:text-red-300":
              danger,
            "rounded-r-none pr-2": caretContent,
          },
          className
        )}
      >
        {icon}
        <span>{label}</span>
      </Button>

      {caretContent && (
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="ghost"
              className="h-11 w-5 rounded-l-none rounded-r-xl px-0 text-white/40 transition-all hover:bg-white/10 hover:text-white/80"
            >
              <ChevronUp className="h-3 w-3" />
            </Button>
          </PopoverTrigger>
          <PopoverContent
            side="top"
            align="center"
            className="w-64 border-white/10 bg-[#1e1e28] p-0 text-white shadow-xl"
          >
            {caretContent}
          </PopoverContent>
        </Popover>
      )}
    </div>
  )

  if (!tooltip) return button

  return (
    <Tooltip>
      <TooltipTrigger asChild>{button}</TooltipTrigger>
      <TooltipContent
        side="top"
        className="border-white/10 bg-[#1e1e28] text-xs text-white"
      >
        {tooltip}
      </TooltipContent>
    </Tooltip>
  )
}

export default function ControlBar() {
  const controller = useMeet((state) => state.controller)
  const localController = useMeet((state) => state.controllerState)
  const timer = useTimer()
  const [menuOpen, setMenuOpen] = useState(false)

  const LOCAL_MENU: MenuItemDef[] = useMemo(
    () => [
      {
        icon: !localController.isCurrentlyRecording ? (
          <Video className="h-3.5 w-3.5" />
        ) : (
          <VideoOff className="h-3.5 w-3.5 text-red-500" />
        ),
        label: !localController.isCurrentlyRecording
          ? "Start Recording"
          : "Stop Recording",
        onClick: () => {
          if (!localController.isCurrentlyRecording) {
            controller?.startRecording()
          } else {
            controller?.stopRecording()
          }
        },
      },
    ],
    [controller, localController.isCurrentlyRecording]
  )

  const audioCaretContent = (
    <div className="py-2">
      <p className="px-3 py-1.5 text-xs font-semibold tracking-wider text-white/40 uppercase">
        Microphone
      </p>
      <Select
        value={localController?.currentAudioDeviceId || ""}
        onValueChange={(value) =>
          controller?.changeAudioDevice(value, { echoCancellation: true })
        }
      >
        <SelectTrigger className="mx-2 mb-1 w-[calc(100%-16px)] border-white/10 bg-white/5 text-sm text-white">
          <SelectValue placeholder="Select microphone" />
        </SelectTrigger>
        <SelectContent className="border-white/10 bg-[#1e1e28] text-white">
          {localController?.availableAudioDevices.map((device) => (
            <SelectItem
              key={device.deviceId}
              value={device.deviceId}
              className="focus:bg-white/10"
            >
              {device.label || `Audio Device ${device.deviceId}`}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )

  const videoCaretContent = (
    <div className="py-2">
      <p className="px-3 py-1.5 text-xs font-semibold tracking-wider text-white/40 uppercase">
        Camera
      </p>
      <Select
        value={localController?.currentVideoDeviceId || ""}
        onValueChange={(value) =>
          controller?.changeVideoDevice(value, {
            height: { ideal: 720 },
            width: { ideal: 1280 },
          })
        }
      >
        <SelectTrigger className="mx-2 mb-1 w-[calc(100%-16px)] border-white/10 bg-white/5 text-sm text-white">
          <SelectValue placeholder="Select camera" />
        </SelectTrigger>
        <SelectContent className="border-white/10 bg-[#1e1e28] text-white">
          {localController?.availableVideoDevices.map((camera) => (
            <SelectItem
              key={camera.deviceId}
              value={camera.deviceId}
              className="focus:bg-white/10"
            >
              {camera.label || `Camera ${camera.deviceId}`}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )

  return (
    <div className="pointer-events-none fixed inset-0 flex items-end justify-center pb-6">
      {/* video here */}
      {/* <video id="localVideo" autoPlay playsInline controls={false} className="absolute inset-0 w-full h-full object-cover" /> */}

      <TooltipProvider delayDuration={300}>
        <div className="pointer-events-auto flex items-center gap-1 rounded-2xl border border-white/8 bg-[#1a1a28]/95 px-4 py-2.5 shadow-2xl backdrop-blur-xl">
          {/* Timer */}
          <span className="min-w-11 text-center text-xs text-white/40 tabular-nums select-none">
            {timer}
          </span>

          <div className="mx-2 h-8 w-px bg-white/10" />

          {/* Mute */}
          <ControlButton
            icon={
              controller?.currentLocalStreamState.audioEnabled ? (
                <Mic className="h-4 w-4" />
              ) : (
                <Mic className="h-4 w-4" />
              )
            }
            label={localController?.audioEnabled ? "Unmute" : "Mute"}
            tooltip={
              localController?.audioEnabled
                ? "Unmute microphone"
                : "Mute microphone"
            }
            danger={!localController?.audioEnabled}
            onClick={controller?.toggleAudio}
            caretContent={audioCaretContent}
          />

          {/* Camera */}
          <ControlButton
            icon={
              localController?.videoEnabled ? (
                <Video className="h-4 w-4" />
              ) : (
                <VideoOff className="h-4 w-4" />
              )
            }
            label={localController?.videoEnabled ? "Stop Video" : "Start Video"}
            tooltip={
              localController?.videoEnabled
                ? "Turn off camera"
                : "Turn on camera"
            }
            danger={!localController?.videoEnabled}
            onClick={controller?.toggleVideo}
            caretContent={videoCaretContent}
          />

          <div className="mx-2 h-8 w-px bg-white/10" />

          {/* Security — dummy */}
          <ControlButton
            icon={<Shield className="h-4 w-4" />}
            label="Security"
            tooltip="Security options"
          />

          {/* Participants — dummy */}
          <ControlButton
            icon={<Users className="h-4 w-4" />}
            label="Participants"
            tooltip="Show participants"
            caretContent={
              <div className="px-3 py-3 text-sm text-white/60">
                <Badge
                  variant="secondary"
                  className="mb-2 border-0 bg-indigo-500/20 text-indigo-300"
                >
                  4 participants
                </Badge>
                <p className="text-xs text-white/30">
                  Participant list coming soon
                </p>
              </div>
            }
          />

          {/* Chat — dummy */}
          <ControlButton
            icon={<MessageSquare className="h-4 w-4" />}
            label="Chat"
            tooltip="Open chat"
          />

          {/* Share Screen */}
          <ControlButton
            icon={<MonitorUp className="h-4 w-4" />}
            label={
              localController?.isCurrentlySharingScreen
                ? "Stop Sharing"
                : "Share Screen"
            }
            tooltip={
              localController?.isCurrentlySharingScreen
                ? "Stop sharing your screen"
                : "Share your screen"
            }
            active={localController?.isCurrentlySharingScreen}
            onClick={() => {
              if (localController?.isCurrentlySharingScreen) {
                controller?.stopScreenShare()
              } else {
                controller?.startScreenShare()
              }
            }}
          />

          {/* Reactions — dummy */}
          <ControlButton
            icon={<SmilePlus className="h-4 w-4" />}
            label="Reactions"
            tooltip="Send a reaction"
            caretContent={
              <div className="flex flex-wrap gap-2 p-3">
                {["👍", "👏", "❤️", "😂", "😮", "🎉", "🙌", "🤔"].map(
                  (emoji) => (
                    <button
                      key={emoji}
                      className="text-xl transition-transform hover:scale-125"
                    >
                      {emoji}
                    </button>
                  )
                )}
              </div>
            }
          />

          {/* More — dummy */}
          {/* More menu */}
          <DropdownMenu onOpenChange={setMenuOpen}>
            <DropdownMenuTrigger asChild>
              <Button className="flex h-10 w-10 items-center justify-center rounded-md bg-black/50 text-white/80 backdrop-blur-sm transition hover:bg-black/70 hover:text-white data-[state=open]:bg-black/70 data-[state=open]:text-white">
                <MoreHorizontal className="h-3.5 w-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              side="bottom"
              align="end"
              className="w-52 border-white/10 bg-zinc-900/95 text-white shadow-2xl backdrop-blur-md"
            >
              {LOCAL_MENU.map((item) => (
                <DropdownMenuItem
                  key={item.label}
                  className="flex cursor-pointer items-center gap-2.5 px-3 py-2 text-sm text-white/80 focus:bg-white/10 focus:text-white"
                  onClick={item.onClick}
                >
                  <span className="text-white/50">{item.icon}</span>
                  {item.label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>

          <div className="mx-2 h-8 w-px bg-white/10" />

          {/* End Call */}
          <Button className="h-11 gap-2 rounded-xl bg-red-500 px-5 text-sm font-semibold text-white transition-colors hover:bg-red-600 active:bg-red-700">
            <Phone className="h-4 w-4 rotate-135" />
            End
          </Button>
        </div>
      </TooltipProvider>
    </div>
  )
}

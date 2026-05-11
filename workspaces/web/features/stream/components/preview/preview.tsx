"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import useMeet from "../../state/meet"
import IRoom from "../../types"
import { LOCAL_STREAM_TYPE } from "../../types/service"
import { MediaStreamController } from "../../services/local"
import { Video, VideoOff } from "lucide-react"

interface IPreviewRoomProps {
  data: IRoom
  clientId?: string
}

export default function PreviewRoom({ data, clientId }: IPreviewRoomProps) {
  const [name, setName] = useState("")
  const [audioOption, setAudioOption] = useState("computer")
  const { controller, setController, controllerState, localStreams } = useMeet()
  const videoRef = useRef<HTMLVideoElement>(null)

  const videoSteam = useMemo(() => {
    if (!controllerState.videoEnabled) return null
    return (
      localStreams.find(
        (stream) => stream?.type === LOCAL_STREAM_TYPE.CAMERA
      ) || null
    )
  }, [controllerState.videoEnabled, localStreams])

  useEffect(() => {
    const localMediaController = new MediaStreamController()
    setController(localMediaController)
    return () => {
      localMediaController.destroy()
    }
  }, [])

  useEffect(() => {
    if (videoRef.current && videoSteam) {
      videoRef.current.srcObject = videoSteam.stream
    }
  }, [videoSteam])

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4">
      {/* Logo + Title */}
      <div className="mb-6 flex flex-col items-center">
        <div className="mb-3 h-12 w-12"></div>
        <h1 className="text-xl font-semibold capitalize">
          {data.name} Meeting Room
        </h1>
      </div>

      {/* Name Input */}
      {data.allow_guests && (
        <div className="mb-4 w-full max-w-2xl">
          <Input
            placeholder="Type your name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="rounded-md border-2 border-gray-300 bg-white p-3 text-gray-900 placeholder:text-gray-500 focus-visible:border-blue-600 focus-visible:ring-1 focus-visible:ring-blue-600 dark:text-white"
          />
        </div>
      )}

      {/* Main Panel */}
      <div className="flex w-full max-w-2xl gap-0 overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
        {/* Left: Camera Preview */}
        <div className="relative flex flex-1 flex-col items-center justify-center bg-gray-200">
          {/* Camera off icon */}
          <div className="flex min-h-44 w-full flex-col items-center justify-center gap-3">
            {!controllerState.videoEnabled && (
              <>
                <VideoOff className="h-10 w-10 text-gray-500" />
                <span className="text-sm font-medium text-gray-600">
                  Your camera is turned off
                </span>
              </>
            )}

            {controllerState.videoEnabled && videoSteam && (
              <video
                ref={videoRef}
                autoPlay
                playsInline
                className="mb-0.5 h-full w-full rounded-lg rounded-tr-none rounded-b-none object-cover"
              />
            )}
          </div>
          {/* Bottom bar */}
          <div className="right-0 bottom-0 left-0 flex w-full items-center justify-start gap-3 bg-gray-300 px-3 py-2">
            <div className="text-background">
              {controllerState.videoEnabled ? <Video /> : <VideoOff />}
            </div>
            <Switch
              checked={controllerState.videoEnabled}
              onCheckedChange={controller?.toggleVideo}
              className="scale-75"
            />
          </div>
        </div>

        {/* Right: Audio Options */}
        <div className="flex w-sm flex-col gap-3 border-l border-gray-200 p-4 text-black">
          <RadioGroup
            value={audioOption}
            onValueChange={setAudioOption}
            className="gap-3"
          >
            {/* Computer Audio */}
            <div className={`rounded-md p-3`}>
              <div className="mb-4 flex items-center gap-2">
                <RadioGroupItem value="computer" id="computer" />
                <Label
                  htmlFor="computer"
                  className="cursor-pointer font-medium"
                >
                  Computer audio
                </Label>
              </div>

              {audioOption === "computer" && (
                <div className="flex flex-col gap-2 pl-1">
                  {/* Microphone */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm text-gray-700">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-4 w-4"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        strokeWidth={1.5}
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          d="M12 18.75a6 6 0 006-6v-1.5m-6 7.5a6 6 0 01-6-6v-1.5m6 7.5v3.75m-3.75 0h7.5M12 15.75a3 3 0 01-3-3V4.5a3 3 0 116 0v8.25a3 3 0 01-3 3z"
                        />
                      </svg>
                      <Select
                        defaultValue={
                          controllerState.currentAudioDeviceId || ""
                        }
                      >
                        <SelectTrigger className="h-auto max-w-60 gap-1 border-none p-0 text-sm shadow-none focus:ring-0">
                          <div className="truncate text-left">
                            <SelectValue placeholder="Microphone" />
                          </div>
                        </SelectTrigger>
                        <SelectContent>
                          {controllerState.availableAudioDevices.map(
                            (device) => (
                              <SelectItem
                                key={device.deviceId}
                                value={device.deviceId}
                              >
                                {device.label ||
                                  "Unknown Device"
                                    .replace(" (Built-in)", "")
                                    .replace(" (USB Audio Device)", "")}
                              </SelectItem>
                            )
                          )}
                        </SelectContent>
                      </Select>
                    </div>
                    <Switch
                      checked={controllerState.audioEnabled}
                      onCheckedChange={controller?.toggleAudio}
                    />
                  </div>
                  <Separator />

                  {/* Speaker */}
                  <div className="flex items-center gap-2 text-sm text-gray-700">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={1.5}
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        d="M19.114 5.636a9 9 0 010 12.728M16.463 8.288a5.25 5.25 0 010 7.424M6.75 8.25l4.72-4.72a.75.75 0 011.28.53v15.88a.75.75 0 01-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.01 9.01 0 012.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75z"
                      />
                    </svg>
                    <Select defaultValue="usb-speaker">
                      <SelectTrigger className="h-auto gap-1 border-none p-0 text-sm shadow-none focus:ring-0">
                        <SelectValue placeholder="Speakers" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="usb-speaker">
                          Speakers (USB Audio Device)
                        </SelectItem>
                        <SelectItem value="default-speaker">
                          Default Speakers
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              )}
            </div>
          </RadioGroup>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="mt-4 flex w-full max-w-2xl justify-end gap-3">
        <Button variant="outline" className="px-6">
          Cancel
        </Button>
        <Button variant="default" className="px-6 text-background">
          Join now
        </Button>
      </div>
    </div>
  )
}

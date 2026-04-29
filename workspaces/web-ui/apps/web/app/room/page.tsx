"use client"

import useStream from "@/hooks/stream"
import { Button } from "@workspace/ui/components/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"

export default function Room() {
  const {
    availableCameras,
    availableAudioDevices,
    selectedCameraId,
    selectedAudioDeviceId,
    currentMediaStream,
    isMuted,
    isVideoEnabled,
    handleCameraChange,
    handleAudioDeviceChange,
    toggleMute,
    toggleVideoStream,
  } = useStream({ videoElementId: "localVideo" })

  return (
    <div>
      <Select
        value={selectedCameraId || ""}
        onValueChange={(value) =>
          handleCameraChange(value, {
            height: { ideal: 720 },
            width: { ideal: 1280 },
          })
        }
      >
        <SelectTrigger>
          <SelectValue placeholder="Select a camera" />
        </SelectTrigger>
        <SelectContent>
          {availableCameras.map((camera) => (
            <SelectItem key={camera.deviceId} value={camera.deviceId}>
              {camera.label || `Camera ${camera.deviceId}`}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        value={selectedAudioDeviceId || ""}
        onValueChange={(value) =>
          handleAudioDeviceChange(value, { echoCancellation: true })
        }
      >
        <SelectTrigger>
          <SelectValue placeholder="Select an audio device" />
        </SelectTrigger>
        <SelectContent>
          {availableAudioDevices.map((device) => (
            <SelectItem key={device.deviceId} value={device.deviceId}>
              {device.label || `Audio Device ${device.deviceId}`}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <video id="localVideo" autoPlay playsInline controls={false} />
      <Button onClick={toggleVideoStream}>
        {isVideoEnabled ? "Disable Video" : "Enable Video"}
      </Button>
      <Button onClick={toggleMute}>{isMuted ? "Unmute" : "Mute"}</Button>
    </div>
  )
}

import useCurrentRoom from "@/features/stream/state"
import {
  AudioOpenSettings,
  CameraOpenSettings,
  StreamVideoEntityType,
} from "@/features/stream/types/stream"
import { useCallback, useEffect, useRef, useState } from "react"

// screen sharing sections
async function requestScreenShare() {
  const mediaStream = await navigator.mediaDevices.getDisplayMedia({
    video: {},
    audio: true,
  })
  return mediaStream
}

// camera and mic sections
async function init(
  onDevicesUpdated?: (
    availableVideoDevices: MediaDeviceInfo[],
    availableAudioDevices: MediaDeviceInfo[]
  ) => void
) {
  const videoCameras = await getConnectedDevices("videoinput")
  const audioDevices = await getConnectedDevices("audioinput")
  if (onDevicesUpdated) {
    onDevicesUpdated(videoCameras, audioDevices)
  }

  navigator.mediaDevices.addEventListener("devicechange", async () => {
    const newVideoCameras = await getConnectedDevices("videoinput")
    const newAudioDevices = await getConnectedDevices("audioinput")
    if (onDevicesUpdated) {
      onDevicesUpdated(newVideoCameras, newAudioDevices)
    }
  })
}

function createBlackVideoTrack(width = 640, height = 480) {
  const canvas = Object.assign(document.createElement("canvas"), {
    width,
    height,
  })
  canvas?.getContext("2d")?.fillRect(0, 0, width, height)
  const stream = canvas.captureStream()
  return stream.getVideoTracks()[0]
}

async function getConnectedDevices(type: string) {
  const devices = await navigator.mediaDevices.enumerateDevices()
  return devices.filter((device) => device.kind === type)
}

async function startStream(
  cameraOptions?: CameraOpenSettings,
  audioOptions?: AudioOpenSettings,
  // Add these parameters to maintain state across device swaps
  videoEnabled = true,
  audioEnabled = true
) {
  const finalStream = new MediaStream()

  // Video Logic
  if (cameraOptions) {
    try {
      const videoStream = await navigator.mediaDevices.getUserMedia({
        video: cameraOptions,
      })
      videoStream.getVideoTracks().forEach((track) => {
        track.enabled = videoEnabled // Sync with state
        finalStream.addTrack(track)
      })
    } catch (e) {
      console.warn("Could not access video device:", e)
      const blackTrack = createBlackVideoTrack()
      if (blackTrack) finalStream.addTrack(blackTrack)
    }
  } else {
    const blackTrack = createBlackVideoTrack()
    if (blackTrack) finalStream.addTrack(blackTrack)
  }

  // Audio Logic
  if (audioOptions) {
    try {
      const audioStream = await navigator.mediaDevices.getUserMedia({
        audio: audioOptions,
      })
      audioStream.getAudioTracks().forEach((track) => {
        track.enabled = audioEnabled // Sync with state
        finalStream.addTrack(track)
      })
    } catch (e) {
      console.warn("Could not access audio device:", e)
    }
  }

  return finalStream
}

const useStream = () => {
  const [availableCameras, setAvailableCameras] = useState<MediaDeviceInfo[]>(
    []
  )
  const [availableAudioDevices, setAvailableAudioDevices] = useState<
    MediaDeviceInfo[]
  >([])

  const [selectedCameraId, setSelectedCameraId] = useState<string | null>(null)
  const [selectedAudioDeviceId, setSelectedAudioDeviceId] = useState<
    string | null
  >(null)

  const [options, setOptions] = useState<{
    camera?: CameraOpenSettings
    audio?: AudioOpenSettings
  }>({})

  const [currentMediaStream, setCurrentMediaStream] =
    useState<MediaStream | null>(null)

  const [isMuted, setIsMuted] = useState(false)
  const [isVideoEnabled, setIsVideoEnabled] = useState(true)

  const [currentScreenShareStream, setCurrentScreenShareStream] =
    useState<MediaStream | null>(null)
  const [isCurrentlyScreenSharing, setIsCurrentlyScreenSharing] =
    useState(false)

  useEffect(() => {
    init((videoDevices, audioDevices) => {
      setAvailableCameras(videoDevices)
      setAvailableAudioDevices(audioDevices)
    })
  }, [])

  const room = useCurrentRoom()

  useEffect(() => {
    if (selectedCameraId || selectedAudioDeviceId) {
      // Pass the current state to the new stream creator
      startStream(options.camera, options.audio, isVideoEnabled, !isMuted).then(
        (stream) => {
          setCurrentMediaStream(stream)
          // Update the room's video streams
          room.setVideoStreams([
            {
              id: "local-video-placeholder",
              stream,
              isLocal: true,
              type: StreamVideoEntityType.SELF,
            },
            ...room.videosStreams.filter(
              (s) => s.type !== StreamVideoEntityType.SELF
            ),
          ])
        }
      )
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [options.camera, options.audio, isVideoEnabled, isMuted])

  const handleCameraChange = useCallback(
    async (id: string, options?: CameraOpenSettings) => {
      setSelectedCameraId(id)
      if (options) {
        setOptions((prev) => ({
          ...prev,
          camera: {
            deviceId: { exact: id },
            ...(options as Record<string, unknown>),
          },
        }))
      }
    },
    []
  )

  const handleAudioDeviceChange = useCallback(
    async (id: string, options?: AudioOpenSettings) => {
      setSelectedAudioDeviceId(id)
      if (options) {
        setOptions((prev) => ({
          ...prev,
          audio: {
            deviceId: { exact: id },
            ...(options as Record<string, unknown>),
          },
        }))
      }
    },
    []
  )

  const toggleVideoStream = useCallback(() => {
    if (currentMediaStream) {
      setIsVideoEnabled((prev) => {
        const newValue = !prev
        currentMediaStream.getVideoTracks().forEach((track) => {
          track.enabled = newValue
        })
        return newValue
      })
    }
  }, [currentMediaStream])

  const toggleMute = useCallback(() => {
    if (currentMediaStream) {
      setIsMuted((prev) => {
        const newValue = !prev
        currentMediaStream.getAudioTracks().forEach((track) => {
          track.enabled = !newValue
        })
        return newValue
      })
    }
  }, [currentMediaStream])

  const stopScreenShare = useCallback(
    (stream: MediaStream) => {
      stream.getTracks().forEach((track) => track.stop())
      setCurrentScreenShareStream(null)
      setIsCurrentlyScreenSharing(false)
      room.setVideoStreams(
        room.videosStreams.filter(
          (s) => s.type !== StreamVideoEntityType.SCREEN_SHARE && s.isLocal
        )
      )
    },
    [room]
  )

  const startScreenShare = useCallback(async () => {
    try {
      const screenStream = await requestScreenShare()
      setCurrentScreenShareStream(screenStream)
      setIsCurrentlyScreenSharing(true)

      room.setVideoStreams([
        {
          id: "local-screen",
          stream: screenStream,
          isLocal: true,
          type: StreamVideoEntityType.SCREEN_SHARE,
        },
        ...room.videosStreams.filter(
          (s) => s.type !== StreamVideoEntityType.SCREEN_SHARE && s.isLocal
        ),
      ])

      if (screenStream.getVideoTracks().length > 0) {
        screenStream.getVideoTracks()[0]!.onended = () => {
          stopScreenShare(screenStream)
        }
      }

      return screenStream
    } catch (e) {
      console.warn("Screen share failed:", e)
      return null
    }
  }, [room, stopScreenShare])

  return {
    availableCameras,
    availableAudioDevices,
    selectedCameraId,
    selectedAudioDeviceId,
    currentMediaStream,
    isMuted,
    isVideoEnabled,
    handleCameraChange,
    toggleMute,
    toggleVideoStream,
    setOptions,
    handleAudioDeviceChange,
    startScreenShare,
    currentScreenShareStream,
    stopScreenShare,
    isCurrentlyScreenSharing,
  }
}

export default useStream

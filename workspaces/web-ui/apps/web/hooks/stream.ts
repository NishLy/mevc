import { AudioOpenSettings, CameraOpenSettings } from "@/types/stream"
import { useCallback, useEffect, useRef, useState } from "react"

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
  element: HTMLVideoElement,
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

  element.srcObject = finalStream
  return finalStream
}

interface UseStreamProps {
  videoElementId: string
}

const useStream = (props: UseStreamProps) => {
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

  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    init((videoDevices, audioDevices) => {
      setAvailableCameras(videoDevices)
      setAvailableAudioDevices(audioDevices)
    })
  }, [])

  useEffect(() => {
    if (videoRef.current) return

    const interval = setInterval(() => {
      const videoElement = document.getElementById(
        props.videoElementId
      ) as HTMLVideoElement | null
      if (videoElement) {
        videoRef.current = videoElement
        clearInterval(interval)
      }
    }, 500)

    return () => clearInterval(interval)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (videoRef.current && (selectedCameraId || selectedAudioDeviceId)) {
      // Pass the current state to the new stream creator
      startStream(
        videoRef.current,
        options.camera,
        options.audio,
        isVideoEnabled,
        !isMuted
      ).then(setCurrentMediaStream)
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
  }
}

export default useStream

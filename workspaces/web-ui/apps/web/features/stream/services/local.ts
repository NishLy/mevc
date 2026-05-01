import {
  AudioOpenSettings,
  CameraOpenSettings,
  MediaStreamItem,
  MediaStreamOptions,
} from "../types/service"

const LOCAL_STREAM_CONSTANT = "local_stream"
const LOCAL_SCREEN_SHARE_STREAM_CONSTRAINT = "local_screen_share_stream"

function createBlackVideoTrack(width = 640, height = 480) {
  const canvas = Object.assign(document.createElement("canvas"), {
    width,
    height,
  })
  canvas?.getContext("2d")?.fillRect(0, 0, width, height)
  const stream = canvas.captureStream()
  return stream.getVideoTracks()[0]
}

export const defaultMediaStreamOptions: MediaStreamOptions = {
  videoEnabled: true,
  audioEnabled: true,
  useAnyAvailableCamera: true,
  useAnyAvailableAudio: true,
  autoStartStream: true,
}

export class MediaStreamController {
  localStreams: MediaStreamItem[] = []
  currentLocalStream: MediaStream | null = null
  currentScreenShareStream: MediaStream | null = null
  availableVideoDevices: MediaDeviceInfo[] = []
  availableAudioDevices: MediaDeviceInfo[] = []
  options: MediaStreamOptions = { ...defaultMediaStreamOptions }

  currentVideoDeviceId: string | null = null
  currentAudioDeviceId: string | null = null

  isCurrentlySharingScreen = false
  isCurrentlyRecording = false

  onDevicesUpdatedCallback?: (
    availableVideoDevices: MediaDeviceInfo[],
    availableAudioDevices: MediaDeviceInfo[]
  ) => void

  onVideoToggleCallback?: (enabled: boolean) => void
  onAudioToggleCallback?: (enabled: boolean) => void
  onVideoDeviceChangeCallback?: (deviceId: string) => void
  onAudioDeviceChangeCallback?: (deviceId: string) => void
  onScreenShareToggleCallback?: (isSharing: boolean) => void
  onLocalStreamUpdateCallback?: (streams: MediaStreamItem[]) => void
  onRecordingToggleCallback?: (isRecording: boolean) => void

  currentLocalStreamState = {
    videoEnabled: true,
    audioEnabled: true,
  }

  constructor(
    public id: string,
    options?: MediaStreamOptions
  ) {
    if (options) {
      this.options = { ...this.options, ...options }
    }
    this.bindMethods()
    this.init()
  }

  private bindMethods() {
    this.getConnectedDevices = this.getConnectedDevices.bind(this)
    this.stopMediaStream = this.stopMediaStream.bind(this)
    this.changeVideoDevice = this.changeVideoDevice.bind(this)
    this.changeAudioDevice = this.changeAudioDevice.bind(this)
    this.toggleVideo = this.toggleVideo.bind(this)
    this.toggleAudio = this.toggleAudio.bind(this)
    this.startLocalStream = this.startLocalStream.bind(this)
    this.stopScreenShare = this.stopScreenShare.bind(this)
    this.startScreenShare = this.startScreenShare.bind(this)
    this.destroy = this.destroy.bind(this)
  }

  private async init() {
    const videoCameras = await this.getConnectedDevices("videoinput")
    const audioDevices = await this.getConnectedDevices("audioinput")

    this.availableVideoDevices = videoCameras
    this.availableAudioDevices = audioDevices

    navigator.mediaDevices.addEventListener("devicechange", async () => {
      const newVideoCameras = await this.getConnectedDevices("videoinput")
      const newAudioDevices = await this.getConnectedDevices("audioinput")

      if (this.onDevicesUpdatedCallback) {
        this.onDevicesUpdatedCallback(newVideoCameras, newAudioDevices)
      }

      this.availableVideoDevices = newVideoCameras
      this.availableAudioDevices = newAudioDevices
    })

    if (this.options.autoStartStream) {
      if (this.options.useAnyAvailableCamera && videoCameras.length > 0) {
        this.currentVideoDeviceId = videoCameras[0]!.deviceId
        if (this.onVideoDeviceChangeCallback) {
          this.onVideoDeviceChangeCallback(this.currentVideoDeviceId)
        }
      }

      if (this.options.useAnyAvailableAudio && audioDevices.length > 0) {
        this.currentAudioDeviceId = audioDevices[0]!.deviceId
        if (this.onAudioDeviceChangeCallback) {
          this.onAudioDeviceChangeCallback(this.currentAudioDeviceId)
        }
      }

      const { id, stream } = await this.startLocalStream(
        {
          ...(this.options.cameraOptions as Record<string, any>),
          deviceId: this.currentVideoDeviceId || undefined,
        },
        {
          ...(this.options.audioOptions as Record<string, any>),
          deviceId: this.currentAudioDeviceId || undefined,
        },
        this.options.videoEnabled,
        this.options.audioEnabled
      )

      this.currentLocalStream = stream

      // Defer setting the local stream to ensure any onLocalStreamUpdateCallback is set
      setTimeout(() => {
        console.log(this.currentLocalStream, this.onLocalStreamUpdateCallback)
        if (this.currentLocalStream) {
          this.setLocalStream(LOCAL_STREAM_CONSTANT, this.currentLocalStream)
        }

        if (this.onDevicesUpdatedCallback) {
          this.onDevicesUpdatedCallback(videoCameras, audioDevices)
        }
      }, 1000)
    }
  }

  private setLocalStream(type = LOCAL_STREAM_CONSTANT, stream: MediaStream) {
    const id = this.id + "_" + type
    const item = this.localStreams.find((s) => s.id === id)

    if (item) {
      item.stream = stream
    } else {
      this.localStreams.push({
        id,
        stream,
        type: "camera",
        isLocal: true,
      })
    }

    if (this.onLocalStreamUpdateCallback) {
      this.onLocalStreamUpdateCallback(this.localStreams)
    }
  }

  private removeLocalStream(id = LOCAL_STREAM_CONSTANT) {
    this.localStreams = this.localStreams.filter((s) => s.id !== id)
    if (this.onLocalStreamUpdateCallback) {
      this.onLocalStreamUpdateCallback(this.localStreams)
    }
  }

  async getConnectedDevices(type: string) {
    // Requesting any stream to ensure device labels are available (some browsers require this)
    await navigator.mediaDevices.getUserMedia({
      video: true,
      audio: true,
    })

    const devices = await navigator.mediaDevices.enumerateDevices()
    return devices.filter((device) => device.kind === type)
  }

  stopMediaStream(stream: MediaStream) {
    stream.getTracks().forEach((track) => track.stop())
  }

  private async requestScreenShare() {
    const mediaStream = await navigator.mediaDevices.getDisplayMedia({
      video: {},
      audio: true,
    })
    return {
      id: this.generateStreamId(),
      stream: mediaStream,
    }
  }

  async changeVideoDevice(newDeviceId: string, options?: CameraOpenSettings) {
    try {
      this.currentVideoDeviceId = newDeviceId
      if (this.currentLocalStream) {
        this.stopMediaStream(this.currentLocalStream)
        this.currentLocalStream = null
      }
      options = options || this.options.cameraOptions
      const stream = await this.startLocalStream(
        { deviceId: newDeviceId, ...(options as Record<string, any>) },
        this.options.audioOptions || {
          deviceId: this.currentAudioDeviceId || undefined,
        },
        this.options.videoEnabled,
        this.options.audioEnabled
      )
      this.currentLocalStream = stream.stream
      this.setLocalStream(LOCAL_STREAM_CONSTANT, this.currentLocalStream)
      if (this.onVideoDeviceChangeCallback) {
        this.onVideoDeviceChangeCallback(newDeviceId)
      }
    } catch (e) {
      console.warn("Could not switch video device:", e)
    }
  }

  async changeAudioDevice(newDeviceId: string, options?: AudioOpenSettings) {
    try {
      this.currentAudioDeviceId = newDeviceId
      if (this.currentLocalStream) {
        this.stopMediaStream(this.currentLocalStream)
        this.currentLocalStream = null
      }
      options = options || this.options.audioOptions
      const stream = await this.startLocalStream(
        this.options.cameraOptions || {
          deviceId: this.currentVideoDeviceId || undefined,
        },
        { deviceId: newDeviceId, ...(options as Record<string, any>) },
        this.options.videoEnabled,
        this.options.audioEnabled
      )
      this.currentLocalStream = stream.stream
      this.setLocalStream(LOCAL_STREAM_CONSTANT, this.currentLocalStream)
      if (this.onAudioDeviceChangeCallback) {
        this.onAudioDeviceChangeCallback(newDeviceId)
      }
    } catch (e) {
      console.warn("Could not switch audio device:", e)
    }
  }

  toggleVideo() {
    if (this.currentLocalStream) {
      this.currentLocalStream.getVideoTracks().forEach((track) => {
        track.enabled = !this.currentLocalStreamState.videoEnabled
      })
      this.currentLocalStreamState.videoEnabled =
        !this.currentLocalStreamState.videoEnabled
      if (this.onVideoToggleCallback) {
        this.onVideoToggleCallback(this.currentLocalStreamState.videoEnabled)
      }
    }
  }

  toggleAudio() {
    if (this.currentLocalStream) {
      this.currentLocalStream.getAudioTracks().forEach((track) => {
        track.enabled = !this.currentLocalStreamState.audioEnabled
      })
      this.currentLocalStreamState.audioEnabled =
        !this.currentLocalStreamState.audioEnabled
      if (this.onAudioToggleCallback) {
        this.onAudioToggleCallback(this.currentLocalStreamState.audioEnabled)
      }
    }
  }

  private generateStreamId() {
    return `stream_${Math.random().toString(36).substr(2, 9)}`
  }

  async startLocalStream(
    cameraOptions?: CameraOpenSettings,
    audioOptions?: AudioOpenSettings,
    // Add these parameters to maintain state across device swaps
    videoEnabled = true,
    audioEnabled = true
  ) {
    const combinedStream = new MediaStream()

    if (cameraOptions) {
      try {
        const videoStream = await navigator.mediaDevices.getUserMedia({
          video: cameraOptions,
        })
        videoStream.getVideoTracks().forEach((track) => {
          track.enabled = videoEnabled // Sync with state
          combinedStream.addTrack(track)
        })
      } catch (e) {
        console.warn("Could not access video device:", e)
        const blackTrack = createBlackVideoTrack()
        if (blackTrack) combinedStream.addTrack(blackTrack)
      }
    } else {
      const blackTrack = createBlackVideoTrack()
      if (blackTrack) combinedStream.addTrack(blackTrack)
    }

    // Audio Logic
    if (audioOptions) {
      try {
        const audioStream = await navigator.mediaDevices.getUserMedia({
          audio: audioOptions,
        })
        audioStream.getAudioTracks().forEach((track) => {
          track.enabled = audioEnabled // Sync with state
          combinedStream.addTrack(track)
        })
      } catch (e) {
        console.warn("Could not access audio device:", e)
      }
    }

    return {
      id: this.generateStreamId(),
      stream: combinedStream,
    }
  }

  stopScreenShare() {
    if (this.currentScreenShareStream) {
      this.stopMediaStream(this.currentScreenShareStream)
      if (this.currentScreenShareStream) {
        this.currentScreenShareStream = null
        this.removeLocalStream(LOCAL_SCREEN_SHARE_STREAM_CONSTRAINT)
      }
    }

    this.isCurrentlySharingScreen = false
    if (this.onScreenShareToggleCallback) {
      this.onScreenShareToggleCallback(false)
    }
  }

  async startScreenShare() {
    const screenStream = await this.requestScreenShare()

    if (screenStream.stream.getVideoTracks().length > 0) {
      screenStream.stream.getVideoTracks()[0]!.onended = () => {
        this.stopScreenShare()
      }
    }

    this.currentScreenShareStream = screenStream.stream
    this.isCurrentlySharingScreen = true
    this.setLocalStream(
      LOCAL_SCREEN_SHARE_STREAM_CONSTRAINT,
      screenStream.stream
    )

    if (this.onScreenShareToggleCallback) {
      this.onScreenShareToggleCallback(true)
    }
  }

  destroy() {
    this.localStreams.forEach((s) => this.stopMediaStream(s.stream))
    this.localStreams = []
    this.currentLocalStream = null
    this.currentScreenShareStream = null
    this.isCurrentlySharingScreen = false
    this.recordedChunks = []
  }

  private recordedChunks: Blob[] = []
  private mediaRecorder: MediaRecorder | null = null

  async startRecording() {
    const displayMediaOptions = {
      video: {
        displaySurface: "browser",
      },
      audio: true,
      preferCurrentTab: true,
    }

    const stream =
      await navigator.mediaDevices.getDisplayMedia(displayMediaOptions)

    // 1. Initialize the recorder
    this.mediaRecorder = new MediaRecorder(stream, { mimeType: "video/webm" })

    // 2. Collect data as it becomes available
    this.mediaRecorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        this.recordedChunks.push(event.data)
      }
    }

    this.mediaRecorder.onstop = this.saveRecording.bind(this)
    this.mediaRecorder.start()
    this.isCurrentlyRecording = true
    if (this.onRecordingToggleCallback) {
      this.onRecordingToggleCallback(true)
    }
  }

  stopRecording() {
    if (this.mediaRecorder && this.mediaRecorder.state !== "inactive") {
      this.mediaRecorder.stop()
    }
  }

  private saveRecording() {
    const blob = new Blob(this.recordedChunks, { type: "video/webm" })
    const url = URL.createObjectURL(blob)

    // Trigger automatic download
    const a = document.createElement("a")
    a.href = url
    a.download = `recording_${new Date().toISOString()}.webm`
    a.click()

    // Cleanup
    URL.revokeObjectURL(url)
    this.recordedChunks = []
    this.isCurrentlyRecording = false
    if (this.onRecordingToggleCallback) {
      this.onRecordingToggleCallback(false)
    }
  }
}

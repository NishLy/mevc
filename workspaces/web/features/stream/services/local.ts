import {
  AudioOpenSettings,
  CameraOpenSettings,
  LOCAL_STREAM_TYPE,
  LocalStreamsTuple,
  MediaStreamOptions,
  TrackMeta,
} from "../types/service"

export function createBlackVideoTrack(width = 640, height = 480) {
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
  useAnyAvailableAudioOutput: true,
  autoStartStream: true,
}

export class MediaStreamController {
  localStreams: LocalStreamsTuple = [null, null]
  activeLocalMediaStream: MediaStream | null = null
  activeScreenShareStream: MediaStream | null = null
  availableVideoDevices: MediaDeviceInfo[] = []
  availableAudioDevices: MediaDeviceInfo[] = []
  availableAudioOutputDevices: MediaDeviceInfo[] = []
  options: MediaStreamOptions = { ...defaultMediaStreamOptions }

  activeVideoDeviceId: string | null = null
  activeAudioDeviceId: string | null = null
  activeAudioOutputDeviceId: string | null = null
  isCurrentlySharingScreen = false
  isCurrentlyRecording = false

  onDevicesUpdatedCallback?: (
    availableVideoDevices: MediaDeviceInfo[],
    availableAudioDevices: MediaDeviceInfo[],
    availableAudioOutputDevices: MediaDeviceInfo[]
  ) => void

  onVideoToggleCallback?: (enabled: boolean) => void
  onAudioToggleCallback?: (enabled: boolean) => void
  onVideoDeviceChangeCallback?: (deviceId: string) => void
  onAudioDeviceChangeCallback?: (deviceId: string) => void
  onAudioOutputDeviceChangeCallback?: (deviceId: string) => void
  onScreenShareToggleCallback?: (isSharing: boolean) => void
  onLocalStreamUpdateCallback?: (streams: LocalStreamsTuple) => void
  onRecordingToggleCallback?: (isRecording: boolean) => void

  activeStreamConfiguration = {
    videoEnabled: true,
    audioEnabled: true,
  }

  constructor(options?: MediaStreamOptions) {
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
    this.initializeMediaStream = this.initializeMediaStream.bind(this)
    this.stopScreenShare = this.stopScreenShare.bind(this)
    this.startScreenShare = this.startScreenShare.bind(this)
    this.destroy = this.destroy.bind(this)
    this.startRecording = this.startRecording.bind(this)
    this.stopRecording = this.stopRecording.bind(this)
    this.saveRecording = this.saveRecording.bind(this)
    this.setLocalStream = this.setLocalStream.bind(this)
    this.removeLocalStream = this.removeLocalStream.bind(this)
  }

  private async init() {
    const videoCameras = await this.getConnectedDevices("videoinput", "video")
    const audioDevices = await this.getConnectedDevices("audioinput", "audio")
    const audioOutputs = await this.getConnectedDevices("audiooutput", "audio")

    // Listen for device changes
    navigator.mediaDevices.addEventListener("devicechange", async () => {
      const newVideoCameras = await this.getConnectedDevices(
        "videoinput",
        "video"
      )
      const newAudioDevices = await this.getConnectedDevices(
        "audioinput",
        "audio"
      )

      const newAudioOutputs = await this.getConnectedDevices(
        "audiooutput",
        "audio"
      )

      if (this.onDevicesUpdatedCallback) {
        this.onDevicesUpdatedCallback(
          newVideoCameras,
          newAudioDevices,
          newAudioOutputs
        )
      }

      this.availableVideoDevices = newVideoCameras
      this.availableAudioDevices = newAudioDevices
      this.availableAudioOutputDevices = newAudioOutputs
    })

    // Auto-start stream if option is enabled
    if (this.options.autoStartStream) {
      // If useAnyAvailableCamera/audio is enabled, set the current device IDs to the first available devices
      if (this.options.useAnyAvailableCamera && videoCameras.length > 0) {
        this.activeVideoDeviceId = videoCameras[0]!.deviceId
        if (this.onVideoDeviceChangeCallback) {
          this.onVideoDeviceChangeCallback(this.activeVideoDeviceId)
        }
      }

      // If useAnyAvailableAudio is enabled, set the current device IDs to the first available devices
      if (this.options.useAnyAvailableAudio && audioDevices.length > 0) {
        this.activeAudioDeviceId = audioDevices[0]!.deviceId
        if (this.onAudioDeviceChangeCallback) {
          this.onAudioDeviceChangeCallback(this.activeAudioDeviceId)
        }
      }

      // If useAnyAvailableAudioOutput is enabled, set the current device IDs to the first available devices
      if (this.options.useAnyAvailableAudioOutput && audioOutputs.length > 0) {
        this.activeAudioOutputDeviceId = audioOutputs[0]!.deviceId
        if (this.onAudioOutputDeviceChangeCallback) {
          this.onAudioOutputDeviceChangeCallback(this.activeAudioOutputDeviceId)
        }
      }

      // Start the local stream with the current device IDs (which may have been set above based on options)
      const { stream } = await this.initializeMediaStream(
        {
          ...(this.options.cameraOptions as Record<string, unknown>),
          deviceId: this.activeVideoDeviceId || undefined,
        },
        {
          ...(this.options.audioOptions as Record<string, unknown>),
          deviceId: this.activeAudioDeviceId || undefined,
        },
        this.options.videoEnabled,
        this.options.audioEnabled
      )

      this.activeLocalMediaStream = stream

      // Ensure the local stream is added to the list of local streams
      this.setLocalStream(LOCAL_STREAM_TYPE.CAMERA, this.activeLocalMediaStream)

      // Store available devices in the instance for later use
      this.onDevicesUpdatedCallback?.(videoCameras, audioDevices, audioOutputs)
    }
  }

  private getLocalStreamsIndex(type: string) {
    return type === LOCAL_STREAM_TYPE.CAMERA ? 0 : 1
  }

  private getIDForStreamType(type: string) {
    return type === LOCAL_STREAM_TYPE.CAMERA
      ? LOCAL_STREAM_TYPE.CAMERA
      : LOCAL_STREAM_TYPE.SCREEN_SHARE
  }

  private generateStreamGroupId() {
    return crypto.randomUUID()
  }

  // Updates the local media stream in the list of local streams, either replacing the existing stream or adding a new entry if it doesn't exist, and triggers an update callback to notify of changes
  private async setLocalStream(
    type = LOCAL_STREAM_TYPE.CAMERA,
    stream: MediaStream
  ) {
    const index = await this.getLocalStreamsIndex(type)
    const item = this.localStreams[index]

    if (item) {
      item.stream = stream
      item.isAudioEnabled = stream
        .getAudioTracks()
        .some((track) => track.enabled)
      item.isVideoEnabled = stream
        .getVideoTracks()
        .some((track) => track.enabled)
    } else {
      const streamGroupId = this.generateStreamGroupId()

      let trackAudio: MediaStreamTrack | null = null
      let trackVideo: MediaStreamTrack | null = null
      const streamId = stream.id

      for (const track of stream.getTracks()) {
        if (track.kind === "audio") {
          trackAudio = track
        } else if (track.kind === "video") {
          trackVideo = track
        }
      }

      const metadataAudio: TrackMeta | null = trackAudio
        ? {
            trackId: trackAudio?.id || "",
            kind: trackAudio?.kind || "",
            clientId: "",
            streamGroupId,
            label: trackAudio?.label || "",
            enabled: trackAudio?.enabled || false,
            username: "",
            streamId: streamId || "",
            transceiverMid: "",
          }
        : null

      const metadataVideo: TrackMeta | null = trackVideo
        ? {
            trackId: trackVideo?.id || "",
            kind: trackVideo?.kind || "",
            clientId: "",
            streamGroupId,
            label: trackVideo?.label || "",
            enabled: trackVideo?.enabled || false,
            username: "",
            streamId: streamId || "",
            transceiverMid: "",
          }
        : null

      this.localStreams[index] = {
        id: this.getIDForStreamType(type),
        stream,
        type,
        isLocal: true,
        isAudioEnabled: stream.getAudioTracks().some((track) => track.enabled),
        isVideoEnabled: stream.getVideoTracks().some((track) => track.enabled),
        metadata: {
          audio: metadataAudio,
          video: metadataVideo,
        },
      }
    }

    this.onLocalStreamUpdateCallback?.(this.localStreams)
  }

  // Removes a local stream from the list of local streams based on the specified type and triggers an update callback to notify of changes
  private async removeLocalStream(type = LOCAL_STREAM_TYPE.CAMERA) {
    const index = await this.getLocalStreamsIndex(type)
    this.localStreams[index] = null
    this.onLocalStreamUpdateCallback?.(this.localStreams)
  }

  // Fetches connected media devices of a specific type (audio or video), requesting permissions if necessary to access device labels
  async getConnectedDevices(type: string, kind: "video" | "audio" | "both") {
    try {
      // Requesting any stream to ensure device labels are available (some browsers require this)
      await navigator.mediaDevices.getUserMedia({
        video: kind === "video" || kind === "both",
        audio: kind === "audio" || kind === "both",
      })

      return (await navigator.mediaDevices.enumerateDevices()).filter(
        (device) => device.kind === type
      )
    } catch (_e) {
      return []
    }
  }

  stopMediaStream(stream: MediaStream) {
    stream.getTracks().forEach((track) => track.stop())
  }

  // Get user media stream based on current device selections and options, throws error if access is denied or fails
  private async requestScreenShare() {
    const mediaStream = await navigator.mediaDevices.getDisplayMedia({
      video: {},
      audio: true,
    })
    return {
      stream: mediaStream,
    }
  }

  // Changes the video input device by stopping the current stream, requesting a new one with the selected device, and updating the local stream reference
  async changeVideoDevice(newDeviceId: string, options?: CameraOpenSettings) {
    try {
      this.activeVideoDeviceId = newDeviceId
      if (this.activeLocalMediaStream) {
        this.stopMediaStream(this.activeLocalMediaStream)
        this.activeLocalMediaStream = null
      }
      options = options || this.options.cameraOptions
      const stream = await this.initializeMediaStream(
        { deviceId: newDeviceId, ...(options as Record<string, unknown>) },
        this.options.audioOptions || {
          deviceId: this.activeAudioDeviceId || undefined,
        },
        this.options.videoEnabled,
        this.options.audioEnabled
      )
      this.activeLocalMediaStream = stream.stream
      await this.setLocalStream(
        LOCAL_STREAM_TYPE.CAMERA,
        this.activeLocalMediaStream
      )
      if (this.onVideoDeviceChangeCallback) {
        this.onVideoDeviceChangeCallback(newDeviceId)
      }
    } catch (e) {
      console.warn("Could not switch video device:", e)
    }
  }

  // Changes the audio input device by stopping the current stream, requesting a new one with the selected device, and updating the local stream reference
  async changeAudioDevice(newDeviceId: string, options?: AudioOpenSettings) {
    try {
      this.activeAudioDeviceId = newDeviceId
      if (this.activeLocalMediaStream) {
        this.stopMediaStream(this.activeLocalMediaStream)
        this.activeLocalMediaStream = null
      }
      options = options || this.options.audioOptions
      const stream = await this.initializeMediaStream(
        this.options.cameraOptions || {
          deviceId: this.activeVideoDeviceId || undefined,
        },
        { deviceId: newDeviceId, ...(options as Record<string, unknown>) },
        this.options.videoEnabled,
        this.options.audioEnabled
      )
      this.activeLocalMediaStream = stream.stream
      this.setLocalStream(LOCAL_STREAM_TYPE.CAMERA, this.activeLocalMediaStream)
      if (this.onAudioDeviceChangeCallback) {
        this.onAudioDeviceChangeCallback(newDeviceId)
      }
    } catch (e) {
      console.warn("Could not switch audio device:", e)
    }
  }

  async changeAudioOutputDevice(newDeviceId: string) {
    try {
      this.activeAudioOutputDeviceId = newDeviceId
      if (this.onAudioOutputDeviceChangeCallback) {
        this.onAudioOutputDeviceChangeCallback(newDeviceId)
      }
    } catch (e) {
      console.warn("Could not switch audio output device:", e)
    }
  }

  // Toggles the enabled state of video tracks in the active local media stream, updating internal state and invoking a callback if provided
  toggleVideo() {
    if (this.activeLocalMediaStream) {
      this.activeLocalMediaStream.getVideoTracks().forEach((track) => {
        track.enabled = !this.activeStreamConfiguration.videoEnabled
      })
      this.activeStreamConfiguration.videoEnabled =
        !this.activeStreamConfiguration.videoEnabled
      if (this.onVideoToggleCallback) {
        this.onVideoToggleCallback(this.activeStreamConfiguration.videoEnabled)
      }

      this.setLocalStream(LOCAL_STREAM_TYPE.CAMERA, this.activeLocalMediaStream)
    }
  }

  // Toggles the enabled state of audio tracks in the active local media stream, updating internal state and invoking a callback if provided
  toggleAudio() {
    if (this.activeLocalMediaStream) {
      this.activeLocalMediaStream.getAudioTracks().forEach((track) => {
        track.enabled = !this.activeStreamConfiguration.audioEnabled
      })
      this.activeStreamConfiguration.audioEnabled =
        !this.activeStreamConfiguration.audioEnabled
      if (this.onAudioToggleCallback) {
        this.onAudioToggleCallback(this.activeStreamConfiguration.audioEnabled)
      }

      this.setLocalStream(LOCAL_STREAM_TYPE.CAMERA, this.activeLocalMediaStream)
    }
  }

  // Initializes a new media stream based on provided camera and audio options, applying the current enabled/disabled state for video and audio tracks to ensure consistency across device changes
  async initializeMediaStream(
    cameraOptions?: CameraOpenSettings,
    audioOptions?: AudioOpenSettings,

    // Add these parameters to maintain state across device swaps
    videoEnabled = true,
    audioEnabled = true
  ) {
    const combinedStream = new MediaStream()

    if (cameraOptions && (cameraOptions as Record<string, unknown>).deviceId) {
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
        const blackTrack = createBlackVideoTrack(1280, 720) // Create a black track with a common resolution
        if (blackTrack) combinedStream.addTrack(blackTrack)
      }
    } else {
      const blackTrack = createBlackVideoTrack(1280, 720) // Create a black track with a common resolution
      if (blackTrack) combinedStream.addTrack(blackTrack)
    }

    // Audio Logic
    if (audioOptions && (audioOptions as Record<string, unknown>).deviceId) {
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
      stream: combinedStream,
    }
  }

  // Screen sharing logic
  stopScreenShare() {
    if (this.activeScreenShareStream) {
      if (this.activeScreenShareStream) {
        this.removeLocalStream(LOCAL_STREAM_TYPE.SCREEN_SHARE)
        this.stopMediaStream(this.activeScreenShareStream)
        this.activeScreenShareStream = null
      }
    }

    this.isCurrentlySharingScreen = false
    if (this.onScreenShareToggleCallback) {
      this.onScreenShareToggleCallback(false)
    }
  }

  // Starts screen sharing and sets up an event listener to detect when the user stops sharing via browser UI
  async startScreenShare() {
    const screenStream = await this.requestScreenShare()

    if (screenStream.stream.getVideoTracks().length > 0) {
      screenStream.stream.getVideoTracks()[0]!.onended = () => {
        this.stopScreenShare()
      }
    }

    this.activeScreenShareStream = screenStream.stream
    this.isCurrentlySharingScreen = true
    this.setLocalStream(LOCAL_STREAM_TYPE.SCREEN_SHARE, screenStream.stream)

    if (this.onScreenShareToggleCallback) {
      this.onScreenShareToggleCallback(true)
    }
  }

  // Cleans up all media streams and resets state, intended to be called when the controller is no longer needed to ensure proper resource management
  destroy() {
    this.localStreams.forEach((stream) => {
      if (stream) {
        this.stopMediaStream(stream.stream)
      }
    })

    this.localStreams = [null, null]
    this.activeLocalMediaStream = null
    this.activeScreenShareStream = null
    this.isCurrentlySharingScreen = false
    this.recordedChunks = []
  }

  private recordedChunks: Blob[] = []
  private mediaRecorder: MediaRecorder | null = null

  // Starts recording the screen using the MediaRecorder API, collecting data chunks as they become available and saving the recording when stopped
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

  // Stops the recording and triggers the save process to download the recorded video
  stopRecording() {
    if (this.mediaRecorder && this.mediaRecorder.state !== "inactive") {
      this.mediaRecorder.stop()
    }
  }

  // Saves the recorded video by creating a Blob from the collected chunks, generating a download link, and triggering an automatic download for the user
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

package rtc

import (
	"fmt"
	"sync"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

type Viewer struct {
	session Session
	track   *webrtc.TrackLocalStaticRTP
	sender  *webrtc.RTPSender
}

// TrackRouter handles a single incoming stream and broadcasts it to many viewers
type TrackRouter struct {
	sync.RWMutex
	streamID      string
	streamGroupId string
	incomingTrack *webrtc.TrackRemote
	publisherPC   *webrtc.PeerConnection // The person sending the video
	viewers       map[string]*Viewer
	metadata      *SessionTrackMetadata
	done          chan struct{}
	hasStarted    bool
	publisherID   string
	priority      int
}

func NewTrackRouter(publisherPC *webrtc.PeerConnection, publisherID string) *TrackRouter {
	router := &TrackRouter{
		incomingTrack: nil, // Set this when the publisher adds the track
		publisherPC:   publisherPC,
		viewers:       make(map[string]*Viewer),
		metadata:      nil,
		publisherID:   publisherID,
		done:          make(chan struct{}),
		priority:      4, // Default to lowest priority
	}

	// START THE SINGLE READER LOOP HERE

	return router
}

func (r *TrackRouter) SetPriority(priority int) {
	r.Lock()
	r.priority = priority
	r.Unlock()
}

func (r *TrackRouter) SetIncomingTrack(track *webrtc.TrackRemote) {
	r.Lock()
	r.incomingTrack = track
	r.streamID = track.StreamID()
	r.Unlock()
}

func (r *TrackRouter) SetStreamID(streamID string) {
	r.Lock()
	r.streamID = streamID
	r.Unlock()
}

func (r *TrackRouter) SetStreamGroupId(streamGroupId string) {
	r.Lock()
	r.streamGroupId = streamGroupId
	r.Unlock()
}

func (r *TrackRouter) SetMetadata(metadata *SessionTrackMetadata) {
	r.Lock()
	r.metadata = metadata
	r.Unlock()
}

// broadcastLoop reads from the remote track exactly ONCE and copies to all viewers
func (r *TrackRouter) broadcastLoop() {
	rtpBuf := make([]byte, 1500)
	for {
		// Read raw bytes from the incoming track
		i, _, err := r.incomingTrack.Read(rtpBuf)
		if err != nil {
			return // Publisher disconnected
		}

		r.RLock()
		for _, viewer := range r.viewers {
			// Write raw bytes. TrackLocalStaticRTP automatically handles
			// the Payload Type translation! No manual check() needed.
			if _, err := viewer.track.Write(rtpBuf[:i]); err != nil {
				// Handle write error (e.g., viewer disconnected)
			}
		}
		r.RUnlock()
	}
}

// AddViewer is called when a new peer wants to watch the stream
func (r *TrackRouter) AddViewer(viewerSession Session) error {
	r.Lock()
	defer r.Unlock()

	if viewerSession.GetClientId() == r.publisherID {
		return fmt.Errorf("publisher cannot be a viewer of their own stream")
	}

	if _, exists := r.viewers[viewerSession.GetClientId()]; exists {
		return fmt.Errorf("viewer already exists")
	}

	viewerPC := viewerSession.GetPeerConnection()

	if r.incomingTrack == nil {
		return fmt.Errorf("incoming track not set yet")
	}

	// 1. Create the outgoing track for this specific viewer
	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		r.incomingTrack.Codec().RTPCodecCapability,
		r.incomingTrack.ID(),
		r.incomingTrack.StreamID(),
	)

	if err != nil {
		return err
	}

	// 2. Add it to the viewer's PeerConnection
	rtpSender, err := viewerPC.AddTrack(localTrack) // AddTrack is simpler if you don't need transceiver control
	if err != nil {
		return err
	}

	// 3. Safely add this track to our broadcaster list
	r.viewers[viewerSession.GetClientId()] = &Viewer{
		session: viewerSession,
		track:   localTrack,
		sender:  rtpSender,
	}

	// 4. CRITICAL: Ask the PUBLISHER for a Keyframe so the new viewer can start decoding
	err = r.publisherPC.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{MediaSSRC: uint32(r.incomingTrack.SSRC())},
	})

	logger.Sugar.Infof("Added viewer %s to stream %s (track ID=%s)", viewerSession.GetClientId(), r.incomingTrack.StreamID(), r.incomingTrack.ID())
	return nil
}

func (r *TrackRouter) Start() {
	if r.incomingTrack == nil || r.metadata == nil {
		return
	}

	go r.broadcastLoop()
	go r.requestKeyframes()

	r.hasStarted = true
	// logger.Sugar.Infof("Started router for stream %s with track %s (kind=%s)", r.incomingTrack.StreamID(), r.incomingTrack.ID(), r.incomingTrack.Kind().String())
}

func (r *TrackRouter) RemoveViewer(viewerID string) error {
	r.Lock()
	defer r.Unlock()

	if viewer, exists := r.viewers[viewerID]; exists {
		// Tell the viewer's PC to stop sending this track
		pc := viewer.session.GetPeerConnection()

		if pc != nil && pc.ConnectionState() != webrtc.PeerConnectionStateClosed && pc.ConnectionState() != webrtc.PeerConnectionStateFailed && pc.ConnectionState() != webrtc.PeerConnectionStateDisconnected {
			err := pc.RemoveTrack(viewer.sender)

			if err != nil {
				return fmt.Errorf("failed to remove track for viewer %s: %v", viewerID, err)
			}
		}
	}

	delete(r.viewers, viewerID)
	return nil
}

func (r *TrackRouter) requestKeyframes() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-r.done:
			// The router was closed! Exit the goroutine.
			return

		case <-ticker.C:
			// Time to ask for a keyframe
			err := r.publisherPC.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(r.incomingTrack.SSRC())},
			})
			if err != nil {
				return // Publisher is gone
			}
		}
	}
}

func (r *TrackRouter) Close() {
	r.Lock()
	defer r.Unlock()

	// 1. Close the done channel to signal the PLI ticker to stop.
	// We use a select to ensure we don't panic if Close() is called twice.
	select {
	case <-r.done:
	default:
		close(r.done)
	}

	for clientId, v := range r.viewers {
		// Tell the viewer's PC to stop sending this track
		pc := v.session.GetPeerConnection()

		if pc != nil && pc.ConnectionState() != webrtc.PeerConnectionStateClosed && pc.ConnectionState() != webrtc.PeerConnectionStateFailed && pc.ConnectionState() != webrtc.PeerConnectionStateDisconnected {
			err := pc.RemoveTrack(v.sender)
			if err != nil {
				logger.Sugar.Warnf("Failed to remove track for viewer %s: %v", clientId, err)
			}
		}
	}

	// 2. Clean up the viewer map so we drop references to their tracks
	r.viewers = make(map[string]*Viewer)
	r.hasStarted = false
}

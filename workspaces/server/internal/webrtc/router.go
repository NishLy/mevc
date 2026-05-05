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
}

// TrackRouter handles a single incoming stream and broadcasts it to many viewers
type TrackRouter struct {
	sync.RWMutex
	incomingTrack *webrtc.TrackRemote
	publisherPC   *webrtc.PeerConnection // The person sending the video
	viewers       map[string]*Viewer
	metadata      *SessionTrackMetadata
	done          chan struct{}
	hasStarted    bool
	publisherID   string
}

func NewTrackRouter(publisherPC *webrtc.PeerConnection, publisherID string) *TrackRouter {
	router := &TrackRouter{
		incomingTrack: nil, // Set this when the publisher adds the track
		publisherPC:   publisherPC,
		viewers:       make(map[string]*Viewer),
		metadata:      nil,
		publisherID:   publisherID,
		done:          make(chan struct{}),
	}

	// START THE SINGLE READER LOOP HERE

	return router
}

func (r *TrackRouter) SetIncomingTrack(track *webrtc.TrackRemote) {
	r.Lock()
	r.incomingTrack = track
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
	if _, err := viewerPC.AddTransceiverFromTrack(localTrack); err != nil {
		return err
	}

	// 3. Safely add this track to our broadcaster list
	r.viewers[viewerSession.GetClientId()] = &Viewer{
		session: viewerSession,
		track:   localTrack,
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
	logger.Sugar.Infof("Started router for stream %s with track %s (kind=%s)", r.incomingTrack.StreamID(), r.incomingTrack.ID(), r.incomingTrack.Kind().String())
}

func (r *TrackRouter) RemoveViewer(viewerID string) {
	r.Lock()
	delete(r.viewers, viewerID)
	r.Unlock()
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

	// 2. Clean up the viewer map so we drop references to their tracks
	r.viewers = make(map[string]*Viewer)
	r.hasStarted = false
}

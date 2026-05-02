package rtc

import (
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/pion/webrtc/v4"
)

func forwardTrack(track *webrtc.TrackRemote, session Session) {
	for _, mt := range session.GetTransceivers() {
		if mt.kind != track.Kind() {
			continue // skip if kind doesn't match
		}

		mt.mu.Lock()
		if mt.busy {
			mt.mu.Unlock()
			continue
		}

		mt.busy = true
		mt.mu.Unlock()

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			track.Codec().RTPCodecCapability,
			track.ID(),
			track.StreamID(),
		)

		if err != nil {
			mt.mu.Lock()
			mt.busy = false
			mt.mu.Unlock()
			return
		}

		if err := mt.t.Sender().ReplaceTrack(localTrack); err != nil {
			logger.Sugar.Warnf("ReplaceTrack failed for track %s: %v", track.ID(), err)
			mt.mu.Lock()
			mt.busy = false
			mt.mu.Unlock()
			continue
		}

		// Renegotiate so the subscriber browser's ontrack fires with the correct
		// MSID (containing the publisher's real track ID) instead of the
		// placeholder ID that was negotiated during the initial offer/answer.
		if err := session.Renegotiate(); err != nil {
			logger.Sugar.Warnf("Renegotiate failed for track %s: %v", track.ID(), err)
		}

		go func() {
			defer func() {
				mt.t.Sender().ReplaceTrack(nil)
				mt.mu.Lock()
				mt.busy = false
				mt.mu.Unlock()
				session.RemoveRemoteTrack(track.ID())
				logger.Sugar.Debugf("Slot freed for track %s", track.ID())
			}()

			buf := make([]byte, 1500)
			for {
				n, _, err := track.Read(buf)
				if err != nil {
					return
				}
				localTrack.Write(buf[:n])
			}
		}()

		session.SetSubscribedTrack(track.ID(), true)
		return
	}

	logger.Sugar.Warnf("No available transceiver slot for session %s to forward track %s (kind=%s)", session.GetClientId(), track.ID(), track.Kind())
}

package rtc

import (
	"github.com/pion/webrtc/v4"
)

func forwardTrack(track *webrtc.TrackRemote, session Session) (*webrtc.RTPTransceiver, error) {
	pc := session.GetPeerConnection()

	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		track.Codec().RTPCodecCapability,
		track.ID(),
		track.StreamID(),
	)
	if err != nil {
		return nil, err
	}

	// create transceiver
	transceiver, err := pc.AddTransceiverFromTrack(localTrack)
	if err != nil {
		return nil, err
	}

	sender := transceiver.Sender()

	// RTCP handling
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := sender.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	// RTP forwarding
	go func() {
		for {
			pkt, _, err := track.ReadRTP()
			if err != nil {
				return
			}

			if err := localTrack.WriteRTP(pkt); err != nil {
				return
			}
		}
	}()

	return transceiver, nil
}

package rtc

import (
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

func createLocalTrancieverAndTrack(track *webrtc.TrackRemote, session Session) (*webrtc.RTPTransceiver, *webrtc.TrackLocalStaticRTP, error) {
	pc := session.GetPeerConnection()

	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		track.Codec().RTPCodecCapability,
		track.ID(),
		track.StreamID(),
	)
	if err != nil {
		return nil, nil, err
	}

	transceiver, err := pc.AddTransceiverFromTrack(localTrack)
	if err != nil {
		return nil, nil, err
	}

	return transceiver, localTrack, nil
}

func check(sender *webrtc.RTPSender, track *webrtc.TrackRemote) uint8 {
	params := sender.GetParameters()

	var targetPT uint8
	for _, codec := range params.Codecs {
		if codec.MimeType == track.Codec().MimeType {
			targetPT = uint8(codec.PayloadType)
			break
		}
	}

	return targetPT
}

func forwardTrack(pc *webrtc.PeerConnection, transceiver *webrtc.RTPTransceiver, track *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) error {
	for {
		if pc.SignalingState() == webrtc.SignalingStateStable {
			params := transceiver.Sender().GetParameters()
			if len(params.Codecs) > 0 {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	go func() {
		pli := []rtcp.Packet{
			&rtcp.PictureLossIndication{
				MediaSSRC: uint32(track.SSRC()),
			},
		}

		for {
			if err := pc.WriteRTCP(pli); err != nil {
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()

	sender := transceiver.Sender()

	// RTCP reader
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := sender.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	// RTP forward (FIXED)
	go func() {
		var targetPT uint8 = 0

		for {
			pkt, _, err := track.ReadRTP()
			if err != nil {
				return
			}

			if targetPT == 0 {
				targetPT = check(sender, track)

				if targetPT == 0 {
					continue
				}

			}

			pkt.PayloadType = targetPT

			if err := localTrack.WriteRTP(pkt); err != nil {
				return
			}
		}
	}()

	return nil
}

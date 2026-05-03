package rtc

import (
	"time"

	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
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

	return transceiver, localTrack, err
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

func forwardTrack(pc *webrtc.PeerConnection, transceiver *webrtc.RTPTransceiver, track *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP, clientID string) error {
	logger.Sugar.Infof("Starting to forward track for client %s: SSRC=%d, PT=%d, Seq=%d", clientID, track.SSRC(), track.PayloadType(), 0)

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
				logger.Sugar.Errorf("ReadRTP error: %v", err)
				return
			}

			if targetPT == 0 {
				targetPT = check(sender, track)

				if targetPT == 0 {
					continue
				}
			}

			// logger.Sugar.Infof("Writing RTP packet: SSRC=%d, PT=%d, Seq=%d",
			// 	pkt.SSRC, pkt.PayloadType, pkt.SequenceNumber)

			pkt.PayloadType = targetPT

			if err := localTrack.WriteRTP(pkt); err != nil {
				logger.Sugar.Errorf("WriteRTP error: %v", err)
				return
			}
		}
	}()

	return nil
}

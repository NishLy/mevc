package rtc

import (
	"errors"
	"fmt"
	"net"
	"os"

	socketio "github.com/googollee/go-socket.io"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type udpConn struct {
	conn        *net.UDPConn
	port        int
	payloadType uint8
}

var udpConns = map[string]*udpConn{
	"audio": {port: 4000, payloadType: 111},
	"video": {port: 4002, payloadType: 96},
}

func WebRTCBootstrap(io *socketio.Server) {
	pc := MustCreatePeerConnection()
	defer pc.Close()

	MustAddTransceivers(pc)
	mustDialUDP()
	registerHandlers(pc, io)

	<-webrtc.GatheringCompletePromise(pc)
	select {}
}

func MustCreatePeerConnection() *webrtc.PeerConnection {
	me := &webrtc.MediaEngine{}
	MustRegisterCodecs(me)

	ir := &interceptor.Registry{}
	pli, err := intervalpli.NewReceiverInterceptor()
	Must(err)
	ir.Add(pli)
	Must(webrtc.RegisterDefaultInterceptors(me, ir))

	api := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithInterceptorRegistry(ir))
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	Must(err)
	return pc
}

func MustRegisterCodecs(me *webrtc.MediaEngine) {
	codecs := []struct {
		mime      string
		clockRate uint32
		kind      webrtc.RTPCodecType
	}{
		{webrtc.MimeTypeVP8, 90000, webrtc.RTPCodecTypeVideo},
		{webrtc.MimeTypeOpus, 48000, webrtc.RTPCodecTypeAudio},
	}
	for _, c := range codecs {
		Must(me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: c.mime, ClockRate: c.clockRate},
		}, c.kind))
	}
}

func MustAddTransceivers(pc *webrtc.PeerConnection) {
	for _, kind := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeAudio, webrtc.RTPCodecTypeVideo} {
		_, err := pc.AddTransceiverFromKind(kind)
		Must(err)
	}
}

func mustDialUDP() {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:")
	Must(err)

	for _, c := range udpConns {
		raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", c.port))
		Must(err)
		c.conn, err = net.DialUDP("udp", laddr, raddr)
		Must(err)
	}
}

func registerHandlers(pc *webrtc.PeerConnection, io *socketio.Server) {
	pc.OnTrack(forwardTrack)

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("ICE state: %s\n", state)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Connection state: %s\n", state)
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			fmt.Println("Done forwarding")
			os.Exit(0)
		}
	})

	io.OnEvent("/", "send_offer", func(c socketio.Conn, id string, offer interface{}) {
		fmt.Printf("Received offer from client %s\n", c.ID())

		pc := MustCreatePeerConnection()
		MustAddTransceivers(pc)

		Must(pc.SetRemoteDescription(offer.(webrtc.SessionDescription)))

		answer, err := pc.CreateAnswer(nil)
		Must(err)
		Must(pc.SetLocalDescription(answer))

		go func() {
			<-webrtc.GatheringCompletePromise(pc)
			fmt.Printf("Sending answer to client %s\n", c.ID())
			c.Emit("receive_answer", pc.LocalDescription()) // send back to same client
		}()
	})
}

func forwardTrack(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	conn, ok := udpConns[track.Kind().String()]
	if !ok {
		return
	}

	buf := make([]byte, 1500)
	pkt := &rtp.Packet{}
	for {
		n, _, err := track.Read(buf)
		if err != nil {
			panic(err)
		}
		Must(pkt.Unmarshal(buf[:n]))
		pkt.PayloadType = conn.payloadType
		n, err = pkt.MarshalTo(buf)
		Must(err)

		if _, err = conn.conn.Write(buf[:n]); err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Err.Error() == "write: connection refused" {
				continue
			}
			panic(err)
		}
	}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

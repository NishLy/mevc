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
	pc := mustCreatePeerConnection()
	defer pc.Close()

	mustAddTransceivers(pc)
	mustDialUDP()
	registerHandlers(pc, io)

	<-webrtc.GatheringCompletePromise(pc)
	select {}
}

func mustCreatePeerConnection() *webrtc.PeerConnection {
	me := &webrtc.MediaEngine{}
	mustRegisterCodecs(me)

	ir := &interceptor.Registry{}
	pli, err := intervalpli.NewReceiverInterceptor()
	must(err)
	ir.Add(pli)
	must(webrtc.RegisterDefaultInterceptors(me, ir))

	api := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithInterceptorRegistry(ir))
	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	must(err)
	return pc
}

func mustRegisterCodecs(me *webrtc.MediaEngine) {
	codecs := []struct {
		mime      string
		clockRate uint32
		kind      webrtc.RTPCodecType
	}{
		{webrtc.MimeTypeVP8, 90000, webrtc.RTPCodecTypeVideo},
		{webrtc.MimeTypeOpus, 48000, webrtc.RTPCodecTypeAudio},
	}
	for _, c := range codecs {
		must(me.RegisterCodec(webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: c.mime, ClockRate: c.clockRate},
		}, c.kind))
	}
}

func mustAddTransceivers(pc *webrtc.PeerConnection) {
	for _, kind := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeAudio, webrtc.RTPCodecTypeVideo} {
		_, err := pc.AddTransceiverFromKind(kind)
		must(err)
	}
}

func mustDialUDP() {
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:")
	must(err)

	for _, c := range udpConns {
		raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", c.port))
		must(err)
		c.conn, err = net.DialUDP("udp", laddr, raddr)
		must(err)
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

		pc := mustCreatePeerConnection()
		mustAddTransceivers(pc)

		must(pc.SetRemoteDescription(offer.(webrtc.SessionDescription)))

		answer, err := pc.CreateAnswer(nil)
		must(err)
		must(pc.SetLocalDescription(answer))

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
		must(pkt.Unmarshal(buf[:n]))
		pkt.PayloadType = conn.payloadType
		n, err = pkt.MarshalTo(buf)
		must(err)

		if _, err = conn.conn.Write(buf[:n]); err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Err.Error() == "write: connection refused" {
				continue
			}
			panic(err)
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

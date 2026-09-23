// Command viewer is a headless client for the board's three video paths, so what
// each one costs the board can be measured with no browser. It counts what
// arrives and discards it. See README.md.
package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

var bytesIn, pkts, audioMsgs atomic.Int64

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mint" {
		mintMain(os.Args[2:])
		return
	}

	mode := flag.String("mode", "webrtc", "webrtc | direct | mjpeg")
	host := flag.String("host", "10.0.0.222", "board")
	secs := flag.Int("secs", 40, "seconds to view")
	tokFile := flag.String("token", "kvm-token", "token file")
	withAudio := flag.Bool("audio", false, "direct: ask for audio and count it apart from video")
	flag.Parse()
	tok, err := os.ReadFile(*tokFile)
	if err != nil {
		panic(err)
	}
	cookie := "nano-kvm-token=" + strings.TrimSpace(string(tok))

	done := make(chan struct{})
	go func() { time.Sleep(time.Duration(*secs) * time.Second); close(done) }()
	go report(done)

	switch *mode {
	case "mjpeg":
		mjpeg(*host, cookie, done)
	case "direct":
		path := "/api/stream/h264/direct"
		if *withAudio {
			path += "?audio=1"
		}
		wsRead(*host, path, cookie, done)
	case "webrtc":
		rtc(*host, cookie, done)
	}
	fmt.Printf("TOTAL mode=%s bytes=%d packets=%d audio=%d\n", *mode, bytesIn.Load(), pkts.Load(), audioMsgs.Load())
}

func report(done chan struct{}) {
	t := time.NewTicker(5 * time.Second)
	var last int64
	for {
		select {
		case <-done:
			return
		case <-t.C:
			b := bytesIn.Load()
			fmt.Printf("%s rx=%dKB/s packets=%d\n", time.Now().Format("15:04:05"), (b-last)/5/1024, pkts.Load())
			last = b
		}
	}
}

func dialer() *websocket.Dialer {
	return &websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, HandshakeTimeout: 10 * time.Second}
}

func hdr(cookie string) http.Header {
	h := http.Header{}
	h.Set("Cookie", cookie)
	return h
}

func mjpeg(host, cookie string, done chan struct{}) {
	c := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	req, _ := http.NewRequest("GET", "https://"+host+"/api/stream/mjpeg", nil)
	req.Header.Set("Cookie", cookie)
	r, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println("mjpeg status", r.Status)
	go func() { <-done; r.Body.Close() }()
	buf := make([]byte, 64<<10)
	for {
		n, err := r.Body.Read(buf)
		bytesIn.Add(int64(n))
		if err != nil {
			return
		}
	}
}

func wsRead(host, path, cookie string, done chan struct{}) {
	ws, resp, err := dialer().Dial("wss://"+host+path, hdr(cookie))
	if err != nil {
		fmt.Println("dial:", err, resp)
		os.Exit(1)
	}
	go func() { <-done; ws.Close() }()
	for {
		_, m, err := ws.ReadMessage()
		if err != nil {
			return
		}
		// Byte 0 is the keyframe flag on video and 0x10 on audio.
		if len(m) > 0 && m[0] == 0x10 {
			audioMsgs.Add(1)
			continue
		}
		bytesIn.Add(int64(len(m)))
		pkts.Add(1)
	}
}

type msg struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

func rtc(host, cookie string, done chan struct{}) {
	ws, resp, err := dialer().Dial("wss://"+host+"/api/stream/h264", hdr(cookie))
	if err != nil {
		fmt.Println("dial:", err, resp)
		os.Exit(1)
	}
	defer ws.Close()

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}
	// The server answers nothing without this extension in the offer.
	if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{
		URI: "http://www.webrtc.org/experiments/rtp-hdrext/playout-delay"}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}
	defer pc.Close()
	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})
	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) { fmt.Println("peer", s) })
	pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		fmt.Println("track", t.Kind(), t.Codec().MimeType)
		buf := make([]byte, 1600)
		for {
			n, _, err := t.Read(buf)
			if err != nil {
				return
			}
			if t.Kind() == webrtc.RTPCodecTypeVideo {
				bytesIn.Add(int64(n))
				pkts.Add(1)
			}
		}
	})

	var wmu = make(chan struct{}, 1)
	send := func(ev, data string) {
		wmu <- struct{}{}
		defer func() { <-wmu }()
		_ = ws.WriteJSON(msg{ev, data})
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		panic(err)
	}
	gather := webrtc.GatheringCompletePromise(pc)
	if err := pc.SetLocalDescription(offer); err != nil {
		panic(err)
	}
	<-gather
	ld, _ := json.Marshal(pc.LocalDescription())
	send("video-offer", string(ld))

	go func() {
		t := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-done:
				return
			case <-t.C:
				send("heartbeat", "")
			}
		}
	}()
	go func() { <-done; ws.Close() }()

	for {
		var mm msg
		if err := ws.ReadJSON(&mm); err != nil {
			if err != io.EOF {
				select {
				case <-done:
				default:
					fmt.Println("ws:", err)
				}
			}
			return
		}
		switch mm.Event {
		case "video-answer":
			var sd webrtc.SessionDescription
			_ = json.Unmarshal([]byte(mm.Data), &sd)
			if err := pc.SetRemoteDescription(sd); err != nil {
				fmt.Println("answer:", err)
			} else {
				fmt.Println("answer set")
			}
		case "video-candidate":
			var c webrtc.ICECandidateInit
			_ = json.Unmarshal([]byte(mm.Data), &c)
			_ = pc.AddICECandidate(c)
		}
	}
}

// mintMain reads the JWT secret line and /etc/kvm/pwd on stdin and writes a
// session token to -out, readable by its owner only.
func mintMain(args []string) {
	fs := flag.NewFlagSet("mint", flag.ExitOnError)
	out := fs.String("out", "kvm-token", "token file to write")
	ttl := fs.Duration("ttl", 3*time.Hour, "how long the token lasts")
	_ = fs.Parse(args)

	secret, pwd, err := readMintInput(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mint:", err)
		os.Exit(1)
	}
	tok, user, err := mintToken(secret, pwd, time.Now(), *ttl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mint:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, []byte(tok), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "mint:", err)
		os.Exit(1)
	}
	fmt.Printf("token for %s written to %s, valid for %s\n", user, *out, *ttl)
}

package backend

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"
)

type A2SPlayer struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Score    int32   `json:"score"`
	Duration float64 `json:"duration"`
}

type A2SResult struct {
	Port       int         `json:"port"`
	Name       string      `json:"name"`
	Map        string      `json:"map"`
	Folder     string      `json:"folder"`
	Game       string      `json:"game"`
	AppID      int         `json:"appId"`
	Players    int         `json:"players"`
	MaxPlayers int         `json:"maxPlayers"`
	Bots       int         `json:"bots"`
	ServerType string      `json:"serverType"`
	OS         string      `json:"os"`
	Passworded bool        `json:"passworded"`
	VAC        bool        `json:"vac"`
	Version    string      `json:"version"`
	PingMs     int64       `json:"pingMs"`
	PlayerList []A2SPlayer `json:"playerList"`
}

var a2sInfoReq = append([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x54}, []byte("Source Engine Query\x00")...)

func a2sExchange(conn net.Conn, req []byte) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write(req); err != nil {
		return nil, err
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if n < 5 || binary.LittleEndian.Uint32(buf[:4]) != 0xFFFFFFFF {
		return nil, errors.New("unsupported A2S response (split packet or bad header)")
	}
	return buf[4:n], nil
}

type a2sReader struct{ b *bytes.Reader }

func (r a2sReader) u8() byte { v, _ := r.b.ReadByte(); return v }
func (r a2sReader) u16() uint16 {
	var v uint16
	binary.Read(r.b, binary.LittleEndian, &v)
	return v
}
func (r a2sReader) i32() int32 {
	var v int32
	binary.Read(r.b, binary.LittleEndian, &v)
	return v
}
func (r a2sReader) f32() float32 {
	var v uint32
	binary.Read(r.b, binary.LittleEndian, &v)
	return math.Float32frombits(v)
}
func (r a2sReader) str() string {
	var out []byte
	for {
		c, err := r.b.ReadByte()
		if err != nil || c == 0 {
			return string(out)
		}
		out = append(out, c)
	}
}

func a2sQuery(host string, port int) (*A2SResult, error) {
	conn, err := net.Dial("udp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	start := time.Now()
	resp, err := a2sExchange(conn, a2sInfoReq)
	if err != nil {
		return nil, err
	}
	ping := time.Since(start).Milliseconds()
	if resp[0] == 0x41 && len(resp) >= 5 {
		resp, err = a2sExchange(conn, append(append([]byte{}, a2sInfoReq...), resp[1:5]...))
		if err != nil {
			return nil, err
		}
		ping = time.Since(start).Milliseconds()
	}
	if resp[0] != 0x49 {
		return nil, fmt.Errorf("unexpected A2S_INFO header 0x%02x", resp[0])
	}
	r := a2sReader{bytes.NewReader(resp[1:])}
	r.u8()
	res := &A2SResult{Port: port, PingMs: ping}
	res.Name, res.Map, res.Folder, res.Game = r.str(), r.str(), r.str(), r.str()
	res.AppID = int(r.u16())
	res.Players, res.MaxPlayers, res.Bots = int(r.u8()), int(r.u8()), int(r.u8())
	switch r.u8() {
	case 'd':
		res.ServerType = "dedicated"
	case 'l':
		res.ServerType = "listen"
	default:
		res.ServerType = "other"
	}
	switch r.u8() {
	case 'w':
		res.OS = "windows"
	case 'l':
		res.OS = "linux"
	default:
		res.OS = "other"
	}
	res.Passworded = r.u8() == 1
	res.VAC = r.u8() == 1
	res.Version = r.str()
	res.PlayerList = []A2SPlayer{}

	chal := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	preq := append([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x55}, chal...)
	pr, err := a2sExchange(conn, preq)
	if err == nil && pr[0] == 0x41 && len(pr) >= 5 {
		pr, err = a2sExchange(conn, append([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x55}, pr[1:5]...))
	}
	if err == nil && pr[0] == 0x44 {
		pr2 := a2sReader{bytes.NewReader(pr[1:])}
		n := int(pr2.u8())
		for i := 0; i < n && pr2.b.Len() > 0; i++ {
			idx := int(pr2.u8())
			name := pr2.str()
			score := pr2.i32()
			dur := float64(pr2.f32())
			res.PlayerList = append(res.PlayerList, A2SPlayer{Index: idx, Name: name, Score: score, Duration: dur})
		}
	}
	return res, nil
}

func (h *Host) queryPortFor(sv Server) int {
	r := h.runtime(sv)
	if r.QueryPort > 0 {
		return r.QueryPort
	}
	if isPal(sv) {
		return flagInt(ParseArkCmd(sv.Cmd), "queryport", 27015)
	}
	return 27015
}

func (h *Host) QueryA2S(id string) (*A2SResult, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	return a2sQuery("127.0.0.1", h.queryPortFor(sv))
}

func (h *Host) QueryA2SAddress(address string, port int) (*A2SResult, error) {
	if address == "" || port <= 0 || port > 65535 {
		return nil, errors.New("address and a valid port are required")
	}
	return a2sQuery(address, port)
}

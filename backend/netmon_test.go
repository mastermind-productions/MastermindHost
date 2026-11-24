package backend

import (
	"bytes"
	"encoding/binary"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
)

func fakeA2S(t *testing.T) int {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 2048)
		hdr := []byte{0xFF, 0xFF, 0xFF, 0xFF}
		chal := []byte{1, 2, 3, 4}
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			req := buf[:n]
			switch req[4] {
			case 0x54:
				if len(req) == 5+20 {
					pc.WriteTo(append(append([]byte{}, hdr...), append([]byte{0x41}, chal...)...), addr)
					continue
				}
				b := bytes.NewBuffer(append([]byte{}, hdr...))
				b.WriteByte(0x49)
				b.WriteByte(17)
				for _, s := range []string{"TestSrv", "TheIsland_WP", "ark", "ARK"} {
					b.WriteString(s)
					b.WriteByte(0)
				}
				binary.Write(b, binary.LittleEndian, uint16(29999))
				b.Write([]byte{2, 70, 0, 'd', 'w', 1, 1})
				b.WriteString("1.2.3")
				b.WriteByte(0)
				pc.WriteTo(b.Bytes(), addr)
			case 0x55:
				if bytes.Equal(req[5:9], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
					pc.WriteTo(append(append([]byte{}, hdr...), append([]byte{0x41}, chal...)...), addr)
					continue
				}
				b := bytes.NewBuffer(append([]byte{}, hdr...))
				b.Write([]byte{0x44, 1, 0})
				b.WriteString("Alice")
				b.WriteByte(0)
				binary.Write(b, binary.LittleEndian, int32(5))
				binary.Write(b, binary.LittleEndian, float32(125))
				pc.WriteTo(b.Bytes(), addr)
			}
		}
	}()
	return pc.LocalAddr().(*net.UDPAddr).Port
}

func TestA2S(t *testing.T) {
	r, err := a2sQuery("127.0.0.1", fakeA2S(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "TestSrv" || r.Map != "TheIsland_WP" || r.Players != 2 || r.MaxPlayers != 70 || !r.Passworded || r.Version != "1.2.3" {
		t.Fatalf("bad info %+v", r)
	}
	if len(r.PlayerList) != 1 || r.PlayerList[0].Name != "Alice" || r.PlayerList[0].Duration != 125 {
		t.Fatalf("bad players %+v", r.PlayerList)
	}
}

func TestScanBoundPorts(t *testing.T) {
	pc, _ := net.ListenPacket("udp", "127.0.0.1:0")
	defer pc.Close()
	port := pc.LocalAddr().(*net.UDPAddr).Port
	m, err := scanBoundPorts()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m["UDP/"+strconv.Itoa(port)]; !ok {
		t.Fatalf("port %d not found in scan", port)
	}
}

func TestAffinity(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skip()
	}
	cmd := exec.Command("ping", "-n", "6", "127.0.0.1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	pid := uint32(cmd.Process.Pid)
	mask, _ := cpusToMask([]int{0, 2})
	if err := setProcessAffinity(pid, mask); err != nil {
		t.Fatal(err)
	}
	got, err := getProcessAffinity(pid)
	if err != nil || got != mask {
		t.Fatalf("got %b err %v", got, err)
	}
}

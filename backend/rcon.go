package backend

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	rconAuth         = 3
	rconExecCommand  = 2
	rconResponseAuth = 2
)

func rconPacket(id, typ int32, body string) []byte {
	var buf bytes.Buffer
	size := int32(len(body) + 10)
	binary.Write(&buf, binary.LittleEndian, size)
	binary.Write(&buf, binary.LittleEndian, id)
	binary.Write(&buf, binary.LittleEndian, typ)
	buf.WriteString(body)
	buf.Write([]byte{0, 0})
	return buf.Bytes()
}

func rconRead(conn net.Conn) (int32, int32, string, error) {
	var size int32
	if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
		return 0, 0, "", err
	}
	if size < 10 || size > 1<<20 {
		return 0, 0, "", fmt.Errorf("invalid rcon packet size %d", size)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(conn, data); err != nil {
		return 0, 0, "", err
	}
	id := int32(binary.LittleEndian.Uint32(data[0:4]))
	typ := int32(binary.LittleEndian.Uint32(data[4:8]))
	body := string(bytes.TrimRight(data[8:], "\x00"))
	return id, typ, body, nil
}

func RconExec(host string, port int, password, command string, timeout time.Duration) (string, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write(rconPacket(1, rconAuth, password)); err != nil {
		return "", err
	}
	for {
		id, typ, _, err := rconRead(conn)
		if err != nil {
			return "", err
		}
		if typ == rconResponseAuth {
			if id == -1 {
				return "", errors.New("rcon authentication failed")
			}
			break
		}
	}
	if command == "" {
		return "", nil
	}
	if _, err := conn.Write(rconPacket(2, rconExecCommand, command)); err != nil {
		return "", err
	}
	_, _, body, err := rconRead(conn)
	if err != nil {
		return "", err
	}
	return body, nil
}

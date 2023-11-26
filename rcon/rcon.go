package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type PacketType int32

const (
	ResponseValue             PacketType = 0
	ExecCommandOrAuthResponse            = 2
	Auth                                 = 3
)

type Packet struct {
	Size int32
	ID   int32
	Type PacketType
	Body []byte
}

func newPacket(id int32, packetType PacketType, body []byte) Packet {
	return Packet{
		Size: int32(len(body) + 10),
		ID:   id,
		Type: packetType,
		Body: body,
	}
}

func (p *Packet) marshal() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(p.Body)+10))

	binary.Write(buf, binary.LittleEndian, p.Size)
	binary.Write(buf, binary.LittleEndian, p.ID)
	binary.Write(buf, binary.LittleEndian, p.Type)
	binary.Write(buf, binary.LittleEndian, p.Body)
	binary.Write(buf, binary.LittleEndian, [2]byte{})

	return buf.Bytes(), nil
}

func unmashalPacket(r io.Reader) (Packet, error) {
	packet := Packet{}

	err := binary.Read(r, binary.LittleEndian, &packet.Size)
	if err != nil {
		return packet, err
	}

	err = binary.Read(r, binary.LittleEndian, &packet.ID)
	if err != nil {
		return packet, err
	}

	err = binary.Read(r, binary.LittleEndian, &packet.Type)
	if err != nil {
		return packet, err
	}

	// IDとType分のバイト数を引く
	packet.Body = make([]byte, packet.Size-8)
	_, err = io.ReadFull(r, packet.Body)
	if err != nil {
		return packet, err
	}

	// 末尾のnullを取り除く
	packet.Body = packet.Body[:len(packet.Body)-2]

	return packet, nil
}

type Conn struct {
	conn net.Conn
}

func New(addr string, password string) (Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return Conn{}, err
	}

	client := Conn{conn}
	err = client.authenticate(password)
	if err != nil {
		return Conn{}, err
	}

	return client, nil
}

func (c Conn) Close() {
	c.conn.Close()
}

func (c Conn) Exec(command string) error {
	p, err := c.send(ExecCommandOrAuthResponse, command)
	if err != nil {
		return err
	}

	fmt.Println("Response: " + string(p.Body))

	return nil
}

func (c Conn) authenticate(password string) error {
	p, err := c.send(Auth, password)
	if err != nil {
		return err
	}

	// 認証に失敗したらIDに-1が入る
	if p.ID == -1 {
		return errors.New("failed auth")
	}

	return err
}

func (c Conn) send(packetType PacketType, bodyStr string) (Packet, error) {
	p := newPacket(0, packetType, []byte(bodyStr))
	b, err := p.marshal()
	if err != nil {
		return Packet{}, err
	}

	_, err = c.conn.Write(b)
	if err != nil {
		return Packet{}, err
	}

	resp, err := unmashalPacket(c.conn)
	if err != nil {
		return Packet{}, err
	}

	return resp, nil
}

package vless

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const (
	Version       = 0
	CommandTCP    = 0x01
	CommandUDP    = 0x02
	AddrTypeIPv4  = 0x01
	AddrTypeDomain = 0x02
	AddrTypeIPv6  = 0x03
)

var ErrInvalidVersion = errors.New("vless: invalid protocol version")
var ErrAuthFailed = errors.New("vless: authentication failed")

type Server struct {
	logger *zap.Logger
	users  map[string]string // uuid -> name
}

type Conn struct {
	net.Conn
	uuid     string
	target   string
	command  byte
}

func NewServer(logger *zap.Logger, users map[string]string) *Server {
	return &Server{
		logger: logger,
		users:  users,
	}
}

func (s *Server) Handle(conn net.Conn) {
	defer conn.Close()

	header, err := ReadHeader(conn)
	if err != nil {
		s.logger.Error("failed to read vless header", zap.Error(err))
		return
	}

	if _, ok := s.users[header.UUID]; !ok {
		s.logger.Warn("auth failed", zap.String("uuid", header.UUID))
		conn.Write([]byte{Version, 0x01, 0x00}) // status 1 = auth fail
		return
	}

	// Send success response
	conn.Write([]byte{Version, 0x00, 0x00})

	s.logger.Info("client connected",
		zap.String("uuid", header.UUID),
		zap.String("target", header.Target),
		zap.String("command", commandName(header.Command)),
	)

	targetConn, err := net.Dial("tcp", header.Target)
	if err != nil {
		s.logger.Error("failed to connect to target", zap.Error(err), zap.String("target", header.Target))
		return
	}
	defer targetConn.Close()

	// Bidirectional copy
	done := make(chan struct{})
	go func() {
		io.Copy(targetConn, conn)
		close(done)
	}()
	go func() {
		io.Copy(conn, targetConn)
		close(done)
	}()
	<-done
}

type Header struct {
	Version  byte
	UUID     string
	Flow     string
	Command  byte
	Target   string
	AddrType byte
}

func ReadHeader(r io.Reader) (*Header, error) {
	// Read version
	var versionBuf [1]byte
	if _, err := io.ReadFull(r, versionBuf[:]); err != nil {
		return nil, err
	}

	version := versionBuf[0]
	if version != Version {
		return nil, ErrInvalidVersion
	}

	// Read UUID (16 bytes)
	var uuidBuf [16]byte
	if _, err := io.ReadFull(r, uuidBuf[:]); err != nil {
		return nil, err
	}
	uuid := formatUUID(uuidBuf[:])

	// Read addons length
	var addonsLenBuf [1]byte
	if _, err := io.ReadFull(r, addonsLenBuf[:]); err != nil {
		return nil, err
	}
	addonsLen := int(addonsLenBuf[0])

	// Read addons (flow)
	flow := ""
	if addonsLen > 0 {
		addonsBuf := make([]byte, addonsLen)
		if _, err := io.ReadFull(r, addonsBuf); err != nil {
			return nil, err
		}
		flow = string(addonsBuf)
	}

	// Read command
	var cmdBuf [1]byte
	if _, err := io.ReadFull(r, cmdBuf[:]); err != nil {
		return nil, err
	}

	// Read address
	addrTypeBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, addrTypeBuf); err != nil {
		return nil, err
	}
	addrType := addrTypeBuf[0]

	// Read port (2 bytes)
	var portBuf [2]byte
	if _, err := io.ReadFull(r, portBuf[:]); err != nil {
		return nil, err
	}
	port := binary.BigEndian.Uint16(portBuf[:])

	// Read address based on type
	var host string
	switch addrType {
	case AddrTypeIPv4:
		ipBuf := make([]byte, 4)
		if _, err := io.ReadFull(r, ipBuf); err != nil {
			return nil, err
		}
		host = net.IP(ipBuf).String()

	case AddrTypeDomain:
		var domainLenBuf [1]byte
		if _, err := io.ReadFull(r, domainLenBuf[:]); err != nil {
			return nil, err
		}
		domainBuf := make([]byte, domainLenBuf[0])
		if _, err := io.ReadFull(r, domainBuf); err != nil {
			return nil, err
		}
		host = string(domainBuf)

	case AddrTypeIPv6:
		ipBuf := make([]byte, 16)
		if _, err := io.ReadFull(r, ipBuf); err != nil {
			return nil, err
		}
		host = net.IP(ipBuf).String()

	default:
		return nil, fmt.Errorf("vless: unknown address type: %d", addrType)
	}

	return &Header{
		Version:  version,
		UUID:     uuid,
		Flow:     flow,
		Command:  cmdBuf[0],
		Target:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		AddrType: addrType,
	}, nil
}

func formatUUID(b []byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func commandName(cmd byte) string {
	switch cmd {
	case CommandTCP:
		return "TCP"
	case CommandUDP:
		return "UDP"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", cmd)
	}
}

func ParseUUID(s string) (string, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return "", fmt.Errorf("vless: invalid UUID length")
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		s[0:8], s[8:12], s[12:16], s[16:20], s[20:32]), nil
}

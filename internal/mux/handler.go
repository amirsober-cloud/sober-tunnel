package mux

import (
	"net"
	"strconv"
	"time"

	"github.com/amirsober-cloud/sober-tunnel/internal/protocol"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) *Handler {

	return &Handler{
		manager: manager,
	}
}

func (h *Handler) Handle(conn net.Conn) {

	session := NewSession(conn)

	id := conn.RemoteAddr().String()

	h.manager.AddSession(
		id,
		session,
	)

	defer func() {

		h.manager.RemoveSession(id)

		session.Close()

	}()

	for {

		frame, err := protocol.ReadFrame(conn)

		if err != nil {
			return
		}

		switch frame.Command {

		case protocol.CmdOpen:

			h.handleOpen(
				session,
				frame,
			)

		case protocol.CmdData:

			h.handleData(
				session,
				frame,
			)

		case protocol.CmdClose:

			h.handleClose(
				session,
				frame,
			)

		default:

			return
		}

	}
}

func (h *Handler) handleOpen(
	session *Session,
	frame *protocol.Frame,
) {

	if session.GetStream(frame.StreamID) != nil {
		return
	}

	targetAddr := net.JoinHostPort(
		frame.Address,
		strconv.Itoa(int(frame.Port)),
	)

	conn, err := net.DialTimeout(
		"tcp",
		targetAddr,
		5*time.Second,
	)

	if err != nil {
		session.WriteFrame(&protocol.Frame{Version: protocol.Version1, Command: protocol.CmdClose, StreamID: frame.StreamID})
		return
	}

	stream := &Stream{

		ID: frame.StreamID,

		Session: session,

		Conn: conn,
	}

	if !session.AddStream(stream) {
		conn.Close()
		return
	}

	go h.readStream(stream)
}
func (h *Handler) readStream(
	stream *Stream,
) {
	defer func() {
		stream.Close()
		stream.Session.RemoveStream(stream.ID)
		stream.Session.WriteFrame(&protocol.Frame{Version: protocol.Version1, Command: protocol.CmdClose, StreamID: stream.ID})
	}()

	buf := make(
		[]byte,
		32*1024,
	)

	for {

		n, err := stream.Conn.Read(buf)

		if n > 0 {
			frame := &protocol.Frame{

				Version: protocol.Version1,

				Command: protocol.CmdData,

				StreamID: stream.ID,

				Payload: append(
					[]byte(nil),
					buf[:n]...,
				),
			}

			writeErr := stream.Session.WriteFrame(
				frame,
			)

			if writeErr != nil {

				stream.Close()

				return
			}
		}
		if err != nil {
			return
		}
	}
}
func (h *Handler) handleData(
	session *Session,
	frame *protocol.Frame,
) {

	stream := session.GetStream(
		frame.StreamID,
	)

	if stream == nil {
		return
	}

	if stream.Conn == nil {
		return
	}

	if len(frame.Payload) > 1024*1024 {
		return
	}

	_, err := stream.Conn.Write(
		frame.Payload,
	)

	if err != nil {

		stream.Close()

		session.RemoveStream(
			frame.StreamID,
		)

		return
	}
}

func (h *Handler) handleClose(
	session *Session,
	frame *protocol.Frame,
) {

	stream := session.GetStream(
		frame.StreamID,
	)

	if stream == nil {
		return
	}

	stream.Close()

	session.RemoveStream(
		frame.StreamID,
	)
}

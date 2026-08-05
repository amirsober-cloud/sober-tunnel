package mux

import (
	"fmt"
	"net"

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

		conn.Close()

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

	targetAddr := fmt.Sprintf(
		"%s:%d",
		frame.Address,
		frame.Port,
	)

	conn, err := net.Dial(
		"tcp",
		targetAddr,
	)

	if err != nil {
		return
	}

	stream := &Stream{

		ID: frame.StreamID,

		Session: session,

		Conn: conn,
	}

	session.AddStream(stream)

	go h.readStream(stream)
}
func (h *Handler) readStream(
	stream *Stream,
) {

	buf := make(
		[]byte,
		32*1024,
	)

	for {

		n, err := stream.Conn.Read(buf)

		if err != nil {

			stream.Close()

			return
		}

		frame := &protocol.Frame{

			Version: protocol.Version1,

			Command: protocol.CmdData,

			StreamID: stream.ID,

			Payload: append(
				[]byte(nil),
				buf[:n]...,
			),
		}

		err = protocol.WriteFrame(
			stream.Session.conn,
			frame,
		)

		if err != nil {

			stream.Close()

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

package websocket

import (
	"io"

	"github.com/gorilla/websocket"
)

func NewWriter(con *websocket.Conn) *binaryWriter {
	return &binaryWriter{
		conn: con,
	}
}

func NewReader(con *websocket.Conn) *binaryReader {
	return &binaryReader{
		conn: con,
	}
}

type binaryWriter struct {
	conn *websocket.Conn
}

func (s *binaryWriter) Write(p []byte) (int, error) {
	w, err := s.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, convert(err)
	}
	defer w.Close()
	n, err := w.Write(p)
	return n, err
}

type binaryReader struct {
	conn   *websocket.Conn
	reader io.Reader
}

func (s *binaryReader) Read(p []byte) (int, error) {
	var msgType int
	var err error
	for {
		if s.reader == nil {
			msgType, s.reader, err = s.conn.NextReader()
			if err != nil {
				s.reader = nil
				return 0, convert(err)
			}
		} else {
			msgType = websocket.BinaryMessage
		}

		switch msgType {
		case websocket.BinaryMessage:
			n, readErr := s.reader.Read(p)
			err = readErr
			if err != nil {
				s.reader = nil
				if err == io.EOF {
					if n == 0 {
						continue
					} else {
						return n, nil
					}
				}
			}
			return n, convert(err)
		case websocket.CloseMessage:
			return 0, io.EOF
		default:
			s.reader = nil
		}
	}
}

func convert(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*websocket.CloseError); ok && e.Code == websocket.CloseNormalClosure {
		return io.EOF
	}
	return err
}

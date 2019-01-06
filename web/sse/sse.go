package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type event struct {
	typ  string
	data interface{}
}

type Stream struct {
	w       http.ResponseWriter
	flusher http.Flusher
	event   chan event
	err     chan error
	closed  bool
}

func Open(w http.ResponseWriter) *Stream {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, _ := w.(http.Flusher)
	s := &Stream{
		w:       w,
		flusher: flusher,
		event:   make(chan event, 128),
		err:     make(chan error),
	}
	go s.encoder()
	return s
}

func (s *Stream) encoder() {
	defer close(s.err)
	var err error
	enc := json.NewEncoder(s.w)
	for e := range s.event {
		_, err = fmt.Fprintf(s.w, "event: %s\ndata: ", e.typ)
		if err != nil {
			break
		}
		err = enc.Encode(e.data)
		if err != nil {
			break
		}
		_, err = fmt.Fprint(s.w, "\n\n")
		if err != nil {
			break
		}
		if s.flusher != nil {
			s.flusher.Flush()
		}
	}
	s.err <- err
}

func (s *Stream) Send(ev string, data interface{}) {
	s.event <- event{
		typ:  ev,
		data: data,
	}
}

func (s *Stream) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	close(s.event)
	return <-s.err
}

func (s *Stream) Err() <-chan error {
	return s.err
}

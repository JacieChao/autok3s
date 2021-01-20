package common

type LogWriter struct {
	ch   chan []byte
	Name string
}

func NewLogWriter(name string) *LogWriter {
	return &LogWriter{
		ch:   make(chan []byte),
		Name: name,
	}
}

func (w *LogWriter) Close() {
	close(w.ch)
}

func (w *LogWriter) Write(b []byte) (int, error) {
	w.ch <- b
	return 0, nil
}

func (w *LogWriter) Read() <-chan []byte {
	return w.ch
}

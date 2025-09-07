package combined

// Daemon executes the injected worker engines concurrently, each in their own
// goroutine. Daemon itself does not block, so that it is the responsibility of
// the calling goroutine to stop the programm to exit prematurely. Any signal
// handling must then be injected directly into the underlying worker engines.
func (w *Worker) Daemon() {
	go w.par.Daemon()
	go w.seq.Daemon()
}

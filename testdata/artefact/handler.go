package artefact

import "time"

type Handler struct{}

func (h *Handler) Cooler() time.Duration {
	return 0
}

func (h *Handler) Ensure() error {
	return nil
}

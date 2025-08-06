package handler

import (
	"time"
)

func (h *Handler) Cooler() time.Duration {
	return h.coo
}

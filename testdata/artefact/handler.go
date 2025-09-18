package artefact

type Handler struct{}

func (h *Handler) Active() bool {
	return true
}

func (h *Handler) Ensure() error {
	return nil
}

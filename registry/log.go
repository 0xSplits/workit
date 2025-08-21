package registry

func (r *Registry) Log(err error) bool {
	return r.fil(err)
}

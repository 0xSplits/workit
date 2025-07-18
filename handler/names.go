package handler

// Names returns the list of worker handler names matching the given list of
// worker handler interfaces. E.g. this function should enable the metrics
// registry to whitelist all injected worker handlers automatically.
func Names(han []Interface) []string {
	var lis []string

	for _, x := range han {
		lis = append(lis, Name(x))
	}

	return lis
}

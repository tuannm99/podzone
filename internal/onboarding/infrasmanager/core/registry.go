package core

// Registry keeps provisioners keyed by InfraType.
// It prevents accidental duplicate registration.
type Registry struct {
	m map[InfraType]InfraProvisioner
}

func NewRegistry(m map[InfraType]InfraProvisioner) (*Registry, error) {
	if m == nil {
		return &Registry{m: map[InfraType]InfraProvisioner{}}, nil
	}
	// validate
	for t, p := range m {
		if t == "" || p == nil {
			return nil, ErrInvalidInput
		}
	}
	return &Registry{m: m}, nil
}

func (r *Registry) Get(t InfraType) (InfraProvisioner, bool) {
	p, ok := r.m[t]
	return p, ok
}

package fiberapp

type HealthChecker func() error

type HealthRegistry struct {
	healthCheckers []HealthChecker
}

func NewHealthRegistry() *HealthRegistry {
	return &HealthRegistry{
		healthCheckers: []HealthChecker{},
	}
}

func (r *HealthRegistry) AddHealthCheckers(checkers ...HealthChecker) {
	r.healthCheckers = append(r.healthCheckers, checkers...)
}

func (r *HealthRegistry) Check() error {
	for _, checker := range r.healthCheckers {
		if err := checker(); err != nil {
			return err
		}
	}
	return nil
}

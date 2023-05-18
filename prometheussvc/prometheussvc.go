package prometheussvc

import "github.com/prometheus/client_golang/prometheus"

func NewPrometheusSvc() prometheus.Registerer {
	return prometheus.DefaultRegisterer
}

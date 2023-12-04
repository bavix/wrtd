package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

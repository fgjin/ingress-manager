package global

import "client/pkg"

const (
	WorkNum  = 5
	MaxRetry = 5
)

var (
	MyConfig         *pkg.IngressYAML
	CustomAnnotation = "ingress/http"
	ConfigPath     string
)

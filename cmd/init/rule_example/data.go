package rule_example

import (
	_ "embed"
)

//go:embed aws.yaml
var Aws string

//go:embed azure.yaml
var Azure string

//go:embed gcp.yaml
var GCP string

//go:embed k8s.yaml
var K8S string

//go:embed default_template.yaml
var DefaultTemplate string

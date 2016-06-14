package data

type CheckResponse struct {
	Metrics  Metrics
	Metadata Metadata
	Response []byte
}

type Metrics map[string]*float64

type Metadata map[string]string

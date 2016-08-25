package data

type CheckResponse struct {
	Metrics  Metrics  `json:"metrics"`
	Metadata Metadata `json:"metadata"`
	Response []byte   `json:"-"`
}

type Metrics map[string]*float64

type Metadata map[string]string

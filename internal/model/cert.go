package model

type CertInfo struct {
	Domain     []string `json:"domain"`
	Issuer     string   `json:"issuer"`
	ValidStart string   `json:"valid_start"`
	ValidStop  string   `json:"valid_stop"`
}

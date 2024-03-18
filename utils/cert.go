package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
)

func ParseCert(data []byte) (*tls.Certificate, error) {
	var (
		cert  = &tls.Certificate{}
		block *pem.Block
	)

	for {
		block, data = pem.Decode(data)
		if block == nil {
			break
		}

		switch block.Type {
		case "CERTIFICATE":
			cert.Certificate = append(cert.Certificate, block.Bytes)
		case "PRIVATE KEY":
			pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			cert.PrivateKey = pk
		case "RSA PRIVATE KEY":
			pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			cert.PrivateKey = pk
		case "EC PRIVATE KEY":
			pk, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			cert.PrivateKey = pk
		}
	}
	return cert, nil
}

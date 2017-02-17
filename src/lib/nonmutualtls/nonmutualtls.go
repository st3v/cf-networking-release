package nonmutualtls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
)

func NewServerTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to load cert or key: %s", err)
	}
	c := &tls.Config{
		Certificates: []tls.Certificate{keyPair},
		MinVersion:   tls.VersionTLS12,
	}
	c.PreferServerCipherSuites = true
	c.CipherSuites = []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256}
	return c, nil
}

func NewClientTLSConfig(caCertFiles ...string) (*tls.Config, error) {
	c := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	var err error
	c.RootCAs, err = newCACertPool(caCertFiles)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func newCACertPool(caCertFiles []string) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()
	for _, certFile := range caCertFiles {
		certBytes, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf("failed read ca cert file: %s", err.Error())
		}

		if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
			return nil, errors.New("Unable to load caCert")
		}
	}
	return caCertPool, nil
}

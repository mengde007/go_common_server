package common

import (
	"net"
	"net/http"
	"time"
)

func CreateTransport() *http.Transport {
	return &http.Transport{
		Dial: func(network, address string) (net.Conn, error) {
			c, err := net.DialTimeout(network, address, time.Second*15)
			if err != nil {
				return nil, err
			}

			c.SetDeadline(time.Now().Add(15 * time.Second))

			return c, nil
		},
		DisableKeepAlives: true,
	}
}

func CreateHttpsTransport() *http.Transport {
	//pool := x509.NewCertPool()
	//caCertPath := "certs/ca.crt"
	//
	//caCrt, err := ioutil.ReadFile(caCertPath)
	//if err != nil {
	//	fmt.Println("ReadFile err:", err)
	//	return nil
	//}
	//pool.AppendCertsFromPEM(caCrt)

	return &http.Transport{
		Dial: func(network, address string) (net.Conn, error) {
			c, err := net.DialTimeout(network, address, time.Second*15)
			if err != nil {
				return nil, err
			}

			c.SetDeadline(time.Now().Add(15 * time.Second))

			return c, nil
		},
		//TLSClientConfig:   &tls.Config{RootCAs: pool},
		DisableKeepAlives: true,
	}
}

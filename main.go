package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"io"
	"net"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

var (
	localType string // tcp tls
	localAddr string

	tlsCert string
	tlsKey  string

	parentType string // tcp tls
	parentAddr string

	debug bool

	localCertRaw, parentCertRaw []byte

	genCert bool
)

func main() {
	flag.StringVar(&localType, "t", "tcp", "local protocol type <tcp|tls>")
	flag.StringVar(&localAddr, "p", "", "local ip:port to listen, such as: 127.0.0.1:1080")
	flag.StringVar(&parentType, "T", "tcp", "parent protocol type <tcp|tls>")
	flag.StringVar(&parentAddr, "P", "", "parent address, such as: '1.2.3.4:21080'")
	flag.StringVar(&tlsCert, "C", "", "cert file for tls")
	flag.StringVar(&tlsKey, "K", "", "key file for tls")
	flag.BoolVar(&debug, "debug", false, "use debug")
	flag.BoolVar(&genCert, "gencert", false, "generate a cert & key for tls")
	flag.Parse()

	if genCert {
		createCert("tcpforward")
		return
	}

	if localAddr == "" || parentAddr == "" {
		flag.Usage()
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	logrus.Infof("Running at %s@%s\n", localAddr, localType)

	var l net.Listener
	var err error

	switch localType {
	case "tls":
		// Load client cert
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			logrus.Errorf("Load cert failed: %v\n", err)
			return
		}
		localCertRaw = cert.Certificate[0]
		// Load CA cert
		// caCert, err := ioutil.ReadFile("ca.crt")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// clientCertPool := x509.NewCertPool()
		// clientCertPool.AppendCertsFromPEM(caCert)

		config := &tls.Config{
			// MaxVersion:   tls.VersionTLS12, // debug
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAnyClientCert,
			// ClientCAs:          clientCertPool,
			InsecureSkipVerify: true,
			VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				logrus.Debug("VerifyPeerCertificate CB:", len(rawCerts))

				// 验证客户端证书
				// fmt.Println(hex.Dump(rawCerts[0]))
				if len(rawCerts) == 0 || !bytes.Equal(rawCerts[0], localCertRaw) {
					return errors.New("tls authentication failed, certificate mismatch")
				}

				return nil
			},
		}
		l, err = tls.Listen("tcp", localAddr, config)
		if err != nil {
			logrus.Errorf("Listen failed: %v\n", err)
			return
		}

	case "tcp":
		l, err = net.Listen("tcp", localAddr)
		if err != nil {
			logrus.Errorf("Listen failed: %v\n", err)
			return
		}

	default:
		logrus.Errorln("unsupported local protocol type")
		return
	}

	for {
		client, err := l.Accept()
		if err != nil {
			logrus.Errorf("Accept failed: %v\n", err)
			continue
		}
		go processTcpForward(client)
	}
}

func processTcpForward(client net.Conn) {
	dest, err := proxy.Dial(context.Background(), "tcp", parentAddr)
	if err != nil {
		logrus.Errorf("connect remote error: %s\n", err)
		client.Close()
		return
	}

	if parentType == "tls" {
		if conn, err := processAuth(dest); err != nil {
			logrus.Errorln("auth error:", err)
			client.Close()
			return
		} else {
			dest = conn
		}
	}

	logrus.Infof("pipe %v <-> %v\n", client.RemoteAddr().String(), dest.RemoteAddr().String())
	tcpForward(client, dest)
}

var parentCert tls.Certificate

func processAuth(dest net.Conn) (conn net.Conn, err error) {
	// Load client cert
	if len(parentCertRaw) == 0 {
		parentCert, err = tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return nil, err
		}
		parentCertRaw = parentCert.Certificate[0]
	}

	// Load CA cert
	// caCert, err := ioutil.ReadFile(tlsCert)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// caCertPool := x509.NewCertPool()
	// caCertPool.AppendCertsFromPEM(caCert)

	tlsConn := tls.Client(dest, &tls.Config{
		Certificates: []tls.Certificate{parentCert},
		// RootCAs:            caCertPool,
		ServerName:         "tcpforward",
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			logrus.Debug("VerifyPeerCertificate CB:", len(rawCerts))

			// 验证服务端证书
			if len(rawCerts) == 0 || !bytes.Equal(rawCerts[0], parentCertRaw) {
				return errors.New("tls server cert unmatch")
			}

			return nil
		},
	})

	err = tlsConn.Handshake()
	if err == nil {
		logrus.Debug("server cert:", tlsConn.ConnectionState().PeerCertificates[0].Subject.String())
	}
	return tlsConn, err
}

func tcpForward(client, target net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(client, target)
	go forward(target, client)
}

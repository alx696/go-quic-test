package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/quic-go/quic-go"
)

func run(ctx context.Context) {
	uc, e := net.ListenUDP("udp4", &net.UDPAddr{Port: 10001})
	if e != nil {
		log.Fatalln(e)
	}
	defer uc.Close()
	t := quic.Transport{
		Conn: uc,
	}
	l, e := t.Listen(generateTLSConfig(), &quic.Config{})
	if e != nil {
		log.Fatalln(e)
	}
	defer l.Close()

	for {
		select {
		case <-ctx.Done():
			log.Println("上下文结束,停止循环")
			return
		default:
			c, e := l.Accept(ctx)
			if e != nil {
				log.Println(e)
				continue
			}
			defer c.CloseWithError(0, "done")
			stream, e := c.AcceptStream(ctx)
			if e != nil {
				log.Println(e)
				continue
			}
			defer stream.Close()

			buf := make([]byte, 1024)
			n, e := stream.Read(buf)
			if e != nil {
				log.Println(e)
				continue
			}

			_, e = stream.Write([]byte(string(buf[:n])))
			if e != nil {
				log.Println(e)
				continue
			}
		}
	}
}

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())

	go run(ctx)

	// 等待关闭信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	stopSignal := <-signalChan
	log.Println("收到关闭信号", stopSignal)

	ctxCancel()
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-example"},
	}
}

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

func run(tag string, ctx context.Context, waitGroup *sync.WaitGroup) error {
	defer waitGroup.Done()

	c, e := quic.DialAddr(ctx, "172.17.0.1:10001", &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"quic-example"}}, nil)
	if e != nil {
		return e
	}
	defer c.CloseWithError(0, "done")
	stream, e := c.OpenStreamSync(ctx)
	if e != nil {
		return e
	}
	defer stream.Close()

	_, e = stream.Write([]byte(tag))
	if e != nil {
		return e
	}

	buf := make([]byte, 1024)
	n, e := stream.Read(buf)
	if e != nil {
		return e
	}
	log.Println(string(buf[:n]))

	return nil
}

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	waitGroup := sync.WaitGroup{}

	timeBegin := time.Now()
	count := 0
	for count < 1200 {
		count++
		waitGroup.Add(1)
		go run(fmt.Sprint(count), ctx, &waitGroup)
	}

	waitGroup.Wait()
	ctxCancel()

	log.Println("耗时", time.Now().UnixMilli()-timeBegin.UnixMilli())
}

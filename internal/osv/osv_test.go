package osv

import (
	"fmt"
	"sync"
	"testing"
)

func TestOSV_Init(t *testing.T) {
	n := 4
	tNum := 1
	osv := make([]*Node, n)
	for i := 0; i < n; i++ {
		osv[i] = NewOSV(n, tNum, i)
		go func(i int) {
			osv[i].Init()
		}(i)
	}
	var wait sync.WaitGroup
	wait.Add(n * (n - 1))
	for i := 0; i < n; i++ {
		go func(i int) {
			for msg := range osv[i].SendChan {
				fmt.Printf("message from %v to %v %v\n", msg.From(), msg.Dest(), msg.Type())
				wait.Done()
			}
		}(i)
	}
	wait.Wait()
}

func routeLocal(osv []*Node) {
	for i := range osv {
		go func(i int) {
			for {
				msg := <-osv[i].SendChan
				osv[msg.DestID].ReceiveChan <- msg
			}
		}(i)
	}
}

func TestOSV_Recv(t *testing.T) {
	n := 4
	tNum := 1
	osv := make([]*Node, n)
	for i := 0; i < n; i++ {
		osv[i] = NewOSV(n, tNum, i)
	}
	go func() {
		for i := 0; i < n-1; i++ {
			go func(i int) {
				osv[i].Init()
			}(i)
		}
	}()
	go routeLocal(osv)
	go func() {
		for i := 0; i < n; i++ {
			go func(i int) {
				osv[i].Run()
			}(i)
		}
	}()
	var wait sync.WaitGroup
	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			<-osv[i].OutPut
			wait.Done()
		}(i)
	}
	wait.Wait()
}

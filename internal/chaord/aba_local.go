package chaord

import (
	"Chaord/internal/osv"
	"log"
	"sync"
)

/*
func ABA1(nodeNum, degree int, nodes []*Node) {
	osvs := make([]*osv.Node, nodeNum)
	msgs := make([][]osv.Message, nodeNum)
	for i := 0; i < nodeNum; i++ {
		nodes[i].osv1 = osv.NewOSV(nodeNum, degree, i)
		msgs[i] = nodes[i].osv1.InitSync()
		osvs[i] = nodes[i].osv1
	}
	log.Printf("ABA1 start")
	fmt.Println(msgs)
	log.Printf("ABA_1 finish")
}
*/

func ABA1(nodeNum, degree int, nodes []*Node) {
	osvs := make([]*osv.Node, nodeNum)
	for i := 0; i < nodeNum; i++ {
		nodes[i].osv1 = osv.NewOSV(nodeNum, degree, i)
		nodes[i].osv1.Init()
		osvs[i] = nodes[i].osv1
	}
	go osv.RouteLocal(osvs)
	log.Printf("ABA1 start")
	go func() {
		for i := 0; i < nodeNum; i++ {
			go func(i int) {
				osvs[i].Run()
			}(i)
		}
	}()
	var wait sync.WaitGroup
	wait.Add(nodeNum)
	for i := 0; i < nodeNum; i++ {
		go func(i int) {
			<-osvs[i].OutPut
			wait.Done()
		}(i)
	}
	wait.Wait()
	log.Printf("ABA1 finish")
}

func ABAtx(nodeNum, degree int, nodes []*Node) {
	osvs := make([]*osv.Node, nodeNum)
	for i := 0; i < nodeNum; i++ {
		nodes[i].osvTX = osv.NewOSV(nodeNum, degree, i)
		nodes[i].osvTX.Init()
		osvs[i] = nodes[i].osvTX
	}
	go osv.RouteLocal(osvs)
	log.Printf("ABA(tx) start")
	go func() {
		for i := 0; i < nodeNum; i++ {
			go func(i int) {
				osvs[i].Run()
			}(i)
		}
	}()
	var wait sync.WaitGroup
	wait.Add(nodeNum)
	for i := 0; i < nodeNum; i++ {
		go func(i int) {
			<-osvs[i].OutPut
			wait.Done()
		}(i)
	}
	wait.Wait()
	log.Printf("ABA(tx) finish")
}

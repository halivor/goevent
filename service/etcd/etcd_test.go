package etcd

import (
	"fmt"
	"testing"
)

func TestEtcd(*testing.T) {
	New([]string{"127.0.0.1:2379"})
	v := Get("foo1")
	fmt.Println("get foo1 =>", string(v))
	/*ch :=*/ Watch("foo1")
	/*go func() {*/
	//for {
	//select {
	//case data := <-ch:
	//fmt.Println("watch foo1 =>", string(data))
	//}
	//}
	/*}()*/
	v2 := Get("foo2")
	fmt.Println("get foo2 =>", string(v2))
	ch2 := Watch("foo2")
	for {
		select {
		case data := <-ch2:
			fmt.Println("watch foo2 =>", string(data))
		}
	}
}

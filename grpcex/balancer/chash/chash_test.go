package chash

import (
	"fmt"
	"testing"
)

func TestAddNode(t *testing.T) {
	p := &chashPicker{}
	p.AddNode(&chashNode{
		hash: 100,
	})
	p.AddNode(&chashNode{
		hash: 200,
	})
	p.AddNode(&chashNode{
		hash: 120,
	})
	p.AddNode(&chashNode{
		hash: 1,
	})
	p.AddNode(&chashNode{
		hash: 300,
	})
	p.AddNode(&chashNode{
		hash: 150,
	})
	if len(p.nodes) != 6 {
		t.Fail()
	}
	if p.nodes[0].hash != 1 {
		t.Fail()
	}
	if p.nodes[5].hash != 300 {
		t.Fail()
	}
	for i := 0; i < len(p.nodes)-1; i++ {
		if p.nodes[i].hash > p.nodes[i+1].hash {
			t.Fail()
		}
	}
}

func TestChash(t *testing.T) {
	addrs := []string{"192.168.2.12:8300", "192.168.2.11:8300"}
	m := uint32(4294967295) / 2
	n := 0
	for _, addr := range addrs {
		for i := 0; i < 4; i++ {
			key := fmt.Sprintf("%d/%s", i, addr)
			hash := bkdrHash32([]byte(key))
			fmt.Printf("key=%s, hash=%d\n", key, hash)
			if hash <= m {
				n++
			}
		}
	}
	fmt.Printf("n=%d\n", n)
	if n != 4 {
		t.Fail()
	}
}

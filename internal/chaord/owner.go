package chaord

import "math/rand"

type Owner struct {
	scale    int
	flagChan chan []bool
}

func (owner *Owner) Init(scale int) {
	owner.scale = scale
}

func (owner *Owner) step1() {
	vec_b := make([]int, owner.scale)
	for i := 0; i < owner.scale; i++ {
		vec_b[i] = rand.Int()
	}
}

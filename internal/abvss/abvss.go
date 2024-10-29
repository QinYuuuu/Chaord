package abvss

import (
	"errors"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	//"go.dedis.ch/kyber/v4/pairing/bls12381/circl"
	"math/big"
	"math/rand"
	"sync"

	"Chaord/pkg/utils/polynomial"
)

type ABVSS struct {
	instanceid int
	nodeid     int
	degree     int
	nodenum    int
	p          *big.Int
	batchsize  int
	vnum       int
	Curve      kyber.Group
	mutex      *sync.Mutex
	*ABVSSD
	*ABVSSR
	*ABVSSV
}

func (vss *ABVSS) GetNodeID() int {
	return vss.nodeid
}
func (vss *ABVSS) GetInstanceID() int {
	return vss.instanceid
}

type ABVSSD struct {
	pk     []kyber.Point
	secret []*big.Int
	polyf  []polynomial.Polynomial
	polyg  []polynomial.Polynomial
	//shares [][]*big.Int
}

type ABVSSR struct {
	sk kyber.Scalar

	zix          [][]kyber.Point
	ziy          [][]kyber.Point
	xix          [][]kyber.Point
	xiy          [][]kyber.Point
	fshares      []*big.Int
	gshares      []*big.Int
	randombeacon *rand.Rand
	r            [][]*big.Int
	Received     chan bool
	complain     bool
	qlist        map[int][]*big.Int
}

type ABVSSV struct {
	Count int
	ilist []struct {
		index int
		lcm   []*big.Int
	}
	jlist []int
	done  bool
	Ready chan bool
}

func NewVSS(index, nodeid, nodenum, degree, batchsize, vnum int, p *big.Int, flag int, mutex *sync.Mutex) (*ABVSS, error) {
	if nodenum < 3*degree+1 {
		return nil, errors.New("must satisfy n >= 3f+1")
	}
	if batchsize <= 0 || vnum <= 0 {
		return nil, errors.New("batchsize/vnum must >= 1")
	}
	var curve kyber.Group
	if flag == 1 {
		curve = edwards25519.NewBlakeSHA256Ed25519()
	} /*else if flag == 2 {
		curve = circl.NewSuiteBLS12381()
	}*/
	return &ABVSS{
		instanceid: index,
		nodeid:     nodeid,
		degree:     degree,
		nodenum:    nodenum,
		p:          p,
		batchsize:  batchsize,
		vnum:       vnum,
		Curve:      curve,
		mutex:      mutex,
	}, nil
}

func (vss *ABVSS) GetN() int {
	return vss.nodenum
}

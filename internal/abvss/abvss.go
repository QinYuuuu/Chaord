package abvss

import (
	"Chaord/pkg/utils/polynomial"
	"errors"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	//"go.dedis.ch/kyber/v4/pairing/bls12381/circl"
	"math/big"
	"math/rand"
)

type ABVSS struct {
	instanceid int
	nodeid     int
	degree     int
	nodenum    int
	p          *big.Int
	batchsize  int
	vnum       int
	curve      kyber.Group
	*Distributor
	*Receiver
	*Verifier
}

func (vss *ABVSS) GetNodeID() int {
	return vss.nodeid
}
func (vss *ABVSS) GetInstanceID() int {
	return vss.instanceid
}

type Distributor struct {
	pk     []kyber.Point
	secret []*big.Int
	polyf  []polynomial.Polynomial
	polyg  []polynomial.Polynomial
	//shares [][]*big.Int
}

type Receiver struct {
	sk           kyber.Scalar
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

type Verifier struct {
	count   int
	shareCh chan struct {
		index int
		lcm   []*big.Int
	}
	ilist []struct {
		index int
		lcm   []*big.Int
	}
	jlist []int
	done  bool
	Ready chan bool
}

func NewVSS(index, nodeid, nodenum, degree, batchsize, vnum int, p *big.Int, flag int) (*ABVSS, error) {
	if nodenum < 3*degree+1 {
		return nil, errors.New("n must satisfy n >= 3f+1")
	}
	if batchsize <= 0 || vnum <= 0 {
		return nil, errors.New("batchsize and vnum must >= 1")
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
		curve:      curve,
	}, nil
}

func (vss *ABVSS) GetN() int {
	return vss.nodenum
}

package abvss

import (
	"Chaord/internal/osv"
	"Chaord/pkg/protobuf"
	"Chaord/pkg/utils/polynomial"
	"errors"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	//"go.dedis.ch/kyber/v4/pairing/bls12381/circl"
	"math/big"
	"math/rand"
)

type ABVSS struct {
	instanceID int
	nodeID     int
	degree     int
	nodeNum    int
	p          *big.Int
	batchSize  int
	vNum       int
	curve      kyber.Group
	*Distributor
	*Receiver
	*Verifier
	*Service

	osv *osv.Node
}

func (vss *ABVSS) GetNodeID() int {
	return vss.nodeID
}
func (vss *ABVSS) GetInstanceID() int {
	return vss.instanceID
}

type Distributor struct {
	pk     []kyber.Point
	secret []*big.Int
	polyf  []*polynomial.Polynomial
	polyg  []*polynomial.Polynomial
	//shares [][]*big.Int
}

type Receiver struct {
	sk           kyber.Scalar
	zix          [][]kyber.Point
	ziy          [][]kyber.Point
	xix          [][]kyber.Point
	xiy          [][]kyber.Point
	fShares      []*big.Int
	gShares      []*big.Int
	randomBeacon *rand.Rand
	r            [][]*big.Int
	Received     chan bool
	complain     bool
	qlist        map[int][]*big.Int
}

type Verifier struct {
	count   int
	shareCh chan LcmTuple
	ilist   []LcmTuple
	jlist   []int
	done    bool
	Ready   chan bool
}

type Service struct {
	DealerBandwidthUsage int
	BandwidthUsage       int
	ReceiveChan          chan *protobuf.Message
	SendChan             []chan *protobuf.Message
}

func NewVSS(index, nodeid, nodenum, degree, batchsize, vnum int, p *big.Int, flag int) (*ABVSS, error) {
	if nodenum < 3*degree+1 {
		return nil, errors.New("n must satisfy n >= 3f+1")
	}
	if batchsize <= 0 || vnum <= 0 {
		return nil, errors.New("batchSize and vNum must >= 1")
	}
	var curve kyber.Group
	if flag == 1 {
		curve = edwards25519.NewBlakeSHA256Ed25519()
	} /*else if flag == 2 {
		curve = circl.NewSuiteBLS12381()
	}*/
	return &ABVSS{
		instanceID: index,
		nodeID:     nodeid,
		degree:     degree,
		nodeNum:    nodenum,
		p:          p,
		batchSize:  batchsize,
		vNum:       vnum,
		curve:      curve,
	}, nil
}

func (vss *ABVSS) GetN() int {
	return vss.nodeNum
}

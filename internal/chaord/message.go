package chaord

// not used in local test

type FlagMsg struct {
	FromID int
	Flag   []bool
}

type BFMsg struct {
	B [][]byte
}

type CHatMsg struct {
	FromID int
	CHat   [][]byte
	BHat   [][]byte
}

type Request struct {
	ConsumerID int
	OwnerID    int
}

type BHatPHatMsg struct {
	BHat [][]byte
	PHat [][]byte
}

package merkle

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"testing"
)

func TestNewMerkleTree(t *testing.T) {
	dataList := [][]byte{[]byte("alice"), []byte("bob"), []byte("cindy"), []byte("david"), []byte("elisa")}
	var roothash []byte
	hasher := md5.New()
	tmp1 := append(hasher.Sum([]byte("alice")), hasher.Sum([]byte("bob"))...)
	tmp2 := append(hasher.Sum([]byte("cindy")), hasher.Sum([]byte("david"))...)
	tmp3 := append(hasher.Sum([]byte("elisa")), hasher.Sum([]byte("elisa"))...)
	tmp4 := append(hasher.Sum(tmp1), hasher.Sum(tmp2)...)
	tmp5 := append(hasher.Sum(tmp3), hasher.Sum(tmp3)...)
	tmp6 := append(hasher.Sum(tmp4), hasher.Sum(tmp5)...)
	roothash = hasher.Sum(tmp6)
	m, _ := NewMerkleTree(dataList, hasher.Sum)
	t.Run("Verify the inorder of merkle tree", func(t *testing.T) {
		got := m.Root().Hash()
		fmt.Println(got)
		if !reflect.DeepEqual(got, roothash) {
			t.Errorf("got %v want %v", got, roothash)
		}
	})
}

func TestCommitment(t *testing.T) {
	hasher := md5.New()
	dataList := [][]byte{[]byte("alice"), []byte("bob"), []byte("cindy"), []byte("david"), []byte("elisa")}
	m, _ := NewMerkleTree(dataList, hasher.Sum)
	//tmp1 := append(hasher.Sum([]byte("alice")), hasher.Sum([]byte("bob"))...)
	tmp2 := append(hasher.Sum([]byte("cindy")), hasher.Sum([]byte("david"))...)
	tmp3 := append(hasher.Sum([]byte("elisa")), hasher.Sum([]byte("elisa"))...)
	//tmp4 := append(hasher.Sum(tmp1), hasher.Sum(tmp2)...)
	tmp5 := append(hasher.Sum(tmp3), hasher.Sum(tmp3)...)
	//tmp6 := append(hasher.Sum(tmp4), hasher.Sum(tmp5)...)
	hashlist := [][]byte{hasher.Sum([]byte("alice")), hasher.Sum(tmp2), hasher.Sum(tmp5)}
	poslist := []bool{true, false, false}
	want := Witness{}
	want.SetHash(hashlist)
	want.SetPos(poslist)
	t.Run("Create the witness", func(t *testing.T) {
		got, _ := CreateWitness(m, 1)
		fmt.Println(want.Hash())
		fmt.Println(got.Hash())
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})
	t.Run("Verify the witness", func(t *testing.T) {
		comm := Commit(m)
		w, _ := CreateWitness(m, 1)
		result, _ := Verify(comm, w, []byte("bob"), hasher.Sum)
		if !result {
			t.Errorf("verify failed")
		}
	})
}

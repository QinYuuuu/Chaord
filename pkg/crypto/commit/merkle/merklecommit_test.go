package merkle

import (
	"crypto/hasher"
	"fmt"
	"reflect"
	"testing"
)

func TestNewMerkleTree(t *testing.T) {
	dataList := [][]byte{[]byte("alice"), []byte("bob"), []byte("cindy"), []byte("david"), []byte("elisa")}
	var roothash []byte
	tmp1 := append(hasher.SHA256Hasher([]byte("alice")), hasher.SHA256Hasher([]byte("bob"))...)
	tmp2 := append(hasher.SHA256Hasher([]byte("cindy")), hasher.SHA256Hasher([]byte("david"))...)
	tmp3 := append(hasher.SHA256Hasher([]byte("elisa")), hasher.SHA256Hasher([]byte("elisa"))...)
	tmp4 := append(hasher.SHA256Hasher(tmp1), hasher.SHA256Hasher(tmp2)...)
	tmp5 := append(hasher.SHA256Hasher(tmp3), hasher.SHA256Hasher(tmp3)...)
	tmp6 := append(hasher.SHA256Hasher(tmp4), hasher.SHA256Hasher(tmp5)...)
	roothash = hasher.SHA256Hasher(tmp6)
	m, _ := NewMerkleTree(dataList, hasher.SHA256Hasher)
	t.Run("Verify the inorder of merkle tree", func(t *testing.T) {
		got := m.Root().Hash()
		if !reflect.DeepEqual(got, roothash) {
			t.Errorf("got %v want %v", got, roothash)
		}
	})
}

func TestCommitment(t *testing.T) {
	dataList := [][]byte{[]byte("alice"), []byte("bob"), []byte("cindy"), []byte("david"), []byte("elisa")}
	m, _ := NewMerkleTree(dataList, hasher.SHA256Hasher)
	//tmp1 := append(hasher.SHA256Hasher([]byte("alice")), hasher.SHA256Hasher([]byte("bob"))...)
	tmp2 := append(hasher.SHA256Hasher([]byte("cindy")), hasher.SHA256Hasher([]byte("david"))...)
	tmp3 := append(hasher.SHA256Hasher([]byte("elisa")), hasher.SHA256Hasher([]byte("elisa"))...)
	//tmp4 := append(hasher.SHA256Hasher(tmp1), hasher.SHA256Hasher(tmp2)...)
	tmp5 := append(hasher.SHA256Hasher(tmp3), hasher.SHA256Hasher(tmp3)...)
	//tmp6 := append(hasher.SHA256Hasher(tmp4), hasher.SHA256Hasher(tmp5)...)
	hashlist := [][]byte{hasher.SHA256Hasher([]byte("alice")), hasher.SHA256Hasher(tmp2), hasher.SHA256Hasher(tmp5)}
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
		result, _ := Verify(comm, w, []byte("bob"), hasher.SHA256Hasher)
		if !result {
			t.Errorf("verify failed")
		}
	})
}

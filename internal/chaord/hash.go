package chaord

import "crypto/MD5"

func MD5hasher(input []byte) []byte {
	hasher := md5.New()
	hasher.Write(input)
	hash := hasher.Sum(nil)
	return hash
}

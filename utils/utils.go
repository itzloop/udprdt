package utils

import "encoding/binary"

func Uint32ToBinary(n uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, n)
	return buf
}

func BinaryToUint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

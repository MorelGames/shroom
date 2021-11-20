package main

import (
	"encoding/binary"
	"hash/crc32"
)

func Rg(seed []byte, counter uint32) int {
	bcounter := make([]byte, 4)
	binary.LittleEndian.PutUint32(bcounter, counter)
	return int(crc32.ChecksumIEEE(append(seed, bcounter...)))
}

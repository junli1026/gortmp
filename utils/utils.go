package utils

import (
	"encoding/binary"
	"math"
)

// ReadUint32 read byte array as uint32
func ReadUint32(data []byte) uint32 {
	if len(data) == 1 {
		return uint32(data[0])
	}
	if len(data) == 2 {
		return uint32(data[0])<<8 | uint32(data[1])
	}
	if len(data) == 3 {
		return uint32(data[0])<<16 |
			uint32(data[1])<<8 |
			uint32(data[2])
	}
	return binary.BigEndian.Uint32(data)
}

// ReadFloat64 read byte array as float64
func ReadFloat64(data []byte) float64 {
	bits := binary.BigEndian.Uint64(data)
	float := math.Float64frombits(bits)
	return float
}

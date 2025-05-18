/* sha3_256.go */
package crypto

// Implementation of SHA3-256 from scratch, following FIPS 202
// Only for educational purposes.

import (
	"encoding/binary"
)

// Rotation offsets for Keccak-f[1600]
var rotOffsets = [5][5]uint{
	{0, 36, 3, 41, 18},
	{1, 44, 10, 45, 2},
	{62, 6, 43, 15, 61},
	{28, 55, 25, 21, 56},
	{27, 20, 39, 8, 14},
}

// Round constants
var roundConstants = [24]uint64{
	0x0000000000000001, 0x0000000000008082,
	0x800000000000808A, 0x8000000080008000,
	0x000000000000808B, 0x0000000080000001,
	0x8000000080008081, 0x8000000000008009,
	0x000000000000008A, 0x0000000000000088,
	0x0000000080008009, 0x000000008000000A,
	0x000000008000808B, 0x800000000000008B,
	0x8000000000008089, 0x8000000000008003,
	0x8000000000008002, 0x8000000000000080,
	0x000000000000800A, 0x800000008000000A,
	0x8000000080008081, 0x8000000000008080,
	0x0000000080000001, 0x8000000080008008,
}

const (
	// SHA3-256 parameters
	rate      = 1088 / 8       // bytes
	capacity  = 512 / 8        // bytes
	outputLen = 256 / 8        // bytes
)

// Keccak-f[1600] permutation
func keccakF(state *[25]uint64) {
	for round := 0; round < 24; round++ {
		// Theta
		var C [5]uint64
		for x := 0; x < 5; x++ {
			C[x] = state[x] ^ state[x+5] ^ state[x+10] ^ state[x+15] ^ state[x+20]
		}
		var D [5]uint64
		for x := 0; x < 5; x++ {
			D[x] = C[(x+4)%5] ^ (C[(x+1)%5] << 1 | C[(x+1)%5] >> (64-1))
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				state[x+5*y] ^= D[x]
			}
		}
		// Rho and Pi
		var B [25]uint64
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				B[y*5+((2*x+3*y)%5)] = state[x+5*y] << rotOffsets[x][y] | state[x+5*y] >> (64-rotOffsets[x][y])
			}
		}
		// Chi
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				state[x+5*y] = B[x+5*y] ^ ((^B[((x+1)%5)+5*y]) & B[((x+2)%5)+5*y])
			}
		}
		// Iota
		state[0] ^= roundConstants[round]
	}
}

// pad10*1 padding: append 0x06 then zeros then final bit 0x80
func padMessage(msg []byte) []byte {
	msgLen := len(msg)
	oCap := rate
	padLen := ((-msgLen - 2) % oCap + oCap) % oCap
	padded := make([]byte, msgLen+2+padLen)
	copy(padded, msg)
	padded[msgLen] = 0x06
	padded[len(padded)-1] |= 0x80
	return padded
}

// Sha3_256 calculates SHA3-256 digest
func Sha3_256(msg []byte) []byte {
	// Initialize state
	var state [25]uint64
	// Absorb
	padded := padMessage(msg)
	for i := 0; i < len(padded); i += rate {
		block := padded[i : i+rate]
		for j := 0; j < rate/8; j++ {
			state[j] ^= binary.LittleEndian.Uint64(block[j*8:])
		}
		keccakF(&state)
	}
	// Squeeze
	hash := make([]byte, outputLen)
	for i := 0; i < outputLen/8; i++ {
		binary.LittleEndian.PutUint64(hash[i*8:], state[i])
	}
	return hash
}

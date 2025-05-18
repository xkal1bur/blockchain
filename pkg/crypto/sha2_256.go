package crypto

import (
	"encoding/binary"
	"math"
)

// ----------------------------------------------------------------------------
// Bitwise functions

func rotr(x uint32, n uint) uint32 {
	return (x >> n) | (x << (32 - n))
}

func shr(x uint32, n uint) uint32 {
	return x >> n
}

func sig0(x uint32) uint32 {
	return rotr(x, 7) ^ rotr(x, 18) ^ shr(x, 3)
}

func sig1(x uint32) uint32 {
	return rotr(x, 17) ^ rotr(x, 19) ^ shr(x, 10)
}

func capsig0(x uint32) uint32 {
	return rotr(x, 2) ^ rotr(x, 13) ^ rotr(x, 22)
}

func capsig1(x uint32) uint32 {
	return rotr(x, 6) ^ rotr(x, 11) ^ rotr(x, 25)
}

func ch(x, y, z uint32) uint32 {
	return (x & y) ^ (^x & z)
}

func maj(x, y, z uint32) uint32 {
	return (x & y) ^ (x & z) ^ (y & z)
}

// ----------------------------------------------------------------------------
// Constants

func fracBin(f float64) uint32 {
	_, frac := math.Modf(f)
	return uint32(frac * (1 << 32))
}

func genPrimes(n int) []uint32 {
	var primes []uint32
	for i := 2; len(primes) < n; i++ {
		prime := true
		for j := 2; j*j <= i; j++ {
			if i%j == 0 {
				prime = false
				break
			}
		}
		if prime {
			primes = append(primes, uint32(i))
		}
	}
	return primes
}

func genK() []uint32 {
	K := make([]uint32, 64)
	primes := genPrimes(64)
	for i := 0; i < 64; i++ {
		K[i] = fracBin(math.Cbrt(float64(primes[i])))
	}
	return K
}

func genH() []uint32 {
	H := make([]uint32, 8)
	primes := genPrimes(8)
	for i := 0; i < 8; i++ {
		H[i] = fracBin(math.Sqrt(float64(primes[i])))
	}
	return H
}

// ----------------------------------------------------------------------------
// Padding

func pad(msg []byte) []byte {
	ml := uint64(len(msg) * 8)
	msg = append(msg, 0x80)

	for len(msg)%64 != 56 {
		msg = append(msg, 0x00)
	}

	lenBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lenBytes, ml)
	return append(msg, lenBytes...)
}

// ----------------------------------------------------------------------------
// Main SHA-256 hash function

func Sha256Edu(msg []byte) [32]byte {
	K := genK()
	H := genH()
	msg = pad(msg)

	for i := 0; i < len(msg); i += 64 {
		block := msg[i : i+64]
		var W [64]uint32

		for t := 0; t < 16; t++ {
			W[t] = binary.BigEndian.Uint32(block[t*4 : t*4+4])
		}
		for t := 16; t < 64; t++ {
			W[t] = sig1(W[t-2]) + W[t-7] + sig0(W[t-15]) + W[t-16]
		}

		a, b, c, d, e, f, g, h := H[0], H[1], H[2], H[3], H[4], H[5], H[6], H[7]

		for t := 0; t < 64; t++ {
			T1 := h + capsig1(e) + ch(e, f, g) + K[t] + W[t]
			T2 := capsig0(a) + maj(a, b, c)
			h = g
			g = f
			f = e
			e = d + T1
			d = c
			c = b
			b = a
			a = T1 + T2
		}

		H[0] += a
		H[1] += b
		H[2] += c
		H[3] += d
		H[4] += e
		H[5] += f
		H[6] += g
		H[7] += h
	}

	var digest [32]byte
	for i, h := range H {
		binary.BigEndian.PutUint32(digest[i*4:], h)
	}
	return digest
}

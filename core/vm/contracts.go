// Copyright 2014 by the Authors
// This file is part of the go-core library.
//
// The go-core library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-core library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-core library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/core-coin/go-goldilocks"
	"golang.org/x/crypto/sha3"
	"math/big"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/math"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/crypto/blake2b"
	"github.com/core-coin/go-core/crypto/bn256"
	"github.com/core-coin/go-core/params"

	//lint:ignore SA1019 Needed for precompile
	"golang.org/x/crypto/ripemd160"
)

// PrecompiledContract is the basic interface for native Go contracts. The implementation
// requires a deterministic energy count based on the input size of the Run method of the
// contract.
type PrecompiledContract interface {
	RequiredEnergy(input []byte) uint64 // RequiredPrice calculates the contract energy use
	Run(input []byte) ([]byte, error)   // Run runs the precompiled contract
}

var PrecompiledContracts = map[common.Address]PrecompiledContract{
	common.Addr1: &ecrecover{},
	common.Addr2: &sha256hash{},
	common.Addr3: &ripemd160hash{},
	common.Addr4: &dataCopy{},
	common.Addr5: &bigModExp{},
	common.Addr6: &bn256Add{},
	common.Addr7: &bn256ScalarMul{},
	common.Addr8: &bn256Pairing{},
	common.Addr9: &blake2F{},
}

// RunPrecompiledContract runs and evaluates the output of a precompiled contract.
// It returns
// - the returned bytes,
// - the _remaining_ energy,
// - any error that occurred
func RunPrecompiledContract(p PrecompiledContract, input []byte, suppliedEnergy uint64) (ret []byte, remainingEnergy uint64, err error) {
	energyCost := p.RequiredEnergy(input)
	if suppliedEnergy < energyCost {
		return nil, 0, ErrOutOfEnergy
	}
	suppliedEnergy -= energyCost
	output, err := p.Run(input)
	return output, suppliedEnergy, err
}

// ECRECOVER implemented as a native contract.
type ecrecover struct{}

func (c *ecrecover) RequiredEnergy(input []byte) uint64 {
	return params.EcrecoverEnergy
}

func (c *ecrecover) Run(input []byte) ([]byte, error) {
	var ecRecoverInputLength = sha3.New256().Size() + crypto.ExtendedSignatureLength // 32 + 171

	input = common.RightPadBytes(input, ecRecoverInputLength)

	pubKey, err := crypto.Ecrecover(input[:32], input[96:267])
	// make sure the public key is a valid one
	if err != nil {
		return nil, nil
	}
	if pubKey != nil {
		return common.LeftPadBytes(crypto.PubkeyToAddress(goldilocks.BytesToPublicKey(pubKey)).Bytes(), 32), nil
	}
	return nil, errors.New("invalid signature")
}

// SHA256 implemented as a native contract.
type sha256hash struct{}

// RequiredEnergy returns the energy required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size energy costs
// required for anything significant is so high it's impossible to pay for.
func (c *sha256hash) RequiredEnergy(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.Sha256PerWordEnergy + params.Sha256BaseEnergy
}
func (c *sha256hash) Run(input []byte) ([]byte, error) {
	h := sha256.Sum256(input)
	return h[:], nil
}

// RIPEMD160 implemented as a native contract.
type ripemd160hash struct{}

// RequiredEnergy returns the energy required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size energy costs
// required for anything significant is so high it's impossible to pay for.
func (c *ripemd160hash) RequiredEnergy(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.Ripemd160PerWordEnergy + params.Ripemd160BaseEnergy
}
func (c *ripemd160hash) Run(input []byte) ([]byte, error) {
	ripemd := ripemd160.New()
	ripemd.Write(input)
	return common.LeftPadBytes(ripemd.Sum(nil), 32), nil
}

// data copy implemented as a native contract.
type dataCopy struct{}

// RequiredEnergy returns the energy required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size energy costs
// required for anything significant is so high it's impossible to pay for.
func (c *dataCopy) RequiredEnergy(input []byte) uint64 {
	return uint64(len(input)+31)/32*params.IdentityPerWordEnergy + params.IdentityBaseEnergy
}
func (c *dataCopy) Run(in []byte) ([]byte, error) {
	return in, nil
}

// bigModExp implements a native big integer exponential modular operation.
type bigModExp struct{}

var (
	big0      = big.NewInt(0)
	big1      = big.NewInt(1)
	big4      = big.NewInt(4)
	big8      = big.NewInt(8)
	big16     = big.NewInt(16)
	big32     = big.NewInt(32)
	big64     = big.NewInt(64)
	big96     = big.NewInt(96)
	big480    = big.NewInt(480)
	big1024   = big.NewInt(1024)
	big3072   = big.NewInt(3072)
	big199680 = big.NewInt(199680)
)

// RequiredEnergy returns the energy required to execute the pre-compiled contract.
func (c *bigModExp) RequiredEnergy(input []byte) uint64 {
	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32))
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32))
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32))
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Retrieve the head 32 bytes of exp for the adjusted exponent length
	var expHead *big.Int
	if big.NewInt(int64(len(input))).Cmp(baseLen) <= 0 {
		expHead = new(big.Int)
	} else {
		if expLen.Cmp(big32) > 0 {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), 32))
		} else {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), expLen.Uint64()))
		}
	}
	// Calculate the adjusted exponent length
	var msb int
	if bitlen := expHead.BitLen(); bitlen > 0 {
		msb = bitlen - 1
	}
	adjExpLen := new(big.Int)
	if expLen.Cmp(big32) > 0 {
		adjExpLen.Sub(expLen, big32)
		adjExpLen.Mul(big8, adjExpLen)
	}
	adjExpLen.Add(adjExpLen, big.NewInt(int64(msb)))

	// Calculate the energy cost of the operation
	energy := new(big.Int).Set(math.BigMax(modLen, baseLen))
	switch {
	case energy.Cmp(big64) <= 0:
		energy.Mul(energy, energy)
	case energy.Cmp(big1024) <= 0:
		energy = new(big.Int).Add(
			new(big.Int).Div(new(big.Int).Mul(energy, energy), big4),
			new(big.Int).Sub(new(big.Int).Mul(big96, energy), big3072),
		)
	default:
		energy = new(big.Int).Add(
			new(big.Int).Div(new(big.Int).Mul(energy, energy), big16),
			new(big.Int).Sub(new(big.Int).Mul(big480, energy), big199680),
		)
	}
	energy.Mul(energy, math.BigMax(adjExpLen, big1))
	energy.Div(energy, new(big.Int).SetUint64(params.ModExpQuadCoeffDiv))

	if energy.BitLen() > 64 {
		return math.MaxUint64
	}
	return energy.Uint64()
}

func (c *bigModExp) Run(input []byte) ([]byte, error) {
	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32)).Uint64()
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32)).Uint64()
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32)).Uint64()
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Handle a special case when both the base and mod length is zero
	if baseLen == 0 && modLen == 0 {
		return []byte{}, nil
	}
	// Retrieve the operands and execute the exponentiation
	var (
		base = new(big.Int).SetBytes(getData(input, 0, baseLen))
		exp  = new(big.Int).SetBytes(getData(input, baseLen, expLen))
		mod  = new(big.Int).SetBytes(getData(input, baseLen+expLen, modLen))
	)
	if mod.BitLen() == 0 {
		// Modulo 0 is undefined, return zero
		return common.LeftPadBytes([]byte{}, int(modLen)), nil
	}
	return common.LeftPadBytes(base.Exp(base, exp, mod).Bytes(), int(modLen)), nil
}

// newCurvePoint unmarshals a binary blob into a bn256 elliptic curve point,
// returning it, or an error if the point is invalid.
func newCurvePoint(blob []byte) (*bn256.G1, error) {
	p := new(bn256.G1)
	if _, err := p.Unmarshal(blob); err != nil {
		return nil, err
	}
	return p, nil
}

// newTwistPoint unmarshals a binary blob into a bn256 elliptic curve point,
// returning it, or an error if the point is invalid.
func newTwistPoint(blob []byte) (*bn256.G2, error) {
	p := new(bn256.G2)
	if _, err := p.Unmarshal(blob); err != nil {
		return nil, err
	}
	return p, nil
}

// runBn256Add implements the Bn256Add precompile
func runBn256Add(input []byte) ([]byte, error) {
	x, err := newCurvePoint(getData(input, 0, 64))
	if err != nil {
		return nil, err
	}
	y, err := newCurvePoint(getData(input, 64, 64))
	if err != nil {
		return nil, err
	}
	res := new(bn256.G1)
	res.Add(x, y)
	return res.Marshal(), nil
}

// bn256Add implements a native elliptic curve point addition conforming to
// consensus rules.
type bn256Add struct{}

// RequiredEnergy returns the energy required to execute the pre-compiled contract.
func (c *bn256Add) RequiredEnergy(input []byte) uint64 {
	return params.Bn256AddEnergy
}

func (c *bn256Add) Run(input []byte) ([]byte, error) {
	return runBn256Add(input)
}

// runBn256ScalarMul implements the Bn256ScalarMul precompile
func runBn256ScalarMul(input []byte) ([]byte, error) {
	p, err := newCurvePoint(getData(input, 0, 64))
	if err != nil {
		return nil, err
	}
	res := new(bn256.G1)
	res.ScalarMult(p, new(big.Int).SetBytes(getData(input, 64, 32)))
	return res.Marshal(), nil
}

// bn256ScalarMul implements a native elliptic curve scalar
// multiplication conforming to consensus rules.
type bn256ScalarMul struct{}

// RequiredEnergy returns the energy required to execute the pre-compiled contract.
func (c *bn256ScalarMul) RequiredEnergy(input []byte) uint64 {
	return params.Bn256ScalarMulEnergy
}

func (c *bn256ScalarMul) Run(input []byte) ([]byte, error) {
	return runBn256ScalarMul(input)
}

var (
	// true32Byte is returned if the bn256 pairing check succeeds.
	true32Byte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	// false32Byte is returned if the bn256 pairing check fails.
	false32Byte = make([]byte, 32)

	// errBadPairingInput is returned if the bn256 pairing input is invalid.
	errBadPairingInput = errors.New("bad elliptic curve pairing size")
)

// runBn256Pairing implements the Bn256Pairing precompile
func runBn256Pairing(input []byte) ([]byte, error) {
	// Handle some corner cases cheaply
	if len(input)%192 > 0 {
		return nil, errBadPairingInput
	}
	// Convert the input into a set of coordinates
	var (
		cs []*bn256.G1
		ts []*bn256.G2
	)
	for i := 0; i < len(input); i += 192 {
		c, err := newCurvePoint(input[i : i+64])
		if err != nil {
			return nil, err
		}
		t, err := newTwistPoint(input[i+64 : i+192])
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
		ts = append(ts, t)
	}
	// Execute the pairing checks and return the results
	if bn256.PairingCheck(cs, ts) {
		return true32Byte, nil
	}
	return false32Byte, nil
}

// bn256Pairing implements a pairing pre-compile for the bn256 curve
type bn256Pairing struct{}

// RequiredEnergy returns the energy required to execute the pre-compiled contract.
func (c *bn256Pairing) RequiredEnergy(input []byte) uint64 {
	return params.Bn256PairingBaseEnergy + uint64(len(input)/192)*params.Bn256PairingPerPointEnergy
}

func (c *bn256Pairing) Run(input []byte) ([]byte, error) {
	return runBn256Pairing(input)
}

type blake2F struct{}

func (c *blake2F) RequiredEnergy(input []byte) uint64 {
	// If the input is malformed, we can't calculate the energy, return 0 and let the
	// actual call choke and fault.
	if len(input) != blake2FInputLength {
		return 0
	}
	return uint64(binary.BigEndian.Uint32(input[0:4]))
}

const (
	blake2FInputLength        = 213
	blake2FFinalBlockBytes    = byte(1)
	blake2FNonFinalBlockBytes = byte(0)
)

var (
	errBlake2FInvalidInputLength = errors.New("invalid input length")
	errBlake2FInvalidFinalFlag   = errors.New("invalid final flag")
)

func (c *blake2F) Run(input []byte) ([]byte, error) {
	// Make sure the input is valid (correct lenth and final flag)
	if len(input) != blake2FInputLength {
		return nil, errBlake2FInvalidInputLength
	}
	if input[212] != blake2FNonFinalBlockBytes && input[212] != blake2FFinalBlockBytes {
		return nil, errBlake2FInvalidFinalFlag
	}
	// Parse the input into the Blake2b call parameters
	var (
		rounds = binary.BigEndian.Uint32(input[0:4])
		final  = (input[212] == blake2FFinalBlockBytes)

		h [8]uint64
		m [16]uint64
		t [2]uint64
	)
	for i := 0; i < 8; i++ {
		offset := 4 + i*8
		h[i] = binary.LittleEndian.Uint64(input[offset : offset+8])
	}
	for i := 0; i < 16; i++ {
		offset := 68 + i*8
		m[i] = binary.LittleEndian.Uint64(input[offset : offset+8])
	}
	t[0] = binary.LittleEndian.Uint64(input[196:204])
	t[1] = binary.LittleEndian.Uint64(input[204:212])

	// Execute the compression function, extract and return the result
	blake2b.F(&h, m, t, final, rounds)

	output := make([]byte, 64)
	for i := 0; i < 8; i++ {
		offset := i * 8
		binary.LittleEndian.PutUint64(output[offset:offset+8], h[i])
	}
	return output, nil
}

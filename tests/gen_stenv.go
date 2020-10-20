// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package tests

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/core-coin/go-core/v2/common"
	"github.com/core-coin/go-core/v2/common/math"
)

var _ = (*stEnvMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (s stEnv) MarshalJSON() ([]byte, error) {
	type stEnv struct {
		Coinbase    common.UnprefixedAddress `json:"currentCoinbase"   gencodec:"required"`
		Difficulty  *math.HexOrDecimal256    `json:"currentDifficulty" gencodec:"required"`
		EnergyLimit math.HexOrDecimal64      `json:"currentEnergyLimit"   gencodec:"required"`
		Number      math.HexOrDecimal64      `json:"currentNumber"     gencodec:"required"`
		Timestamp   math.HexOrDecimal64      `json:"currentTimestamp"  gencodec:"required"`
	}
	var enc stEnv
	enc.Coinbase = common.UnprefixedAddress(s.Coinbase)
	enc.Difficulty = (*math.HexOrDecimal256)(s.Difficulty)
	enc.EnergyLimit = math.HexOrDecimal64(s.EnergyLimit)
	enc.Number = math.HexOrDecimal64(s.Number)
	enc.Timestamp = math.HexOrDecimal64(s.Timestamp)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (s *stEnv) UnmarshalJSON(input []byte) error {
	type stEnv struct {
		Coinbase    *common.UnprefixedAddress `json:"currentCoinbase"   gencodec:"required"`
		Difficulty  *math.HexOrDecimal256     `json:"currentDifficulty" gencodec:"required"`
		EnergyLimit *math.HexOrDecimal64      `json:"currentEnergyLimit"   gencodec:"required"`
		Number      *math.HexOrDecimal64      `json:"currentNumber"     gencodec:"required"`
		Timestamp   *math.HexOrDecimal64      `json:"currentTimestamp"  gencodec:"required"`
	}
	var dec stEnv
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Coinbase == nil {
		return errors.New("missing required field 'currentCoinbase' for stEnv")
	}
	s.Coinbase = common.Address(*dec.Coinbase)
	if dec.Difficulty == nil {
		return errors.New("missing required field 'currentDifficulty' for stEnv")
	}
	s.Difficulty = (*big.Int)(dec.Difficulty)
	if dec.EnergyLimit == nil {
		return errors.New("missing required field 'currentEnergyLimit' for stEnv")
	}
	s.EnergyLimit = uint64(*dec.EnergyLimit)
	if dec.Number == nil {
		return errors.New("missing required field 'currentNumber' for stEnv")
	}
	s.Number = uint64(*dec.Number)
	if dec.Timestamp == nil {
		return errors.New("missing required field 'currentTimestamp' for stEnv")
	}
	s.Timestamp = uint64(*dec.Timestamp)
	return nil
}

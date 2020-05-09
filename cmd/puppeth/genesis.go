// Copyright 2017 The go-core Authors
// This file is part of go-core.
//
// go-core is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-core is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-core. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	math2 "github.com/core-coin/go-core/common/math"
	"github.com/core-coin/go-core/consensus/cryptore"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/params"
)

// alethGenesisSpec represents the genesis specification format used by the
// C++ Core implementation.
type alethGenesisSpec struct {
	SealEngine string `json:"sealEngine"`
	Params     struct {
		AccountStartNonce          math2.HexOrDecimal64   `json:"accountStartNonce"`
		MaximumExtraDataSize       hexutil.Uint64         `json:"maximumExtraDataSize"`
		HomesteadForkBlock         *hexutil.Big           `json:"homesteadForkBlock,omitempty"`
		DaoHardforkBlock           math2.HexOrDecimal64   `json:"daoHardforkBlock"`
		EIP150ForkBlock            *hexutil.Big           `json:"EIP150ForkBlock,omitempty"`
		EIP158ForkBlock            *hexutil.Big           `json:"EIP158ForkBlock,omitempty"`
		ByzantiumForkBlock         *hexutil.Big           `json:"byzantiumForkBlock,omitempty"`
		ConstantinopleForkBlock    *hexutil.Big           `json:"constantinopleForkBlock,omitempty"`
		ConstantinopleFixForkBlock *hexutil.Big           `json:"constantinopleFixForkBlock,omitempty"`
		IstanbulForkBlock          *hexutil.Big           `json:"istanbulForkBlock,omitempty"`
		MinEnergyLimit                hexutil.Uint64         `json:"minEnergyLimit"`
		MaxEnergyLimit                hexutil.Uint64         `json:"maxEnergyLimit"`
		TieBreakingEnergy             bool                   `json:"tieBreakingEnergy"`
		EnergyLimitBoundDivisor       math2.HexOrDecimal64   `json:"energyLimitBoundDivisor"`
		MinimumDifficulty          *hexutil.Big           `json:"minimumDifficulty"`
		DifficultyBoundDivisor     *math2.HexOrDecimal256 `json:"difficultyBoundDivisor"`
		DurationLimit              *math2.HexOrDecimal256 `json:"durationLimit"`
		BlockReward                *hexutil.Big           `json:"blockReward"`
		NetworkID                  hexutil.Uint64         `json:"networkID"`
		ChainID                    hexutil.Uint64         `json:"chainID"`
		AllowFutureBlocks          bool                   `json:"allowFutureBlocks"`
	} `json:"params"`

	Genesis struct {
		Nonce      types.BlockNonce `json:"nonce"`
		Difficulty *hexutil.Big     `json:"difficulty"`
		MixHash    common.Hash      `json:"mixHash"`
		Author     common.Address   `json:"author"`
		Timestamp  hexutil.Uint64   `json:"timestamp"`
		ParentHash common.Hash      `json:"parentHash"`
		ExtraData  hexutil.Bytes    `json:"extraData"`
		EnergyLimit   hexutil.Uint64   `json:"energyLimit"`
	} `json:"genesis"`

	Accounts map[common.UnprefixedAddress]*alethGenesisSpecAccount `json:"accounts"`
}

// alethGenesisSpecAccount is the prefunded genesis account and/or precompiled
// contract definition.
type alethGenesisSpecAccount struct {
	Balance     *math2.HexOrDecimal256   `json:"balance,omitempty"`
	Nonce       uint64                   `json:"nonce,omitempty"`
	Precompiled *alethGenesisSpecBuiltin `json:"precompiled,omitempty"`
}

// alethGenesisSpecBuiltin is the precompiled contract definition.
type alethGenesisSpecBuiltin struct {
	Name          string                         `json:"name,omitempty"`
	StartingBlock *hexutil.Big                   `json:"startingBlock,omitempty"`
	Linear        *alethGenesisSpecLinearPricing `json:"linear,omitempty"`
}

type alethGenesisSpecLinearPricing struct {
	Base uint64 `json:"base"`
	Word uint64 `json:"word"`
}

// newAlethGenesisSpec converts a go-core genesis block into a Aleth-specific
// chain specification format.
func newAlethGenesisSpec(network string, genesis *core.Genesis) (*alethGenesisSpec, error) {
	// Only cryptore is currently supported between go-core and aleth
	if genesis.Config.Cryptore == nil {
		return nil, errors.New("unsupported consensus engine")
	}
	// Reconstruct the chain spec in Aleth format
	spec := &alethGenesisSpec{
		SealEngine: "Cryptore",
	}
	// Some defaults
	spec.Params.AccountStartNonce = 0
	spec.Params.TieBreakingEnergy = false
	spec.Params.AllowFutureBlocks = false

	// Dao hardfork block is a special one. The fork block is listed as 0 in the
	// config but aleth will sync with ETC clients up until the actual dao hard
	// fork block.
	spec.Params.DaoHardforkBlock = 0

	if num := genesis.Config.HomesteadBlock; num != nil {
		spec.Params.HomesteadForkBlock = (*hexutil.Big)(num)
	}
	if num := genesis.Config.EIP150Block; num != nil {
		spec.Params.EIP150ForkBlock = (*hexutil.Big)(num)
	}
	if num := genesis.Config.EIP158Block; num != nil {
		spec.Params.EIP158ForkBlock = (*hexutil.Big)(num)
	}
	if num := genesis.Config.ByzantiumBlock; num != nil {
		spec.Params.ByzantiumForkBlock = (*hexutil.Big)(num)
	}
	if num := genesis.Config.ConstantinopleBlock; num != nil {
		spec.Params.ConstantinopleForkBlock = (*hexutil.Big)(num)
	}
	if num := genesis.Config.PetersburgBlock; num != nil {
		spec.Params.ConstantinopleFixForkBlock = (*hexutil.Big)(num)
	}
	if num := genesis.Config.IstanbulBlock; num != nil {
		spec.Params.IstanbulForkBlock = (*hexutil.Big)(num)
	}
	spec.Params.NetworkID = (hexutil.Uint64)(genesis.Config.ChainID.Uint64())
	spec.Params.ChainID = (hexutil.Uint64)(genesis.Config.ChainID.Uint64())
	spec.Params.MaximumExtraDataSize = (hexutil.Uint64)(params.MaximumExtraDataSize)
	spec.Params.MinEnergyLimit = (hexutil.Uint64)(params.MinEnergyLimit)
	spec.Params.MaxEnergyLimit = (hexutil.Uint64)(math.MaxInt64)
	spec.Params.MinimumDifficulty = (*hexutil.Big)(params.MinimumDifficulty)
	spec.Params.DifficultyBoundDivisor = (*math2.HexOrDecimal256)(params.DifficultyBoundDivisor)
	spec.Params.EnergyLimitBoundDivisor = (math2.HexOrDecimal64)(params.EnergyLimitBoundDivisor)
	spec.Params.DurationLimit = (*math2.HexOrDecimal256)(params.DurationLimit)
	spec.Params.BlockReward = (*hexutil.Big)(cryptore.FrontierBlockReward)

	spec.Genesis.Nonce = types.EncodeNonce(genesis.Nonce)
	spec.Genesis.MixHash = genesis.Mixhash
	spec.Genesis.Difficulty = (*hexutil.Big)(genesis.Difficulty)
	spec.Genesis.Author = genesis.Coinbase
	spec.Genesis.Timestamp = (hexutil.Uint64)(genesis.Timestamp)
	spec.Genesis.ParentHash = genesis.ParentHash
	spec.Genesis.ExtraData = (hexutil.Bytes)(genesis.ExtraData)
	spec.Genesis.EnergyLimit = (hexutil.Uint64)(genesis.EnergyLimit)

	for address, account := range genesis.Alloc {
		spec.setAccount(address, account)
	}

	spec.setPrecompile(1, &alethGenesisSpecBuiltin{Name: "ecrecover",
		Linear: &alethGenesisSpecLinearPricing{Base: 3000}})
	spec.setPrecompile(2, &alethGenesisSpecBuiltin{Name: "sha256",
		Linear: &alethGenesisSpecLinearPricing{Base: 60, Word: 12}})
	spec.setPrecompile(3, &alethGenesisSpecBuiltin{Name: "ripemd160",
		Linear: &alethGenesisSpecLinearPricing{Base: 600, Word: 120}})
	spec.setPrecompile(4, &alethGenesisSpecBuiltin{Name: "identity",
		Linear: &alethGenesisSpecLinearPricing{Base: 15, Word: 3}})
	if genesis.Config.ByzantiumBlock != nil {
		spec.setPrecompile(5, &alethGenesisSpecBuiltin{Name: "modexp",
			StartingBlock: (*hexutil.Big)(genesis.Config.ByzantiumBlock)})
		spec.setPrecompile(6, &alethGenesisSpecBuiltin{Name: "alt_bn128_G1_add",
			StartingBlock: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Linear:        &alethGenesisSpecLinearPricing{Base: 500}})
		spec.setPrecompile(7, &alethGenesisSpecBuiltin{Name: "alt_bn128_G1_mul",
			StartingBlock: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Linear:        &alethGenesisSpecLinearPricing{Base: 40000}})
		spec.setPrecompile(8, &alethGenesisSpecBuiltin{Name: "alt_bn128_pairing_product",
			StartingBlock: (*hexutil.Big)(genesis.Config.ByzantiumBlock)})
	}
	if genesis.Config.IstanbulBlock != nil {
		if genesis.Config.ByzantiumBlock == nil {
			return nil, errors.New("invalid genesis, istanbul fork is enabled while byzantium is not")
		}
		spec.setPrecompile(6, &alethGenesisSpecBuiltin{
			Name:          "alt_bn128_G1_add",
			StartingBlock: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
		}) // Aleth hardcoded the energy policy
		spec.setPrecompile(7, &alethGenesisSpecBuiltin{
			Name:          "alt_bn128_G1_mul",
			StartingBlock: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
		}) // Aleth hardcoded the energy policy
		spec.setPrecompile(9, &alethGenesisSpecBuiltin{
			Name:          "blake2_compression",
			StartingBlock: (*hexutil.Big)(genesis.Config.IstanbulBlock),
		})
	}
	return spec, nil
}

func (spec *alethGenesisSpec) setPrecompile(address byte, data *alethGenesisSpecBuiltin) {
	if spec.Accounts == nil {
		spec.Accounts = make(map[common.UnprefixedAddress]*alethGenesisSpecAccount)
	}
	addr := common.UnprefixedAddress(common.BytesToAddress([]byte{address}))
	if _, exist := spec.Accounts[addr]; !exist {
		spec.Accounts[addr] = &alethGenesisSpecAccount{}
	}
	spec.Accounts[addr].Precompiled = data
}

func (spec *alethGenesisSpec) setAccount(address common.Address, account core.GenesisAccount) {
	if spec.Accounts == nil {
		spec.Accounts = make(map[common.UnprefixedAddress]*alethGenesisSpecAccount)
	}

	a, exist := spec.Accounts[common.UnprefixedAddress(address)]
	if !exist {
		a = &alethGenesisSpecAccount{}
		spec.Accounts[common.UnprefixedAddress(address)] = a
	}
	a.Balance = (*math2.HexOrDecimal256)(account.Balance)
	a.Nonce = account.Nonce

}

// parityChainSpec is the chain specification format used by Parity.
type parityChainSpec struct {
	Name    string `json:"name"`
	Datadir string `json:"dataDir"`
	Engine  struct {
		Cryptore struct {
			Params struct {
				MinimumDifficulty      *hexutil.Big      `json:"minimumDifficulty"`
				DifficultyBoundDivisor *hexutil.Big      `json:"difficultyBoundDivisor"`
				DurationLimit          *hexutil.Big      `json:"durationLimit"`
				BlockReward            map[string]string `json:"blockReward"`
				DifficultyBombDelays   map[string]string `json:"difficultyBombDelays"`
				HomesteadTransition    hexutil.Uint64    `json:"homesteadTransition"`
				EIP100bTransition      hexutil.Uint64    `json:"eip100bTransition"`
			} `json:"params"`
		} `json:"Cryptore"`
	} `json:"engine"`

	Params struct {
		AccountStartNonce         hexutil.Uint64       `json:"accountStartNonce"`
		MaximumExtraDataSize      hexutil.Uint64       `json:"maximumExtraDataSize"`
		MinEnergyLimit               hexutil.Uint64       `json:"minEnergyLimit"`
		EnergyLimitBoundDivisor      math2.HexOrDecimal64 `json:"energyLimitBoundDivisor"`
		NetworkID                 hexutil.Uint64       `json:"networkID"`
		ChainID                   hexutil.Uint64       `json:"chainID"`
		MaxCodeSize               hexutil.Uint64       `json:"maxCodeSize"`
		MaxCodeSizeTransition     hexutil.Uint64       `json:"maxCodeSizeTransition"`
		EIP98Transition           hexutil.Uint64       `json:"eip98Transition"`
		EIP150Transition          hexutil.Uint64       `json:"eip150Transition"`
		EIP160Transition          hexutil.Uint64       `json:"eip160Transition"`
		EIP161abcTransition       hexutil.Uint64       `json:"eip161abcTransition"`
		EIP161dTransition         hexutil.Uint64       `json:"eip161dTransition"`
		EIP155Transition          hexutil.Uint64       `json:"eip155Transition"`
		EIP140Transition          hexutil.Uint64       `json:"eip140Transition"`
		EIP211Transition          hexutil.Uint64       `json:"eip211Transition"`
		EIP214Transition          hexutil.Uint64       `json:"eip214Transition"`
		EIP658Transition          hexutil.Uint64       `json:"eip658Transition"`
		EIP145Transition          hexutil.Uint64       `json:"eip145Transition"`
		EIP1014Transition         hexutil.Uint64       `json:"eip1014Transition"`
		EIP1052Transition         hexutil.Uint64       `json:"eip1052Transition"`
		EIP1283Transition         hexutil.Uint64       `json:"eip1283Transition"`
		EIP1283DisableTransition  hexutil.Uint64       `json:"eip1283DisableTransition"`
		EIP1283ReenableTransition hexutil.Uint64       `json:"eip1283ReenableTransition"`
		EIP1344Transition         hexutil.Uint64       `json:"eip1344Transition"`
		EIP1884Transition         hexutil.Uint64       `json:"eip1884Transition"`
		EIP2028Transition         hexutil.Uint64       `json:"eip2028Transition"`
	} `json:"params"`

	Genesis struct {
		Seal struct {
			Core struct {
				Nonce   types.BlockNonce `json:"nonce"`
				MixHash hexutil.Bytes    `json:"mixHash"`
			} `json:"core"`
		} `json:"seal"`

		Difficulty *hexutil.Big   `json:"difficulty"`
		Author     common.Address `json:"author"`
		Timestamp  hexutil.Uint64 `json:"timestamp"`
		ParentHash common.Hash    `json:"parentHash"`
		ExtraData  hexutil.Bytes  `json:"extraData"`
		EnergyLimit   hexutil.Uint64 `json:"energyLimit"`
	} `json:"genesis"`

	Nodes    []string                                             `json:"nodes"`
	Accounts map[common.UnprefixedAddress]*parityChainSpecAccount `json:"accounts"`
}

// parityChainSpecAccount is the prefunded genesis account and/or precompiled
// contract definition.
type parityChainSpecAccount struct {
	Balance math2.HexOrDecimal256   `json:"balance"`
	Nonce   math2.HexOrDecimal64    `json:"nonce,omitempty"`
	Builtin *parityChainSpecBuiltin `json:"builtin,omitempty"`
}

// parityChainSpecBuiltin is the precompiled contract definition.
type parityChainSpecBuiltin struct {
	Name       string       `json:"name"`                  // Each builtin should has it own name
	Pricing    interface{}  `json:"pricing"`               // Each builtin should has it own price strategy
	ActivateAt *hexutil.Big `json:"activate_at,omitempty"` // ActivateAt can't be omitted if empty, default means no fork
}

// parityChainSpecPricing represents the different pricing models that builtin
// contracts might advertise using.
type parityChainSpecPricing struct {
	Linear *parityChainSpecLinearPricing `json:"linear,omitempty"`
	ModExp *parityChainSpecModExpPricing `json:"modexp,omitempty"`

	// Before the https://github.com/paritytech/parity-core/pull/11039,
	// Parity uses this format to config bn pairing price policy.
	AltBnPairing *parityChainSepcAltBnPairingPricing `json:"alt_bn128_pairing,omitempty"`

	// Blake2F is the price per round of Blake2 compression
	Blake2F *parityChainSpecBlakePricing `json:"blake2_f,omitempty"`
}

type parityChainSpecLinearPricing struct {
	Base uint64 `json:"base"`
	Word uint64 `json:"word"`
}

type parityChainSpecModExpPricing struct {
	Divisor uint64 `json:"divisor"`
}

// parityChainSpecAltBnConstOperationPricing defines the price
// policy for bn const operation(used after istanbul)
type parityChainSpecAltBnConstOperationPricing struct {
	Price uint64 `json:"price"`
}

// parityChainSepcAltBnPairingPricing defines the price policy
// for bn pairing.
type parityChainSepcAltBnPairingPricing struct {
	Base uint64 `json:"base"`
	Pair uint64 `json:"pair"`
}

// parityChainSpecBlakePricing defines the price policy for blake2 f
// compression.
type parityChainSpecBlakePricing struct {
	EnergyPerRound uint64 `json:"energy_per_round"`
}

type parityChainSpecAlternativePrice struct {
	AltBnConstOperationPrice *parityChainSpecAltBnConstOperationPricing `json:"alt_bn128_const_operations,omitempty"`
	AltBnPairingPrice        *parityChainSepcAltBnPairingPricing        `json:"alt_bn128_pairing,omitempty"`
}

// parityChainSpecVersionedPricing represents a single version price policy.
type parityChainSpecVersionedPricing struct {
	Price *parityChainSpecAlternativePrice `json:"price,omitempty"`
	Info  string                           `json:"info,omitempty"`
}

// newParityChainSpec converts a go-core genesis block into a Parity specific
// chain specification format.
func newParityChainSpec(network string, genesis *core.Genesis, bootnodes []string) (*parityChainSpec, error) {
	// Only cryptore is currently supported between go-core and Parity
	if genesis.Config.Cryptore == nil {
		return nil, errors.New("unsupported consensus engine")
	}
	// Reconstruct the chain spec in Parity's format
	spec := &parityChainSpec{
		Name:    network,
		Nodes:   bootnodes,
		Datadir: strings.ToLower(network),
	}
	spec.Engine.Cryptore.Params.BlockReward = make(map[string]string)
	spec.Engine.Cryptore.Params.DifficultyBombDelays = make(map[string]string)
	// Frontier
	spec.Engine.Cryptore.Params.MinimumDifficulty = (*hexutil.Big)(params.MinimumDifficulty)
	spec.Engine.Cryptore.Params.DifficultyBoundDivisor = (*hexutil.Big)(params.DifficultyBoundDivisor)
	spec.Engine.Cryptore.Params.DurationLimit = (*hexutil.Big)(params.DurationLimit)
	spec.Engine.Cryptore.Params.BlockReward["0x0"] = hexutil.EncodeBig(cryptore.FrontierBlockReward)

	// Homestead
	spec.Engine.Cryptore.Params.HomesteadTransition = hexutil.Uint64(genesis.Config.HomesteadBlock.Uint64())

	// Tangerine Whistle : 150
	// https://github.com/core-coin/EIPs/blob/master/EIPS/eip-608.md
	spec.Params.EIP150Transition = hexutil.Uint64(genesis.Config.EIP150Block.Uint64())

	// Spurious Dragon: 155, 160, 161, 170
	// https://github.com/core-coin/EIPs/blob/master/EIPS/eip-607.md
	spec.Params.EIP155Transition = hexutil.Uint64(genesis.Config.EIP155Block.Uint64())
	spec.Params.EIP160Transition = hexutil.Uint64(genesis.Config.EIP155Block.Uint64())
	spec.Params.EIP161abcTransition = hexutil.Uint64(genesis.Config.EIP158Block.Uint64())
	spec.Params.EIP161dTransition = hexutil.Uint64(genesis.Config.EIP158Block.Uint64())

	// Byzantium
	if num := genesis.Config.ByzantiumBlock; num != nil {
		spec.setByzantium(num)
	}
	// Constantinople
	if num := genesis.Config.ConstantinopleBlock; num != nil {
		spec.setConstantinople(num)
	}
	// ConstantinopleFix (remove eip-1283)
	if num := genesis.Config.PetersburgBlock; num != nil {
		spec.setConstantinopleFix(num)
	}
	// Istanbul
	if num := genesis.Config.IstanbulBlock; num != nil {
		spec.setIstanbul(num)
	}
	spec.Params.MaximumExtraDataSize = (hexutil.Uint64)(params.MaximumExtraDataSize)
	spec.Params.MinEnergyLimit = (hexutil.Uint64)(params.MinEnergyLimit)
	spec.Params.EnergyLimitBoundDivisor = (math2.HexOrDecimal64)(params.EnergyLimitBoundDivisor)
	spec.Params.NetworkID = (hexutil.Uint64)(genesis.Config.ChainID.Uint64())
	spec.Params.ChainID = (hexutil.Uint64)(genesis.Config.ChainID.Uint64())
	spec.Params.MaxCodeSize = params.MaxCodeSize
	// gcore has it set from zero
	spec.Params.MaxCodeSizeTransition = 0

	// Disable this one
	spec.Params.EIP98Transition = math.MaxInt64

	spec.Genesis.Seal.Core.Nonce = types.EncodeNonce(genesis.Nonce)
	spec.Genesis.Seal.Core.MixHash = (genesis.Mixhash[:])
	spec.Genesis.Difficulty = (*hexutil.Big)(genesis.Difficulty)
	spec.Genesis.Author = genesis.Coinbase
	spec.Genesis.Timestamp = (hexutil.Uint64)(genesis.Timestamp)
	spec.Genesis.ParentHash = genesis.ParentHash
	spec.Genesis.ExtraData = (hexutil.Bytes)(genesis.ExtraData)
	spec.Genesis.EnergyLimit = (hexutil.Uint64)(genesis.EnergyLimit)

	spec.Accounts = make(map[common.UnprefixedAddress]*parityChainSpecAccount)
	for address, account := range genesis.Alloc {
		bal := math2.HexOrDecimal256(*account.Balance)

		spec.Accounts[common.UnprefixedAddress(address)] = &parityChainSpecAccount{
			Balance: bal,
			Nonce:   math2.HexOrDecimal64(account.Nonce),
		}
	}
	spec.setPrecompile(1, &parityChainSpecBuiltin{Name: "ecrecover",
		Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 3000}}})

	spec.setPrecompile(2, &parityChainSpecBuiltin{
		Name: "sha256", Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 60, Word: 12}},
	})
	spec.setPrecompile(3, &parityChainSpecBuiltin{
		Name: "ripemd160", Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 600, Word: 120}},
	})
	spec.setPrecompile(4, &parityChainSpecBuiltin{
		Name: "identity", Pricing: &parityChainSpecPricing{Linear: &parityChainSpecLinearPricing{Base: 15, Word: 3}},
	})
	if genesis.Config.ByzantiumBlock != nil {
		spec.setPrecompile(5, &parityChainSpecBuiltin{
			Name:       "modexp",
			ActivateAt: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Pricing: &parityChainSpecPricing{
				ModExp: &parityChainSpecModExpPricing{Divisor: 20},
			},
		})
		spec.setPrecompile(6, &parityChainSpecBuiltin{
			Name:       "alt_bn128_add",
			ActivateAt: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Pricing: &parityChainSpecPricing{
				Linear: &parityChainSpecLinearPricing{Base: 500, Word: 0},
			},
		})
		spec.setPrecompile(7, &parityChainSpecBuiltin{
			Name:       "alt_bn128_mul",
			ActivateAt: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Pricing: &parityChainSpecPricing{
				Linear: &parityChainSpecLinearPricing{Base: 40000, Word: 0},
			},
		})
		spec.setPrecompile(8, &parityChainSpecBuiltin{
			Name:       "alt_bn128_pairing",
			ActivateAt: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Pricing: &parityChainSpecPricing{
				AltBnPairing: &parityChainSepcAltBnPairingPricing{Base: 100000, Pair: 80000},
			},
		})
	}
	if genesis.Config.IstanbulBlock != nil {
		if genesis.Config.ByzantiumBlock == nil {
			return nil, errors.New("invalid genesis, istanbul fork is enabled while byzantium is not")
		}
		spec.setPrecompile(6, &parityChainSpecBuiltin{
			Name:       "alt_bn128_add",
			ActivateAt: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Pricing: map[*hexutil.Big]*parityChainSpecVersionedPricing{
				(*hexutil.Big)(big.NewInt(0)): {
					Price: &parityChainSpecAlternativePrice{
						AltBnConstOperationPrice: &parityChainSpecAltBnConstOperationPricing{Price: 500},
					},
				},
				(*hexutil.Big)(genesis.Config.IstanbulBlock): {
					Price: &parityChainSpecAlternativePrice{
						AltBnConstOperationPrice: &parityChainSpecAltBnConstOperationPricing{Price: 150},
					},
				},
			},
		})
		spec.setPrecompile(7, &parityChainSpecBuiltin{
			Name:       "alt_bn128_mul",
			ActivateAt: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Pricing: map[*hexutil.Big]*parityChainSpecVersionedPricing{
				(*hexutil.Big)(big.NewInt(0)): {
					Price: &parityChainSpecAlternativePrice{
						AltBnConstOperationPrice: &parityChainSpecAltBnConstOperationPricing{Price: 40000},
					},
				},
				(*hexutil.Big)(genesis.Config.IstanbulBlock): {
					Price: &parityChainSpecAlternativePrice{
						AltBnConstOperationPrice: &parityChainSpecAltBnConstOperationPricing{Price: 6000},
					},
				},
			},
		})
		spec.setPrecompile(8, &parityChainSpecBuiltin{
			Name:       "alt_bn128_pairing",
			ActivateAt: (*hexutil.Big)(genesis.Config.ByzantiumBlock),
			Pricing: map[*hexutil.Big]*parityChainSpecVersionedPricing{
				(*hexutil.Big)(big.NewInt(0)): {
					Price: &parityChainSpecAlternativePrice{
						AltBnPairingPrice: &parityChainSepcAltBnPairingPricing{Base: 100000, Pair: 80000},
					},
				},
				(*hexutil.Big)(genesis.Config.IstanbulBlock): {
					Price: &parityChainSpecAlternativePrice{
						AltBnPairingPrice: &parityChainSepcAltBnPairingPricing{Base: 45000, Pair: 34000},
					},
				},
			},
		})
		spec.setPrecompile(9, &parityChainSpecBuiltin{
			Name:       "blake2_f",
			ActivateAt: (*hexutil.Big)(genesis.Config.IstanbulBlock),
			Pricing: &parityChainSpecPricing{
				Blake2F: &parityChainSpecBlakePricing{EnergyPerRound: 1},
			},
		})
	}
	return spec, nil
}

func (spec *parityChainSpec) setPrecompile(address byte, data *parityChainSpecBuiltin) {
	if spec.Accounts == nil {
		spec.Accounts = make(map[common.UnprefixedAddress]*parityChainSpecAccount)
	}
	a := common.UnprefixedAddress(common.BytesToAddress([]byte{address}))
	if _, exist := spec.Accounts[a]; !exist {
		spec.Accounts[a] = &parityChainSpecAccount{}
	}
	spec.Accounts[a].Builtin = data
}

func (spec *parityChainSpec) setByzantium(num *big.Int) {
	spec.Engine.Cryptore.Params.BlockReward[hexutil.EncodeBig(num)] = hexutil.EncodeBig(cryptore.ByzantiumBlockReward)
	spec.Engine.Cryptore.Params.DifficultyBombDelays[hexutil.EncodeBig(num)] = hexutil.EncodeUint64(3000000)
	n := hexutil.Uint64(num.Uint64())
	spec.Engine.Cryptore.Params.EIP100bTransition = n
	spec.Params.EIP140Transition = n
	spec.Params.EIP211Transition = n
	spec.Params.EIP214Transition = n
	spec.Params.EIP658Transition = n
}

func (spec *parityChainSpec) setConstantinople(num *big.Int) {
	spec.Engine.Cryptore.Params.BlockReward[hexutil.EncodeBig(num)] = hexutil.EncodeBig(cryptore.ConstantinopleBlockReward)
	spec.Engine.Cryptore.Params.DifficultyBombDelays[hexutil.EncodeBig(num)] = hexutil.EncodeUint64(2000000)
	n := hexutil.Uint64(num.Uint64())
	spec.Params.EIP145Transition = n
	spec.Params.EIP1014Transition = n
	spec.Params.EIP1052Transition = n
	spec.Params.EIP1283Transition = n
}

func (spec *parityChainSpec) setConstantinopleFix(num *big.Int) {
	spec.Params.EIP1283DisableTransition = hexutil.Uint64(num.Uint64())
}

func (spec *parityChainSpec) setIstanbul(num *big.Int) {
	spec.Params.EIP1344Transition = hexutil.Uint64(num.Uint64())
	spec.Params.EIP1884Transition = hexutil.Uint64(num.Uint64())
	spec.Params.EIP2028Transition = hexutil.Uint64(num.Uint64())
	spec.Params.EIP1283ReenableTransition = hexutil.Uint64(num.Uint64())
}

// pyCoreGenesisSpec represents the genesis specification format used by the
// Python Core implementation.
type pyCoreGenesisSpec struct {
	Nonce      types.BlockNonce  `json:"nonce"`
	Timestamp  hexutil.Uint64    `json:"timestamp"`
	ExtraData  hexutil.Bytes     `json:"extraData"`
	EnergyLimit   hexutil.Uint64    `json:"energyLimit"`
	Difficulty *hexutil.Big      `json:"difficulty"`
	Mixhash    common.Hash       `json:"mixhash"`
	Coinbase   common.Address    `json:"coinbase"`
	Alloc      core.GenesisAlloc `json:"alloc"`
	ParentHash common.Hash       `json:"parentHash"`
}

// newPyCoreGenesisSpec converts a go-core genesis block into a Parity specific
// chain specification format.
func newPyCoreGenesisSpec(network string, genesis *core.Genesis) (*pyCoreGenesisSpec, error) {
	// Only cryptore is currently supported between go-core and pycore
	if genesis.Config.Cryptore == nil {
		return nil, errors.New("unsupported consensus engine")
	}
	spec := &pyCoreGenesisSpec{
		Nonce:      types.EncodeNonce(genesis.Nonce),
		Timestamp:  (hexutil.Uint64)(genesis.Timestamp),
		ExtraData:  genesis.ExtraData,
		EnergyLimit:   (hexutil.Uint64)(genesis.EnergyLimit),
		Difficulty: (*hexutil.Big)(genesis.Difficulty),
		Mixhash:    genesis.Mixhash,
		Coinbase:   genesis.Coinbase,
		Alloc:      genesis.Alloc,
		ParentHash: genesis.ParentHash,
	}
	return spec, nil
}

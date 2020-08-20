// Copyright 2020 The CORE FOUNDATION, nadacia
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

package params

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/crypto"
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0x3450eb7d800473633e546b5e466e8d84c36de8f844f4c9562ae0b2c24e6864d2")
	DevinGenesisHash   = common.HexToHash("0xb2da344c413f2cffadc3a2d9a08ecaa380e65e39f4a3cf0b48b4f6158912507f")
	KolibaGenesisHash  = common.HexToHash("0xbf7e331f7f7c1dd2e05159666b3bf8bc7a8a3a9eb1d518969eab529dd9b88c1a")
)

// TrustedCheckpoints associates each known checkpoint with the genesis hash of
// the chain it belongs to.
var TrustedCheckpoints = map[common.Hash]*TrustedCheckpoint{
	MainnetGenesisHash: MainnetTrustedCheckpoint,
	DevinGenesisHash:   DevinTrustedCheckpoint,
	KolibaGenesisHash:  KolibaTrustedCheckpoint,
}

// CheckpointOracles associates each known checkpoint oracles with the genesis hash of
// the chain it belongs to.
var CheckpointOracles = map[common.Hash]*CheckpointOracleConfig{
	MainnetGenesisHash: MainnetCheckpointOracle,
	DevinGenesisHash:   DevinCheckpointOracle,
	KolibaGenesisHash:  KolibaCheckpointOracle,
}

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainID:  big.NewInt(1),
		Cryptore: new(CryptoreConfig),
	}

	// MainnetTrustedCheckpoint contains the light client trusted checkpoint for the main network.
	MainnetTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 289,
		SectionHead:  common.HexToHash("0x5a95eed1a6e01d58b59f86c754cda88e8d6bede65428530eb0bec03267cda6a9"),
		CHTRoot:      common.HexToHash("0x6d4abf2b0f3c015952e6a3cbd5cc9885aacc29b8e55d4de662d29783c74a62bf"),
		BloomRoot:    common.HexToHash("0x1af2a8abbaca8048136b02f782cb6476ab546313186a1d1bd2b02df88ea48e7e"),
	}

	// MainnetCheckpointOracle contains a set of configs for the main network oracle.
	MainnetCheckpointOracle = &CheckpointOracleConfig{
		Address:   common.Address{},
		Signers:   []common.Address{},
		Threshold: 2,
	}

	// DevinChainConfig contains the chain parameters to run a node on the Devin test network.
	DevinChainConfig = &ChainConfig{
		ChainID:  big.NewInt(3),
		Cryptore: new(CryptoreConfig),
	}

	// DevinTrustedCheckpoint contains the light client trusted checkpoint for the Devin test network.
	DevinTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 223,
		SectionHead:  common.HexToHash("0x9aa51ca383f5075f816e0b8ce7125075cd562b918839ee286c03770722147661"),
		CHTRoot:      common.HexToHash("0x755c6a5931b7bd36e55e47f3f1e81fa79c930ae15c55682d3a85931eedaf8cf2"),
		BloomRoot:    common.HexToHash("0xabc37762d11b29dc7dde11b89846e2308ba681eeb015b6a202ef5e242bc107e8"),
	}

	// DevinCheckpointOracle contains a set of configs for the Devin test network oracle.
	DevinCheckpointOracle = &CheckpointOracleConfig{
		Address:   common.Address{},
		Signers:   []common.Address{},
		Threshold: 2,
	}

	// KolibaChainConfig contains the chain parameters to run a node on the Koliba test network.
	KolibaChainConfig = &ChainConfig{
		ChainID: big.NewInt(4),
		Clique: &CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}

	// KolibaTrustedCheckpoint contains the light client trusted checkpoint for the Koliba test network.
	KolibaTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 66,
		SectionHead:  common.HexToHash("0xeea3a7b2cb275956f3049dd27e6cdacd8a6ef86738d593d556efee5361019475"),
		CHTRoot:      common.HexToHash("0x11712af50b4083dc5910e452ca69fbfc0f2940770b9846200a573f87a0af94e6"),
		BloomRoot:    common.HexToHash("0x331b7a7b273e81daeac8cafb9952a16669d7facc7be3b0ebd3a792b4d8b95cc5"),
	}

	// KolibaCheckpointOracle contains a set of configs for the Koliba test network oracle.
	KolibaCheckpointOracle = &CheckpointOracleConfig{
		Address:   common.Address{},
		Signers:   []common.Address{},
		Threshold: 2,
	}

	// AllCryptoreProtocolChanges contains every protocol change (CIPs) introduced
	// and accepted by the Core core developers into the Cryptore consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCryptoreProtocolChanges = &ChainConfig{big.NewInt(1), nil, new(CryptoreConfig), nil}

	// AllCliqueProtocolChanges contains every protocol change (CIPs) introduced
	// and accepted by the Core core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{big.NewInt(1), nil, nil, &CliqueConfig{Period: 0, Epoch: 30000}}

	TestChainConfig = &ChainConfig{big.NewInt(2), nil, new(CryptoreConfig), nil}
	TestRules       = TestChainConfig.Rules(new(big.Int))
)

// TrustedCheckpoint represents a set of post-processed trie roots (CHT and
// BloomTrie) associated with the appropriate section index and head hash. It is
// used to start light syncing from this checkpoint and avoid downloading the
// entire header chain while still being able to securely access old headers/logs.
type TrustedCheckpoint struct {
	SectionIndex uint64      `json:"sectionIndex"`
	SectionHead  common.Hash `json:"sectionHead"`
	CHTRoot      common.Hash `json:"chtRoot"`
	BloomRoot    common.Hash `json:"bloomRoot"`
}

// HashEqual returns an indicator comparing the itself hash with given one.
func (c *TrustedCheckpoint) HashEqual(hash common.Hash) bool {
	if c.Empty() {
		return hash == common.Hash{}
	}
	return c.Hash() == hash
}

// Hash returns the hash of checkpoint's four key fields(index, sectionHead, chtRoot and bloomTrieRoot).
func (c *TrustedCheckpoint) Hash() common.Hash {
	buf := make([]byte, 8+3*common.HashLength)
	binary.BigEndian.PutUint64(buf, c.SectionIndex)
	copy(buf[8:], c.SectionHead.Bytes())
	copy(buf[8+common.HashLength:], c.CHTRoot.Bytes())
	copy(buf[8+2*common.HashLength:], c.BloomRoot.Bytes())
	return crypto.SHA3Hash(buf)
}

// Empty returns an indicator whether the checkpoint is regarded as empty.
func (c *TrustedCheckpoint) Empty() bool {
	return c.SectionHead == (common.Hash{}) || c.CHTRoot == (common.Hash{}) || c.BloomRoot == (common.Hash{})
}

// CheckpointOracleConfig represents a set of checkpoint contract(which acts as an oracle)
// config which used for light client checkpoint syncing.
type CheckpointOracleConfig struct {
	Address   common.Address   `json:"address"`
	Signers   []common.Address `json:"signers"`
	Threshold uint64           `json:"threshold"`
}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	EWASMBlock *big.Int `json:"ewasmBlock,omitempty"` // EWASM switch block (nil = no fork, 0 = already activated)

	// Various consensus engines
	Cryptore *CryptoreConfig `json:"cryptore,omitempty"`
	Clique   *CliqueConfig   `json:"clique,omitempty"`
}

// CryptoreConfig is the consensus engine configs for proof-of-work based sealing.
type CryptoreConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *CryptoreConfig) String() string {
	return "cryptore"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Cryptore != nil:
		engine = c.Cryptore
	case c.Clique != nil:
		engine = c.Clique
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v, Engine: %v}",
		c.ChainID,
		engine,
	)
}

// IsEWASM returns whether num represents a block number after the EWASM fork
func (c *ChainConfig) IsEWASM(num *big.Int) bool {
	return isForked(c.EWASMBlock, num)
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

// CheckConfigForkOrder checks that we don't "skip" any forks, gocore isn't pluggable enough
// to guarantee that forks can be implemented in a different order than on official networks
func (c *ChainConfig) CheckConfigForkOrder() error {
	type fork struct {
		name  string
		block *big.Int
	}
	var lastFork fork
	for _, cur := range []fork{} {
		if lastFork.name != "" {
			// Next one must be higher number
			if lastFork.block == nil && cur.block != nil {
				return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at %v",
					lastFork.name, cur.name, cur.block)
			}
			if lastFork.block != nil && cur.block != nil {
				if lastFork.block.Cmp(cur.block) > 0 {
					return fmt.Errorf("unsupported fork ordering: %v enabled at %v, but %v enabled at %v",
						lastFork.name, lastFork.block, cur.name, cur.block)
				}
			}
		}
		lastFork = cur
	}
	return nil
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	if isForkIncompatible(c.EWASMBlock, newcfg.EWASMBlock, head) {
		return newCompatError("ewasm fork block", c.EWASMBlock, newcfg.EWASMBlock)
	}
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID *big.Int
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Rules{
		ChainID: new(big.Int).Set(chainID),
	}
}

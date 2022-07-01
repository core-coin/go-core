// Copyright 2019 by the Authors
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
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/core-coin/go-core/cmd/utils"
	"github.com/core-coin/go-core/common"
	"github.com/core-coin/go-core/common/hexutil"
	"github.com/core-coin/go-core/common/math"
	"github.com/core-coin/go-core/consensus"
	"github.com/core-coin/go-core/consensus/cryptore"
	"github.com/core-coin/go-core/core"
	"github.com/core-coin/go-core/core/rawdb"
	"github.com/core-coin/go-core/core/state"
	"github.com/core-coin/go-core/core/types"
	"github.com/core-coin/go-core/core/vm"
	"github.com/core-coin/go-core/crypto"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/node"
	"github.com/core-coin/go-core/params"
	"github.com/core-coin/go-core/rlp"
	"github.com/core-coin/go-core/rpc"
	"github.com/core-coin/go-core/trie"
	"github.com/core-coin/go-core/xcbdb"

	cli "gopkg.in/urfave/cli.v1"
)

var (
	rpcPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: node.DefaultHTTPPort,
	}
	retestxcbCommand = cli.Command{
		Action:      utils.MigrateFlags(retestxcb),
		Name:        "retestxcb",
		Usage:       "Launches gocore in retestxcb mode",
		ArgsUsage:   "",
		Flags:       []cli.Flag{rpcPortFlag},
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `Launches gocore in retestxcb mode (no database, no network, only retestxcb RPC interface)`,
	}
)

type RetestxcbTestAPI interface {
	SetChainParams(ctx context.Context, chainParams ChainParams) (bool, error)
	MineBlocks(ctx context.Context, number uint64) (bool, error)
	ModifyTimestamp(ctx context.Context, interval uint64) (bool, error)
	ImportRawBlock(ctx context.Context, rawBlock hexutil.Bytes) (common.Hash, error)
	RewindToBlock(ctx context.Context, number uint64) (bool, error)
	GetLogHash(ctx context.Context, txHash common.Hash) (common.Hash, error)
}

type RetestxcbXcbAPI interface {
	SendRawTransaction(ctx context.Context, rawTx hexutil.Bytes) (common.Hash, error)
	BlockNumber(ctx context.Context) (uint64, error)
	GetBlockByNumber(ctx context.Context, blockNr math.HexOrDecimal64, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error)
	GetBalance(ctx context.Context, address common.Address, blockNr math.HexOrDecimal64) (*math.HexOrDecimal256, error)
	GetCode(ctx context.Context, address common.Address, blockNr math.HexOrDecimal64) (hexutil.Bytes, error)
	GetTransactionCount(ctx context.Context, address common.Address, blockNr math.HexOrDecimal64) (uint64, error)
}

type RetestxcbDebugAPI interface {
	AccountRange(ctx context.Context,
		blockHashOrNumber *math.HexOrDecimal256, txIndex uint64,
		addressHash *math.HexOrDecimal256, maxResults uint64,
	) (AccountRangeResult, error)
	StorageRangeAt(ctx context.Context,
		blockHashOrNumber *math.HexOrDecimal256, txIndex uint64,
		address common.Address,
		begin *math.HexOrDecimal256, maxResults uint64,
	) (StorageRangeResult, error)
}

type RetestWeb3API interface {
	ClientVersion(ctx context.Context) (string, error)
}

type RetestxcbAPI struct {
	xcbDb         xcbdb.Database
	db            state.Database
	chainConfig   *params.ChainConfig
	author        common.Address
	extraData     []byte
	genesisHash   common.Hash
	engine        *NoRewardEngine
	blockchain    *core.BlockChain
	txMap         map[common.Address]map[uint64]*types.Transaction // Sender -> Nonce -> Transaction
	txSenders     map[common.Address]struct{}                      // Set of transaction senders
	blockInterval uint64
}

type ChainParams struct {
	SealEngine string                            `json:"sealEngine"`
	Params     CParamsParams                     `json:"params"`
	Genesis    CParamsGenesis                    `json:"genesis"`
	Accounts   map[common.Address]CParamsAccount `json:"accounts"`
}

type CParamsParams struct {
	AccountStartNonce       math.HexOrDecimal64   `json:"accountStartNonce"`
	NetworkID               *math.HexOrDecimal256 `json:"networkID"`
	MaximumExtraDataSize    math.HexOrDecimal64   `json:"maximumExtraDataSize"`
	TieBreakingEnergy       bool                  `json:"tieBreakingEnergy"`
	MinEnergyLimit          math.HexOrDecimal64   `json:"minEnergyLimit"`
	MaxEnergyLimit          math.HexOrDecimal64   `json:"maxEnergyLimit"`
	EnergyLimitBoundDivisor math.HexOrDecimal64   `json:"energyLimitBoundDivisor"`
	MinimumDifficulty       math.HexOrDecimal256  `json:"minimumDifficulty"`
	DifficultyBoundDivisor  math.HexOrDecimal256  `json:"difficultyBoundDivisor"`
	DurationLimit           math.HexOrDecimal256  `json:"durationLimit"`
	BlockReward             math.HexOrDecimal256  `json:"blockReward"`
}

type CParamsGenesis struct {
	Nonce       math.HexOrDecimal64   `json:"nonce"`
	Difficulty  *math.HexOrDecimal256 `json:"difficulty"`
	Author      common.Address        `json:"author"`
	Timestamp   math.HexOrDecimal64   `json:"timestamp"`
	ParentHash  common.Hash           `json:"parentHash"`
	ExtraData   hexutil.Bytes         `json:"extraData"`
	EnergyLimit math.HexOrDecimal64   `json:"energyLimit"`
}

type CParamsAccount struct {
	Balance     *math.HexOrDecimal256 `json:"balance"`
	Precompiled *CPAccountPrecompiled `json:"precompiled"`
	Code        hexutil.Bytes         `json:"code"`
	Storage     map[string]string     `json:"storage"`
	Nonce       *math.HexOrDecimal64  `json:"nonce"`
}

type CPAccountPrecompiled struct {
	Name          string                `json:"name"`
	StartingBlock math.HexOrDecimal64   `json:"startingBlock"`
	Linear        *CPAPrecompiledLinear `json:"linear"`
}

type CPAPrecompiledLinear struct {
	Base uint64 `json:"base"`
	Word uint64 `json:"word"`
}

type AccountRangeResult struct {
	AddressMap map[common.Hash]common.Address `json:"addressMap"`
	NextKey    common.Hash                    `json:"nextKey"`
}

type StorageRangeResult struct {
	Complete bool                   `json:"complete"`
	Storage  map[common.Hash]SRItem `json:"storage"`
}

type SRItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type NoRewardEngine struct {
	inner     consensus.Engine
	rewardsOn bool
}

func (e *NoRewardEngine) Author(header *types.Header) (common.Address, error) {
	return e.inner.Author(header)
}

func (e *NoRewardEngine) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {
	return e.inner.VerifyHeader(chain, header, seal)
}

func (e *NoRewardEngine) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	return e.inner.VerifyHeaders(chain, headers, seals)
}

func (e *NoRewardEngine) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return e.inner.VerifyUncles(chain, block)
}

func (e *NoRewardEngine) VerifySeal(chain consensus.ChainHeaderReader, header *types.Header) error {
	return e.inner.VerifySeal(chain, header)
}

func (e *NoRewardEngine) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	return e.inner.Prepare(chain, header)
}

func (e *NoRewardEngine) accumulateRewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// Simply touch miner and uncle coinbase accounts
	reward := big.NewInt(0)
	for _, uncle := range uncles {
		state.AddBalance(uncle.Coinbase, reward)
	}
	state.AddBalance(header.Coinbase, reward)
}

func (e *NoRewardEngine) Finalize(chain consensus.ChainHeaderReader, header *types.Header, statedb *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header) {
	if e.rewardsOn {
		e.inner.Finalize(chain, header, statedb, txs, uncles)
	} else {
		e.accumulateRewards(chain.Config(), statedb, header, uncles)
		header.Root = statedb.IntermediateRoot(true)
	}
}

func (e *NoRewardEngine) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, statedb *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	if e.rewardsOn {
		return e.inner.FinalizeAndAssemble(chain, header, statedb, txs, uncles, receipts)
	} else {
		e.accumulateRewards(chain.Config(), statedb, header, uncles)
		header.Root = statedb.IntermediateRoot(true)

		// Header seems complete, assemble into a block and return
		return types.NewBlock(header, txs, uncles, receipts, new(trie.Trie)), nil
	}
}

func (e *NoRewardEngine) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	return e.inner.Seal(chain, block, results, stop)
}

func (e *NoRewardEngine) SealHash(header *types.Header) common.Hash {
	return e.inner.SealHash(header)
}

func (e *NoRewardEngine) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return e.inner.CalcDifficulty(chain, time, parent)
}

func (e *NoRewardEngine) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return e.inner.APIs(chain)
}

func (e *NoRewardEngine) Close() error {
	return e.inner.Close()
}

func (api *RetestxcbAPI) SetChainParams(ctx context.Context, chainParams ChainParams) (bool, error) {
	// Clean up
	if api.blockchain != nil {
		api.blockchain.Stop()
	}
	if api.engine != nil {
		api.engine.Close()
	}
	if api.xcbDb != nil {
		api.xcbDb.Close()
	}
	xcbDb := rawdb.NewMemoryDatabase()
	accounts := make(core.GenesisAlloc)
	for address, account := range chainParams.Accounts {
		balance := big.NewInt(0)
		if account.Balance != nil {
			balance.Set((*big.Int)(account.Balance))
		}
		var nonce uint64
		if account.Nonce != nil {
			nonce = uint64(*account.Nonce)
		}
		if account.Precompiled == nil || account.Balance != nil {
			storage := make(map[common.Hash]common.Hash)
			for k, v := range account.Storage {
				storage[common.HexToHash(k)] = common.HexToHash(v)
			}
			accounts[address] = core.GenesisAccount{
				Balance: balance,
				Code:    account.Code,
				Nonce:   nonce,
				Storage: storage,
			}
		}
	}
	networkId := big.NewInt(1)
	if chainParams.Params.NetworkID != nil {
		networkId.Set((*big.Int)(chainParams.Params.NetworkID))
	}

	genesis := &core.Genesis{
		Config: &params.ChainConfig{
			NetworkID: networkId,
		},
		Nonce:       uint64(chainParams.Genesis.Nonce),
		Timestamp:   uint64(chainParams.Genesis.Timestamp),
		ExtraData:   chainParams.Genesis.ExtraData,
		EnergyLimit: uint64(chainParams.Genesis.EnergyLimit),
		Difficulty:  big.NewInt(0).Set((*big.Int)(chainParams.Genesis.Difficulty)),
		Coinbase:    chainParams.Genesis.Author,
		ParentHash:  chainParams.Genesis.ParentHash,
		Alloc:       accounts,
	}
	chainConfig, genesisHash, err := core.SetupGenesisBlock(xcbDb, genesis)
	if err != nil {
		return false, err
	}
	fmt.Printf("Chain config: %v\n", chainConfig)

	var inner consensus.Engine
	switch chainParams.SealEngine {
	case "NoProof", "NoReward":
		inner = cryptore.NewFaker()
	case "Cryptore":
		inner = cryptore.New(cryptore.Config{}, nil, false)
	default:
		return false, fmt.Errorf("unrecognised seal engine: %s", chainParams.SealEngine)
	}
	engine := &NoRewardEngine{inner: inner, rewardsOn: chainParams.SealEngine != "NoReward"}

	blockchain, err := core.NewBlockChain(xcbDb, nil, chainConfig, engine, vm.Config{}, nil, nil)
	if err != nil {
		return false, err
	}

	api.chainConfig = chainConfig
	api.genesisHash = genesisHash
	api.author = chainParams.Genesis.Author
	api.extraData = chainParams.Genesis.ExtraData
	api.xcbDb = xcbDb
	api.engine = engine
	api.blockchain = blockchain
	api.db = state.NewDatabase(api.xcbDb)
	api.txMap = make(map[common.Address]map[uint64]*types.Transaction)
	api.txSenders = make(map[common.Address]struct{})
	api.blockInterval = 0
	return true, nil
}

func (api *RetestxcbAPI) SendRawTransaction(ctx context.Context, rawTx hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(rawTx, tx); err != nil {
		// Return nil is not by mistake - some tests include sending transaction where energyLimit overflows uint64
		return common.Hash{}, nil
	}
	signer := types.MakeSigner(api.blockchain.Config().NetworkID)
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return common.Hash{}, err
	}
	if nonceMap, ok := api.txMap[sender]; ok {
		nonceMap[tx.Nonce()] = tx
	} else {
		nonceMap = make(map[uint64]*types.Transaction)
		nonceMap[tx.Nonce()] = tx
		api.txMap[sender] = nonceMap
	}
	api.txSenders[sender] = struct{}{}
	return tx.Hash(), nil
}

func (api *RetestxcbAPI) MineBlocks(ctx context.Context, number uint64) (bool, error) {
	for i := 0; i < int(number); i++ {
		if err := api.mineBlock(); err != nil {
			return false, err
		}
	}
	fmt.Printf("Mined %d blocks\n", number)
	return true, nil
}

func (api *RetestxcbAPI) currentNumber() uint64 {
	if current := api.blockchain.CurrentBlock(); current != nil {
		return current.NumberU64()
	}
	return 0
}

func (api *RetestxcbAPI) mineBlock() error {
	number := api.currentNumber()
	parentHash := rawdb.ReadCanonicalHash(api.xcbDb, number)
	parent := rawdb.ReadBlock(api.xcbDb, parentHash, number)
	var timestamp uint64
	if api.blockInterval == 0 {
		timestamp = uint64(time.Now().Unix())
	} else {
		timestamp = parent.Time() + api.blockInterval
	}
	energyLimit := core.CalcEnergyLimit(parent, 9223372036854775807, 9223372036854775807)
	header := &types.Header{
		ParentHash:  parent.Hash(),
		Number:      big.NewInt(int64(number + 1)),
		EnergyLimit: energyLimit,
		Extra:       api.extraData,
		Time:        timestamp,
	}
	header.Coinbase = api.author
	if api.engine != nil {
		api.engine.Prepare(api.blockchain, header)
	}
	statedb, err := api.blockchain.StateAt(parent.Root())
	if err != nil {
		return err
	}
	energyPool := new(core.EnergyPool).AddEnergy(header.EnergyLimit)
	txCount := 0
	var txs []*types.Transaction
	var receipts []*types.Receipt
	var blockFull = energyPool.Energy() < params.TxEnergy
	for address := range api.txSenders {
		if blockFull {
			break
		}
		m := api.txMap[address]
		for nonce := statedb.GetNonce(address); ; nonce++ {
			if tx, ok := m[nonce]; ok {
				// Try to apply transactions to the state
				statedb.Prepare(tx.Hash(), common.Hash{}, txCount)
				snap := statedb.Snapshot()

				receipt, err := core.ApplyTransaction(
					api.chainConfig,
					api.blockchain,
					&api.author,
					energyPool,
					statedb,
					header, tx, &header.EnergyUsed, *api.blockchain.GetVMConfig(),
				)
				if err != nil {
					statedb.RevertToSnapshot(snap)
					break
				}
				txs = append(txs, tx)
				receipts = append(receipts, receipt)
				delete(m, nonce)
				if len(m) == 0 {
					// Last tx for the sender
					delete(api.txMap, address)
					delete(api.txSenders, address)
				}
				txCount++
				if energyPool.Energy() < params.TxEnergy {
					blockFull = true
					break
				}
			} else {
				break // Gap in the nonces
			}
		}
	}
	block, err := api.engine.FinalizeAndAssemble(api.blockchain, header, statedb, txs, []*types.Header{}, receipts)
	if err != nil {
		return err
	}
	return api.importBlock(block)
}

func (api *RetestxcbAPI) importBlock(block *types.Block) error {
	if _, err := api.blockchain.InsertChain([]*types.Block{block}); err != nil {
		return err
	}
	fmt.Printf("Imported block %d,  head is %d\n", block.NumberU64(), api.currentNumber())
	return nil
}

func (api *RetestxcbAPI) ModifyTimestamp(ctx context.Context, interval uint64) (bool, error) {
	api.blockInterval = interval
	return true, nil
}

func (api *RetestxcbAPI) ImportRawBlock(ctx context.Context, rawBlock hexutil.Bytes) (common.Hash, error) {
	block := new(types.Block)
	if err := rlp.DecodeBytes(rawBlock, block); err != nil {
		return common.Hash{}, err
	}
	fmt.Printf("Importing block %d with parent hash: %x, genesisHash: %x\n", block.NumberU64(), block.ParentHash(), api.genesisHash)
	if err := api.importBlock(block); err != nil {
		return common.Hash{}, err
	}
	return block.Hash(), nil
}

func (api *RetestxcbAPI) RewindToBlock(ctx context.Context, newHead uint64) (bool, error) {
	if err := api.blockchain.SetHead(newHead); err != nil {
		return false, err
	}
	// When we rewind, the transaction pool should be cleaned out.
	api.txMap = make(map[common.Address]map[uint64]*types.Transaction)
	api.txSenders = make(map[common.Address]struct{})
	return true, nil
}

var emptyListHash common.Hash = common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")

func (api *RetestxcbAPI) GetLogHash(ctx context.Context, txHash common.Hash) (common.Hash, error) {
	receipt, _, _, _ := rawdb.ReadReceipt(api.xcbDb, txHash, api.chainConfig)
	if receipt == nil {
		return emptyListHash, nil
	} else {
		if logListRlp, err := rlp.EncodeToBytes(receipt.Logs); err != nil {
			return common.Hash{}, err
		} else {
			return common.BytesToHash(crypto.SHA3(logListRlp)), nil
		}
	}
}

func (api *RetestxcbAPI) BlockNumber(ctx context.Context) (uint64, error) {
	return api.currentNumber(), nil
}

func (api *RetestxcbAPI) GetBlockByNumber(ctx context.Context, blockNr math.HexOrDecimal64, fullTx bool) (map[string]interface{}, error) {
	block := api.blockchain.GetBlockByNumber(uint64(blockNr))
	if block != nil {
		response, err := RPCMarshalBlock(block, true, fullTx)
		if err != nil {
			return nil, err
		}
		response["author"] = response["miner"]
		response["totalDifficulty"] = (*hexutil.Big)(api.blockchain.GetTd(block.Hash(), uint64(blockNr)))
		return response, err
	}
	return nil, fmt.Errorf("block %d not found", blockNr)
}

func (api *RetestxcbAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block := api.blockchain.GetBlockByHash(blockHash)
	if block != nil {
		response, err := RPCMarshalBlock(block, true, fullTx)
		if err != nil {
			return nil, err
		}
		response["author"] = response["miner"]
		response["totalDifficulty"] = (*hexutil.Big)(api.blockchain.GetTd(block.Hash(), block.Number().Uint64()))
		return response, err
	}
	return nil, fmt.Errorf("block 0x%x not found", blockHash)
}

func (api *RetestxcbAPI) AccountRange(ctx context.Context,
	blockHashOrNumber *math.HexOrDecimal256, txIndex uint64,
	addressHash *math.HexOrDecimal256, maxResults uint64,
) (AccountRangeResult, error) {
	var (
		header *types.Header
		block  *types.Block
	)
	if (*big.Int)(blockHashOrNumber).Cmp(big.NewInt(math.MaxInt64)) > 0 {
		blockHash := common.BigToHash((*big.Int)(blockHashOrNumber))
		header = api.blockchain.GetHeaderByHash(blockHash)
		block = api.blockchain.GetBlockByHash(blockHash)
		//fmt.Printf("Account range: %x, txIndex %d, start: %x, maxResults: %d\n", blockHash, txIndex, common.BigToHash((*big.Int)(addressHash)), maxResults)
	} else {
		blockNumber := (*big.Int)(blockHashOrNumber).Uint64()
		header = api.blockchain.GetHeaderByNumber(blockNumber)
		block = api.blockchain.GetBlockByNumber(blockNumber)
		//fmt.Printf("Account range: %d, txIndex %d, start: %x, maxResults: %d\n", blockNumber, txIndex, common.BigToHash((*big.Int)(addressHash)), maxResults)
	}
	parentHeader := api.blockchain.GetHeaderByHash(header.ParentHash)
	var root common.Hash
	var statedb *state.StateDB
	var err error
	if parentHeader == nil || int(txIndex) >= len(block.Transactions()) {
		root = header.Root
		statedb, err = api.blockchain.StateAt(root)
		if err != nil {
			return AccountRangeResult{}, err
		}
	} else {
		root = parentHeader.Root
		statedb, err = api.blockchain.StateAt(root)
		if err != nil {
			return AccountRangeResult{}, err
		}
		// Recompute transactions up to the target index.
		signer := types.MakeSigner(api.blockchain.Config().NetworkID)
		for idx, tx := range block.Transactions() {
			// Assemble the transaction call message and return if the requested offset
			msg, _ := tx.AsMessage(signer)
			context := core.NewCVMContext(msg, block.Header(), api.blockchain, nil)
			// Not yet the searched for transaction, execute on top of the current state
			vmenv := vm.NewCVM(context, statedb, api.blockchain.Config(), vm.Config{})
			if _, _, _, err := core.ApplyMessage(vmenv, msg, new(core.EnergyPool).AddEnergy(tx.Energy())); err != nil {
				return AccountRangeResult{}, fmt.Errorf("transaction %#x failed: %v", tx.Hash(), err)
			}
			// Ensure any modifications are committed to the state
			root = statedb.IntermediateRoot(true)
			if idx == int(txIndex) {
				// This is to make sure root can be opened by OpenTrie
				root, err = statedb.Commit(true)
				if err != nil {
					return AccountRangeResult{}, err
				}
				break
			}
		}
	}
	accountTrie, err := statedb.Database().OpenTrie(root)
	if err != nil {
		return AccountRangeResult{}, err
	}
	it := trie.NewIterator(accountTrie.NodeIterator(common.BigToHash((*big.Int)(addressHash)).Bytes()))
	result := AccountRangeResult{AddressMap: make(map[common.Hash]common.Address)}
	for i := 0; i < int(maxResults) && it.Next(); i++ {
		if preimage := accountTrie.GetKey(it.Key); preimage != nil {
			result.AddressMap[common.BytesToHash(it.Key)] = common.BytesToAddress(preimage)
		}
	}
	//fmt.Printf("Number of entries returned: %d\n", len(result.AddressMap))
	// Add the 'next key' so clients can continue downloading.
	if it.Next() {
		next := common.BytesToHash(it.Key)
		result.NextKey = next
	}
	return result, nil
}

func (api *RetestxcbAPI) GetBalance(ctx context.Context, address common.Address, blockNr math.HexOrDecimal64) (*math.HexOrDecimal256, error) {
	//fmt.Printf("GetBalance %x, block %d\n", address, blockNr)
	header := api.blockchain.GetHeaderByNumber(uint64(blockNr))
	statedb, err := api.blockchain.StateAt(header.Root)
	if err != nil {
		return nil, err
	}
	return (*math.HexOrDecimal256)(statedb.GetBalance(address)), nil
}

func (api *RetestxcbAPI) GetCode(ctx context.Context, address common.Address, blockNr math.HexOrDecimal64) (hexutil.Bytes, error) {
	header := api.blockchain.GetHeaderByNumber(uint64(blockNr))
	statedb, err := api.blockchain.StateAt(header.Root)
	if err != nil {
		return nil, err
	}
	return statedb.GetCode(address), nil
}

func (api *RetestxcbAPI) GetTransactionCount(ctx context.Context, address common.Address, blockNr math.HexOrDecimal64) (uint64, error) {
	header := api.blockchain.GetHeaderByNumber(uint64(blockNr))
	statedb, err := api.blockchain.StateAt(header.Root)
	if err != nil {
		return 0, err
	}
	return statedb.GetNonce(address), nil
}

func (api *RetestxcbAPI) StorageRangeAt(ctx context.Context,
	blockHashOrNumber *math.HexOrDecimal256, txIndex uint64,
	address common.Address,
	begin *math.HexOrDecimal256, maxResults uint64,
) (StorageRangeResult, error) {
	var (
		header *types.Header
		block  *types.Block
	)
	if (*big.Int)(blockHashOrNumber).Cmp(big.NewInt(math.MaxInt64)) > 0 {
		blockHash := common.BigToHash((*big.Int)(blockHashOrNumber))
		header = api.blockchain.GetHeaderByHash(blockHash)
		block = api.blockchain.GetBlockByHash(blockHash)
		//fmt.Printf("Storage range: %x, txIndex %d, addr: %x, start: %x, maxResults: %d\n",
		//	blockHash, txIndex, address, common.BigToHash((*big.Int)(begin)), maxResults)
	} else {
		blockNumber := (*big.Int)(blockHashOrNumber).Uint64()
		header = api.blockchain.GetHeaderByNumber(blockNumber)
		block = api.blockchain.GetBlockByNumber(blockNumber)
		//fmt.Printf("Storage range: %d, txIndex %d, addr: %x, start: %x, maxResults: %d\n",
		//	blockNumber, txIndex, address, common.BigToHash((*big.Int)(begin)), maxResults)
	}
	parentHeader := api.blockchain.GetHeaderByHash(header.ParentHash)
	var root common.Hash
	var statedb *state.StateDB
	var err error
	if parentHeader == nil || int(txIndex) >= len(block.Transactions()) {
		root = header.Root
		statedb, err = api.blockchain.StateAt(root)
		if err != nil {
			return StorageRangeResult{}, err
		}
	} else {
		root = parentHeader.Root
		statedb, err = api.blockchain.StateAt(root)
		if err != nil {
			return StorageRangeResult{}, err
		}
		// Recompute transactions up to the target index.
		signer := types.MakeSigner(api.blockchain.Config().NetworkID)
		for idx, tx := range block.Transactions() {
			// Assemble the transaction call message and return if the requested offset
			msg, _ := tx.AsMessage(signer)
			context := core.NewCVMContext(msg, block.Header(), api.blockchain, nil)
			// Not yet the searched for transaction, execute on top of the current state
			vmenv := vm.NewCVM(context, statedb, api.blockchain.Config(), vm.Config{})
			if _, _, _, err := core.ApplyMessage(vmenv, msg, new(core.EnergyPool).AddEnergy(tx.Energy())); err != nil {
				return StorageRangeResult{}, fmt.Errorf("transaction %#x failed: %v", tx.Hash(), err)
			}
			// Ensure any modifications are committed to the state
			_ = statedb.IntermediateRoot(true)
			if idx == int(txIndex) {
				// This is to make sure root can be opened by OpenTrie
				_, err = statedb.Commit(true)
				if err != nil {
					return StorageRangeResult{}, err
				}
			}
		}
	}
	storageTrie := statedb.StorageTrie(address)
	it := trie.NewIterator(storageTrie.NodeIterator(common.BigToHash((*big.Int)(begin)).Bytes()))
	result := StorageRangeResult{Storage: make(map[common.Hash]SRItem)}
	for i := 0; /*i < int(maxResults) && */ it.Next(); i++ {
		if preimage := storageTrie.GetKey(it.Key); preimage != nil {
			key := (*math.HexOrDecimal256)(big.NewInt(0).SetBytes(preimage))
			v, _, err := rlp.SplitString(it.Value)
			if err != nil {
				return StorageRangeResult{}, err
			}
			value := (*math.HexOrDecimal256)(big.NewInt(0).SetBytes(v))
			ks, _ := key.MarshalText()
			vs, _ := value.MarshalText()
			if len(ks)%2 != 0 {
				ks = append(append(append([]byte{}, ks[:2]...), byte('0')), ks[2:]...)
			}
			if len(vs)%2 != 0 {
				vs = append(append(append([]byte{}, vs[:2]...), byte('0')), vs[2:]...)
			}
			result.Storage[common.BytesToHash(it.Key)] = SRItem{
				Key:   string(ks),
				Value: string(vs),
			}
		}
	}
	if it.Next() {
		result.Complete = false
	} else {
		result.Complete = true
	}
	return result, nil
}

func (api *RetestxcbAPI) ClientVersion(ctx context.Context) (string, error) {
	return "Gocore-" + params.VersionWithTag(gitTag, gitCommit, gitDate), nil
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}

func retestxcb(ctx *cli.Context) error {
	log.Info("Welcome to retestxcb!")
	// register signer API with server
	var (
		extapiURL string
	)
	apiImpl := &RetestxcbAPI{}
	var testApi RetestxcbTestAPI = apiImpl
	var xcbApi RetestxcbXcbAPI = apiImpl
	var debugApi RetestxcbDebugAPI = apiImpl
	var web3Api RetestWeb3API = apiImpl
	rpcAPI := []rpc.API{
		{
			Namespace: "test",
			Public:    true,
			Service:   testApi,
			Version:   "1.0",
		},
		{
			Namespace: "xcb",
			Public:    true,
			Service:   xcbApi,
			Version:   "1.0",
		},
		{
			Namespace: "debug",
			Public:    true,
			Service:   debugApi,
			Version:   "1.0",
		},
		{
			Namespace: "web3",
			Public:    true,
			Service:   web3Api,
			Version:   "1.0",
		},
	}
	vhosts := splitAndTrim(ctx.GlobalString(utils.HTTPVirtualHostsFlag.Name))
	cors := splitAndTrim(ctx.GlobalString(utils.HTTPCORSDomainFlag.Name))
	// register apis and create handler stack
	srv := rpc.NewServer()
	err := node.RegisterApisFromWhitelist(rpcAPI, []string{"test", "eth", "debug", "web3"}, srv, false)
	if err != nil {
		utils.Fatalf("Could not register RPC apis: %w", err)
	}
	handler := node.NewHTTPHandlerStack(srv, cors, vhosts)

	// start http server
	var RetestethHTTPTimeouts = rpc.HTTPTimeouts{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	httpEndpoint := fmt.Sprintf("%s:%d", ctx.GlobalString(utils.RPCListenAddrFlag.Name), ctx.Int(rpcPortFlag.Name))
	listener, err := node.StartHTTPEndpoint(httpEndpoint, RetestethHTTPTimeouts, handler)
	if err != nil {
		utils.Fatalf("Could not start RPC api: %v", err)
	}
	extapiURL = fmt.Sprintf("http://%s", httpEndpoint)
	log.Info("HTTP endpoint opened", "url", extapiURL)

	defer func() {
		listener.Close()
		log.Info("HTTP endpoint closed", "url", httpEndpoint)
	}()

	abortChan := make(chan os.Signal, 11)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan
	log.Info("Exiting...", "signal", sig)
	return nil
}

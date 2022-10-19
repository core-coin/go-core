// Copyright 2016 by the Authors
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

package bind

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/core-coin/go-core/common"
)

var bindTests = []struct {
	name     string
	contract string
	bytecode []string
	abi      []string
	imports  string
	tester   string
	fsigs    []map[string]string
	libs     map[string]string
	aliases  map[string]string
	types    []string
}{
	// Test that the binding is available in combined and separate forms too
	{
		`Empty`,
		`contract NilContract {}`,
		[]string{`606060405260068060106000396000f3606060405200`},
		[]string{`[]`},
		`"github.com/core-coin/go-core/common"`,
		`
			if b, err := NewEmpty(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("combined binding (%v) nil or error (%v) not nil", b, nil)
			}
			if b, err := NewEmptyCaller(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("caller binding (%v) nil or error (%v) not nil", b, nil)
			}
			if b, err := NewEmptyTransactor(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("transactor binding (%v) nil or error (%v) not nil", b, nil)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Test that all the official sample contracts bind correctly
	{
		`Token`,
		`https://coreblockchain.cc/token`,
		[]string{`60606040526040516107fd3803806107fd83398101604052805160805160a05160c051929391820192909101600160a060020a0333166000908152600360209081526040822086905581548551838052601f6002600019610100600186161502019093169290920482018390047f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e56390810193919290918801908390106100e857805160ff19168380011785555b506101189291505b8082111561017157600081556001016100b4565b50506002805460ff19168317905550505050610658806101a56000396000f35b828001600101855582156100ac579182015b828111156100ac5782518260005055916020019190600101906100fa565b50508060016000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061017557805160ff19168380011785555b506100c89291506100b4565b5090565b82800160010185558215610165579182015b8281111561016557825182600050559160200191906001019061018756606060405236156100775760e060020a600035046306fdde03811461007f57806323b872dd146100dc578063313ce5671461010e57806370a082311461011a57806395d89b4114610132578063a9059cbb1461018e578063cae9ca51146101bd578063dc3080f21461031c578063dd62ed3e14610341575b610365610002565b61036760008054602060026001831615610100026000190190921691909104601f810182900490910260809081016040526060828152929190828280156104eb5780601f106104c0576101008083540402835291602001916104eb565b6103d5600435602435604435600160a060020a038316600090815260036020526040812054829010156104f357610002565b6103e760025460ff1681565b6103d560043560036020526000908152604090205481565b610367600180546020600282841615610100026000190190921691909104601f810182900490910260809081016040526060828152929190828280156104eb5780601f106104c0576101008083540402835291602001916104eb565b610365600435602435600160a060020a033316600090815260036020526040902054819010156103f157610002565b60806020604435600481810135601f8101849004909302840160405260608381526103d5948235946024803595606494939101919081908382808284375094965050505050505060006000836004600050600033600160a060020a03168152602001908152602001600020600050600087600160a060020a031681526020019081526020016000206000508190555084905080600160a060020a0316638f4ffcb1338630876040518560e060020a0281526004018085600160a060020a0316815260200184815260200183600160a060020a03168152602001806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f1680156102f25780820380516001836020036101000a031916815260200191505b50955050505050506000604051808303816000876161da5a03f11561000257505050509392505050565b6005602090815260043560009081526040808220909252602435815220546103d59081565b60046020818152903560009081526040808220909252602435815220546103d59081565b005b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f1680156103c75780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60408051918252519081900360200190f35b6060908152602090f35b600160a060020a03821660009081526040902054808201101561041357610002565b806003600050600033600160a060020a03168152602001908152602001600020600082828250540392505081905550806003600050600084600160a060020a0316815260200190815260200160002060008282825054019250508190555081600160a060020a031633600160a060020a03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a35050565b820191906000526020600020905b8154815290600101906020018083116104ce57829003601f168201915b505050505081565b600160a060020a03831681526040812054808301101561051257610002565b600160a060020a0380851680835260046020908152604080852033949094168086529382528085205492855260058252808520938552929052908220548301111561055c57610002565b816003600050600086600160a060020a03168152602001908152602001600020600082828250540392505081905550816003600050600085600160a060020a03168152602001908152602001600020600082828250540192505081905550816005600050600086600160a060020a03168152602001908152602001600020600050600033600160a060020a0316815260200190815260200160002060008282825054019250508190555082600160a060020a031633600160a060020a03167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a3939250505056`},
		[]string{`[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"success","type":"bool"}],"type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"},{"name":"_extraData","type":"bytes"}],"name":"approveAndCall","outputs":[{"name":"success","type":"bool"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"spentAllowance","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"},{"name":"","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"inputs":[{"name":"initialSupply","type":"uint256"},{"name":"tokenName","type":"string"},{"name":"decimalUnits","type":"uint8"},{"name":"tokenSymbol","type":"string"}],"type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`},
		`"github.com/core-coin/go-core/common"`,
		`
			if b, err := NewToken(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	{
		`Crowdsale`,
		`https://coreblockchain.cc/crowdsale`,
		[]string{`606060408190526007805460ff1916905560a0806105a883396101006040529051608051915160c05160e05160008054600160a060020a03199081169095178155670de0b6b3a7640000958602600155603c9093024201600355930260045560058054909216909217905561052f90819061007990396000f36060604052361561006c5760e060020a600035046301cb3b20811461008257806329dcb0cf1461014457806338af3eed1461014d5780636e66f6e91461015f5780637a3a0e84146101715780637b3e5e7b1461017a578063a035b1fe14610183578063dc0d3dff1461018c575b61020060075460009060ff161561032357610002565b61020060035460009042106103205760025460015490106103cb576002548154600160a060020a0316908290606082818181858883f150915460025460408051600160a060020a039390931683526020830191909152818101869052517fe842aea7a5f1b01049d752008c53c52890b1a6daf660cf39e8eec506112bbdf6945090819003909201919050a15b60405160008054600160a060020a039081169230909116319082818181858883f150506007805460ff1916600117905550505050565b6103a160035481565b6103ab600054600160a060020a031681565b6103ab600554600160a060020a031681565b6103a160015481565b6103a160025481565b6103a160045481565b6103be60043560068054829081101561000257506000526002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d3f8101547ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d409190910154600160a060020a03919091169082565b005b505050815481101561000257906000526020600020906002020160005060008201518160000160006101000a815481600160a060020a030219169083021790555060208201518160010160005055905050806002600082828250540192505081905550600560009054906101000a9004600160a060020a0316600160a060020a031663a9059cbb3360046000505484046040518360e060020a0281526004018083600160a060020a03168152602001828152602001925050506000604051808303816000876161da5a03f11561000257505060408051600160a060020a03331681526020810184905260018183015290517fe842aea7a5f1b01049d752008c53c52890b1a6daf660cf39e8eec506112bbdf692509081900360600190a15b50565b5060a0604052336060908152346080819052600680546001810180835592939282908280158290116102025760020281600202836000526020600020918201910161020291905b8082111561039d57805473ffffffffffffffffffffffffffffffffffffffff19168155600060019190910190815561036a565b5090565b6060908152602090f35b600160a060020a03166060908152602090f35b6060918252608052604090f35b5b60065481101561010e576006805482908110156100025760009182526002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d3f0190600680549254600160a060020a0316928490811015610002576002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d40015460405190915082818181858883f19350505050507fe842aea7a5f1b01049d752008c53c52890b1a6daf660cf39e8eec506112bbdf660066000508281548110156100025760008290526002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d3f01548154600160a060020a039190911691908490811015610002576002027ff652222313e28459528d920b65115c16c04f3efc82aaedc97be59f3f377c0d40015460408051600160a060020a0394909416845260208401919091526000838201525191829003606001919050a16001016103cc56`},
		[]string{`[{"constant":false,"inputs":[],"name":"checkGoalReached","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"deadline","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"beneficiary","outputs":[{"name":"","type":"address"}],"type":"function"},{"constant":true,"inputs":[],"name":"tokenReward","outputs":[{"name":"","type":"address"}],"type":"function"},{"constant":true,"inputs":[],"name":"fundingGoal","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"amountRaised","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"price","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"uint256"}],"name":"funders","outputs":[{"name":"addr","type":"address"},{"name":"amount","type":"uint256"}],"type":"function"},{"inputs":[{"name":"ifSuccessfulSendTo","type":"address"},{"name":"fundingGoalInCores","type":"uint256"},{"name":"durationInMinutes","type":"uint256"},{"name":"coreCostOfEachToken","type":"uint256"},{"name":"addressOfTokenUsedAsReward","type":"address"}],"type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"backer","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"isContribution","type":"bool"}],"name":"FundTransfer","type":"event"}]`},
		`"github.com/core-coin/go-core/common"`,
		`
			if b, err := NewCrowdsale(common.Address{}, nil); b == nil || err != nil {
				t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Test that named and anonymous inputs are handled correctly
	{
		`InputChecker`, ``, []string{``},
		[]string{`
			[
				{"type":"function","name":"noInput","constant":true,"inputs":[],"outputs":[]},
				{"type":"function","name":"namedInput","constant":true,"inputs":[{"name":"str","type":"string"}],"outputs":[]},
				{"type":"function","name":"anonInput","constant":true,"inputs":[{"name":"","type":"string"}],"outputs":[]},
				{"type":"function","name":"namedInputs","constant":true,"inputs":[{"name":"str1","type":"string"},{"name":"str2","type":"string"}],"outputs":[]},
				{"type":"function","name":"anonInputs","constant":true,"inputs":[{"name":"","type":"string"},{"name":"","type":"string"}],"outputs":[]},
				{"type":"function","name":"mixedInputs","constant":true,"inputs":[{"name":"","type":"string"},{"name":"str","type":"string"}],"outputs":[]}
			]
		`},
		`
			"fmt"

			"github.com/core-coin/go-core/common"
		`,
		`if b, err := NewInputChecker(common.Address{}, nil); b == nil || err != nil {
			 t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
		 } else if false { // Don't run, just compile and test types
			 var err error

			 err = b.NoInput(nil)
			 err = b.NamedInput(nil, "")
			 err = b.AnonInput(nil, "")
			 err = b.NamedInputs(nil, "", "")
			 err = b.AnonInputs(nil, "", "")
			 err = b.MixedInputs(nil, "", "")

			 fmt.Println(err)
		 }`,
		nil,
		nil,
		nil,
		nil,
	},
	// Test that named and anonymous outputs are handled correctly
	{
		`OutputChecker`, ``, []string{``},
		[]string{`
			[
				{"type":"function","name":"noOutput","constant":true,"inputs":[],"outputs":[]},
				{"type":"function","name":"namedOutput","constant":true,"inputs":[],"outputs":[{"name":"str","type":"string"}]},
				{"type":"function","name":"anonOutput","constant":true,"inputs":[],"outputs":[{"name":"","type":"string"}]},
				{"type":"function","name":"namedOutputs","constant":true,"inputs":[],"outputs":[{"name":"str1","type":"string"},{"name":"str2","type":"string"}]},
				{"type":"function","name":"collidingOutputs","constant":true,"inputs":[],"outputs":[{"name":"str","type":"string"},{"name":"Str","type":"string"}]},
				{"type":"function","name":"anonOutputs","constant":true,"inputs":[],"outputs":[{"name":"","type":"string"},{"name":"","type":"string"}]},
				{"type":"function","name":"mixedOutputs","constant":true,"inputs":[],"outputs":[{"name":"","type":"string"},{"name":"str","type":"string"}]}
			]
		`},
		`
			"fmt"

			"github.com/core-coin/go-core/common"
		`,
		`if b, err := NewOutputChecker(common.Address{}, nil); b == nil || err != nil {
			 t.Fatalf("binding (%v) nil or error (%v) not nil", b, nil)
		 } else if false { // Don't run, just compile and test types
			 var str1, str2 string
			 var err error

			 err              = b.NoOutput(nil)
			 str1, err        = b.NamedOutput(nil)
			 str1, err        = b.AnonOutput(nil)
			 res, _          := b.NamedOutputs(nil)
			 str1, str2, err  = b.CollidingOutputs(nil)
			 str1, str2, err  = b.AnonOutputs(nil)
			 str1, str2, err  = b.MixedOutputs(nil)

			 fmt.Println(str1, str2, res.Str1, res.Str2, err)
		 }`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that named, anonymous and indexed events are handled correctly
	{
		`EventChecker`, ``, []string{``},
		[]string{`
			[
				{"type":"event","name":"empty","inputs":[]},
				{"type":"event","name":"indexed","inputs":[{"name":"addr","type":"address","indexed":true},{"name":"num","type":"int256","indexed":true}]},
				{"type":"event","name":"mixed","inputs":[{"name":"addr","type":"address","indexed":true},{"name":"num","type":"int256"}]},
				{"type":"event","name":"anonymous","anonymous":true,"inputs":[]},
				{"type":"event","name":"dynamic","inputs":[{"name":"idxStr","type":"string","indexed":true},{"name":"idxDat","type":"bytes","indexed":true},{"name":"str","type":"string"},{"name":"dat","type":"bytes"}]},
				{"type":"event","name":"unnamed","inputs":[{"name":"","type":"uint256","indexed": true},{"name":"","type":"uint256","indexed":true}]}
			]
		`},
		`
			"fmt"
			"math/big"
			"reflect"

			"github.com/core-coin/go-core/common"
		`,
		`if e, err := NewEventChecker(common.Address{}, nil); e == nil || err != nil {
			 t.Fatalf("binding (%v) nil or error (%v) not nil", e, nil)
		 } else if false { // Don't run, just compile and test types
			 var (
				 err  error
			   res  bool
				 str  string
				 dat  []byte
				 hash common.Hash
			 )
			 _, err = e.FilterEmpty(nil)
			 _, err = e.FilterIndexed(nil, []common.Address{}, []*big.Int{})

			 mit, err := e.FilterMixed(nil, []common.Address{})

			 res = mit.Next()  // Make sure the iterator has a Next method
			 err = mit.Error() // Make sure the iterator has an Error method
			 err = mit.Close() // Make sure the iterator has a Close method

			 fmt.Println(mit.Event.Raw.BlockHash) // Make sure the raw log is contained within the results
			 fmt.Println(mit.Event.Num)           // Make sure the unpacked non-indexed fields are present
			 fmt.Println(mit.Event.Addr)          // Make sure the reconstructed indexed fields are present

			 dit, err := e.FilterDynamic(nil, []string{}, [][]byte{})

			 str  = dit.Event.Str    // Make sure non-indexed strings retain their type
			 dat  = dit.Event.Dat    // Make sure non-indexed bytes retain their type
			 hash = dit.Event.IdxStr // Make sure indexed strings turn into hashes
			 hash = dit.Event.IdxDat // Make sure indexed bytes turn into hashes

			 sink := make(chan *EventCheckerMixed)
			 sub, err := e.WatchMixed(nil, sink, []common.Address{})
			 defer sub.Unsubscribe()

			 event := <-sink
			 fmt.Println(event.Raw.BlockHash) // Make sure the raw log is contained within the results
			 fmt.Println(event.Num)           // Make sure the unpacked non-indexed fields are present
			 fmt.Println(event.Addr)          // Make sure the reconstructed indexed fields are present

			 fmt.Println(res, str, dat, hash, err)

			 oit, err := e.FilterUnnamed(nil, []*big.Int{}, []*big.Int{})
			 arg0  := oit.Event.Arg0    // Make sure unnamed arguments are handled correctly
			 arg1  := oit.Event.Arg1    // Make sure unnamed arguments are handled correctly
			 fmt.Println(arg0, arg1)
		 }
		 // Run a tiny reflection test to ensure disallowed methods don't appear
		 if _, ok := reflect.TypeOf(&EventChecker{}).MethodByName("FilterAnonymous"); ok {
		 	t.Errorf("binding has disallowed method (FilterAnonymous)")
		 }`,
		nil,
		nil,
		nil,
		nil,
	},
	// Test that contract interactions (deploy, transact and call) generate working code
	{
		`Interactor`,
		`
			contract Interactor {
				string public deployString;
				string public transactString;

				function Interactor(string str) {
				  deployString = str;
				}

				function transact(string str) {
				  transactString = str;
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b506040516105d63803806105d68339818101604052602081101561003357600080fd5b810190808051604051939291908464010000000082111561005357600080fd5b8382019150602082018581111561006957600080fd5b825186600182028301116401000000008211171561008657600080fd5b8083526020830192505050908051906020019080838360005b838110156100ba57808201518184015260208101905061009f565b50505050905090810190601f1680156100e75780820380516001836020036101000a031916815260200191505b50604052505050806000908051906020019061010492919061010b565b50506101b0565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061014c57805160ff191683800117855561017a565b8280016001018555821561017a579182015b8281111561017957825182559160200191906001019061015e565b5b509050610187919061018b565b5090565b6101ad91905b808211156101a9576000816000905550600101610191565b5090565b90565b610417806101bf6000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c806305af2c4414610046578063b50b8137146100c9578063e53adee614610142575b600080fd5b61004e6101c5565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561008e578082015181840152602081019050610073565b50505050905090810190601f1680156100bb5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b610140600480360360208110156100df57600080fd5b81019080803590602001906401000000008111156100fc57600080fd5b82018360208201111561010e57600080fd5b8035906020019184600183028401116401000000008311171561013057600080fd5b9091929391929390505050610263565b005b61014a610279565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561018a57808201518184015260208101905061016f565b50505050905090810190601f1680156101b75780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60008054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561025b5780601f106102305761010080835404028352916020019161025b565b820191906000526020600020905b81548152906001019060200180831161023e57829003601f168201915b505050505081565b818160019190610274929190610317565b505050565b60018054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561030f5780601f106102e45761010080835404028352916020019161030f565b820191906000526020600020905b8154815290600101906020018083116102f257829003601f168201915b505050505081565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061035857803560ff1916838001178555610386565b82800160010185558215610386579182015b8281111561038557823582559160200191906001019061036a565b5b5090506103939190610397565b5090565b6103b991905b808211156103b557600081600090555060010161039d565b5090565b9056fea2646970667358221220fe9fbbc6f5583d4eb2da05b0eb9b416d01de4d65956935093567097cb56882f364736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"inputs":[{"internalType":"string","name":"str","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"deployString","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"str","type":"string"}],"name":"transact","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"transactString","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"}]`},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy an interaction tester contract and call a transaction on it
			_, _, interactor, err := DeployInteractor(auth, sim, "Deploy string")
			if err != nil {
				t.Fatalf("Failed to deploy interactor contract: %v", err)
			}
			if _, err := interactor.Transact(auth, "Transact string"); err != nil {
				t.Fatalf("Failed to transact with interactor contract: %v", err)
			}
			// Commit all pending transactions in the simulator and check the contract state
			sim.Commit()

			if str, err := interactor.DeployString(nil); err != nil {
				t.Fatalf("Failed to retrieve deploy string: %v", err)
			} else if str != "Deploy string" {
				t.Fatalf("Deploy string mismatch: have '%s', want 'Deploy string'", str)
			}
			if str, err := interactor.TransactString(nil); err != nil {
				t.Fatalf("Failed to retrieve transact string: %v", err)
			} else if str != "Transact string" {
				t.Fatalf("Transact string mismatch: have '%s', want 'Transact string'", str)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that plain values can be properly returned and deserialized
	{
		`Getter`,
		`
			contract Getter {
				function getter() constant returns (string, int, bytes32) {
					return ("Hi", 1, sha3(""));
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b5061017a806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80632c149b2414610030575b600080fd5b6100386100c1565b6040518080602001848152602001838152602001828103825285818151815260200191508051906020019080838360005b83811015610084578082015181840152602081019050610069565b50505050905090810190601f1680156100b15780820380516001836020036101000a031916815260200191505b5094505050505060405180910390f35b6060600080600160405180600001905060405180910390206040518060400160405280600281526020017f4869000000000000000000000000000000000000000000000000000000000000815250919081915092509250925090919256fea26469706673582212209c2f29b00aec02b8784ec578fae7fbcc31cbb48fc7a131cb9d9ade670f0e4ee164736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"constant":true,"inputs":[],"name":"getter","outputs":[{"name":"","type":"string"},{"name":"","type":"int256"},{"name":"","type":"bytes32"}],"type":"function"}]`},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a tuple tester contract and execute a structured call on it
			_, _, getter, err := DeployGetter(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy getter contract: %v", err)
			}
			sim.Commit()

			if str, num, _, err := getter.Getter(nil); err != nil {
				t.Fatalf("Failed to call anonymous field retriever: %v", err)
			} else if str != "Hi" || num.Cmp(big.NewInt(1)) != 0 {
				t.Fatalf("Retrieved value mismatch: have %v/%v, want %v/%v", str, num, "Hi", 1)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that tuples can be properly returned and deserialized
	{
		`Tupler`,
		`
			contract Tupler {
				function tuple() constant returns (string a, int b, bytes32 c) {
					return ("Hi", 1, sha3(""));
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b5061017a806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80636147673f14610030575b600080fd5b6100386100c1565b6040518080602001848152602001838152602001828103825285818151815260200191508051906020019080838360005b83811015610084578082015181840152602081019050610069565b50505050905090810190601f1680156100b15780820380516001836020036101000a031916815260200191505b5094505050505060405180910390f35b6060600080600160405180600001905060405180910390206040518060400160405280600281526020017f4869000000000000000000000000000000000000000000000000000000000000815250919081915092509250925090919256fea26469706673582212204a42146ac40edce5673003b66f9e14ebf3abd2508008a6a0d0ecfecdd4c2f56f64736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"inputs":[],"name":"tuple","outputs":[{"internalType":"string","name":"a","type":"string"},{"internalType":"int256","name":"b","type":"int256"},{"internalType":"bytes32","name":"c","type":"bytes32"}],"stateMutability":"view","type":"function"}]`},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a tuple tester contract and execute a structured call on it
			_, _, tupler, err := DeployTupler(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy tupler contract: %v", err)
			}
			sim.Commit()

			if res, err := tupler.Tuple(nil); err != nil {
				t.Fatalf("Failed to call structure retriever: %v", err)
			} else if res.A != "Hi" || res.B.Cmp(big.NewInt(1)) != 0 {
				t.Fatalf("Retrieved value mismatch: have %v/%v, want %v/%v", res.A, res.B, "Hi", 1)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that arrays/slices can be properly returned and deserialized.
	// Only addresses are tested, remainder just compiled to keep the test small.
	{
		`Slicer`,
		`
			contract Slicer {
				function echoAddresses(address[] input) constant returns (address[] output) {
					return input;
				}
				function echoInts(int[] input) constant returns (int[] output) {
					return input;
				}
				function echoFancyInts(uint24[23] input) constant returns (uint24[23] output) {
					return input;
				}
				function echoBools(bool[] input) constant returns (bool[] output) {
					return input;
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b50610375806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80632d0473421461005157806366fee4701461010c578063707f9239146101c757806398b4388514610282575b600080fd5b6100c86004803603602081101561006757600080fd5b810190808035906020019064010000000081111561008457600080fd5b82018360208201111561009657600080fd5b803590602001918460208302840111640100000000831117156100b857600080fd5b90919293919293905050506102e1565b60405180806020018281038252848482818152602001925060200280828437600081840152601f19601f820116905080830192505050935050505060405180910390f35b6101836004803603602081101561012257600080fd5b810190808035906020019064010000000081111561013f57600080fd5b82018360208201111561015157600080fd5b8035906020019184602083028401116401000000008311171561017357600080fd5b90919293919293905050506102f1565b60405180806020018281038252848482818152602001925060200280828437600081840152601f19601f820116905080830192505050935050505060405180910390f35b61023e600480360360208110156101dd57600080fd5b81019080803590602001906401000000008111156101fa57600080fd5b82018360208201111561020c57600080fd5b8035906020019184602083028401116401000000008311171561022e57600080fd5b9091929391929390505050610301565b60405180806020018281038252848482818152602001925060200280828437600081840152601f19601f820116905080830192505050935050505060405180910390f35b6102b060048036036102e081101561029957600080fd5b81019080806102e001909192919290505050610311565b6040518082601760200280828437600081840152601f19601f82011690508083019250505091505060405180910390f35b3660008383915091509250929050565b3660008383915091509250929050565b3660008383915091509250929050565b3681905091905056fea2646970667358221220040db7f0f01def1525f4579447c9f0ef9d867163a4457522283d2b759ecc856d64736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"inputs":[{"internalType":"address[]","name":"input","type":"address[]"}],"name":"echoAddresses","outputs":[{"internalType":"address[]","name":"output","type":"address[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bool[]","name":"input","type":"bool[]"}],"name":"echoBools","outputs":[{"internalType":"bool[]","name":"output","type":"bool[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint24[23]","name":"input","type":"uint24[23]"}],"name":"echoFancyInts","outputs":[{"internalType":"uint24[23]","name":"output","type":"uint24[23]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"int256[]","name":"input","type":"int256[]"}],"name":"echoInts","outputs":[{"internalType":"int256[]","name":"output","type":"int256[]"}],"stateMutability":"view","type":"function"}]`},
		`
			"math/big"
			"reflect"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/common"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a slice tester contract and execute a n array call on it
			_, _, slicer, err := DeploySlicer(auth, sim)
			if err != nil {
					t.Fatalf("Failed to deploy slicer contract: %v", err)
			}
			sim.Commit()

			if out, err := slicer.EchoAddresses(nil, []common.Address{auth.From, common.Address{}}); err != nil {
					t.Fatalf("Failed to call slice echoer: %v", err)
			} else if !reflect.DeepEqual(out, []common.Address{auth.From, common.Address{}}) {
					t.Fatalf("Slice return mismatch: have %v, want %v", out, []common.Address{auth.From, common.Address{}})
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that anonymous default methods can be correctly invoked
	{
		`Defaulter`,
		`
			contract Defaulter {
				address public caller;

				function() {
					caller = msg.sender;
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b50610148806100206000396000f3fe608060405260043610601f5760003560e01c806354062aff14606e576069565b36606957336000806101000a81548175ffffffffffffffffffffffffffffffffffffffffffff021916908375ffffffffffffffffffffffffffffffffffffffffffff160217905550005b600080fd5b348015607957600080fd5b50608060c6565b604051808275ffffffffffffffffffffffffffffffffffffffffffff1675ffffffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000809054906101000a900475ffffffffffffffffffffffffffffffffffffffffffff168156fea26469706673582212200efb1676cee26bac51a701c165e2e38ecc5f0c315203e7c752648ca986a48d8164736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"inputs":[],"name":"caller","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"stateMutability":"payable","type":"receive"}]`},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a default method invoker contract and execute its default method
			_, _, defaulter, err := DeployDefaulter(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy defaulter contract: %v", err)
			}
			if _, err := (&DefaulterRaw{defaulter}).Transfer(auth); err != nil {
				t.Fatalf("Failed to invoke default method: %v", err)
			}
			sim.Commit()

			if caller, err := defaulter.Caller(nil); err != nil {
				t.Fatalf("Failed to call address retriever: %v", err)
			} else if (caller != auth.From) {
				t.Fatalf("Address mismatch: have %s, want %s", caller.Hex(), auth.From.Hex())
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that non-existent contracts are reported as such (though only simulator test)
	{
		`NonExistent`,
		`
			contract NonExistent {
				function String() constant returns(string) {
					return "I don't exist";
				}
			}
		`,
		[]string{`6060604052609f8060106000396000f3606060405260e060020a6000350463f97a60058114601a575b005b600060605260c0604052600d60809081527f4920646f6e27742065786973740000000000000000000000000000000000000060a052602060c0908152600d60e081905281906101009060a09080838184600060046012f15050815172ffffffffffffffffffffffffffffffffffffff1916909152505060405161012081900392509050f3`},
		[]string{`[{"constant":true,"inputs":[],"name":"String","outputs":[{"name":"","type":"string"}],"type":"function"}]`},
		`
			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/common"
			"github.com/core-coin/go-core/core"
		`,
		`
			// Create a simulator and wrap a non-deployed contract

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{}, uint64(10000000000))
			defer sim.Close()

			nonexistent, err := NewNonExistent(common.Address{}, sim)
			if err != nil {
				t.Fatalf("Failed to access non-existent contract: %v", err)
			}
			// Ensure that contract calls fail with the appropriate error
			if res, err := nonexistent.String(nil); err == nil {
				t.Fatalf("Call succeeded on non-existent contract: %v", res)
			} else if (err != bind.ErrNoCode) {
				t.Fatalf("Error mismatch: have %v, want %v", err, bind.ErrNoCode)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that energy estimation works for contracts with weird energy mechanics too.
	{
		`FunkyEnergyPattern`,
		`
			contract FunkyEnergyPattern {
				string public field;

				function SetField(string value) {
					// This check will screw energy estimation! Good, good!
					if (msg.energy < 100000) {
						throw;
					}
					field = value;
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b506102fb806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80635f10819c1461003b578063944e3cd2146100be575b600080fd5b610043610137565b6040518080602001828103825283818151815260200191508051906020019080838360005b83811015610083578082015181840152602081019050610068565b50505050905090810190601f1680156100b05780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b610135600480360360208110156100d457600080fd5b81019080803590602001906401000000008111156100f157600080fd5b82018360208201111561010357600080fd5b8035906020019184600183028401116401000000008311171561012557600080fd5b90919293919293905050506101d5565b005b60008054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156101cd5780601f106101a2576101008083540402835291602001916101cd565b820191906000526020600020905b8154815290600101906020018083116101b057829003601f168201915b505050505081565b620186a05a10156101e557600080fd5b8181600091906101f69291906101fb565b505050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061023c57803560ff191683800117855561026a565b8280016001018555821561026a579182015b8281111561026957823582559160200191906001019061024e565b5b509050610277919061027b565b5090565b61029d91905b80821115610299576000816000905550600101610281565b5090565b9056fea26469706673582212201b44f850fb5acb3fc3e391ce7fbae403bc47615e6faf470c82f419047d674cd564736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"inputs":[{"internalType":"string","name":"value","type":"string"}],"name":"SetField","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"field","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"}]`},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a funky energy pattern contract
			_, _, limiter, err := DeployFunkyEnergyPattern(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy funky contract: %v", err)
			}
			sim.Commit()

			// Set the field with automatic estimation and check that it succeeds
			if _, err := limiter.SetField(auth, "automatic"); err != nil {
				t.Fatalf("Failed to call automatically energyed transaction: %v", err)
			}
			sim.Commit()

			if field, _ := limiter.Field(nil); field != "automatic" {
				t.Fatalf("Field mismatch: have %v, want %v", field, "automatic")
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Test that constant functions can be called from an (optional) specified address
	{
		`CallFrom`,
		`
			contract CallFrom {
				function callFrom() constant returns(address) {
					return msg.sender;
				}
			}
		`, []string{`608060405234801561001057600080fd5b5060dc8061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063ebdaa57614602d575b600080fd5b60336079565b604051808275ffffffffffffffffffffffffffffffffffffffffffff1675ffffffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60003390509056fea2646970667358221220779b05929c101dfbece322a04ab710112ec8900380f839eaf669d52934a9d33d64736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"inputs":[],"name":"callFrom","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`},
		`	
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/common"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a sender tester contract and execute a structured call on it
			_, _, callfrom, err := DeployCallFrom(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy sender contract: %v", err)
			}
			sim.Commit()

			if res, err := callfrom.CallFrom(nil); err != nil {
				t.Errorf("Failed to call constant function: %v", err)
			} else if res != (common.Address{}) {
				t.Errorf("Invalid address returned, want: %x, got: %x", (common.Address{}), res)
			}
			addr1, _ := common.HexToAddress("cb540000000000000000000000000000000000000000")
			addr2, _ := common.HexToAddress("cb270000000000000000000000000000000000000001")
			addr3, _ := common.HexToAddress("cb970000000000000000000000000000000000000002")
			for _, addr := range []common.Address{addr1,addr2,addr3} {
				if res, err := callfrom.CallFrom(&bind.CallOpts{From: addr}); err != nil {
					t.Fatalf("Failed to call constant function: %v", err)
				} else if res != addr {
					t.Fatalf("Invalid address returned, want: %x, got: %x", addr, res)
				}
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that methods and returns with underscores inside work correctly.
	{
		`Underscorer`,
		`
		contract Underscorer {
			function UnderscoredOutput() constant returns (int _int, string _string) {
				return (314, "pi");
			}
			function LowerLowerCollision() constant returns (int _res, int res) {
				return (1, 2);
			}
			function LowerUpperCollision() constant returns (int _res, int Res) {
				return (1, 2);
			}
			function UpperLowerCollision() constant returns (int _Res, int res) {
				return (1, 2);
			}
			function UpperUpperCollision() constant returns (int _Res, int Res) {
				return (1, 2);
			}
			function PurelyUnderscoredOutput() constant returns (int _, int res) {
				return (1, 2);
			}
			function AllPurelyUnderscoredOutput() constant returns (int _, int __) {
				return (1, 2);
			}
			function _under_scored_func() constant returns (int _int) {
				return 0;
			}
		}
		`, []string{`608060405234801561001057600080fd5b5061033c806100206000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c806334256b111161005b57806334256b111461011a57806373d501f81461013f5780638a1f5b3714610164578063f05f7df3146101ee57610088565b80630abdd04d1461008d57806324356ce1146100b25780632a5e10e6146100d05780632cc8ce31146100f5575b600080fd5b610095610213565b604051808381526020018281526020019250505060405180910390f35b6100ba610228565b6040518082815260200191505060405180910390f35b6100d8610230565b604051808381526020018281526020019250505060405180910390f35b6100fd610245565b604051808381526020018281526020019250505060405180910390f35b61012261025a565b604051808381526020018281526020019250505060405180910390f35b61014761026f565b604051808381526020018281526020019250505060405180910390f35b61016c610284565b6040518083815260200180602001828103825283818151815260200191508051906020019080838360005b838110156101b2578082015181840152602081019050610197565b50505050905090810190601f1680156101df5780820380516001836020036101000a031916815260200191505b50935050505060405180910390f35b6101f66102cc565b604051808381526020018281526020019250505060405180910390f35b60008060016002819150809050915091509091565b600080905090565b60008060016002819150809050915091509091565b60008060016002819150809050915091509091565b60008060016002819150809050915091509091565b60008060016002819150809050915091509091565b6000606061013a8090506040518060400160405280600281526020017f7069000000000000000000000000000000000000000000000000000000000000815250915091509091565b6000806001600281915080905091509150909156fea264697066735822122022eed0039b9944c4bd5969e8fbbed2e59430f2d6b797659aec4f6ff8d5ee21cb64736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"constant":true,"inputs":[],"name":"LowerUpperCollision","outputs":[{"name":"_res","type":"int256"},{"name":"Res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_under_scored_func","outputs":[{"name":"_int","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"UnderscoredOutput","outputs":[{"name":"_int","type":"int256"},{"name":"_string","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"PurelyUnderscoredOutput","outputs":[{"name":"_","type":"int256"},{"name":"res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"UpperLowerCollision","outputs":[{"name":"_Res","type":"int256"},{"name":"res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"AllPurelyUnderscoredOutput","outputs":[{"name":"_","type":"int256"},{"name":"__","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"UpperUpperCollision","outputs":[{"name":"_Res","type":"int256"},{"name":"Res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"LowerLowerCollision","outputs":[{"name":"_res","type":"int256"},{"name":"res","type":"int256"}],"payable":false,"stateMutability":"view","type":"function"}]`},
		`
			"fmt"
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a underscorer tester contract and execute a structured call on it
			_, _, underscorer, err := DeployUnderscorer(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy underscorer contract: %v", err)
			}
			sim.Commit()

			// Verify that underscored return values correctly parse into structs
			if res, err := underscorer.UnderscoredOutput(nil); err != nil {
				t.Errorf("Failed to call constant function: %v", err)
			} else if res.Int.Cmp(big.NewInt(314)) != 0 || res.String != "pi" {
				t.Errorf("Invalid result, want: {314, \"pi\"}, got: %+v", res)
			}
			// Verify that underscored and non-underscored name collisions force tuple outputs
			var a, b *big.Int

			a, b, _ = underscorer.LowerLowerCollision(nil)
			a, b, _ = underscorer.LowerUpperCollision(nil)
			a, b, _ = underscorer.UpperLowerCollision(nil)
			a, b, _ = underscorer.UpperUpperCollision(nil)
			a, b, _ = underscorer.PurelyUnderscoredOutput(nil)
			a, b, _ = underscorer.AllPurelyUnderscoredOutput(nil)
			a, _ = underscorer.UnderScoredFunc(nil)

			fmt.Println(a, b, err)
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Tests that logs can be successfully filtered and decoded.
	{
		`Eventer`,
		`
			contract Eventer {
				event SimpleEvent (
					address indexed Addr,
					bytes32 indexed Id,
					bool    indexed Flag,
					uint    Value
				);
				function raiseSimpleEvent(address addr, bytes32 id, bool flag, uint value) {
					SimpleEvent(addr, id, flag, value);
				}

				event NodataEvent (
					uint   indexed Number,
					int16  indexed Short,
					uint32 indexed Long
				);
				function raiseNodataEvent(uint number, int16 short, uint32 long) {
					NodataEvent(number, short, long);
				}

				event DynamicEvent (
					string indexed IndexedString,
					bytes  indexed IndexedBytes,
					string NonIndexedString,
					bytes  NonIndexedBytes
				);
				function raiseDynamicEvent(string str, bytes blob) {
					DynamicEvent(str, blob, str, blob);
				}

				event FixedBytesEvent (
					bytes24 indexed IndexedBytes,
					bytes24 NonIndexedBytes
				);
				function raiseFixedBytesEvent(bytes24 blob) {
					FixedBytesEvent(blob, blob);
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b50610432806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80630976d98314610051578063752dae501461008a5780639a00028c14610158578063e89703ba146101be575b600080fd5b6100886004803603602081101561006757600080fd5b81019080803567ffffffffffffffff19169060200190929190505050610209565b005b610156600480360360408110156100a057600080fd5b81019080803590602001906401000000008111156100bd57600080fd5b8201836020820111156100cf57600080fd5b803590602001918460018302840111640100000000831117156100f157600080fd5b90919293919293908035906020019064010000000081111561011257600080fd5b82018360208201111561012457600080fd5b8035906020019184600183028401116401000000008311171561014657600080fd5b9091929391929390505050610265565b005b6101bc6004803603608081101561016e57600080fd5b81019080803575ffffffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080351515906020019092919080359060200190929190505050610340565b005b610207600480360360608110156101d457600080fd5b8101908080359060200190929190803560010b9060200190929190803563ffffffff16906020019092919050505061039a565b005b8067ffffffffffffffff19167fb93054074d8a957aa255759770eb3375329d307d93093101da7f1f2676e1c56882604051808267ffffffffffffffff191667ffffffffffffffff1916815260200191505060405180910390a250565b81816040518083838082843780830192505050925050506040518091039020848460405180838380828437808301925050509250505060405180910390207f5c32738e8bc8054eb118d02c285e22d40b0ec203d9442d0d2f6b0fa9808f482e868686866040518080602001806020018381038352878782818152602001925080828437600081840152601f19601f8201169050808301925050508381038252858582818152602001925080828437600081840152601f19601f820116905080830192505050965050505050505060405180910390a350505050565b811515838575ffffffffffffffffffffffffffffffffffffffffffff167f7e114930cce781478e313eee1a5bf3fff0d713486daf7f5ddf0ede922e4dca24846040518082815260200191505060405180910390a450505050565b8063ffffffff168260010b847f68f66847d823f0f4a51a4dcb16820f2d98955b5987aaa17fb0df063016e8b6c160405160405180910390a450505056fea264697066735822122086fa2dad41353aa0d62d7fbb244829892bab75eecbefa724f8549f9bfc890d2564736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"string","name":"IndexedString","type":"string"},{"indexed":true,"internalType":"bytes","name":"IndexedBytes","type":"bytes"},{"indexed":false,"internalType":"string","name":"NonIndexedString","type":"string"},{"indexed":false,"internalType":"bytes","name":"NonIndexedBytes","type":"bytes"}],"name":"DynamicEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes24","name":"IndexedBytes","type":"bytes24"},{"indexed":false,"internalType":"bytes24","name":"NonIndexedBytes","type":"bytes24"}],"name":"FixedBytesEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"Number","type":"uint256"},{"indexed":true,"internalType":"int16","name":"Short","type":"int16"},{"indexed":true,"internalType":"uint32","name":"Long","type":"uint32"}],"name":"NodataEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"Addr","type":"address"},{"indexed":true,"internalType":"bytes32","name":"Id","type":"bytes32"},{"indexed":true,"internalType":"bool","name":"Flag","type":"bool"},{"indexed":false,"internalType":"uint256","name":"Value","type":"uint256"}],"name":"SimpleEvent","type":"event"},{"inputs":[{"internalType":"string","name":"str","type":"string"},{"internalType":"bytes","name":"blob","type":"bytes"}],"name":"raiseDynamicEvent","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes24","name":"blob","type":"bytes24"}],"name":"raiseFixedBytesEvent","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"number","type":"uint256"},{"internalType":"int16","name":"short","type":"int16"},{"internalType":"uint32","name":"long","type":"uint32"}],"name":"raiseNodataEvent","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"addr","type":"address"},{"internalType":"bytes32","name":"id","type":"bytes32"},{"internalType":"bool","name":"flag","type":"bool"},{"internalType":"uint256","name":"value","type":"uint256"}],"name":"raiseSimpleEvent","outputs":[],"stateMutability":"nonpayable","type":"function"}]`},
		`
			"math/big"
			"time"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/common"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy an eventer contract
			_, _, eventer, err := DeployEventer(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy eventer contract: %v", err)
			}
			sim.Commit()

			// Inject a few events into the contract, gradually more in each block
			addr0, _ := common.HexToAddress("cb540000000000000000000000000000000000000000")
			addr1, _ := common.HexToAddress("cb270000000000000000000000000000000000000001")
			addr2, _ := common.HexToAddress("cb970000000000000000000000000000000000000002")
			addr3, _ := common.HexToAddress("cb700000000000000000000000000000000000000003")
			addrs := []common.Address{addr0,addr1,addr2,addr3}
			for i := 1; i <= 3; i++ {
				for j := 1; j <= i; j++ {
					if _, err := eventer.RaiseSimpleEvent(auth, addrs[j], [32]byte{byte(j)}, true, big.NewInt(int64(10*i+j))); err != nil {
						t.Fatalf("block %d, event %d: raise failed: %v", i, j, err)
					}
				}
				sim.Commit()
			}
			// Test filtering for certain events and ensure they can be found
			sit, err := eventer.FilterSimpleEvent(nil, []common.Address{addrs[1], addrs[3]}, [][32]byte{{byte(1)}, {byte(2)}, {byte(3)}}, []bool{true})
			if err != nil {
				t.Fatalf("failed to filter for simple events: %v", err)
			}
			defer sit.Close()

			sit.Next()
			if sit.Event.Value.Uint64() != 11 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {11, true}", sit.Event)
			}
			sit.Next()
			if sit.Event.Value.Uint64() != 21 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {21, true}", sit.Event)
			}
			sit.Next()
			if sit.Event.Value.Uint64() != 31 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {31, true}", sit.Event)
			}
			sit.Next()
			if sit.Event.Value.Uint64() != 33 || !sit.Event.Flag {
				t.Errorf("simple log content mismatch: have %v, want {33, true}", sit.Event)
			}

			if sit.Next() {
				t.Errorf("unexpected simple event found: %+v", sit.Event)
			}
			if err = sit.Error(); err != nil {
				t.Fatalf("simple event iteration failed: %v", err)
			}
			// Test raising and filtering for an event with no data component
			if _, err := eventer.RaiseNodataEvent(auth, big.NewInt(314), 141, 271); err != nil {
				t.Fatalf("failed to raise nodata event: %v", err)
			}
			sim.Commit()

			nit, err := eventer.FilterNodataEvent(nil, []*big.Int{big.NewInt(314)}, []int16{140, 141, 142}, []uint32{271})
			if err != nil {
				t.Fatalf("failed to filter for nodata events: %v", err)
			}
			defer nit.Close()

			if !nit.Next() {
				t.Fatalf("nodata log not found: %v", nit.Error())
			}
			if nit.Event.Number.Uint64() != 314 {
				t.Errorf("nodata log content mismatch: have %v, want 314", nit.Event.Number)
			}
			if nit.Next() {
				t.Errorf("unexpected nodata event found: %+v", nit.Event)
			}
			if err = nit.Error(); err != nil {
				t.Fatalf("nodata event iteration failed: %v", err)
			}
			// Test raising and filtering for events with dynamic indexed components
			if _, err := eventer.RaiseDynamicEvent(auth, "Hello", []byte("World")); err != nil {
				t.Fatalf("failed to raise dynamic event: %v", err)
			}
			sim.Commit()

			dit, err := eventer.FilterDynamicEvent(nil, []string{"Hi", "Hello", "Bye"}, [][]byte{[]byte("World")})
			if err != nil {
				t.Fatalf("failed to filter for dynamic events: %v", err)
			}
			defer dit.Close()

			if !dit.Next() {
				t.Fatalf("dynamic log not found: %v", dit.Error())
			}
			if dit.Event.NonIndexedString != "Hello" || string(dit.Event.NonIndexedBytes) != "World" ||	dit.Event.IndexedString != common.HexToHash("0x8ca66ee6b2fe4bb928a8e3cd2f508de4119c0895f22e011117e22cf9b13de7ef") || dit.Event.IndexedBytes != common.HexToHash("0x9916ad7f05695e7eb02cbd48e1aa93f04b8104a7d8698f6d150669128f4b174a") {
				t.Errorf("dynamic log content mismatch: have %v, want {'0x06b3dfaec148fb1bb2b066f10ec285e7c9bf402ab32aa78a5d38e34566810cd2, '0xf2208c967df089f60420785795c0a9ba8896b0f6f1867fa7f1f12ad6f79c1a18', 'Hello', 'World'}", dit.Event)
			}
			if dit.Next() {
				t.Errorf("unexpected dynamic event found: %+v", dit.Event)
			}
			if err = dit.Error(); err != nil {
				t.Fatalf("dynamic event iteration failed: %v", err)
			}
			// Test raising and filtering for events with fixed bytes components
			var fblob [24]byte
			copy(fblob[:], []byte("Fixed Bytes"))

			if _, err := eventer.RaiseFixedBytesEvent(auth, fblob); err != nil {
				t.Fatalf("failed to raise fixed bytes event: %v", err)
			}
			sim.Commit()

			fit, err := eventer.FilterFixedBytesEvent(nil, [][24]byte{fblob})
			if err != nil {
				t.Fatalf("failed to filter for fixed bytes events: %v", err)
			}
			defer fit.Close()

			if !fit.Next() {
				t.Fatalf("fixed bytes log not found: %v", fit.Error())
			}
			if fit.Event.NonIndexedBytes != fblob || fit.Event.IndexedBytes != fblob {
				t.Errorf("fixed bytes log content mismatch: have %v, want {'%x', '%x'}", fit.Event, fblob, fblob)
			}
			if fit.Next() {
				t.Errorf("unexpected fixed bytes event found: %+v", fit.Event)
			}
			if err = fit.Error(); err != nil {
				t.Fatalf("fixed bytes event iteration failed: %v", err)
			}
			// Test subscribing to an event and raising it afterwards
			ch := make(chan *EventerSimpleEvent, 16)
			sub, err := eventer.WatchSimpleEvent(nil, ch, nil, nil, nil)
			if err != nil {
				t.Fatalf("failed to subscribe to simple events: %v", err)
			}
			addr255, _ := common.HexToAddress("cb970000000000000000000000000000000000000255")
			if _, err := eventer.RaiseSimpleEvent(auth, addr255, [32]byte{255}, true, big.NewInt(255)); err != nil {
				t.Fatalf("failed to raise subscribed simple event: %v", err)
			}
			sim.Commit()

			select {
			case event := <-ch:
				if event.Value.Uint64() != 255 {
					t.Errorf("simple log content mismatch: have %v, want 255", event)
				}
			case <-time.After(250 * time.Millisecond):
				t.Fatalf("subscribed simple event didn't arrive")
			}
			// Unsubscribe from the event and make sure we're not delivered more
			sub.Unsubscribe()

			addr254, _ := common.HexToAddress("cb970000000000000000000000000000000000000254")
			if _, err := eventer.RaiseSimpleEvent(auth, addr254, [32]byte{254}, true, big.NewInt(254)); err != nil {
				t.Fatalf("failed to raise subscribed simple event: %v", err)
			}
			sim.Commit()

			select {
			case event := <-ch:
				t.Fatalf("unsubscribed simple event arrived: %v", event)
			case <-time.After(250 * time.Millisecond):
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	{
		`DeeplyNestedArray`,
		`
			contract DeeplyNestedArray {
				uint64[3][4][5] public deepUint64Array;
				function storeDeepUintArray(uint64[3][4][5] arr) public {
					deepUint64Array = arr;
				}
				function retrieveDeepArray() public view returns (uint64[3][4][5]) {
					return deepUint64Array;
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b506105c6806100206000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c80634332d8431461004657806395c896e4146100b0578063bcfc9ffe146100e0575b600080fd5b6100866004803603606081101561005c57600080fd5b8101908080359060200190929190803590602001909291908035906020019092919050505061016c565b604051808267ffffffffffffffff1667ffffffffffffffff16815260200191505060405180910390f35b6100de60048036036107808110156100c757600080fd5b8101908080610780019091929192905050506101bc565b005b6100e86101d1565b604051808260056000925b8184101561015c578284602002015160046000925b8184101561014e5782846020020151600360200280838360005b8381101561013d578082015181840152602081019050610122565b505050509050019260010192610108565b9250505092600101926100f3565b9250505091505060405180910390f35b6000836005811061017957fe5b60040201826004811061018857fe5b01816003811061019457fe5b6004918282040191900660080292509250509054906101000a900467ffffffffffffffff1681565b8060009060056101cd9291906102bd565b5050565b6101d961030f565b6000600580602002604051908101604052809291906000905b828210156102b457838260040201600480602002604051908101604052809291906000905b828210156102a15783820160038060200260405190810160405280929190826003801561028d576020028201916000905b82829054906101000a900467ffffffffffffffff1667ffffffffffffffff16815260200190600801906020826007010492830192600103820291508084116102485790505b505050505081526020019060010190610217565b50505050815260200190600101906101f2565b50505050905090565b82600560040281019282156102fe57916101800282015b828111156102fd5782829060046102ec92919061033c565b5091610180019190600401906102d4565b5b50905061030b9190610389565b5090565b6040518060a001604052806005905b6103266103b5565b81526020019060019003908161031e5790505090565b8260048101928215610378579160600282015b828111156103775782829060036103679291906103e2565b509160600191906001019061034f565b5b509050610385919061049a565b5090565b6103b291905b808211156103ae57600081816103a591906104c6565b5060040161038f565b5090565b90565b60405180608001604052806004905b6103cc61050b565b8152602001906001900390816103c45790505090565b82600380016004900481019282156104895791602002820160005b8382111561045357833567ffffffffffffffff1683826101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555092602001926008016020816007010492830192600103026103fd565b80156104875782816101000a81549067ffffffffffffffff0219169055600801602081600701049283019260010302610453565b505b509050610496919061052d565b5090565b6104c391905b808211156104bf57600081816104b69190610564565b506001016104a0565b5090565b90565b50600081816104d59190610564565b50600101600081816104e79190610564565b50600101600081816104f99190610564565b5060010160006105099190610564565b565b6040518060600160405280600390602082028036833780820191505090505090565b61056191905b8082111561055d57600081816101000a81549067ffffffffffffffff021916905550600101610533565b5090565b90565b506000905556fea26469706673582212208d4276d9b1c5d1a17ce559421325aaaf0580df8b014e41ffae1a07656cf0d7d664736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"constant":false,"inputs":[{"name":"arr","type":"uint64[3][4][5]"}],"name":"storeDeepUintArray","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"retrieveDeepArray","outputs":[{"name":"","type":"uint64[3][4][5]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"","type":"uint256"},{"name":"","type":"uint256"},{"name":"","type":"uint256"}],"name":"deepUint64Array","outputs":[{"name":"","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"}]`},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			//deploy the test contract
			_, _, testContract, err := DeployDeeplyNestedArray(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy test contract: %v", err)
			}

			// Finish deploy.
			sim.Commit()

			//Create coordinate-filled array, for testing purposes.
			testArr := [5][4][3]uint64{}
			for i := 0; i < 5; i++ {
				testArr[i] = [4][3]uint64{}
				for j := 0; j < 4; j++ {
					testArr[i][j] = [3]uint64{}
					for k := 0; k < 3; k++ {
						//pack the coordinates, each array value will be unique, and can be validated easily.
						testArr[i][j][k] = uint64(i) << 16 | uint64(j) << 8 | uint64(k)
					}
				}
			}

			if _, err := testContract.StoreDeepUintArray(&bind.TransactOpts{
				From: auth.From,
				Signer: auth.Signer,
			}, testArr); err != nil {
				t.Fatalf("Failed to store nested array in test contract: %v", err)
			}

			sim.Commit()

			retrievedArr, err := testContract.RetrieveDeepArray(&bind.CallOpts{
				From: auth.From,
				Pending: false,
			})
			if err != nil {
				t.Fatalf("Failed to retrieve nested array from test contract: %v", err)
			}

			//quick check to see if contents were copied
			// (See accounts/abi/unpack_test.go for more extensive testing)
			if retrievedArr[4][3][2] != testArr[4][3][2] {
				t.Fatalf("Retrieved value does not match expected value! got: %d, expected: %d. %v", retrievedArr[4][3][2], testArr[4][3][2], err)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	{
		`CallbackParam`,
		`
			contract FunctionPointerTest {
				function test(function(uint256) external callback) external {
					callback(1);
				}
			}
		`,
		[]string{`608060405234801561001057600080fd5b5061015e806100206000396000f3fe60806040526004361061003b576000357c010000000000000000000000000000000000000000000000000000000090048063d7a5aba214610040575b600080fd5b34801561004c57600080fd5b506100be6004803603602081101561006357600080fd5b810190808035806c0100000000000000000000000090049068010000000000000000900463ffffffff1677ffffffffffffffffffffffffffffffffffffffffffffffff169091602001919093929190939291905050506100c0565b005b818160016040518263ffffffff167c010000000000000000000000000000000000000000000000000000000002815260040180828152602001915050600060405180830381600087803b15801561011657600080fd5b505af115801561012a573d6000803e3d6000fd5b50505050505056fea165627a7a7230582062f87455ff84be90896dbb0c4e4ddb505c600d23089f8e80a512548440d7e2580029`},
		[]string{`[
			{
				"constant": false,
				"inputs": [
					{
						"name": "callback",
						"type": "function"
					}
				],
				"name": "test",
				"outputs": [],
				"payable": false,
				"stateMutability": "nonpayable",
				"type": "function"
			}
		]`}, `
			"strings"
		`,
		`
			if strings.Compare("test(function)", CallbackParamFuncSigs["d7a5aba2"]) != 0 {
				t.Fatalf("")
			}
		`,
		[]map[string]string{
			{
				"test(function)": "d7a5aba2",
			},
		},
		nil,
		nil,
		nil,
	}, {
		`Tuple`,
		`
		pragma solidity >=0.4.19 <0.6.0;
		pragma experimental ABIEncoderV2;

		contract Tuple {
			struct S { uint a; uint[] b; T[] c; }
			struct T { uint x; uint y; }
			struct P { uint8 x; uint8 y; }
			struct Q { uint16 x; uint16 y; }
			event TupleEvent(S a, T[2][] b, T[][2] c, S[] d, uint[] e);
			event TupleEvent2(P[]);

			function func1(S memory a, T[2][] memory b, T[][2] memory c, S[] memory d, uint[] memory e) public pure returns (S memory, T[2][] memory, T[][2] memory, S[] memory, uint[] memory) {
				return (a, b, c, d, e);
			}
			function func2(S memory a, T[2][] memory b, T[][2] memory c, S[] memory d, uint[] memory e) public {
				emit TupleEvent(a, b, c, d, e);
			}
			function func3(Q[] memory) public pure {} // call function, nothing to return
		}
		`,
		[]string{`608060405234801561001057600080fd5b50610f03806100206000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c8063771c58b2146100465780637cf0a6cf1461007a57806391c5468b14610096575b600080fd5b610060600480360381019061005b9190610660565b6100b2565b604051610071959493929190610b64565b60405180910390f35b610094600480360381019061008f9190610660565b6100e3565b005b6100b060048036038101906100ab919061061f565b610129565b005b6100ba61012c565b60606100c461014d565b6060808989898989945094509450945094509550955095509550959050565b7fb9874eef117b46e9a589b15684e216dff362d2a28dbae79091c1b7ff8cbbbffb858585858560405161011a959493929190610b64565b60405180910390a15050505050565b50565b60405180606001604052806000815260200160608152602001606081525090565b60405180604001604052806002905b606081526020019060019003908161015c5790505090565b600082601f83011261018557600080fd5b813561019861019382610c07565b610bda565b915081818352602084019350602081019050838560808402820111156101bd57600080fd5b60005b838110156101ed57816101d38882610353565b8452602084019350608083019250506001810190506101c0565b5050505092915050565b600082601f83011261020857600080fd5b600261021b61021682610c2f565b610bda565b9150818360005b83811015610252578135860161023888826103c7565b845260208401935060208301925050600181019050610222565b5050505092915050565b600082601f83011261026d57600080fd5b813561028061027b82610c51565b610bda565b915081818352602084019350602081019050838560408402820111156102a557600080fd5b60005b838110156102d557816102bb88826104cd565b8452602084019350604083019250506001810190506102a8565b5050505092915050565b600082601f8301126102f057600080fd5b81356103036102fe82610c79565b610bda565b9150818183526020840193506020810190508360005b83811015610349578135860161032f8882610519565b845260208401935060208301925050600181019050610319565b5050505092915050565b600082601f83011261036457600080fd5b600261037761037282610ca1565b610bda565b9150818385604084028201111561038d57600080fd5b60005b838110156103bd57816103a388826105a9565b845260208401935060408301925050600181019050610390565b5050505092915050565b600082601f8301126103d857600080fd5b81356103eb6103e682610cc3565b610bda565b9150818183526020840193506020810190508385604084028201111561041057600080fd5b60005b83811015610440578161042688826105a9565b845260208401935060408301925050600181019050610413565b5050505092915050565b600082601f83011261045b57600080fd5b813561046e61046982610ceb565b610bda565b9150818183526020840193506020810190508385602084028201111561049357600080fd5b60005b838110156104c357816104a9888261060a565b845260208401935060208301925050600181019050610496565b5050505092915050565b6000604082840312156104df57600080fd5b6104e96040610bda565b905060006104f9848285016105f5565b600083015250602061050d848285016105f5565b60208301525092915050565b60006060828403121561052b57600080fd5b6105356060610bda565b905060006105458482850161060a565b600083015250602082013567ffffffffffffffff81111561056557600080fd5b6105718482850161044a565b602083015250604082013567ffffffffffffffff81111561059157600080fd5b61059d848285016103c7565b60408301525092915050565b6000604082840312156105bb57600080fd5b6105c56040610bda565b905060006105d58482850161060a565b60008301525060206105e98482850161060a565b60208301525092915050565b60008135905061060481610e7a565b92915050565b60008135905061061981610e91565b92915050565b60006020828403121561063157600080fd5b600082013567ffffffffffffffff81111561064b57600080fd5b6106578482850161025c565b91505092915050565b600080600080600060a0868803121561067857600080fd5b600086013567ffffffffffffffff81111561069257600080fd5b61069e88828901610519565b955050602086013567ffffffffffffffff8111156106bb57600080fd5b6106c788828901610174565b945050604086013567ffffffffffffffff8111156106e457600080fd5b6106f0888289016101f7565b935050606086013567ffffffffffffffff81111561070d57600080fd5b610719888289016102df565b925050608086013567ffffffffffffffff81111561073657600080fd5b6107428882890161044a565b9150509295509295909350565b600061075b8383610907565b60808301905092915050565b6000610773838361095e565b905092915050565b60006107878383610a78565b905092915050565b600061079b8383610b26565b60408301905092915050565b60006107b38383610b55565b60208301905092915050565b60006107ca82610d67565b6107d48185610df7565b93506107df83610d13565b8060005b838110156108105781516107f7888261074f565b975061080283610da9565b9250506001810190506107e3565b5085935050505092915050565b600061082882610d72565b6108328185610e08565b93508360208202850161084485610d23565b8060005b8581101561088057848403895281516108618582610767565b945061086c83610db6565b925060208a01995050600181019050610848565b50829750879550505050505092915050565b600061089d82610d7d565b6108a78185610e13565b9350836020820285016108b985610d2d565b8060005b858110156108f557848403895281516108d6858261077b565b94506108e183610dc3565b925060208a019950506001810190506108bd565b50829750879550505050505092915050565b61091081610d88565b61091a8184610e24565b925061092582610d3d565b8060005b8381101561095657815161093d878261078f565b965061094883610dd0565b925050600181019050610929565b505050505050565b600061096982610d93565b6109738185610e2f565b935061097e83610d47565b8060005b838110156109af578151610996888261078f565b97506109a183610ddd565b925050600181019050610982565b5085935050505092915050565b60006109c782610d9e565b6109d18185610e40565b93506109dc83610d57565b8060005b83811015610a0d5781516109f488826107a7565b97506109ff83610dea565b9250506001810190506109e0565b5085935050505092915050565b6000610a2582610d9e565b610a2f8185610e51565b9350610a3a83610d57565b8060005b83811015610a6b578151610a5288826107a7565b9750610a5d83610dea565b925050600181019050610a3e565b5085935050505092915050565b6000606083016000830151610a906000860182610b55565b5060208301518482036020860152610aa882826109bc565b91505060408301518482036040860152610ac2828261095e565b9150508091505092915050565b6000606083016000830151610ae76000860182610b55565b5060208301518482036020860152610aff82826109bc565b91505060408301518482036040860152610b19828261095e565b9150508091505092915050565b604082016000820151610b3c6000850182610b55565b506020820151610b4f6020850182610b55565b50505050565b610b5e81610e70565b82525050565b600060a0820190508181036000830152610b7e8188610acf565b90508181036020830152610b9281876107bf565b90508181036040830152610ba6818661081d565b90508181036060830152610bba8185610892565b90508181036080830152610bce8184610a1a565b90509695505050505050565b6000604051905081810181811067ffffffffffffffff82111715610bfd57600080fd5b8060405250919050565b600067ffffffffffffffff821115610c1e57600080fd5b602082029050602081019050919050565b600067ffffffffffffffff821115610c4657600080fd5b602082029050919050565b600067ffffffffffffffff821115610c6857600080fd5b602082029050602081019050919050565b600067ffffffffffffffff821115610c9057600080fd5b602082029050602081019050919050565b600067ffffffffffffffff821115610cb857600080fd5b602082029050919050565b600067ffffffffffffffff821115610cda57600080fd5b602082029050602081019050919050565b600067ffffffffffffffff821115610d0257600080fd5b602082029050602081019050919050565b6000819050602082019050919050565b6000819050919050565b6000819050602082019050919050565b6000819050919050565b6000819050602082019050919050565b6000819050602082019050919050565b600081519050919050565b600060029050919050565b600081519050919050565b600060029050919050565b600081519050919050565b600081519050919050565b6000602082019050919050565b6000602082019050919050565b6000602082019050919050565b6000602082019050919050565b6000602082019050919050565b6000602082019050919050565b600082825260208201905092915050565b600081905092915050565b600082825260208201905092915050565b600081905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b600061ffff82169050919050565b6000819050919050565b610e8381610e62565b8114610e8e57600080fd5b50565b610e9a81610e70565b8114610ea557600080fd5b5056fea264697066735822122075c328d99613c649a8b87157e96eeb6e18dbaf09156e63dce6abf75bc68f785c64736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`
[{"anonymous":false,"inputs":[{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"indexed":false,"internalType":"struct Tuple.S","name":"a","type":"tuple"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"indexed":false,"internalType":"struct Tuple.T[2][]","name":"b","type":"tuple[2][]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"indexed":false,"internalType":"struct Tuple.T[][2]","name":"c","type":"tuple[][2]"},{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"indexed":false,"internalType":"struct Tuple.S[]","name":"d","type":"tuple[]"},{"indexed":false,"internalType":"uint256[]","name":"e","type":"uint256[]"}],"name":"TupleEvent","type":"event"},{"anonymous":false,"inputs":[{"components":[{"internalType":"uint8","name":"x","type":"uint8"},{"internalType":"uint8","name":"y","type":"uint8"}],"indexed":false,"internalType":"struct Tuple.P[]","name":"","type":"tuple[]"}],"name":"TupleEvent2","type":"event"},{"constant":true,"inputs":[{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"internalType":"struct Tuple.S","name":"a","type":"tuple"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[2][]","name":"b","type":"tuple[2][]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[][2]","name":"c","type":"tuple[][2]"},{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"internalType":"struct Tuple.S[]","name":"d","type":"tuple[]"},{"internalType":"uint256[]","name":"e","type":"uint256[]"}],"name":"func1","outputs":[{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"internalType":"struct Tuple.S","name":"","type":"tuple"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[2][]","name":"","type":"tuple[2][]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[][2]","name":"","type":"tuple[][2]"},{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"internalType":"struct Tuple.S[]","name":"","type":"tuple[]"},{"internalType":"uint256[]","name":"","type":"uint256[]"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":false,"inputs":[{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"internalType":"struct Tuple.S","name":"a","type":"tuple"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[2][]","name":"b","type":"tuple[2][]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[][2]","name":"c","type":"tuple[][2]"},{"components":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256[]","name":"b","type":"uint256[]"},{"components":[{"internalType":"uint256","name":"x","type":"uint256"},{"internalType":"uint256","name":"y","type":"uint256"}],"internalType":"struct Tuple.T[]","name":"c","type":"tuple[]"}],"internalType":"struct Tuple.S[]","name":"d","type":"tuple[]"},{"internalType":"uint256[]","name":"e","type":"uint256[]"}],"name":"func2","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"components":[{"internalType":"uint16","name":"x","type":"uint16"},{"internalType":"uint16","name":"y","type":"uint16"}],"internalType":"struct Tuple.Q[]","name":"","type":"tuple[]"}],"name":"func3","outputs":[],"payable":false,"stateMutability":"pure","type":"function"}]
		`},
		`
			"math/big"
			"reflect"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,

		`
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			_, _, contract, err := DeployTuple(auth, sim)
			if err != nil {
				t.Fatalf("deploy contract failed %v", err)
			}
			sim.Commit()

			check := func(a, b interface{}, errMsg string) {
				if !reflect.DeepEqual(a, b) {
					t.Fatal(errMsg)
				}
			}

			a := TupleS{
				A: big.NewInt(1),
				B: []*big.Int{big.NewInt(2), big.NewInt(3)},
				C: []TupleT{
					{
						X: big.NewInt(4),
						Y: big.NewInt(5),
					},
					{
						X: big.NewInt(6),
						Y: big.NewInt(7),
					},
				},
			}

			b := [][2]TupleT{
				{
					{
						X: big.NewInt(8),
						Y: big.NewInt(9),
					},
					{
						X: big.NewInt(10),
						Y: big.NewInt(11),
					},
				},
			}

			c := [2][]TupleT{
				{
					{
						X: big.NewInt(12),
						Y: big.NewInt(13),
					},
					{
						X: big.NewInt(14),
						Y: big.NewInt(15),
					},
				},
				{
					{
						X: big.NewInt(16),
						Y: big.NewInt(17),
					},
				},
			}

			d := []TupleS{a}

			e := []*big.Int{big.NewInt(18), big.NewInt(19)}
			ret1, ret2, ret3, ret4, ret5, err := contract.Func1(nil, a, b, c, d, e)
			if err != nil {
				t.Fatalf("invoke contract failed, err %v", err)
			}
			check(ret1, a, "ret1 mismatch")
			check(ret2, b, "ret2 mismatch")
			check(ret3, c, "ret3 mismatch")
			check(ret4, d, "ret4 mismatch")
			check(ret5, e, "ret5 mismatch")

			_, err = contract.Func2(auth, a, b, c, d, e)
			if err != nil {
				t.Fatalf("invoke contract failed, err %v", err)
			}
			sim.Commit()

			iter, err := contract.FilterTupleEvent(nil)
			if err != nil {
				t.Fatalf("failed to create event filter, err %v", err)
			}
			defer iter.Close()

			iter.Next()
			check(iter.Event.A, a, "field1 mismatch")
			check(iter.Event.B, b, "field2 mismatch")
			check(iter.Event.C, c, "field3 mismatch")
			check(iter.Event.D, d, "field4 mismatch")
			check(iter.Event.E, e, "field5 mismatch")

			err = contract.Func3(nil, nil)
			if err != nil {
				t.Fatalf("failed to call function which has no return, err %v", err)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	{
		`UseLibrary`,
		`
		library Math {
    		function add(uint a, uint b) public view returns(uint) {
        		return a + b;
    		}
		}

		contract UseLibrary {
			function add (uint c, uint d) public view returns(uint) {
        		return Math.add(c,d);
    		}
		}
		`,
		[]string{
			// Bytecode for the UseLibrary contract
			`608060405234801561001057600080fd5b50610175806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80630644fb8714610030575b600080fd5b6100666004803603604081101561004657600080fd5b81019080803590602001909291908035906020019092919050505061007c565b6040518082815260200191505060405180910390f35b600075____$935ab3a9a8503b27f98999a9c1b2d26ac2$____630644fb8784846040518363ffffffff1660e01b8152600401808381526020018281526020019250505060206040518083038186803b1580156100d757600080fd5b505af41580156100eb573d6000803e3d6000fd5b505050506040513d602081101561010157600080fd5b810190808051906020019092919050505090509291505056fea26469706673582212208c934e06d25cda8668cef90d6969b9cb2934e1aef7d285c066c823e1216a3b6c64736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`,
			// Bytecode for the Math contract
			`60eb610025600b82828239805160001a60751461001857fe5b30600052607581538281f3fe7500000000000000000000000000000000000000000000301460806040526004361060355760003560e01c80630644fb8714603a575b600080fd5b606d60048036036040811015604e57600080fd5b8101908080359060200190929190803590602001909291905050506083565b6040518082815260200191505060405180910390f35b600081830190509291505056fea2646970667358221220fa48ff66ffc52d00e199243efef6a6b64a7ddac257e3633b45cb667b13c9cc6c64736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`,
		},
		[]string{
			`[{"inputs":[{"internalType":"uint256","name":"c","type":"uint256"},{"internalType":"uint256","name":"d","type":"uint256"}],"name":"add","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`,
			`[{"inputs":[{"internalType":"uint256","name":"a","type":"uint256"},{"internalType":"uint256","name":"b","type":"uint256"}],"name":"add","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`,
		},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			//deploy the test contract
			_, _, testContract, err := DeployUseLibrary(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy test contract: %v", err)
			}

			// Finish deploy.
			sim.Commit()

			// Check that the library contract has been deployed
			// by calling the contract's add function.
			res, err := testContract.Add(&bind.CallOpts{
				From: auth.From,
				Pending: false,
			}, big.NewInt(1), big.NewInt(2))
			if err != nil {
				t.Fatalf("Failed to call linked contract: %v", err)
			}
			if res.Cmp(big.NewInt(3)) != 0 {
				t.Fatalf("Add did not return the correct result: %d != %d", res, 3)
			}
		`,
		nil,
		map[string]string{
			"935ab3a9a8503b27f98999a9c1b2d26ac2": "Math",
		},
		nil,
		[]string{"UseLibrary", "Math"},
	}, {
		"Overload",
		`
		pragma solidity ^0.5.10;

		contract overload {
		  mapping(address => uint256) balances;

		  event bar(uint256 i);
		  event bar(uint256 i, uint256 j);

		  function foo(uint256 i) public {
			  emit bar(i);
		  }
		  function foo(uint256 i, uint256 j) public {
			  emit bar(i, j);
		  }
		}
		`,
		[]string{`608060405234801561001057600080fd5b50610179806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c806304f679b31461003b578063cdf9cddd14610073575b600080fd5b6100716004803603604081101561005157600080fd5b8101908080359060200190929190803590602001909291905050506100a1565b005b61009f6004803603602081101561008957600080fd5b81019080803590602001909291905050506100e4565b005b7f518309ce5086a0cda08afa90c6cb5200afca1b03d526de025f1be10c53b6ee888282604051808381526020018281526020019250505060405180910390a15050565b7f125a553e559dfecbc97626e51cc98157117d04e0953bc98d811c240faed62ed0816040518082815260200191505060405180910390a15056fea2646970667358221220f65bf2aaf72f5f9717b6511d7675a93d4cb2cf33bf49d151d4f8b2d4bf14cfa464736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"constant":false,"inputs":[{"name":"i","type":"uint256"},{"name":"j","type":"uint256"}],"name":"foo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"i","type":"uint256"}],"name":"foo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"name":"i","type":"uint256"}],"name":"bar","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"name":"i","type":"uint256"},{"indexed":false,"name":"j","type":"uint256"}],"name":"bar","type":"event"}]`},
		`
		"math/big"
		"time"
		"crypto/rand"

		"github.com/core-coin/go-core/accounts/abi/bind"
		"github.com/core-coin/go-core/accounts/abi/bind/backends"
		"github.com/core-coin/go-core/core"
		"github.com/core-coin/go-core/crypto"
		`,
		`
		// Initialize test accounts
		key, _ := crypto.GenerateKey(rand.Reader)
		auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))
		sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
		defer sim.Close()

		// deploy the test contract
		_, _, contract, err := DeployOverload(auth, sim)
		if err != nil {
			t.Fatalf("Failed to deploy contract: %v", err)
		}
		// Finish deploy.
		sim.Commit()

		resCh, stopCh := make(chan uint64), make(chan struct{})

		go func() {
			barSink := make(chan *OverloadBar)
			sub, _ := contract.WatchBar(nil, barSink)
			defer sub.Unsubscribe()

			bar0Sink := make(chan *OverloadBar0)
			sub0, _ := contract.WatchBar0(nil, bar0Sink)
			defer sub0.Unsubscribe()

			for {
				select {
				case ev := <-barSink:
					resCh <- ev.I.Uint64()
				case ev := <-bar0Sink:
					resCh <- ev.I.Uint64() + ev.J.Uint64()
				case <-stopCh:
					return
				}
			}
		}()
		contract.Foo(auth, big.NewInt(1), big.NewInt(2))
		sim.Commit()
		select {
		case n := <-resCh:
			if n != 3 {
				t.Fatalf("Invalid bar0 event")
			}
		case <-time.NewTimer(3 * time.Second).C:
			t.Fatalf("Wait bar0 event timeout")
		}

		contract.Foo0(auth, big.NewInt(1))
		sim.Commit()
		select {
		case n := <-resCh:
			if n != 1 {
				t.Fatalf("Invalid bar event")
			}
		case <-time.NewTimer(3 * time.Second).C:
			t.Fatalf("Wait bar event timeout")
		}
		close(stopCh)
		`,
		nil,
		nil,
		nil,
		nil,
	},
	{
		"IdentifierCollision",
		`
		pragma solidity >=0.4.19 <0.6.0;

		contract IdentifierCollision {
			uint public _myVar;

			function MyVar() public view returns (uint) {
				return _myVar;
			}
		}
		`,
		[]string{"60806040523480156100115760006000fd5b50610017565b60c3806100256000396000f3fe608060405234801560105760006000fd5b506004361060365760003560e01c806301ad4d8714603c5780634ef1f0ad146058576036565b60006000fd5b60426074565b6040518082815260200191505060405180910390f35b605e607d565b6040518082815260200191505060405180910390f35b60006000505481565b60006000600050549050608b565b9056fea265627a7a7231582067c8d84688b01c4754ba40a2a871cede94ea1f28b5981593ab2a45b46ac43af664736f6c634300050c0032"},
		[]string{`[{"constant":true,"inputs":[],"name":"MyVar","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"_myVar","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`},
		`
		"math/big"
		"crypto/rand"

		"github.com/core-coin/go-core/accounts/abi/bind"
		"github.com/core-coin/go-core/accounts/abi/bind/backends"
	eddsa "github.com/core-coin/go-goldilocks"
		"github.com/core-coin/go-core/crypto"
		"github.com/core-coin/go-core/core"
		`,
		`
		// Initialize test accounts
		key, _ := crypto.GenerateKey(rand.Reader)
		pub := eddsa.Ed448DerivePublicKey(*key)
		addr := crypto.PubkeyToAddress(pub)

		// Deploy registrar contract
		sim := backends.NewSimulatedBackend(core.GenesisAlloc{addr: {Balance: big.NewInt(1000000000)}}, 10000000)
		defer sim.Close()

		transactOpts, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))
		_, _, _, err := DeployIdentifierCollision(transactOpts, sim)
		if err != nil {
			t.Fatalf("failed to deploy contract: %v", err)
		}
		`,
		nil,
		nil,
		map[string]string{"_myVar": "pubVar"}, // alias MyVar to PubVar
		nil,
	},
	{
		"MultiContracts",
		`
		pragma solidity ^0.5.11;
		pragma experimental ABIEncoderV2;

		library ExternalLib {
			struct SharedStruct{
				uint256 f1;
				bytes32 f2;
			}
		}

		contract ContractOne {
			function foo(ExternalLib.SharedStruct memory s) pure public {
				// Do stuff
			}
		}

		contract ContractTwo {
			function bar(ExternalLib.SharedStruct memory s) pure public {
				// Do stuff
			}
		}
        `,
		[]string{
			`608060405234801561001057600080fd5b50610221806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80631d5a0f3a14610030575b600080fd5b61004a600480360381019061004591906100c5565b61004c565b005b50565b60008135905061005e81610198565b92915050565b60006040828403121561007657600080fd5b61008060406100ee565b90506000610090848285016100b0565b60008301525060206100a48482850161004f565b60208301525092915050565b6000813590506100bf816101af565b92915050565b6000604082840312156100d757600080fd5b60006100e584828501610064565b91505092915050565b60006100f8610109565b90506101048282610127565b919050565b6000604051905090565b6000819050919050565b6000819050919050565b61013082610187565b810181811067ffffffffffffffff8211171561014f5761014e610158565b5b80604052505050565b7f4b1f2ce300000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000601f19601f8301169050919050565b6101a181610113565b81146101ac57600080fd5b50565b6101b88161011d565b81146101c357600080fd5b5056fea26469706673582212202b7f837fcad4d1f0e5fbda29459d3ce2d219eacce9f106c9a6808eb46890638b64736f6c637827302e382e342d646576656c6f702e323032322e382e32322b636f6d6d69742e37383961353965650058`,
			`608060405234801561001057600080fd5b50610221806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80635ca5a69a14610030575b600080fd5b61004a600480360381019061004591906100c5565b61004c565b005b50565b60008135905061005e81610198565b92915050565b60006040828403121561007657600080fd5b61008060406100ee565b90506000610090848285016100b0565b60008301525060206100a48482850161004f565b60208301525092915050565b6000813590506100bf816101af565b92915050565b6000604082840312156100d757600080fd5b60006100e584828501610064565b91505092915050565b60006100f8610109565b90506101048282610127565b919050565b6000604051905090565b6000819050919050565b6000819050919050565b61013082610187565b810181811067ffffffffffffffff8211171561014f5761014e610158565b5b80604052505050565b7f4b1f2ce300000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000601f19601f8301169050919050565b6101a181610113565b81146101ac57600080fd5b50565b6101b88161011d565b81146101c357600080fd5b5056fea26469706673582212209958ba244d4c44a2c902dcd8b58dcf986c2496e74e042eb9d973892565fcd89264736f6c637827302e382e342d646576656c6f702e323032322e382e32322b636f6d6d69742e37383961353965650058`,
			`607d6050600b82828239805160001a6075146043577f4b1f2ce300000000000000000000000000000000000000000000000000000000600052600060045260246000fd5b30600052607581538281f3fe750000000000000000000000000000000000000000000030146080604052600080fdfea26469706673582212203361e723284432509462565df724e59c278ac7f9ef170c8731c1eef7e0e7f52164736f6c637827302e382e342d646576656c6f702e323032322e382e32322b636f6d6d69742e37383961353965650058`,
		},
		[]string{
			`[{"inputs":[{"components":[{"internalType":"uint256","name":"f1","type":"uint256"},{"internalType":"bytes32","name":"f2","type":"bytes32"}],"internalType":"struct ExternalLib.SharedStruct","name":"s","type":"tuple"}],"name":"foo","outputs":[],"stateMutability":"pure","type":"function"}]`,
			`[{"inputs":[{"components":[{"internalType":"uint256","name":"f1","type":"uint256"},{"internalType":"bytes32","name":"f2","type":"bytes32"}],"internalType":"struct ExternalLib.SharedStruct","name":"s","type":"tuple"}],"name":"bar","outputs":[],"stateMutability":"pure","type":"function"}]`,
			`[]`,
		},
		`
		"math/big"		
		"crypto/rand"

		eddsa "github.com/core-coin/go-goldilocks"
		"github.com/core-coin/go-core/accounts/abi/bind"
		"github.com/core-coin/go-core/accounts/abi/bind/backends"
		"github.com/core-coin/go-core/crypto"
		"github.com/core-coin/go-core/core"
        `,
		`
		key, _ := crypto.GenerateKey(rand.Reader)
		pub := eddsa.Ed448DerivePublicKey(*key)
		addr := crypto.PubkeyToAddress(pub)

		// Deploy registrar contract
		sim := backends.NewSimulatedBackend(core.GenesisAlloc{addr: {Balance: big.NewInt(1000000000)}}, 10000000)
		defer sim.Close()

		transactOpts, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))
		_, _, c1, err := DeployContractOne(transactOpts, sim)
		if err != nil {
			t.Fatal("Failed to deploy contract")
		}
		sim.Commit()
		err = c1.Foo(nil, ExternalLibSharedStruct{
			F1: big.NewInt(100),
			F2: [32]byte{0x01, 0x02, 0x03},
		})
		if err != nil {
			t.Fatal("Failed to invoke function:", err)
		}
		_, _, c2, err := DeployContractTwo(transactOpts, sim)
		if err != nil {
			t.Fatal("Failed to deploy contract")
		}
		sim.Commit()
		err = c2.Bar(nil, ExternalLibSharedStruct{
			F1: big.NewInt(100),
			F2: [32]byte{0x01, 0x02, 0x03},
		})
		if err != nil {
			t.Fatal("Failed to invoke function:", err)
		}
        `,
		nil,
		nil,
		nil,
		[]string{"ContractOne", "ContractTwo", "ExternalLib"},
	},
	// Test the existence of the free retrieval calls
	{
		`PureAndView`,
		`pragma solidity >=0.6.0;
		contract PureAndView {
			function PureFunc() public pure returns (uint) {
				return 42;
			}
			function ViewFunc() public view returns (uint) {
				return block.number;
			}
		}
		`,
		[]string{`608060405234801561001057600080fd5b5060db8061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c806339432fb3146037578063ea35ae56146053575b600080fd5b603d606f565b6040518082815260200191505060405180910390f35b60596077565b6040518082815260200191505060405180910390f35b600043905090565b6000602a90509056fea2646970667358221220785a3db68fcbfa43c2fa822b86a6dc607c839125af38ff797c16a38458784dc664736f6c637827302e362e392d646576656c6f702e323032302e372e32312b636f6d6d69742e33633832373333370058`},
		[]string{`[{"inputs": [],"name": "PureFunc","outputs": [{"internalType": "uint256","name": "","type": "uint256"}],"stateMutability": "pure","type": "function"},{"inputs": [],"name": "ViewFunc","outputs": [{"internalType": "uint256","name": "","type": "uint256"}],"stateMutability": "view","type": "function"}]`},
		`
			"math/big"
			"crypto/rand"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
		`,
		`
			// Generate a new random account and a funded simulator
			key, _ := crypto.GenerateKey(rand.Reader)
			auth, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))

			sim := backends.NewSimulatedBackend(core.GenesisAlloc{auth.From: {Balance: big.NewInt(10000000000)}}, 10000000)
			defer sim.Close()

			// Deploy a tester contract and execute a structured call on it
			_, _, pav, err := DeployPureAndView(auth, sim)
			if err != nil {
				t.Fatalf("Failed to deploy PureAndView contract: %v", err)
			}
			sim.Commit()

			// This test the existence of the free retreiver call for view and pure functions
			if num, err := pav.PureFunc(nil); err != nil {
				t.Fatalf("Failed to call anonymous field retriever: %v", err)
			} else if num.Cmp(big.NewInt(42)) != 0 {
				t.Fatalf("Retrieved value mismatch: have %v, want %v", num, 42)
			}
			if num, err := pav.ViewFunc(nil); err != nil {
				t.Fatalf("Failed to call anonymous field retriever: %v", err)
			} else if num.Cmp(big.NewInt(1)) != 0 {
				t.Fatalf("Retrieved value mismatch: have %v, want %v", num, 1)
			}
		`,
		nil,
		nil,
		nil,
		nil,
	},
	// Test fallback separation introduced in v0.6.0
	{
		`NewFallbacks`,
		`
		pragma solidity >=0.6.0 <0.7.0;
	
		contract NewFallbacks {
			event Fallback(bytes data);
			fallback() external {
				bytes memory data;
				assembly {
					calldatacopy(data, 0, calldatasize())
				}
				emit Fallback(data);
			}
	
			event Received(address addr, uint value);
			receive() external payable {
				emit Received(msg.sender, msg.value);
			}
		}
	   `,
		[]string{"608060405234801561001057600080fd5b50610230806100206000396000f3fe608060405236610044577fb4764187bbbca84b57e9671514c33bdb82a80b7ae801dc5bbeab272a07868ce3333460405161003a9291906100e9565b60405180910390a1005b34801561005057600080fd5b50606036600082377fc5f892623b9cf327459605db591333292717e25ed9606e17f41a7a395784aaf8816040516100879190610112565b60405180910390a150005b61009b81610150565b82525050565b60006100ac82610134565b6100b6818561013f565b93506100c681856020860161018e565b6100cf816101c1565b840191505092915050565b6100e381610184565b82525050565b60006040820190506100fe6000830185610092565b61010b60208301846100da565b9392505050565b6000602082019050818103600083015261012c81846100a1565b905092915050565b600081519050919050565b600082825260208201905092915050565b600061015b82610162565b9050919050565b600075ffffffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b60005b838110156101ac578082015181840152602081019050610191565b838111156101bb576000848401525b50505050565b6000601f19601f830116905091905056fea264697066735822122068573f19c872bcc9beb1e92de46f5bac58be7c913a517086ec850e7d387c81ff64736f6c63782a302e382e342d646576656c6f702e323032322e372e362b636f6d6d69742e30353336326564342e6d6f64005b"},
		[]string{`[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes","name":"data","type":"bytes"}],"name":"Fallback","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"addr","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Received","type":"event"},{"stateMutability":"nonpayable","type":"fallback"},{"stateMutability":"payable","type":"receive"}]`},
		`
			"bytes"
			"math/big"
			"crypto/rand"
	
			eddsa "github.com/core-coin/go-goldilocks"

			"github.com/core-coin/go-core/accounts/abi/bind"
			"github.com/core-coin/go-core/accounts/abi/bind/backends"
			"github.com/core-coin/go-core/core"
			"github.com/core-coin/go-core/crypto"
	   `,
		`
			key, _ := crypto.GenerateKey(rand.Reader)
			pub := eddsa.Ed448DerivePublicKey(*key)
			addr := crypto.PubkeyToAddress(pub)
	
			sim := backends.NewSimulatedBackend(core.GenesisAlloc{addr: {Balance: big.NewInt(1000000000)}}, 1000000)
			defer sim.Close()
	
			opts, _ := bind.NewKeyedTransactorWithNetworkID(key, big.NewInt(1337))
			_, _, c, err := DeployNewFallbacks(opts, sim)
			if err != nil {
				t.Fatalf("Failed to deploy contract: %v", err)
			}
			sim.Commit()
	
			// Test receive function
			opts.Value = big.NewInt(100)
			c.Receive(opts)
			sim.Commit()
	
			var gotEvent bool
			iter, _ := c.FilterReceived(nil)
			defer iter.Close()
			for iter.Next() {
				if iter.Event.Addr != addr {
					t.Fatal("Msg.sender mismatch")
				}
				if iter.Event.Value.Uint64() != 100 {
					t.Fatal("Msg.value mismatch")
				}
				gotEvent = true
				break
			}
			if !gotEvent {
				t.Fatal("Expect to receive event emitted by receive")
			}
	
			// Test fallback function
			opts.Value = nil
			calldata := []byte{0x01, 0x02, 0x03}
			c.Fallback(opts, calldata)
			sim.Commit()
	
			iter2, _ := c.FilterFallback(nil)
			defer iter2.Close()
			for iter2.Next() {
				if !bytes.Equal(iter2.Event.Data, calldata) {
					t.Fatal("calldata mismatch")
				}
				gotEvent = true
				break
			}
			if !gotEvent {
				t.Fatal("Expect to receive event emitted by fallback")
			}
	   `,
		nil,
		nil,
		nil,
		nil,
	},
}

// Tests that packages generated by the binder can be successfully compiled and
// the requested tester run against it.
func TestGolangBindings(t *testing.T) {
	// Skip the test if no Go command can be found
	gocmd := runtime.GOROOT() + "/bin/go"
	if !common.FileExist(gocmd) {
		t.Skip("go sdk not found for testing")
	}
	// Create a temporary workspace for the test suite
	ws, err := ioutil.TempDir("", "binding-test")
	if err != nil {
		t.Fatalf("failed to create temporary workspace: %v", err)
	}
	defer os.RemoveAll(ws)

	pkg := filepath.Join(ws, "bindtest")
	if err = os.MkdirAll(pkg, 0700); err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
	// Generate the test suite for all the contracts
	for i, tt := range bindTests {
		var types []string
		if tt.types != nil {
			types = tt.types
		} else {
			types = []string{tt.name}
		}
		// Generate the binding and create a Go source file in the workspace
		bind, err := Bind(types, tt.abi, tt.bytecode, tt.fsigs, "bindtest", LangGo, tt.libs, tt.aliases)
		if err != nil {
			t.Fatalf("test %d: failed to generate binding: %v", i, err)
		}
		if err = ioutil.WriteFile(filepath.Join(pkg, strings.ToLower(tt.name)+".go"), []byte(bind), 0600); err != nil {
			t.Fatalf("test %d: failed to write binding: %v", i, err)
		}
		// Generate the test file with the injected test code
		code := fmt.Sprintf(`
			package bindtest

			import (
				"testing"
				%s
			)

			func Test%s(t *testing.T) {
				%s
			}
		`, tt.imports, tt.name, tt.tester)
		if err := ioutil.WriteFile(filepath.Join(pkg, strings.ToLower(tt.name)+"_test.go"), []byte(code), 0600); err != nil {
			t.Fatalf("test %d: failed to write tests: %v", i, err)
		}
	}
	// Convert the package to go modules and use the current source for go-core
	moder := exec.Command(gocmd, "mod", "init", "bindtest")
	moder.Dir = pkg
	if out, err := moder.CombinedOutput(); err != nil {
		t.Fatalf("failed to convert binding test to modules: %v\n%s", err, out)
	}
	pwd, _ := os.Getwd()
	replacer := exec.Command(gocmd, "mod", "edit", "-x", "-require", "github.com/core-coin/go-core@v0.0.0", "-replace", "github.com/core-coin/go-core="+filepath.Join(pwd, "..", "..", "..")) // Repo root
	replacer.Dir = pkg
	if out, err := replacer.CombinedOutput(); err != nil {
		t.Fatalf("failed to replace binding test dependency to current source tree: %v\n%s", err, out)
	}
	tidier := exec.Command(gocmd, "mod", "tidy")
	tidier.Dir = pkg
	if out, err := tidier.CombinedOutput(); err != nil {
		t.Fatalf("failed to tidy Go module file: %v\n%s", err, out)
	}
	// Test the entire package and report any failures
	cmd := exec.Command(gocmd, "test", "-v", "-count", "1")
	cmd.Dir = pkg
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run binding test: %v\n%s", err, out)
	}
}

## Go Core

Official Golang implementation of the CORE protocol.

Automated builds are available for stable releases and the unstable master branch. Binary archives are published at [CORE website > downloads](https://coreblockchain.cc/downloads).

## Building the source

Building `gocore` requires both a Go (version 1.13 or later) and a C compiler. You can install them using your favorite package manager. Once the dependencies are installed, run

```shell
make gocore
```

or, to build the full suite of utilities:

```shell
make all
```

## ICAN network prefixes

CORE Client implements ICAN-based addresses with following formats:

Name | Prefix | Length | Format
--- | --- | --- | ---
Mainnet | CB | 44 | H40
Testnets | AB | 44 | H40
Privatenets | CE | 44 | H40

## Executables

The go-core project comes with several wrappers/executables found in the `cmd`
directory.

Command | Description
--- | ---
gocore | Main Core CLI client. It is the entry point into the CORE network (main-, "testnets" or private net), capable of running as a full node (default), archive node (retaining all historical state), or a light node (retrieving data live). It can be used by other processes as a gateway into the CORE network via JSON RPC endpoints exposed on top of HTTP, WebSocket, and/or IPC transports. Type `gocore --help` for command-line options.
abigen | Source code generator to convert CORE contract definitions into easy to use, compile-time type-safe Go packages. It operates on plain CORE contract ABIs with expanded functionality if the contract bytecode is also available. However, it also accepts Ylem source files, making development much more streamlined.
bootnode | Stripped down version of our CORE client implementation that only takes part in the network node discovery protocol, but does not run any of the higher-level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks.
cvm | Developer utility version of the CVM (CORE Virtual Machine) that is capable of running bytecode snippets within a configurable environment and execution mode. Its purpose is to allow isolated, fine-grained debugging of CVM opcodes.
gocorerpctest | Developer utility tool to support our core/rpc-test test suite which validates baseline conformity to the CORE JSON RPC specs.
rlpdump | Developer utility tool to convert binary RLP (Recursive Length Prefix) dumps (data encoding used by the CORE protocol both network as well as consensus wise) to user-friendlier hierarchical representation.

### Full node on the main CORE network

By far the most common scenario is people wanting to simply interact with the CORE network: create accounts; transfer funds; deploy and interact with contracts. For this particular use-case the user doesn't care about years-old historical data, so we can fast-sync quickly to the current state of the network. To do so:

```shell
$ gocore console
```

This command will:
 * Start `gocore` in fast sync mode (default, can be changed with the `--syncmode` flag), causing it to download more data in exchange for avoiding processing the entire history of the CORE network, which is very CPU intensive.
 * Startup `gocore`'s built-in interactive JavaScript console, (via the trailing `console` subcommand) through which you can invoke all official `web3` methods as well as `gocore`'s own management APIs. This tool is optional and if you leave it out you can always attach to an already running `gocore` instance with `gocore attach`.

### A Full node on the Devin network (PoW)

Transitioning towards developers, if you'd like to play around with creating CORE contracts, you almost certainly would like to do that without any real money involved until you get the hang of the entire system. In other words, instead of attaching to the main network, you want to join one of the tests networks (Devin, Koliba) with your node, which is fully equivalent to the main network, but with play-Core only.

```shell
$ gocore --devin console
```

The `console` subcommand has the exact same meaning as above and they are equally useful on the devin too. Please see above for their explanations if you've skipped here.

Specifying the `--devin` flag, however, will reconfigure your `gocore` instance a bit:

 * Instead of using the default data directory (`~/core` on Linux for example), `gocore` will nest itself one level deeper into a `devin` subfolder (`~/core/devin` on Linux). Note, on OSX and Linux this also means that attaching to a running devin node requires the use of a custom endpoint since `gocore attach` will try to attach to a production node endpoint by default.
 * Instead of connecting the main Core network, the client will connect to the test network, which uses different P2P bootnodes, different network IDs, and genesis states.

### Full node on the Koliba test network (PoA)

Gocore also supports connecting to a proof-of-authority (PoA) based test network called Koliba (operated by members of the community). This network is lighter, more secure.

```shell
$ gocore --koliba console
```

### Configuration

As an alternative to passing the numerous flags to the `gocore` binary, you can also pass a configuration file via:

```shell
$ gocore --config /path/to/your_config.toml
```

To get an idea of how the file should look like you can use the `dumpconfig` subcommand to export your existing configuration:

```shell
$ gocore --your-favourite-flags dumpconfig
```

#### Docker quick start

One of the quickest ways to get Core up and running on your machine is by using
Docker:

```shell
docker run -d --name core-node -v /Users/robocop/core-coin:/root \
           -p 8545:8545 -p 30300:30300 \
           core-coin/client-go
```

This will start `gocore` in fast-sync mode with a DB memory allowance of 1GB just as the above command does. It will also create a persistent volume in your home directory for saving your blockchain as well as map the default ports.

Do not forget `--rpcaddr 0.0.0.0`, if you want to access RPC from other containers and/or hosts. By default, `gocore` binds to the local interface, and RPC endpoints are not accessible from the outside.

### Programmatically interfacing gocore nodes

As a developer, sooner rather than later you'll want to start interacting with `gocore` and the CORE network via your own programs and not manually through the console. To aid this, `gocore` has built-in support for JSON-RPC based APIs and `gocore` specific APIs. These can be exposed via HTTP, WebSockets, and IPC (UNIX sockets on UNIX based platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by `gocore`, whereas the HTTP and WS interfaces need to manually be enabled and only expose a subset of APIs due to security reasons. These can be turned on/off and configured as you'd expect.

HTTP based JSON-RPC API options:

  * `--rpc` Enable the HTTP-RPC server
  * `--rpcaddr` HTTP-RPC server listening interface (default: `localhost`)
  * `--rpcport` HTTP-RPC server listening port (default: `8545`)
  * `--rpcapi` API's offered over the HTTP-RPC interface (default: `xcb,net,web3`)
  * `--rpccorsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--wsaddr` WS-RPC server listening interface (default: `localhost`)
  * `--wsport` WS-RPC server listening port (default: `8546`)
  * `--wsapi` API's offered over the WS-RPC interface (default: `xcb,net,web3`)
  * `--wsorigins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: `admin,debug,xcb,miner,net,personal,shh,txpool,web3`)
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to connect via HTTP, WS, or IPC to a `gocore` node configured with the above flags and you'll need to speak JSON-RPC on all transports. You can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based transport before doing so! Hackers on the internet are actively trying to subvert CORE nodes with exposed APIs! Further, all browser tabs can access locally running web servers, so malicious web pages could try to subvert locally available APIs!**

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for granted in the official networks need to be manually set up.

#### Defining the private genesis state

First, you'll need to create the genesis state of your networks, which all nodes need to be aware of and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):

```json
{
  "config": {
    "chainId": "<arbitrary positive integer>"
  },
  "alloc": {},
  "coinbase": "0x0000000000000000000000000000000000000000",
  "difficulty": "0x20000",
  "extraData": "",
  "energyLimit": "0x2fefd8",
  "nonce": "0x0000000000000042",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp": "0x00"
}
```

The above fields should be fine for most purposes, although we'd recommend changing the `nonce` to some random value so you prevent unknown remote nodes from being able to connect to you. If you'd like to pre-fund some accounts for easier testing, create the accounts and populate the `alloc` field with their addresses.

```json
{
  "alloc": {
    "ce180000000000000000000000000000000000000001": {
      "balance": "111111111"
    },
    "ce880000000000000000000000000000000000000002": {
      "balance": "222222222"
    }
  }
}
```

With the genesis state defined in the above JSON file, you'll need to initialize **every** `gocore` node with it prior to starting it up to ensure all blockchain parameters are correctly set:

```shell
$ gocore init path/to/genesis.json
```

#### Creating the rendezvous point

With all nodes that you want to run initialized to the desired genesis state, you'll need to start a bootstrap node that others can use to find each other in your network and/or over the internet. The clean way is to configure and run a dedicated bootnode:

```shell
$ bootnode --genkey=boot.key
$ bootnode --nodekey=boot.key
```

With the bootnode online, it will display an `enode` URL that other nodes can use to connect to it and exchange peer information. Make sure to replace the displayed IP address information (most probably `[::]`) with your externally
accessible IP to get the actual `enode` URL.

*Note: You could also use a full-fledged `gocore` node as a bootnode, but it's the less recommended way.*

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try
`telnet <ip> <port>` to ensure it's indeed reachable), start every subsequent `gocore` node pointed to the bootnode for peer discovery via the `--bootnodes` flag. It will probably also be desirable to keep the data directory of your private network separated, so do also specify a custom `--datadir` flag.

```shell
$ gocore --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

*Note: Since your network will be completely cut off from the main and test networks, you'll also need to configure a miner to process transactions and create new blocks for you.*

#### Running a private miner

In a private network setting, however, a single CPU miner instance is more than enough for practical purposes as it can produce a stable stream of blocks at the correct intervals without needing heavy resources (consider running on a single thread, no need for multiple ones either). To start a `gocore` instance for mining, run it with all your usual flags, extended by:

```shell
$ gocore <usual-flags> --mine --miner.threads=1 --corebase=0x0000000000000000000000000000000000000000
```

Which will start mining blocks and transactions on a single CPU thread, crediting all proceedings to the account specified by `--corebase`. You can further tune the mining by changing the default energy limit blocks converge to (`--targetenergylimit`) and the price transactions are accepted at (`--energyprice`).

## Issue Labels

### Priority

Label | Meaning (SLA)
--- | ---
P1 Urgent | The current release + potentially immediate hotfix
P2 High | The next release
P3 Medium | Within the next 3 releases
P4 Low | Anything outside the next 3 releases

### Severity

Label | Impact
--- | ---
S1 Blocker | Outage, broken feature with no workaround
S2 Critical | Broken feature, workaround too complex & unacceptable
S3 Major | Broken feature, workaround acceptable
S4 Low | Functionality inconvenience or cosmetic issue

## Contribution

Thank you for considering helping out with the source code! We welcome contributions from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-core, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main codebase. If you wish to submit more complex changes though, please check up with the core devs first on the [Core ◆ Talk](https://coretalk.info) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "xcb, rpc: make trace configs optional"

## Security vulnerability disclosure

Please report suspected security vulnerabilities in private following the [Security manual](SECURITY.md). Do NOT create publicly viewable issues for suspected security vulnerabilities.

## License

The go-core library (i.e. all code outside of the `cmd` directory) is licensed under the [GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also included in our repository in the `COPYING.LESSER` file.

The go-core binaries (i.e. all code inside of the `cmd` directory) are licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included in our repository in the `COPYING` file.

## Community

[![Core ◆ Talk](https://img.shields.io/badge/Core%20%E2%97%86%20Talk-Protocol%20and%20Client-green)](https://coretalk.info/c/protocol-and-client/8/)
[![Discord](https://img.shields.io/discord/469227792538533888?label=Discord)](https://discord.gg/b3c5FvZ)

# Go Core

CORE protocol — Official Golang implementation.

The stable releases and the unstable master branch of the automated builds are both accessible in binary packages at [CORE website > downloads](https://coreblockchain.net/download).

## Source Building

To build `gocore`, it is necessary to download a Go (version 1.14 or later) and C/C++ compiler. Any package manager is capable of installing these onto your device. Once installed, run

### To build gocore on Linux or Mac

```shell
make gocore
```

or, to build the full suite of utilities:

```shell
make all
```

### To build or run on Windows

**Note: It is important to note that gocore requires mingw to run and be built on Windows.**

To install mingw:

```shell
choco install mingw
```

## ICAN Network Prefixes

The CORE Client implements ICAN-based addresses with the following formats:

Name | Prefix | Length | Format
--- | --- | --- | ---
Mainnet | CB | 44 | hh!kk!hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh!
Testnets | AB | 44 | hh!kk!hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh!
Privatenets | CE | 44 | hh!kk!hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh!

## Executables

Below, find the wrappers and executables in the `cmd` directory that the go-core project contains.

Command | Description
--- | ---
gocore | The main Core CLI client that provides the network with a point of entry (mainnet, testnet, or private net), running as a full node (default), archive node (that maintains all historical states), or a light node (live data recollection). Capable of being utilized by other processes as a gateway into the CORE network via JSON RPC endpoints exposed on top of HTTP, WebSocket, and/or IPC transports. To access command-line options, type `gocore --help`.
abigen | Generator of source code capable of converting definitions of CORE contracts into user-friendly, compatible, and secure Go packages. Working with basic CORE contract ABIs, it offers enhanced features when accompanied by the contract bytecode. Additionally, it accommodates Ylem source files, simplifying the development process considerably.
bootnode | This streamlined version of our CORE client implementation solely engages in the network node discovery protocol, abstaining from executing any of the more advanced application protocols. It serves as a nimble bootstrap node, facilitating the discovery of peers within private networks.
cvm | An iteration tailored for developers, the CVM (CORE Virtual Machine) utility possesses the ability to execute bytecode snippets within a customizable environment and mode of execution. Its primary function is to enable precise, isolated debugging of CVM opcodes.
gocorerpctest | A developer utility tool designed to bolster our core/rpc-test test suite, ensuring adherence to the fundamental specifications outlined in the CORE JSON RPC standards.
rlpdump | A specialized developer tool that transforms binary RLP (Recursive Length Prefix) dumps — utilized by the CORE protocol in both network and consensus contexts — into a more intuitive, user-friendly hierarchical format.

### Full Node on the Main CORE Network

The prevalent scenario encountered involves individuals seeking basic interaction with the CORE network. This includes tasks such as setting up accounts, conducting fund transfers, and deploying or engaging with contracts. In this specific situation, there's no interest in delving into extensive historical data from years past. Therefore, a swift fast-sync to the present state of the network suffices. To initiate this process:

```shell
gocore console
```

Executing this command will:

* Initiate `gocore` in rapid synchronization mode (which is the default setting, but adjustable using the `--syncmode` parameter). This mode entails downloading more data in exchange for bypassing the processing of the complete historical record of the CORE network, a task demanding significant CPU resources.
* Launch gocore's integrated interactive JavaScript console (accessed via the appended `console` subcommand). Through this interface, you gain the ability to call upon all official `web3` methods as well as gocore's proprietary management APIs. It's worth noting that this tool is optional, and if omitted, you can always link it to a currently active gocore instance using the command `gocore attach`.

### A Full Node on the Devín Network (PoW)

For developers delving into CORE contract development, it's advisable to initially explore the process without the involvement of actual currency. Familiarity with the system is key before committing real resources. Rather than connecting to the primary network, consider affiliating your node with the test network known as Devín. This environment mirrors the main network in functionality but operates exclusively with simulated play-Core funds. This approach provides a secure space for developers to experiment and refine their contract-building skills without any financial risk. By engaging with the Devín test network, you're working in an environment that faithfully replicates the main network's capabilities, using play-Core as your testing currency. This way, you can gain proficiency and confidence in CORE contract creation before venturing into live transactions.

```shell
gocore --devin console
```

Equally significant within the Devín network, the `console` subcommand retains its identical significance as previously outlined. Should you have overlooked the explanations provided here, we encourage you to refer back to the preceding content for a comprehensive understanding.

Let’s describe the `--devin` flag while reconfiguring your gocore instance:

* When opting for gocore, it veers from the standard data directory path, residing a level deeper within a devin subdirectory (such as `~/core/devin` on Linux, as opposed to the default `~/core`). It's essential to bear in mind that on both macOS and Linux platforms, this configuration necessitates the specification of a custom endpoint when attaching to a running devin node. This deviation arises because the default behavior of `gocore attach` is to seek a connection with a production node endpoint.
* Furthermore, instead of interfacing with the primary Core network, the client seamlessly integrates with the test network. This transition entails a shift in P2P bootnodes, network IDs, and genesis states, ensuring a distinct environment tailored for testing and development purposes.

### Configuration

Instead of burdening the `gocore` binary with an array of flags, there's an alternative approach. You can provide a configuration file using the following syntax:

```shell
gocore --config /path/to/your_config.toml
```

For a visual reference on how the file's structure should be, consider using the `dumpconfig` subcommand. This enables you to export your current configuration effortlessly, even incorporating your preferred flags:

```shell
gocore --your-favourite-flags dumpconfig
```

#### Docker Quick Start

Swiftly initiating Core on your system can be effortlessly achieved through Docker. Execute the following command:

```shell
docker run -d --name core-node -v /Users/robocop/core-coin:/root \
           -p 8545:8545 -p 30300:30300 \
           core-coin/client-go
```

This command not only initiates gocore in fast-sync mode with a 1GB DB memory allocation, mirroring the earlier approach, but also establishes a lasting volume in your home directory for preserving your blockchain data. Furthermore, it facilitates the mapping of default ports.

It's crucial to remember to include `--rpcaddr 0.0.0.0` if you intend to access RPC from other containers or hosts. By default, gocore is bound to the local interface, restricting external accessibility to RPC endpoints.

### Programmatically Interfacing with Gocore Nodes

For developers, transitioning from manual console interaction to programmatic interfacing with gocore nodes is a natural progression. To facilitate this shift, gocore offers built-in support for APIs based on JSON-RPC, as well as specific gocore APIs. These interfaces can be accessed through HTTP, WebSockets, and IPC. The IPC interface is activated by default, granting access to the complete range of gocore APIs.

On the other hand, the HTTP and WS interfaces require manual activation and provide a limited subset of APIs due to security considerations. These interfaces can be toggled on or off and customized to suit your requirements.

If you opt for HTTP-based JSON-RPC API, you have a variety of options to configure:

* `--rpc` Enables the HTTP-RPC server.
* `--rpcaddr` Specifies the interface on which the HTTP-RPC server listens (default: `localhost`).
* `--rpcport` Sets the port for the HTTP-RPC server (default: `8545`).
* `--rpcapi` Defines the APIs accessible via the HTTP-RPC interface (default: `xcb,net,web3,sc`).
* `--rpccorsdomain` Designates a comma-separated list of domains from which cross-origin requests are accepted (subject to browser enforcement).
* `--ws` Activates the WS-RPC server.
* `--wsaddr` Specifies the interface for the WS-RPC server to listen on (default: `localhost`).
* `--wsport` Sets the port for the WS-RPC server (default: `8546`).
* `--wsapi` Specifies the APIs accessible through the WS-RPC interface (default: `xcb,net,web3,sc`).
* `--wsorigins` Indicates origins from which websocket requests are accepted.
* `--ipcdisable` Disables the IPC-RPC server.
* `--ipcapi` Specifies the APIs accessible via the IPC-RPC interface (default: `admin,debug,xcb,miner,net,personal,txpool,web3,sc`).
* `--ipcpath` Specifies the filename for the IPC socket/pipe within the datadir (explicit paths require escaping).

To establish connections via HTTP, WS, or IPC to a gocore node configured with these flags, you'll need to leverage the capabilities of your programming environment (libraries, tools, etc.) and communicate using JSON-RPC on all transports.

Importantly, you can reuse the same connection for multiple requests. It's crucial, however, to be mindful of the security implications of exposing an HTTP/WS-based transport, as there are active attempts by malicious actors to compromise CORE nodes with accessible APIs.

Additionally, be aware that all browser tabs have access to locally running web servers, potentially opening avenues for malicious web pages to exploit locally available APIs.

### Private Network Operation

Establishing and managing your private network requires a more hands-on approach, as many configurations typically handled automatically in official networks must now be set up manually.

#### Defining the Private Genesis State

The initial step involves crafting the genesis state for your network, a crucial consensus point that all nodes must acknowledge and concur upon. This involves creating a concise JSON file (let's name it `genesis.json`):

```json
{
  "config": {
    "networkId": "<arbitrary positive integer>"
  },
  "alloc": {},
  "corebase": "ce450000000000000000000000000000000000000000",
  "difficulty": "0x20000",
  "extraData": "",
  "energyLimit": "0x2fefd8",
  "nonce": "0x0000000000000047",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp": "0x00"
}
```

While the provided fields suffice for most cases, it's advisable to alter the `nonce` to a random value to thwart potential connections from unknown remote nodes. If you wish to pre-fund specific accounts for streamlined testing, generate the accounts and populate the `alloc` field with their respective addresses:

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

After defining the genesis state in the provided JSON file, it's imperative to initialize every gocore node with this configuration before commencing operations to ensure all blockchain parameters are accurately configured:

```shell
gocore init path/to/genesis.json
```

#### Setting Up the Meeting Point

After initializing all the nodes you intend to run with the desired genesis state, the next step is to launch a bootstrap node that others can utilize for discovering one another within your network or even over the internet. The preferred method is to configure and activate a dedicated bootnode:

```shell
bootnode --genkey=boot.key
bootnode --nodekey=boot.key
```

With the bootnode now active, it will present an `enode` URL for other nodes to connect to and exchange peer details. Remember to substitute the IP address details displayed (most likely [::]) with your externally accessible IP to obtain the actual `enode` URL.

*Note: While you could potentially utilize a fully-fledged gocore node as a bootnode, it's considered the less recommended approach.*

#### Launching Your Member Nodes

With the bootnode up and externally reachable (you can verify its accessibility with `telnet <ip> <port>`), initiate each subsequent gocore node directed towards the bootnode for peer discovery through the `--bootnodes` flag. It's also advisable to keep the data directory of your private network distinct, so be sure to specify a custom `--datadir` flag.

```shell
gocore --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

*Note: Since your network will be completely isolated from the main and test networks, you'll also need to configure a miner to process transactions and generate new blocks for you.*

#### Running a Private Miner

In a private network context, a single CPU miner instance suffices for practical purposes as it can generate a steady flow of blocks at the appropriate intervals without requiring substantial resources (consider running on a single thread, multiple threads are unnecessary). To initiate a gocore instance for mining, run it with your customary flags, extended by:

```shell
gocore <usual-flags> --mine --miner.threads=1 --corebase=ce450000000000000000000000000000000000000000
```

This will commence mining blocks and transactions on a single CPU thread, attributing all earnings to the account specified by `--corebase`. Further adjustments to the mining process can be made by altering the default energy limit blocks converge to (`--targetenergylimit`) and the price at which transactions are accepted (`--energyprice`).

#### Initiating the Transaction

Here's a guide on how to execute a transaction using the go-core client.

* Launch the go-core Client
   `./gocore --verbosity 2 --nat any console`
* Retrieve the Latest Transaction for Nonce [0x0] (0 for the first; 1 for the second; and so on)
   `web3.xcb.getTransactionCount("cb…")`
   (Note: Transactions with the same nonce but a higher fee may serve as replacements)
   > Expected result: a number
* Compose the Transaction
   `var tx = {nonce: '0x0', energy: 21000, energyPrice: 1000000000, to : "cb…", value: web3.toOre(1), from: "cb…"}`
   (Note: web3.toOre(1) represents the value in Cores)
   > Expected result: undefined
* Unlock the Account
   `personal.unlockAccount("cb…")`
   > Enter Passphrase or leave it blank.
   > Expected result: true
* Sign the Transaction with the Private Key
   `var txSigned = xcb.signTransaction(tx)`
   > Expected result: undefined
* Obtain the Raw Transaction for future broadcasting
   `txSigned.raw`
   > Expected result: Raw transaction
* Broadcast the Transaction (while online)
   `xcb.sendRawTransaction(txSigned.raw)`

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
S1 Critical | Outage, broken feature with no workaround
S2 High | Broken feature, workaround too complex & unacceptable
S3 Moderate | Broken feature, workaround acceptable
S4 Low | Functionality inconvenience or cosmetic issue
S5 Note | Note about feature and/or code

## Contribution

We greatly appreciate your interest in lending a hand to enhance the source code! We extend a warm welcome to contributions from individuals across the internet, and value even the tiniest of adjustments!

If you're inclined to make a contribution to go-core, kindly commence by forking the repository, effecting the necessary fixes, committing your changes, and subsequently dispatching a pull request. This allows our maintainers to scrutinize and integrate your alterations into the primary codebase. However, if you plan to propose more intricate modifications, we recommend reaching out to the core developers first via the [Core ◆ Talk](https://coretalk.space) forum. This step ensures that your proposed changes align with the overarching philosophy of the project, and also grants you the opportunity to receive early feedback. This can streamline both your efforts and our subsequent review and integration procedures.

We kindly request that your contributions align with our coding principles:

* Code should conform to the established [Go formatting](https://golang.org/doc/effective_go.html#formatting) standards, as outlined in the official Go guidelines (i.e., it should be formatted using [gofmt](https://golang.org/cmd/gofmt)).
* Code must be thoroughly documented, adhering to the official [Go commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
* Pull requests should be established based on, and directed towards, the `master` branch.
* Commit messages ought to be prefixed with the package(s) they pertain to.
  * For instance, "xcb, rpc: implement optional trace configurations."

## Security Declaration

Please report suspected security vulnerabilities in private following the [Security manual](https://dev.coreblockchain.net/docs/bug). Do NOT create publicly viewable issues for suspected security vulnerabilities. For more information, please look into [Security recommendations](SECURITY.md).

## License

Licensed under the [CORE License](LICENSE).

## References

* [CORE Blockchain Postman Collection](https://www.postman.com/core-labs/core-blockchain/overview)
* [CORE Improvement Proposals](https://cip.coreblockchain.net)

## Community

[![Developer Portal](https://img.shields.io/badge/Developer-dev.coreblockchain.cc-46b549)](https://dev.coreblockchain.net/)
[![Core ◆ Talk](https://img.shields.io/badge/Core%20%E2%97%86%20Talk-Protocol%20and%20Client-green)](https://coretalk.space)

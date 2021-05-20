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

// This file is a test-utility for testing clef-functionality
//
// Start clef with
//
// build/bin/clef --4bytedb=./cmd/clef/4byte.json --rpc
//
// Start gocore with
//
// build/bin/gocore --nodiscover --maxpeers 0 --signer http://localhost:8550 console --preload=cmd/clef/tests/testsigner.js
//
// and in the console simply invoke
//
// > test()
//
// You can reload the file via `reload()`

function reload(){
	loadScript("./cmd/clef/tests/testsigner.js");
}

function init(){
    if (typeof accts == 'undefined' || accts.length == 0){
        accts = xcb.accounts
        console.log("Got accounts ", accts);
    }
}
init()
function testTx(){
    if( accts && accts.length > 0) {
        var a = accts[0]
        var txdata = xcb.signTransaction({from: a, to: a, value: 1, nonce: 1, energy: 1, energyPrice: 1})
        var v = parseInt(txdata.tx.v)
        console.log("V value: ", v)
        if (v == 37 || v == 38){
            console.log("Mainnet 155-protected networkid was used")
        }
        if (v == 27 || v == 28){
            throw new Error("Mainnet networkid was used, but without replay protection!")
        }
    }
}
function testSignText(){
    if( accts && accts.length > 0){
        var a = accts[0]
        var r = xcb.sign(a, "0x68656c6c6f20776f726c64"); //hello world
        console.log("signing response",  r)
    }
}
function testClique(){
    if( accts && accts.length > 0){
        var a = accts[0]
        var r = debug.testSignCliqueBlock(a, 0); // Sign genesis
        console.log("signing response",  r)
        if( a != r){
            throw new Error("Requested signing by "+a+ " but got sealer "+r)
        }
    }
}

function test(){
    var tests = [
        testTx,
        testSignText,
        testClique,
    ]
    for( i in tests){
        try{
            tests[i]()
        }catch(err){
            console.log(err)
        }
    }
 }

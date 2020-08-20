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

const TodoList = artifacts.require('./OpCodes.sol')
const assert = require('assert')
let contractInstance
const Web3 = require('web3');
const web3 = new Web3(new Web3.providers.HttpProvider('http://localhost:8545'));
// const web3 = new Web3(new Web3.providers.HttpProvider('http://localhost:9545'));

contract('OpCodes', (accounts) => {
   beforeEach(async () => {
      contractInstance = await TodoList.deployed()
   })
   it('Should run without errors the majorit of opcodes', async () => {
     await contractInstance.test()
     await contractInstance.test_stop()

   })

   it('Should throw invalid op code', async () => {
    try{
      await contractInstance.test_invalid()
    }
    catch(error) {
      console.error(error);
    }
   })

   it('Should revert', async () => {
    try{
      await contractInstance.test_revert()    }
    catch(error) {
      console.error(error);
    }
   })
})

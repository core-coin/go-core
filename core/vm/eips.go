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

package vm

func opSelfBalance(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	balance := interpreter.intPool.get().Set(interpreter.cvm.StateDB.GetBalance(contract.Address()))
	stack.push(balance)
	return nil, nil
}

// opChainID implements CHAINID opcode
func opChainID(pc *uint64, interpreter *CVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	chainId := interpreter.intPool.get().Set(interpreter.cvm.chainConfig.ChainID)
	stack.push(chainId)
	return nil, nil
}
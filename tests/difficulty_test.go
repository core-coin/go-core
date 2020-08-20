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

package tests

import (
	"math/big"
	"testing"

	"github.com/core-coin/go-core/params"
)

var (
	mainnetChainConfig = params.ChainConfig{
		ChainID: big.NewInt(1),
	}
)

func TestDifficulty(t *testing.T) {
	t.Parallel()

	dt := new(testMatcher)
	// Not difficulty-tests
	dt.skipLoad("hexencodetest.*")
	dt.skipLoad("crypto.*")
	dt.skipLoad("blockgenesistest\\.json")
	dt.skipLoad("genesishashestest\\.json")
	dt.skipLoad("keyaddrtest\\.json")
	dt.skipLoad("txtest\\.json")

	dt.config("Devin", *params.DevinChainConfig)
	dt.config("Morden", *params.DevinChainConfig)

	dt.config("MainNetwork", mainnetChainConfig)
	dt.config("CustomMainNetwork", mainnetChainConfig)
	dt.config("difficulty.json", mainnetChainConfig)

	dt.walk(t, difficultyTestDir, func(t *testing.T, name string, test *DifficultyTest) {
		cfg := dt.findConfig(name)
		if test.ParentDifficulty.Cmp(params.MinimumDifficulty) < 0 {
			t.Skip("difficulty below minimum")
			return
		}
		if err := dt.checkFailure(t, name, test.Run(cfg)); err != nil {
			t.Error(err)
		}
	})
}

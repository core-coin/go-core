// Copyright 2015 The go-core Authors
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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/core-coin/go-core/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("xcc/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("xcc/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("xcc/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("xcc/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("xcc/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("xcc/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("xcc/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("xcc/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("xcc/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("xcc/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("xcc/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("xcc/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("xcc/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("xcc/downloader/states/drop", nil)
)

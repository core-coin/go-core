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
	headerInMeter      = metrics.NewRegisteredMeter("xce/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("xce/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("xce/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("xce/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("xce/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("xce/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("xce/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("xce/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("xce/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("xce/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("xce/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("xce/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("xce/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("xce/downloader/states/drop", nil)
)

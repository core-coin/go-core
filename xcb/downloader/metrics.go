// Copyright 2020 by the Authors
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
	"github.com/core-coin/go-core/v2/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("xcb/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("xcb/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("xcb/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("xcb/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("xcb/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("xcb/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("xcb/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("xcb/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("xcb/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("xcb/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("xcb/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("xcb/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("xcb/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("xcb/downloader/states/drop", nil)
)

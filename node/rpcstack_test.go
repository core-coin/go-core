// Copyright 2015 by the Authors
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

package node

import (
	"bytes"
	"github.com/core-coin/go-core/internal/testlog"
	"github.com/core-coin/go-core/log"
	"github.com/core-coin/go-core/rpc"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// TestCorsHandler makes sure CORS are properly handled on the http server.
func TestCorsHandler(t *testing.T) {
	srv := createAndStartServer(t, &httpConfig{CorsAllowedOrigins: []string{"test", "test.com"}}, false, &wsConfig{})
	defer srv.stop()
	url := "http://" + srv.listenAddr()

	resp := rpcRequest(t, url, "origin", "test.com")
	assert.Equal(t, "test.com", resp.Header.Get("Access-Control-Allow-Origin"))

	resp2 := rpcRequest(t, url, "origin", "bad")
	assert.Equal(t, "", resp2.Header.Get("Access-Control-Allow-Origin"))
}

// TestVhosts makes sure vhosts are properly handled on the http server.
func TestVhosts(t *testing.T) {
	srv := createAndStartServer(t, &httpConfig{Vhosts: []string{"test"}}, false, &wsConfig{})
	defer srv.stop()
	url := "http://" + srv.listenAddr()

	resp := rpcRequest(t, url, "host", "test")
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	resp2 := rpcRequest(t, url, "host", "bad")
	assert.Equal(t, resp2.StatusCode, http.StatusForbidden)
}

// TestWebsocketOrigins makes sure the websocket origins are properly handled on the websocket server.
func TestWebsocketOrigins(t *testing.T) {
	srv := createAndStartServer(t, &httpConfig{}, true, &wsConfig{Origins: []string{"test"}})
	defer srv.stop()

	dialer := websocket.DefaultDialer
	_, _, err := dialer.Dial("ws://"+srv.listenAddr(), http.Header{
		"Content-type":          []string{"application/json"},
		"Sec-WebSocket-Version": []string{"13"},
		"Origin":                []string{"test"},
	})
	assert.NoError(t, err)

	_, _, err = dialer.Dial("ws://"+srv.listenAddr(), http.Header{
		"Content-type":          []string{"application/json"},
		"Sec-WebSocket-Version": []string{"13"},
		"Origin":                []string{"bad"},
	})
	assert.Error(t, err)
}

func Test_checkPath(t *testing.T) {
	tests := []struct {
		req      *http.Request
		prefix   string
		expected bool
	}{
		{
			req:      &http.Request{URL: &url.URL{Path: "/test"}},
			prefix:   "/test",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/testing"}},
			prefix:   "/test",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/"}},
			prefix:   "/test",
			expected: false,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/fail"}},
			prefix:   "/test",
			expected: false,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/"}},
			prefix:   "",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/fail"}},
			prefix:   "",
			expected: false,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/"}},
			prefix:   "/",
			expected: true,
		},
		{
			req:      &http.Request{URL: &url.URL{Path: "/testing"}},
			prefix:   "/",
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expected, checkPath(tt.req, tt.prefix))
		})
	}
}

func createAndStartServer(t *testing.T, conf *httpConfig, ws bool, wsConf *wsConfig) *httpServer {
	t.Helper()

	srv := newHTTPServer(testlog.Logger(t, log.LvlDebug), rpc.DefaultHTTPTimeouts)
	assert.NoError(t, srv.enableRPC(nil, *conf))
	if ws {
		assert.NoError(t, srv.enableWS(nil, *wsConf))
	}
	assert.NoError(t, srv.setListenAddr("localhost", 0))
	assert.NoError(t, srv.start())
	return srv
}

// wsRequest attempts to open a WebSocket connection to the given URL.
func wsRequest(t *testing.T, url, browserOrigin string) error {
	t.Helper()
	t.Logf("checking WebSocket on %s (origin %q)", url, browserOrigin)

	headers := make(http.Header)
	if browserOrigin != "" {
		headers.Set("Origin", browserOrigin)
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if conn != nil {
		conn.Close()
	}
	return err
}

// rpcRequest performs a JSON-RPC request to the given URL.
func rpcRequest(t *testing.T, url string, extraHeaders ...string) *http.Response {
	t.Helper()

	// Create the request.
	body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"rpc_modules","params":[]}`))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatal("could not create http request:", err)
	}
	req.Header.Set("content-type", "application/json")

	// Apply extra headers.
	if len(extraHeaders)%2 != 0 {
		panic("odd extraHeaders length")
	}
	for i := 0; i < len(extraHeaders); i += 2 {
		key, value := extraHeaders[i], extraHeaders[i+1]
		if strings.ToLower(key) == "host" {
			req.Host = value
		} else {
			req.Header.Set(key, value)
		}
	}

	// Perform the request.
	t.Logf("checking RPC/HTTP on %s %v", url, extraHeaders)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// Code generated by go-bindata.
// sources:
// faucet.html
// DO NOT EDIT!

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _faucetHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x5a\xdb\x6e\xeb\x3a\x7a\xbe\x9f\xa7\xe0\x78\xf6\xc2\x72\x9a\x44\xb2\x1d\xe7\xb0\x12\x3b\x53\xdb\xb1\x73\xb4\x13\xc7\x76\x4e\x0b\x0b\x03\x4a\xa2\x24\xc6\x12\x29\x93\x94\x6c\x25\x63\x60\x5e\xa0\x68\x8b\xe9\x4d\x3b\x17\x1b\x98\x8b\xe9\xed\x14\xbd\xe9\xf3\xec\x17\xe8\x7e\x84\x82\x94\xe4\x63\x56\xd6\x42\x67\xa6\xe8\x06\xb6\x96\x44\xfe\xfc\xf9\xfd\x47\xfe\xfc\xe3\xca\x2f\x4f\xae\x1b\xfd\xc7\x9b\x26\x70\x85\xef\x1d\xff\xa2\x22\xff\x01\x1e\x24\x4e\x35\x87\x48\x4e\x0e\x20\x68\x1d\xff\x02\x80\x8a\x8f\x04\x04\xae\x10\xc1\x36\x1a\x85\x38\xaa\xe6\x1a\x94\x08\x44\xc4\x76\x3f\x0e\x50\x0e\x98\xc9\x57\x35\x27\xd0\x44\xe8\x92\xcd\x11\x30\x5d\xc8\x38\x12\xd5\x41\xbf\xb5\x7d\x90\x9b\x73\xc9\xc6\x73\xa1\xb0\x97\x26\x08\xf4\x51\x35\x17\x61\x34\x0e\x28\x13\x0b\x4c\xc7\xd8\x12\x6e\xd5\x42\x11\x36\xd1\xb6\xfa\xd8\x02\x98\x60\x81\xa1\xb7\xcd\x4d\xe8\xa1\x6a\x31\x61\x23\xb0\xf0\xd0\xf1\xeb\xab\xd6\x41\x62\x4c\xd9\xb0\x2f\xbf\xa7\x53\xf0\xf3\x8f\xff\xf6\x67\xd0\xa0\x0c\x81\x5a\x28\x5c\x44\x04\x36\xa1\x40\x16\x68\xc1\xd0\x44\xa2\xa2\x27\xeb\x24\x07\x0f\x93\x21\x60\xc8\xab\xe6\xb8\x88\x3d\xc4\x5d\x84\x44\x0e\xb8\x0c\xd9\xd5\x9c\x14\x9f\x1f\xea\xba\x69\x11\xed\x99\x5b\xc8\xc3\x11\xd3\x08\x12\xba\xe3\xea\x26\x8b\x03\x41\xdd\xd0\xd8\xb6\xb0\x83\x05\xf4\x74\x93\x9a\xf4\xef\x8b\x5a\x51\x2b\xe9\x16\xe6\x42\x37\x39\x57\x63\x9a\x8f\x89\x66\x72\x9e\x03\x98\x08\xe4\x30\x2c\xe2\x6a\x8e\xbb\x70\xe7\xa0\xbc\x7d\x11\x79\x57\x25\xaf\x85\x5c\x13\x11\x5f\x5c\x61\xfb\xda\xe8\xfb\x4f\xa3\x6b\xef\xf4\x49\x74\x42\x7b\x18\x9d\x7b\x17\x2d\x71\xbe\x79\x3e\xde\x7f\xd9\xed\xdd\x96\x9e\x7a\x9d\xf8\xea\xea\xec\xd1\xd8\x9b\x20\xbf\x1b\x5c\xe4\x80\xc9\x28\xe7\x94\x61\x07\x93\x6a\x0e\x12\x4a\x62\x9f\x86\x3c\x07\xf4\xb9\x74\x89\x30\x16\x14\xf0\x10\xfb\xd0\x41\x7a\x40\x9c\x23\x03\x72\xb4\x57\xde\xc2\x77\xf5\xeb\xdb\x71\xe1\xf2\xd4\xa1\xb5\x5a\xad\xd6\xe9\x0d\xdc\xe6\xc0\xa9\xd5\x6a\x0d\xf9\x59\x73\x1a\xb5\xc7\x5a\xad\x56\x7f\x41\x84\x15\xe4\x40\xe3\xa1\x7e\x7e\xff\xd0\xae\xd5\x6a\xbc\x57\xab\xd5\xae\x9a\xcc\x59\x6f\xe7\x61\xff\x40\x4e\xb9\x4d\xaf\xd9\xbd\xbb\x7d\x74\xcd\xbb\x87\xb8\xbc\x63\xb4\xeb\xf5\x4f\x4d\x1b\x22\xeb\xea\x04\x35\x46\x93\x41\xb3\x66\x3b\xb7\xfd\xc9\xee\xe5\x81\x1b\xd6\x0b\xb4\x35\xe8\x07\xe3\x4b\xea\xf6\xbb\x46\xf7\xe1\xa0\x77\xee\xf4\x1c\x1b\x35\xfc\x8b\x1d\x7e\xc9\x6a\xe3\x5e\x7c\x7a\x32\xdc\xb9\xc1\x9d\xae\xdf\xad\xef\x3b\x8d\x42\x71\xb4\xe3\xbc\xd8\xf6\xcd\xf3\x08\x9d\x8c\x71\xfd\xc0\x78\xea\x8c\xeb\xf7\x7a\x23\xda\xfd\xb4\x3b\xba\xb9\x71\x2e\xca\x35\x56\x3b\x69\xdf\xdc\xde\x15\x4b\x4f\x9f\x5a\xcd\x46\xcf\x39\x6b\x95\x6a\xb0\x56\xe7\xb5\x72\x3b\x78\xae\x35\x4a\xb5\xab\xc7\x4f\xad\xfb\x1d\xfb\xa5\x59\x2f\xc7\xe3\x27\x3d\xe8\xb3\xe7\xc7\xc6\x6d\xaf\xdf\x7b\x69\x9c\x35\xca\xf5\xf1\xc9\x55\xf4\x52\x43\x65\x6c\xb5\xf8\xe5\xe6\x5e\x10\x77\xeb\xe5\xd2\xcb\xc5\x84\xdc\x9c\x90\x6e\xd4\x8b\x7b\xd4\x37\x07\x8f\x4e\x61\xe7\x66\x6f\xd0\x8e\xf6\xbb\x41\xad\xbc\x3f\xae\x9b\x17\x07\x7b\x68\x68\xe0\xcb\x0e\xeb\xf6\x27\xd8\x6f\xf7\xef\x3e\x15\xf9\x7d\xf3\xe6\xde\xdf\x6c\xf3\xb8\xb7\x79\x72\x30\x42\x17\x8f\xe2\x65\xff\xd6\xba\x6e\xf6\x4f\xae\x9b\x51\x18\xc5\xfc\xf4\xc6\xb6\xee\xc4\xc5\xf3\x78\xdf\x31\x6e\xfc\xbb\xd2\x95\xc1\x06\xfa\xdd\xa5\x3e\xbe\x88\xee\x0d\x27\x1a\x6d\xee\x19\x8f\xfb\x93\x7a\xbd\x5c\x1e\x93\x68\xb3\x3c\xb8\x75\xcc\xb9\xa9\x47\x07\xf5\xf2\x7e\xed\xd9\xda\x2c\x16\x6d\xab\x2f\x02\x1e\x5b\xdd\x86\x7e\x60\x05\x37\x2f\x1e\xec\xdc\xb5\x4b\x61\xe4\x0d\xfb\xf7\x63\xfc\xdc\x1d\x5e\x46\x9b\xed\x5a\x67\xf3\xe9\xe2\xe4\xa5\x13\x5e\x74\xbd\x5e\xa9\x25\xfa\xf5\x21\x69\xb5\xc7\x7b\x93\x11\x6e\xda\xc3\xdd\x49\x69\xd4\x6b\x37\x37\x9d\xcb\x16\xbf\xda\x9b\x3c\x14\x27\x1e\x8f\x83\xe7\xba\x79\x7d\x7a\x6a\x3f\x8c\x1f\xee\x77\xbc\xd6\x10\xd6\x7a\xe4\x21\x18\xb8\xc1\x75\xfb\xb6\x13\x9f\xb8\x57\xbb\xfb\x07\x67\xbd\xcd\xcd\x36\xd7\x45\xe3\xbc\x57\x73\xad\x89\xde\xe8\x34\x79\x79\x50\x37\xcc\x89\x7e\x60\x34\x42\x7f\xb3\x65\x18\x4f\xdd\xb2\x11\xbf\xdc\x5d\xc1\x51\xf3\xae\xd6\x88\xd0\xfd\x20\xe6\x67\x03\xef\xfe\x8e\x9c\xb1\x7b\xa7\xd1\x2b\xc3\xd2\xfe\xd5\xc5\x4e\xfd\xb2\xd7\x1d\xb0\x09\x2b\xe0\xee\x25\xe9\xed\x3b\x34\x1c\x3e\x5d\x90\x56\xf9\xfe\x1c\x1e\xb4\xc5\xe8\xee\xe5\xf1\x66\xe7\xb2\x3f\x44\xa3\xb3\x4e\xbd\x75\x1b\x84\x31\x0a\x37\x5b\xcd\xc6\x4e\x29\x12\x77\x7a\x29\xa6\x56\xb1\x77\xfd\xb2\x7f\x6e\x4f\x5a\x36\x34\x26\xf5\xab\xee\x7d\xcf\xdb\xbb\x1b\x99\xad\x5e\xb7\x74\xb9\x17\x8e\xce\xa2\xa1\x63\xd4\xea\xac\x7c\xe1\x6f\xba\xce\xe9\xa6\x70\xdd\xa8\x75\xdf\xb6\xf7\x5b\xd7\xc5\x41\x13\xef\xdf\xe3\x7a\x6f\xa4\x3b\x3b\xcd\xd2\x6d\xf9\x04\xa1\x76\x09\xf5\x36\x6f\x9e\xce\xad\xdb\xe6\x4d\xbd\x33\xfa\x34\x19\xd8\x67\x4e\xcf\xbb\xda\x77\x5b\xc5\xb2\xbf\x7b\x70\x5a\x28\x9f\x16\xdc\xfe\x81\x75\x4b\x5b\xfe\x79\xbb\x31\xdc\x7b\xba\x1e\xdc\xd9\xfd\x96\x0d\xc3\xb8\xf1\x70\xdb\xd8\x3b\x37\x1a\x25\x6f\x74\x43\xe3\x47\xb7\x25\xf4\x83\xb3\xe7\xa7\xb3\xf3\x61\xbc\x33\x34\x3a\x24\x8e\xcc\xc9\xc1\xa8\x3e\xd9\x7f\xbe\x28\xd9\x97\xd7\xf0\xd3\x73\xe8\x3e\xf6\xfc\xbb\xa7\xfb\xc6\xed\xed\x68\xf0\x7c\xd0\x39\x2d\x92\x13\xe3\x04\xd6\x71\x6d\x1c\xdc\xf8\xbb\x8f\x9b\x9f\x86\xad\xc6\x41\x3d\xb2\x9a\x26\xeb\xf7\xf4\xd2\xe5\xe4\x51\x05\x66\x6f\x70\x77\x7d\x7b\xb9\xdb\x78\x3c\x3f\xaf\xe6\x92\xfc\x85\x4d\x4a\x72\x40\xc4\x01\xaa\xe6\x66\x71\x9e\x25\x02\x6e\x32\x1c\x08\xc0\x99\xb9\x94\xd6\x9e\xb9\x66\x7a\x34\xb4\x6c\x0f\x32\xa4\x99\xd4\xd7\xe1\x33\x9c\xe8\x1e\x36\xb8\xfe\x3c\x0a\x11\x8b\xf5\x1d\x6d\x57\x2b\xa4\x1f\x2a\x95\x3d\xaf\x65\xb2\xd2\xee\xde\xf6\xa4\xf3\xd2\x29\xc1\xb2\x27\x86\xf5\x72\xb9\x6d\xea\x17\x2f\x3b\x41\xbf\x8c\x07\x45\xd3\x47\xb7\x85\xd6\xf0\x81\x97\x03\x16\xea\x17\x13\xd8\xad\x7e\x35\x6b\x1d\x57\xf4\x04\xe8\x5f\x86\x79\x9b\x50\x11\xeb\x25\xad\xac\x15\x33\xe0\x72\xe4\x1d\xf4\xfa\xd0\x23\xe5\x13\x7c\x72\x56\x2f\xd8\x4f\x67\x03\x6b\x80\x9e\xea\xf1\xfe\xc3\x43\xbf\x7d\x71\x7d\xf0\x58\xb6\x37\x4f\xef\xca\x48\x14\x26\xa5\x6e\xfb\x6f\x88\xde\xa7\x3e\x22\x42\x7b\xe6\x7a\x49\x2b\x95\xb5\x42\x36\xf0\x75\xdc\x65\xdc\x7d\xda\xab\xdf\x5d\x95\x47\x9d\x4b\xaf\x5b\xda\xef\x37\x27\x4d\xb7\xde\x29\x9e\xb5\x6e\xa2\x5a\x54\x6b\x1b\x2d\x18\x5d\x5e\x52\xde\xbb\xff\x7e\xad\xcb\x63\x50\xbe\x01\xa0\xb9\x88\x51\xf0\xaa\xde\x01\xf0\x21\x73\x30\xd9\x36\xa8\x10\xd4\x3f\x04\x45\x6d\x17\xf9\x47\x6a\x6e\xaa\x9e\x16\x8e\x34\x68\x59\x0c\x71\xfe\xb5\x35\x05\x6d\xff\x8d\x35\x3e\x0d\x89\xf8\xea\x36\xcb\x0b\xa0\x61\xb0\x19\x69\x40\x39\x16\x98\x92\x43\x19\x00\x50\xe0\x08\x1d\xa5\x33\x66\xc8\x38\x65\x87\xc0\x45\x5e\xb0\xba\xfc\xd0\xa5\x11\x62\x87\x87\xd0\x16\xe8\x2d\x5e\xd0\xe0\xd4\x0b\xc5\x8c\xd7\x0c\x49\xa1\xf0\x21\x1b\x53\x35\xc9\x21\x28\x15\x32\x74\x00\x04\xd0\xb2\x30\x71\x16\x10\x03\x60\x40\x73\xe8\x30\x1a\x12\xeb\x10\xfc\xca\xb6\xed\x19\xbc\xa4\xd0\x39\x04\x50\x08\x96\x57\xc5\xc8\x46\x36\xf7\xb2\x8d\x89\x85\x26\x87\xa0\x38\x07\xc0\x2c\xc4\x0e\xc1\x4e\x30\x01\xcc\x31\xf2\xfb\xe5\x2d\x90\xfc\xbf\x01\x2c\x2a\x04\xb2\xe6\x22\x56\xf4\xd4\x7e\x15\x3d\x29\xe5\x2a\x06\xb5\x62\x60\x7a\x90\xf3\x6a\xce\x85\x7c\x9b\x0b\x6c\x0e\xe3\x6d\x9b\x52\x81\x58\x5a\x8c\x41\x4c\x12\x8b\x57\x38\x32\xa5\x16\x66\x0b\xa4\x03\x60\xbe\x1d\x30\xec\x43\x16\xe7\x8e\x53\x48\x15\x0b\x47\x8b\x34\xdb\x72\x97\xd9\xec\xf2\xbc\x14\x16\x62\x92\x6e\x36\xa3\x70\x8b\x19\x81\x92\x7f\x69\x12\x24\xb5\x5b\xdd\xa3\xe6\xd0\x74\x21\x26\x69\xe1\xa6\xea\xba\x45\x26\xba\x5b\x5c\x66\x5a\xca\x98\xf2\xd0\x78\x8b\xef\xbc\x54\x9c\x4e\xdf\xac\x0d\x97\xb9\x97\x16\x44\xd2\x2d\x1c\xcd\xe4\x9f\x7f\x54\xf4\x54\x67\xcb\x1a\x7c\x4b\x53\x6f\x69\xa2\xb2\xc0\x75\x95\x3e\x8d\xa6\x15\x11\x2a\x1e\x34\x90\x07\x6c\xca\xaa\x39\x4e\x4d\x0c\xbd\x5c\xb6\x40\xcd\xe4\x8e\x7b\x6a\x14\x90\x44\x50\x30\x60\xde\xeb\x2b\xb6\x81\xd6\xa1\x52\xe2\xe9\x14\xfc\x16\xdc\x43\xcf\x43\xe2\xf5\x15\x11\x6b\x3a\xad\xe8\x6a\xe1\xca\x36\x0b\x40\x6c\x8c\x3c\x0b\x48\xef\x81\x96\x45\xc9\x2a\xa2\x75\x29\x19\xf5\xa4\xdb\xa0\x49\x00\x89\x85\xac\x35\x7a\x00\x2a\x98\x04\xa1\x00\xd8\x5a\x13\x42\x4d\x64\x07\x99\xbc\x5e\xe4\x40\xe0\x41\x13\xb9\xd4\xb3\x10\x9b\x27\xd3\x9f\x7e\xf7\xa7\x35\xb1\x26\x66\x51\x0d\x2b\xb1\xd2\xb3\x6f\x69\x57\x7d\x45\xdd\x5f\x01\xff\x16\x60\x98\xd1\x18\xa1\x10\x94\x48\xf9\x78\x68\x9a\xd2\x40\x80\x12\xd3\xc3\xe6\xb0\x9a\x53\x88\x7e\xd0\x6e\x91\x09\x03\x61\xba\x70\x3a\x75\x58\xf6\xae\xa1\x09\x32\x43\x81\xf2\x1b\xaf\xaf\xc8\xe3\x68\x3a\xe5\xa1\xe1\x63\x91\x7f\x7d\xfd\x01\x5b\x93\xe9\x74\x23\x03\xbe\xbe\x3b\x00\xad\x90\x58\xc0\x47\xeb\xb8\x74\xf8\x1d\x52\xbe\x35\x14\xcc\x23\xd8\x0b\xde\xf2\x1a\x90\x3a\x2c\x26\x0e\x88\x69\xc8\x92\xa8\x5c\x0a\xa1\xc4\x8f\x40\xea\xa9\x2b\xf6\xa0\x6c\x65\x1e\x60\xc1\x91\x67\xcf\xdc\x2e\x58\xf2\xfc\x55\x88\x4b\xa1\xa0\x0e\x89\x77\x22\x21\x21\x98\x25\xaa\x95\x88\xa8\xa9\xd9\x6f\xfa\xf9\xd7\xac\xff\xfa\xca\x20\x71\x10\x90\x66\xda\x02\x3f\xa4\x27\xd6\x61\x15\x68\x09\x63\x3e\x9d\xae\x9a\x20\x81\x96\xf2\x65\xd0\xc2\xf4\xfd\x20\x58\xc5\x9f\xb8\x7f\xb2\x30\xbd\x28\x0b\x8c\xd8\x1b\x3e\x2d\xe1\xa5\x90\xa6\x53\xf0\xf9\xf5\x55\x9d\x1f\xe0\x07\xed\x06\x31\x4c\x2d\x0e\x12\xe7\xfa\xb2\xe6\x24\x6f\x28\x43\xf2\x52\xb6\x79\xdf\x79\xd6\x06\x12\xc3\x2f\x78\xfd\xa2\x56\x9d\xed\x59\x08\xe4\x80\xbc\x86\x6e\x73\x2c\xd0\x10\xc5\x32\x5c\x16\x17\xa5\xb3\x26\xf4\x3c\x79\x72\xaa\x24\xee\x63\x31\x5b\xf4\x22\x2b\x5b\x12\x61\x8e\x0d\x99\xd8\x13\x14\xab\x80\x57\xb0\x2d\x02\xe1\x02\x8a\xe5\xdc\xb5\x38\x2b\xa0\xc3\x01\x64\x68\xdb\x47\x16\x0e\xfd\x55\x5f\xe3\x01\x24\x0b\xa4\xb9\xe3\x9f\x7f\xfc\xc3\x9f\xd3\x61\x69\xc0\x00\x21\x96\xd4\x52\x01\x24\xc7\x40\x7d\xa6\x1f\xdf\x64\xf4\xef\xff\xb1\xc0\xc8\x90\xa7\xde\x9c\x91\xfa\xfc\x6e\x4e\x3f\xfe\x71\x81\x93\x1d\x12\x6b\x01\x92\x49\x19\xfa\x7e\x48\xff\xb9\xc2\x48\x66\xf1\x8c\x53\xf2\xbd\xce\x6a\x55\xf7\x2b\x9f\x2e\x9b\x7b\xef\x5f\xf7\x04\x5d\x9c\x47\x44\xc8\xdc\xfc\x86\x15\x2b\x6e\xe9\xf8\x8c\x8e\x81\x45\x11\x07\x76\x52\x51\xc8\x3c\xf6\xeb\xe5\x93\x5e\xa6\xc6\x65\x05\xf5\x5d\x3c\x5b\x80\x39\x60\x21\x51\x39\x91\x12\x20\xdc\x95\x84\x98\xa6\x4f\x0d\xf4\x29\x08\x18\x8a\x24\x1a\x1f\x7a\xd8\xc4\x34\xe4\x00\x9a\x82\x32\x0e\x6c\x46\x7d\x80\x26\x2e\x0c\xb9\x90\x8c\xa0\xe7\x01\x18\x41\xec\x41\xc3\x43\x4a\xbb\x5c\x26\x4f\x68\x9a\xa1\x1f\xca\xca\x96\x38\x00\x11\x1a\x3a\xae\x4a\xc1\x1c\x08\x0a\x92\x1c\xe4\x51\xe2\xcc\xf0\xf0\x00\xfa\xb2\xa4\x84\xe6\x90\x6f\x01\x86\x46\x21\xe2\x42\xf9\x34\x10\x18\x59\x72\x95\x49\x7d\x9f\x12\xb0\xc3\x2c\x10\x40\x26\x62\xc0\x97\xb3\x3e\x34\x4d\x95\xd0\x34\x50\x23\x31\x25\x08\xb8\x30\x52\x08\x41\x7f\x8c\x85\x2c\x98\x29\x03\x2d\x68\x22\x83\xd2\x19\x35\xf0\x61\x9c\x6d\x97\xa2\x1f\x63\xe1\xe2\x44\x3d\x01\x62\xbe\x5c\x6a\x01\x0f\xfb\x58\x70\x6d\xc9\x61\x96\x8f\x80\x70\x35\x39\x7b\x78\xed\x70\xe3\x82\x51\xe2\x1c\xff\xfc\xe3\x3f\xfd\x2b\x38\xc5\xc2\x0d\x0d\xe0\x60\x2e\x64\xe5\xab\x26\xa4\xe6\x97\xc1\x44\x18\x2e\x52\x6e\x01\x1f\x0e\x11\x80\xf2\x48\x5f\x6e\xf4\xc9\x59\xcd\x51\x94\xf2\x6a\x96\x03\x02\x32\x07\x89\x6a\xee\x37\x86\x07\xc9\x30\x77\x1c\x84\x86\x87\xcd\x74\x3f\x78\xac\xc4\x5c\x38\x1a\xb3\x63\x2e\x80\x5c\xca\x8b\x89\xa0\x4a\x05\xa9\x5b\x72\x90\xe7\x21\x53\xf7\x01\xa9\x53\x59\xdb\x28\x5f\x24\x1f\xa5\x0a\xa5\x76\x37\xb4\x8a\xc1\xf4\xe3\x06\x0d\xe2\x6d\xc5\x44\x2d\xaf\xa8\x5b\x8f\x2a\x67\xab\xb9\x86\x2c\x34\x32\xd7\xfb\xd8\x60\x08\x0a\x04\x16\x70\x7d\x04\x90\x58\xc0\xa4\x41\xac\x8e\x70\x68\x31\x05\x49\xb9\x9c\x42\x6a\x30\x3a\xe6\x32\x82\x24\xb5\xa4\xa9\xe8\x92\xff\xf1\x1c\x2e\x34\x68\x84\x40\x72\x36\x19\x74\xa2\x18\xda\x58\xca\x37\x86\xf1\x2f\x57\x8e\x86\x55\x0b\xbd\x6f\xb2\x3f\x65\x7e\xf4\x0d\x73\xa5\x54\x5f\x37\x95\x48\x08\xd4\x0d\x1a\x2b\xed\xea\x62\x8c\x90\xf8\xb5\xd4\x6a\xf5\x36\x61\x88\x89\xf3\xa1\x54\x48\x02\x57\xbe\x48\xf6\x1f\x4a\x05\x29\xe8\x87\x52\x61\x62\x16\x3f\x34\x4b\x1f\x0e\x0a\x1f\x6a\x7b\x1f\x4a\x05\x4a\x3e\x94\x0a\xc2\x45\x1f\x4a\x85\x0f\xa5\x9d\xc5\xa0\x4e\x46\xa4\x85\x25\x05\xe2\x92\x57\x16\xe7\xeb\x3e\xa2\x60\xfc\xbf\xf0\x8e\xb6\x6a\x1f\xa4\xe6\x23\x96\xac\xbe\xc0\xe0\xf6\x0a\xa4\x81\x99\xe1\x31\x20\x03\xd4\x4e\x80\x8e\x91\x91\x39\xc8\x96\x0c\x75\x1a\xa0\x84\xd8\x47\x24\xcc\xf8\x0a\x1a\x00\x86\x1d\x57\xc8\x65\xf3\x8d\xb6\x12\xcf\x53\x08\x24\x46\xa0\xba\xcf\x4a\x48\xcc\x53\xbd\xf0\xff\x63\x87\xfb\xe7\x7f\x98\x65\xab\x6f\x78\x5c\x46\xb6\x95\x04\x13\x77\x01\x04\x04\x8d\x41\x05\xf9\x69\xd8\x57\x74\xe4\x1f\x83\x80\x72\xf1\x35\xcb\x22\xdf\x40\x96\xf5\x86\x6d\xff\xea\xa6\xbd\x91\x28\xfa\xd8\x47\x5c\x40\x3f\xc8\xec\x30\x4b\xcc\x0a\xe4\xdc\x1a\x94\x24\x06\xb9\x92\x06\xa9\x65\xb7\x4a\x49\xf4\x37\xb0\xc7\xf2\x25\xe0\xfb\x4d\xf5\x2f\xbf\x07\xf7\x58\xb8\x34\x14\x00\xce\x2f\xe6\x98\x92\x77\x0c\x27\x8d\x33\xfe\xca\x22\xe4\x1f\x6f\x01\x8e\xfd\xc0\x8b\x55\x3a\x4c\xf5\xba\x6e\xb5\xef\x14\x5b\x19\x48\x55\x02\x3e\xb5\x90\xac\x03\x78\xc8\x4d\x14\x08\x59\x8d\x4a\x37\xaf\xc7\x2f\x90\x08\x4c\x50\x76\x06\x6b\xdf\xd4\xd4\x6a\xad\x5d\xd1\x17\x4f\xc0\xef\x6a\x39\x54\xf4\xac\x7d\x53\x49\x7a\x3a\x29\xcd\xbb\x55\xd3\xf2\xac\x17\xfa\x84\xcb\x9a\xc9\x81\x81\xb7\xdc\x70\x58\x27\x54\xed\xa0\xa4\xd7\xe5\xad\x36\x27\x32\x63\x2e\x74\x6b\xe6\xc6\x93\xb6\x7a\x23\x91\xcb\xaa\xd4\x98\x35\x7a\x34\xd3\x5c\xcf\xa8\x2b\xcd\x20\x95\x5b\xed\xf4\x2f\x79\xc8\xff\x7a\xbd\xf9\x2d\xec\xaa\xa3\x21\x83\x71\x3b\xc9\x66\x98\x6f\xbb\xd8\xb2\x10\xd9\x16\x34\x34\xdd\x15\xd1\x7c\xe8\x79\xc7\x3f\xfd\xe1\xf7\xff\xfd\x5f\xff\x08\x1a\x94\x10\x64\x0a\x55\x53\x1d\x82\xb5\xbf\x3f\x56\xf4\x84\xfa\xbd\x1b\xd2\x12\xef\x85\xfa\x7b\x0e\xc2\x42\x7c\x28\x68\xb0\x4d\x89\x17\xbf\x77\x1b\xb1\x70\x84\x2d\x69\xdc\x9f\x7e\xf7\xc7\xb7\xcb\x7b\x05\x26\x60\x54\x24\x98\x0d\x59\xad\x35\x12\x20\x6f\x40\x5d\xe7\xf1\xee\x1d\x6b\xc9\x31\xd3\xd7\x8a\x3e\xf7\xc5\xca\xbc\xb9\x0c\x80\xae\x83\x53\x8f\x1a\xd0\x03\x11\x64\x58\x96\xbb\xaa\x98\x75\xa9\x67\x25\xf9\x32\x64\x4c\xe6\x4b\x79\x55\x0b\x79\x96\xd7\xec\x79\x6b\x2e\x82\x4c\x86\x17\xf2\x03\x01\xaa\xa0\x70\x34\x1b\xe4\x88\x45\x88\xcd\xbf\xe5\x6d\x79\x99\x62\x56\x10\x57\xc1\xe7\x2f\x47\xbf\xc8\xf0\x34\x28\x89\x10\x13\x80\x84\xbe\x81\x98\x44\x43\xb0\x89\x18\xb0\x29\xf3\xe1\x7c\xd3\xde\xf9\x6f\x7a\x8f\xed\xfa\xf5\xd5\x6f\x9a\x0f\x7d\xc9\x22\x97\xdb\x02\xb9\xa1\x7c\xb4\xe5\xe3\x54\x3e\xfa\xf2\x71\x23\x1f\x4d\xf9\x78\x92\x8f\x47\xf9\xb8\x95\x8f\x6e\x2e\xdb\xd6\x0e\x49\xd2\x67\x95\x89\x17\x45\x18\x0a\xd4\x51\xdb\xe7\x13\x14\x1b\xb3\xbe\xb4\xdc\xda\x43\x11\xf2\x40\x15\xb4\xa1\x70\x35\x8f\x3a\xc5\xc2\x8c\x4c\x07\x3b\xe0\xb7\x99\x90\x00\x60\x3b\x9f\x12\x57\x41\x61\x03\x30\x24\x42\x46\x52\xc9\x8e\x16\x38\xf2\xd0\xb6\xf1\x04\x54\x97\xa5\xfa\xac\xd6\x7e\x59\x22\x34\xa1\x87\xb2\xad\x03\x3a\xce\x17\x0b\x5b\x29\x9e\xbf\x03\x3b\x1b\x6b\xa4\x16\xa8\x66\x8a\xd4\x93\x91\x8c\x24\xc5\x92\x50\x69\x82\xb6\xf0\x04\x59\xf9\xd2\x06\xd8\x4c\xd1\x64\x4d\xec\xcc\x2e\x27\xc8\x56\xb9\x74\xae\x2c\xe1\x42\x01\x4c\x55\xf1\x72\x60\x7a\x94\x87\xe9\x5d\xc8\x62\x34\x00\xd2\x85\x32\x13\xcf\xac\x26\x67\x02\xe5\x08\x19\x97\xbc\x0b\xb9\x3b\xd7\x6f\x0a\x6b\x36\x3b\x9f\x01\xd2\x03\x40\x5e\x72\xc1\xd5\xc2\x11\xc0\x95\x8c\xbb\xe6\x21\xe2\x08\xf7\x08\xe0\xcd\xcd\x45\x7a\xa9\x7f\x90\xcf\xa8\x3e\xe3\x2f\x9a\x98\x68\x72\x3b\x69\x8e\xe5\x6d\xb3\xcd\x53\x86\x3c\xf0\xb0\x89\xf2\x78\x0b\x14\x67\x3a\x4d\xfe\x33\x18\x82\xc3\xc5\xa1\x79\x14\x66\x6f\xc9\xbf\xd3\xa3\x55\xcd\xa9\xc8\x59\xd2\x5d\xd2\x4f\xe1\x00\xaa\x9b\x02\x08\x99\x07\xd2\xa3\x2f\x89\x9e\x79\x30\x29\xc2\x45\xad\xad\xe5\xaf\xf4\x25\x4d\x0d\x73\xc9\x12\x4e\x1a\x47\xc4\xca\x5f\xf4\xae\x3b\x1a\x17\x0c\x13\x07\xdb\x71\xfe\x35\x64\xde\x21\xf8\x21\x9f\xfb\x55\xda\xf2\xdd\xf8\x5c\xf8\xa2\x45\xd0\x0b\xd1\x96\x8a\x57\x35\xa9\xce\xdf\xcf\xaa\xe5\xf5\x51\x0e\x7e\xfc\x72\x68\xba\xc8\x1c\x22\x2b\xb7\x21\x89\xf3\x1b\x6b\x58\xb6\x40\xfa\x7a\x08\x96\x61\x4d\x37\x66\xea\x5c\x5b\xb4\xd0\x98\x65\x88\x23\x91\xdf\x38\x5a\xcc\x73\xeb\xfa\x84\xc0\x47\xc2\xa5\xea\x1e\xcd\x90\x99\x1c\x01\x20\x0c\x28\x49\x85\x06\x1e\xe5\x7c\x21\xdd\x64\x24\xd5\xb7\xdc\x2b\x5d\x52\x55\x45\xe6\x3d\x32\x7a\xd4\x1c\x22\x91\xcf\xe7\xc7\x98\x58\x74\xac\x79\x34\x29\x64\x34\x99\xb7\xa9\x49\x65\x4c\x57\x41\x7a\x6c\xe6\x36\xc0\xaf\x41\x6e\xcc\xe5\x01\x9a\x03\x87\xf2\x55\xbe\xc9\x60\x5a\x5d\xee\xca\x52\x70\x13\xe4\x74\x18\xe0\xdc\x46\x2a\xd4\xcc\x4c\x94\xf8\x88\x73\xe8\xa0\x45\x90\xaa\x6b\xb1\xe8\xab\x52\x1c\x9f\x3b\xa0\x0a\x94\x45\x03\xc8\x38\x4a\xa8\x34\x0b\x0a\xb8\xe0\xb2\x32\x00\x14\x65\x55\xe6\x01\xcf\x5b\xf6\xf8\x24\xd8\x8e\x16\x3c\x78\x75\xa1\x96\xd4\x75\xbf\xac\x56\x41\x48\x2c\xa5\x77\x6b\x99\x87\x74\x9f\xa4\xb1\xb5\xa1\xc9\xb3\x3b\xbf\x96\x42\xd3\x7f\x66\xdc\x36\x36\x36\x8e\xd6\x82\x66\x79\x47\x64\x7d\xcf\x96\xca\x03\xd5\x9e\xf3\x75\xef\xb2\x56\x4d\xc0\x6f\x71\x4e\x1a\x87\x0b\x8c\xd5\xc0\xbb\x7c\xd3\x0c\xfb\x0d\xc6\x49\x23\x31\x65\xac\x6c\x76\x4e\xc4\xc2\xfa\x2d\x50\xdc\x7b\x5f\x35\x88\x31\xfa\xee\x36\x84\x8a\x38\xff\xea\xc1\x98\x86\xe2\x10\x7c\x14\x34\x68\x20\x22\x10\xfb\xb8\xa5\x2e\x39\x87\x60\xc6\x65\x4b\x75\xb4\x0f\xc1\x47\xf5\x25\xe7\xb1\x8f\xd4\xaa\xdd\x42\xa1\xb0\x05\x02\x46\x1d\x59\x8d\xd7\x21\x3b\x04\x82\x85\x68\xfa\x2e\xb2\xf4\x2f\x2f\x7f\x31\xb6\x94\xcf\x0c\x5d\xfa\xfd\xbf\xc4\x37\x5d\x0d\x2e\x79\x48\xa1\x95\xf8\x07\x1c\x89\x7e\xc2\x3c\x3f\xcb\x11\x5b\x60\xa7\x50\x28\x6c\x1c\xc9\xb4\x03\x96\x8e\xc1\x26\x17\x30\xbb\x8f\x8e\x91\xc1\x55\x9a\x00\xe9\x32\x95\xd9\x93\x0c\x5e\xbb\x39\x5f\xcc\xe2\x33\xce\x79\x85\x73\xf1\x4f\xfc\xeb\xcd\xfa\xb7\x7e\xab\x30\x1e\x8f\x35\x87\x52\xc7\x4b\x7e\xa5\x30\x4b\x96\x32\x8f\xa8\x5f\x23\x40\x1e\x13\x13\x58\xc8\x46\x6c\xfe\x0b\x82\x2c\x83\x56\x74\x83\x5a\xb1\xfa\x33\xb4\xfa\xa5\xe1\xff\x04\x00\x00\xff\xff\x95\xc6\x21\xf3\x7a\x28\x00\x00")

func faucetHtmlBytes() ([]byte, error) {
	return bindataRead(
		_faucetHtml,
		"faucet.html",
	)
}

func faucetHtml() (*asset, error) {
	bytes, err := faucetHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "faucet.html", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"faucet.html": faucetHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"faucet.html": {faucetHtml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

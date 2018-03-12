package keystore

import (
	"crypto/rand"
	"reflect"
	"testing"
)

func TestZeroing(t *testing.T) {
	type zeroingItem struct {
		input []byte
	}
	type zeroingTable []zeroingItem

	var table zeroingTable
	for i := 0; i < 20; i++ {
		buf := make([]byte, 4096)
		rand.Read(buf)
		table = append(table, zeroingItem{input: buf})
	}

	for _, tt := range table {
		zeroing(tt.input)
		for i := range tt.input {
			if tt.input[i] != 0 {
				t.Errorf("Invalid zeroing '%v'", tt.input)
			}
		}
	}
}

func TestPasswordBytes(t *testing.T) {
	type passwordBytesItem struct {
		input  []byte
		output []byte
	}
	var table []passwordBytesItem
	for i := 0; i < 20; i++ {
		ibuf := make([]byte, 1024)
		rand.Read(ibuf)
		obuf := make([]byte, len(ibuf)*2)
		for j, k := 0, 0; j < len(obuf); j, k = j+2, k+1 {
			obuf[j] = 0
			obuf[j+1] = ibuf[k]
		}
		table = append(table, passwordBytesItem{input: ibuf, output: obuf})
	}
	for _, tt := range table {
		output := passwordBytes(tt.input)
		if !reflect.DeepEqual(output, tt.output) {
			t.Errorf("Invalid output '%v', '%v'", output, tt.output)
		}
	}
}

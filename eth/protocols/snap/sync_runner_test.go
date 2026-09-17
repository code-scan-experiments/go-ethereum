// Copyright 2026 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package snap

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

func TestAccountFullRLPMatchesSlimRoundTrip(t *testing.T) {
	codeHash := bytes.Repeat([]byte{0x11}, 32)
	accounts := []struct {
		name    string
		account types.StateAccount
	}{
		{
			name: "empty root and code hash",
			account: types.StateAccount{
				Balance:  uint256.NewInt(1),
				Root:     types.EmptyRootHash,
				CodeHash: types.EmptyCodeHash[:],
			},
		},
		{
			name: "non-empty root and code hash",
			account: types.StateAccount{
				Nonce:    1,
				Balance:  uint256.NewInt(2),
				Root:     common.HexToHash("0x1234"),
				CodeHash: codeHash,
			},
		},
		{
			name: "zero balance",
			account: types.StateAccount{
				Nonce:    3,
				Balance:  new(uint256.Int),
				Root:     types.EmptyRootHash,
				CodeHash: types.EmptyCodeHash[:],
			},
		},
		{
			name: "large balance",
			account: types.StateAccount{
				Nonce:    4,
				Balance:  new(uint256.Int).SetBytes(bytes.Repeat([]byte{0xff}, 32)),
				Root:     types.EmptyRootHash,
				CodeHash: types.EmptyCodeHash[:],
			},
		},
		{
			name: "large nonce",
			account: types.StateAccount{
				Nonce:    ^uint64(0),
				Balance:  uint256.NewInt(5),
				Root:     types.EmptyRootHash,
				CodeHash: types.EmptyCodeHash[:],
			},
		},
	}

	for _, test := range accounts {
		t.Run(test.name, func(t *testing.T) {
			full, err := rlp.EncodeToBytes(&test.account)
			if err != nil {
				t.Fatalf("failed to encode account: %v", err)
			}
			account := new(types.StateAccount)
			if err := rlp.DecodeBytes(full, account); err != nil {
				t.Fatalf("failed to decode account: %v", err)
			}
			got, err := rlp.EncodeToBytes(account)
			if err != nil {
				t.Fatalf("failed to encode decoded account: %v", err)
			}
			want, err := types.FullAccountRLP(types.SlimAccountRLP(*account))
			if err != nil {
				t.Fatalf("failed to expand slim account: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("full account RLP mismatch: got %x, want %x", got, want)
			}
		})
	}
}

func BenchmarkAccountFullRLPRoundTrip(b *testing.B) {
	account := nonEmptyAccount()
	b.ReportAllocs()
	for b.Loop() {
		slim := types.SlimAccountRLP(*account)
		if _, err := types.FullAccountRLP(slim); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAccountFullRLPDirect(b *testing.B) {
	account := nonEmptyAccount()
	b.ReportAllocs()
	for b.Loop() {
		slim := types.SlimAccountRLP(*account)
		if _, err := rlp.EncodeToBytes(account); err != nil {
			b.Fatal(err)
		}
		_ = slim
	}
}

func nonEmptyAccount() *types.StateAccount {
	return &types.StateAccount{
		Nonce:    42,
		Balance:  new(uint256.Int).SetBytes(bytes.Repeat([]byte{0xab}, 32)),
		Root:     common.HexToHash("0x1234"),
		CodeHash: bytes.Repeat([]byte{0xcd}, 32),
	}
}

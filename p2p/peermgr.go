// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package p2p

import (
	"sync"

	net "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

var lock sync.RWMutex
var peerStatusMessages = make(map[peer.ID]*StatusMessage)

// TODO: create goroutine that checks peer status every 5 minutes

// Check if peer status messages are compatible
func ExchangeStatusMessages(s net.Stream) bool {
	ps1 := GetPeerStatus(s.Conn().LocalPeer())
	ps2 := GetPeerStatus(s.Conn().RemotePeer())
	switch {
	case ps1.ProtocolVersion != ps2.ProtocolVersion:
		return false
	case ps1.MinSupportedVersion != ps2.MinSupportedVersion:
		return false
	case ps1.GenesisHash != ps2.GenesisHash:
		return false
	default:
		return true
	}
}

// Get peer status message
func GetPeerStatus(id peer.ID) *StatusMessage {
	lock.RLock()
	defer lock.RUnlock()
	if status, ok := peerStatusMessages[id]; ok {
		return status
	}
	return nil
}

// Set peer status message
func SetPeerStatus(id peer.ID, status *StatusMessage) {
	lock.Lock()
	defer lock.Unlock()
	if status, ok := peerStatusMessages[id]; ok {
		peerStatusMessages[id] = status
		return
	}
	peerStatusMessages[id] = status
}

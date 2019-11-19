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
	"context"

	"github.com/ChainSafe/gossamer/common"
	log "github.com/ChainSafe/log15"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-core/connmgr"
	net "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

// ConnManager implement connmgr.ConnManager
// https://godoc.org/github.com/libp2p/go-libp2p-core/connmgr#ConnManager
type ConnManager struct{}

// Notifee is used to monitor changes to a connection
func (cm ConnManager) Notifee() net.Notifiee {
	nb := new(net.NotifyBundle)
	nb.ListenF = Listen
	nb.ListenCloseF = ListenClose
	nb.ConnectedF = Connected
	nb.DisconnectedF = Disconnected
	nb.OpenedStreamF = OpenedStream
	nb.ClosedStreamF = ClosedStream
	return nb
}

func (_ ConnManager) TagPeer(peer.ID, string, int)             {}
func (_ ConnManager) UntagPeer(peer.ID, string)                {}
func (_ ConnManager) UpsertTag(peer.ID, string, func(int) int) {}
func (_ ConnManager) GetTagInfo(peer.ID) *connmgr.TagInfo      { return &connmgr.TagInfo{} }
func (_ ConnManager) TrimOpenConns(ctx context.Context)        {}
func (_ ConnManager) Protect(peer.ID, string)                  {}
func (_ ConnManager) Unprotect(peer.ID, string) bool           { return false }
func (_ ConnManager) Close() error                             { return nil }

// Called when network starts listening on an address
func Listen(n net.Network, ma ma.Multiaddr) {
	log.Debug("listen", "network", n, "address", ma)
}

// Called when network stops listening on an address
func ListenClose(n net.Network, ma ma.Multiaddr) {
	log.Debug("listen close", "network", n, "address", ma)
}

// Called when a connection opened
func Connected(n net.Network, c net.Conn) {
	log.Debug("connected", "network", n, "connection", c)

	// TODO: replace dummy status message with current state
	status := &StatusMessage{
		ProtocolVersion:     0,
		MinSupportedVersion: 0,
		Roles:               0,
		BestBlockNumber:     0,
		BestBlockHash:       common.Hash{0x00},
		GenesisHash:         common.Hash{0x00},
		ChainStatus:         []byte{0},
	}

	// TODO: check if we should set the peer status message
	// do we need to check? check happens with status message exchange
	peer := n.LocalPeer()
	SetPeerStatus(peer, status)

	// TODO: start peer status goroutine that checks peer status every 5 minutes

	log.Info("connected", "status", status)
}

// Called when a connection closed
func Disconnected(n net.Network, c net.Conn) {
	log.Debug("disconnected", "network", n, "connection", c)

	// TODO: clean up peer status messages

	log.Info("disconnected")
}

// Called when a stream opened
func OpenedStream(n net.Network, s net.Stream) {
	log.Debug("opened stream", "network", n, "stream", s)

	// s.SetProtocol(DefaultProtocolId)

	// stat := s.Stat()
	// log.Debug("opened stream", "stat", stat)

	peer := s.Conn().RemotePeer()
	protocol := s.Protocol()

	// FIX: protocol is always blank on opened stream
	log.Debug("opened stream", "peer", peer, "protocol", protocol)

	// TODO: use state protocol id instead of default
	if protocol == DefaultProtocolId {
		// Drop peer if status messages are not compatible
		// if !ExchangeStatusMessages(s) {
		// 	log.Debug("peer status message not compatible", "peer", peer)
		// 	n.ClosePeer(peer)
		// }
	}
}

// Called when a stream closed
func ClosedStream(n net.Network, s net.Stream) {
	log.Debug("closed stream", "network", n, "stream", s)

	// s.SetProtocol(DefaultProtocolId)

	peer := s.Conn().RemotePeer()
	protocol := s.Protocol()

	// FIX: protocol is always blank on opened stream but not blank on closed stream
	// log.Debug("closed stream", "peer", peer, "protocol", protocol)

	// TODO: use state protocol id instead of default
	if protocol == DefaultProtocolId {
		log.Info("closed stream", "peer", peer, "protocol", protocol)
	}
}

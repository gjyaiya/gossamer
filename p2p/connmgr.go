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
	nb.ConnectedF = Connected
	nb.OpenedStreamF = OpenedStream
	nb.ClosedStreamF = ClosedStream
	nb.DisconnectedF = Disconnected
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

func Connected(n net.Network, c net.Conn) {
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
	log.Info("connected", "status", status)
}

func OpenedStream(n net.Network, s net.Stream) {
	if s.Protocol() == DefaultProtocolId {
		log.Info("opened stream", "peer", s.Conn().RemotePeer(), "protocol", s.Protocol())
	}
}

func ClosedStream(n net.Network, s net.Stream) {
	if s.Protocol() == DefaultProtocolId {
		log.Info("closed stream", "peer", s.Conn().RemotePeer(), "protocol", s.Protocol())
	}
}

func Disconnected(n net.Network, c net.Conn) {
	log.Info("disconnected")
}

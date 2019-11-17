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

package core

import (
	"bytes"
	// "fmt"

	"github.com/ChainSafe/gossamer/internal/services"
	log "github.com/ChainSafe/log15"

	"github.com/ChainSafe/gossamer/common"
	"github.com/ChainSafe/gossamer/common/optional"
	tx "github.com/ChainSafe/gossamer/common/transaction"
	"github.com/ChainSafe/gossamer/consensus/babe"
	"github.com/ChainSafe/gossamer/core/types"
	"github.com/ChainSafe/gossamer/p2p"
	"github.com/ChainSafe/gossamer/runtime"
)

var _ services.Service = &Service{}

// Service is a overhead layer that allows for communication between the runtime, BABE, and the p2p layer.
// It deals with the validation of transactions and blocks by calling their respective validation functions
// in the runtime.
type Service struct {
	rt *runtime.Runtime
	b  *babe.Session

	recChan  <-chan []byte
	sendChan chan<- []byte
}

// NewService returns a Service that connects the runtime, BABE, and the p2p messages.
func NewService(rt *runtime.Runtime, b *babe.Session, recChan <-chan []byte, sendChan chan<- []byte) *Service {
	return &Service{
		rt:       rt,
		b:        b,
		recChan:  recChan,
		sendChan: sendChan,
	}
}

// Start begins the service. This begins watching the message channel for new block or transaction messages.
func (s *Service) Start() error {
	e := make(chan error)
	go s.start(e)
	return <- e
}

func (s *Service) start(e chan error) {
	e <- nil

	for {
		msg, ok := <- s.recChan
		if !ok {
			log.Warn("core service message watcher", "error", "channel closed")
			break
		}

		msgType := msg[0]
		switch msgType {
		case p2p.TransactionMsgType:
			// process tx
			err := s.ProcessTransaction(msg[1:])
			if err != nil {
				log.Error("core service", "error", err)
				e <- err
			}
			e <- nil
		case p2p.BlockAnnounceMsgType:
			// process block announce
			err := s.ProcessBlockAnnounce(msg[1:])
			if err != nil {
				log.Error("core service", "error", err)
				e <- err
			}
			e <- nil
		case p2p.BlockResponseMsgType:
			// process block response
			err := s.ProcessBlockResponse(msg[1:])
			if err != nil {
				log.Error("core service", "error", err)
				e <- err
			}
			e <- nil
		default:
			log.Error("core service", "error", "got unsupported message type")
		}
	}
}

func (s *Service) Stop() error {
	if s.rt != nil {
		s.rt.Stop()
	}
	if s.sendChan != nil {
		close(s.sendChan)
	}
	return nil
}

func (s *Service) StorageRoot() (common.Hash, error) {
	return s.rt.StorageRoot()
}

// ProcessTransaction attempts to validates the transaction
// if it is validated, it is added to the transaction pool of the BABE session
func (s *Service) ProcessTransaction(e types.Extrinsic) error {
	validity, err := s.validateTransaction(e)
	if err != nil {
		log.Error("ProcessTransaction", "error", err)
		return err
	}

	vtx := tx.NewValidTransaction(&e, validity)
	s.b.PushToTxQueue(vtx)

	return nil
}

// ProcessBlockAnnounce attempts to get block body required for `core_execute_block` and creates BlockResponse
func (s *Service) ProcessBlockAnnounce(b []byte) error {

	// TODO:

	// 1. Decode block message
	// 2. Request block message body
	// 3. Create block response

	// Should this be handled in the p2p package?

	// fmt.Println("\n*** BlockAnnounceMessage ***\n", b, " \n ")

	mba := new(p2p.BlockAnnounceMessage)
	buf := new(bytes.Buffer)

	buf.Write(b)
	mba.Decode(buf)

	// fmt.Println("\n*** Decoded BlockAnnounceMessage ***\n", mba, " \n ")

	bh := common.NewHash(b)

	mbr := &p2p.BlockRequestMessage{
		ID:            1,	// TODO: use increment or random number
		RequestedData: 2,
		StartingBlock: b,
		EndBlockHash:  optional.NewHash(true, bh),
		Direction:     1,
		Max:           optional.NewUint32(true, 1),
	}

	// fmt.Println("\n*** BlockRequestMessage ***\n", mbr, " \n ")

	msg, err := mbr.Encode()
	if err != nil {
		log.Error("ProcessBlockAnnounce", "error", err)
		return err
	}

	// fmt.Println("\n*** Encoded BlockRequestMessage ***\n", msg, " \n ")

	err = s.validateBlock(msg)
	return err
}

// ProcessBlockResponse attempts to add a block to the chain by calling `core_execute_block`
// if the block is validated, it is stored in the block DB and becomes part of the canonical chain
func (s *Service) ProcessBlockResponse(b []byte) error {

	// TODO: check that the BlockResponse matches our request

	err := s.validateBlock(b)
	return err
}

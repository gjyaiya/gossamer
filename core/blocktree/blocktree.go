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

package blocktree

import (
	"fmt"
	"math/big"

	"github.com/ChainSafe/gossamer/core/types"

	"github.com/ChainSafe/gossamer/polkadb"

	"github.com/ChainSafe/gossamer/common"
	log "github.com/ChainSafe/log15"
	"github.com/disiqueira/gotree"
)

type Hash = common.Hash

// BlockTree represents the current state with all possible blocks
type BlockTree struct {
	head            *node
	leaves          leafMap
	finalizedBlocks []*node
	Db              *polkadb.BlockDB
}

// NewBlockTreeFromGenesis initializes a blocktree with a genesis block.
// Currently passes in arrival time as a parameter instead of setting it as time of instanciation
func NewBlockTreeFromGenesis(genesis types.Block, db *polkadb.BlockDB) *BlockTree {
	head := &node{
		hash:        genesis.Header.Hash,
		number:      genesis.Header.Number,
		parent:      nil,
		children:    []*node{},
		depth:       big.NewInt(0),
		arrivalTime: genesis.GetBlockArrivalTime(),
	}
	return &BlockTree{
		head:            head,
		finalizedBlocks: []*node{},
		leaves:          leafMap{head.hash: head},
		Db:              db,
	}
}

// AddBlock inserts the block as child of its parent node
// Note: Assumes block has no children
func (bt *BlockTree) AddBlock(block types.Block) {
	parent := bt.GetNode(block.Header.ParentHash)
	// Check if it already exists
	// TODO: Can shortcut this by checking DB
	// TODO: Write blockData to db
	// TODO: Create getter functions to check if blockNum is greater than best block stored

	n := bt.GetNode(block.Header.Hash)
	if n != nil {
		log.Debug("Attempted to add block to tree that already exists", "Hash", n.hash)
		return
	}

	depth := big.NewInt(0)
	depth.Add(parent.depth, big.NewInt(1))

	n = &node{
		hash:        block.Header.Hash,
		number:      block.Header.Number,
		parent:      parent,
		children:    []*node{},
		depth:       depth,
		arrivalTime: block.GetBlockArrivalTime(),
	}
	parent.addChild(n)

	bt.leaves.Replace(parent, n)
}

// GetNode finds and returns a node based on its Hash. Returns nil if not found.
func (bt *BlockTree) GetNode(h Hash) *node {
	if bt.head.hash == h {
		return bt.head
	}

	for _, child := range bt.head.children {
		if n := child.getNode(h); n != nil {
			return n
		}
	}

	return nil
}

// GetBlockFromBlockNumber finds and returns a block from its number
// TODO: Grab block details from Db, this currently constructs and returns a block from node info
func (bt *BlockTree) GetBlockFromBlockNumber(b *big.Int) *types.Block {
	return bt.getNodeFromBlockNumber(b).getBlockFromNode()

}

// GetBNodeFromBlockNumber finds and returns a node from its number
func (bt *BlockTree) getNodeFromBlockNumber(b *big.Int) *node {
	if b.Cmp(bt.head.number) == 0 {
		return bt.head
	}

	for _, child := range bt.head.children {
		if n := child.getNodeFromBlockNumber(b); n != nil {
			return n
		}
	}

	return nil

}

// String utilizes github.com/disiqueira/gotree to create a printable tree
func (bt *BlockTree) String() string {
	// Construct tree
	tree := gotree.New(bt.head.String())
	for _, child := range bt.head.children {
		sub := tree.Add(child.String())
		child.createTree(sub)
	}

	// Format leaves
	var leaves string
	for k := range bt.leaves {
		leaves = leaves + fmt.Sprintf("0x%X ", k)
	}

	metadata := fmt.Sprintf("Leaves: %v", leaves)

	return fmt.Sprintf("%s\n%s\n", metadata, tree.Print())
}

// LongestPath returns the path from the root to leftmost deepest leaf in BlockTree BT
func (bt *BlockTree) LongestPath() []*node {
	dl := bt.DeepestLeaf()
	var path []*node
	for curr := dl; ; curr = curr.parent {
		path = append([]*node{curr}, path...)
		if curr.parent == nil {
			return path
		}
	}
}

// SubChain returns the path from the node with Hash start to the node with Hash end
func (bt *BlockTree) SubChain(start Hash, end Hash) []*node {
	sn := bt.GetNode(start)
	en := bt.GetNode(end)
	return sn.subChain(en)
}

// SubChain returns the path from the node with Hash start to the node with Hash end
func (bt *BlockTree) SubBlockchain(start *big.Int, end *big.Int) []*types.Block {
	s := bt.getNodeFromBlockNumber(start)
	e := bt.getNodeFromBlockNumber(end)
	sc := bt.SubChain(s.hash, e.hash)
	var bc []*types.Block
	for _, node := range sc {
		bc = append(bc, node.getBlockFromNode())
	}
	return bc

}

// DeepestLeaf returns leftmost deepest leaf in BlockTree BT
func (bt *BlockTree) DeepestLeaf() *node {
	return bt.leaves.DeepestLeaf()
}

// DeepestLeaf returns leftmost deepest block in BlockTree BT
func (bt *BlockTree) DeepestBlock() *types.Block {
	b := bt.leaves.DeepestLeaf().getBlockFromNode()
	return b
}

// computes the slot for a block from genesis
// helper for now, there's a better way to do this
func (bt *BlockTree) ComputeSlotForBlock(b *types.Block, sd uint64) uint64 {
	gt := bt.head.arrivalTime
	nt := b.GetBlockArrivalTime()

	sp := uint64(0)
	for gt < nt {
		gt += sd
		sp += 1
	}

	return sp
}

package main

import (
	"fmt"

	"github.com/ChainSafe/gossamer/cmd/utils"
	"github.com/ChainSafe/gossamer/common"
	cfg "github.com/ChainSafe/gossamer/config"
	"github.com/ChainSafe/gossamer/config/genesis"
	"github.com/ChainSafe/gossamer/polkadb"
	"github.com/ChainSafe/gossamer/trie"
	log "github.com/ChainSafe/log15"
	"github.com/urfave/cli"
)

func loadGenesis(ctx *cli.Context) error {
	fig, err := getConfig(ctx)
	if err != nil {
		return err
	}

	// read genesis file
	fp := getGenesisPath(ctx)
	log.Debug("Loading genesis", "genesisfile", fp, "datadir", fig.Global.DataDir)

	gen, err := genesis.LoadGenesisData(fp)
	if err != nil {
		return err
	}

	log.Info("🕸\t Initializing node", "name", gen.Name, "id", gen.Id, "protocolID", gen.ProtocolId, "bootnodes", common.BytesToStringArray(gen.Bootnodes))

	// DB: Create database dir and initialize stateDB and blockDB
	dbSrv, err := polkadb.NewDbService(fig.Global.DataDir)
	if err != nil {
		return err
	}

	err = dbSrv.Start()
	if err != nil {
		return err
	}

	defer func() {
		err = dbSrv.Stop()
		if err != nil {
			log.Error("error stopping database service")
		}
	}()

	tdb := &trie.Database{
		Db: dbSrv.StateDB.Db,
	}

	// create and load storage trie with initial genesis state
	t := trie.NewEmptyTrie(tdb)

	err = t.Load(gen.GenesisFields().Raw)
	if err != nil {
		return fmt.Errorf("cannot load trie with initial state: %s", err)
	}

	// write initial genesis data to DB
	err = t.StoreInDB()
	if err != nil {
		return fmt.Errorf("cannot store genesis data in db: %s", err)
	}

	err = t.StoreHash()
	if err != nil {
		return fmt.Errorf("cannot store genesis hash in db: %s", err)
	}

	// store node name, ID, p2p protocol, bootnodes in DB
	return t.Db().StoreGenesisData(gen)
}

// getGenesisPath gets the path to the genesis file
func getGenesisPath(ctx *cli.Context) string {
	if file := ctx.GlobalString(utils.GenesisFlag.Name); file != "" {
		return file
	} else {
		return cfg.DefaultGenesisPath
	}
}

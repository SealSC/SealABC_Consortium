/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the Licegogo nse at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"syscall"
	"time"

	"github.com/SealSC/SealABC/account"
	"github.com/SealSC/SealABC/cli"
	"github.com/SealSC/SealABC/common/utility"
	"github.com/SealSC/SealABC/config"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/crypto/ciphers"
	"github.com/SealSC/SealABC/crypto/hashes"
	"github.com/SealSC/SealABC/crypto/signers"
	"github.com/SealSC/SealABC/crypto/signers/ecdsa/secp256k1"
	"github.com/SealSC/SealABC/engine"
	"github.com/SealSC/SealABC/engine/engineStartup"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/applicationCommonConfig"
	"github.com/SealSC/SealABC/network/topology/p2p/fullyConnect"
	"github.com/SealSC/SealABC/service/application/basicAssets"
	"github.com/SealSC/SealABC/service/application/memo"
	"github.com/SealSC/SealABC/service/application/smartAssets"
	"github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsLedger"
	"github.com/SealSC/SealABC/service/system"
	"github.com/SealSC/SealABC/service/system/blockchain"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"github.com/SealSC/SealABC/storage/db"
	"github.com/SealSC/SealABC/storage/db/dbDrivers/levelDB"
	"github.com/SealSC/SealABC/storage/db/dbDrivers/simpleMysql"
	"github.com/SealSC/SealABC/storage/db/dbInterface"
	"github.com/SealSC/SealEVM"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func runtimeInit() {
	//runtime.GOMAXPROCS(1)
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	//runtime.SetCPUProfileRate(1)
}

func main() {
	cli.Run()

	runtimeInit()

	err := config.StaticConfigs.Load()
	if err != nil {
		fmt.Println("load config file error, ", err)
		return
	}

	utility.Load()

	crypto.Load()

	log.SetUpLogger(log.Config{
		Level:   logrus.Level(config.StaticConfigs.LogConf.LogLevel),
		LogFile: config.StaticConfigs.LogConf.LogFile,
	})

	//pprof
	go func() {
		log.Log.Info(http.ListenAndServe(config.StaticConfigs.DebugConf.PProfPort, nil))
	}()

	//TODO deprecated
	//load sql db
	sqlStorage, _ := db.NewSimpleSQLDatabaseDriver(dbInterface.MySQL, simpleMysql.Config{
		User:          config.StaticConfigs.MySQLConf.MySQLUser,
		Password:      config.StaticConfigs.MySQLConf.MySQLPwd,
		DBName:        config.StaticConfigs.MySQLConf.MySQLDBName,
		Charset:       simpleMysql.Charsets.UTF8.String(),
		MaxConnection: 0,
	})

	//common crypto tools
	//TODO check hashType, signerType, cipherType
	cryptoTools := crypto.Tools{
		HashCalculator:  hashes.HashCalculatorByAlgorithmType(config.StaticConfigs.CryptoConf.HashType),
		Cipher:          ciphers.CipherByAlgorithmType(config.StaticConfigs.CryptoConf.CipherType),
		SignerGenerator: signers.SignerGeneratorByAlgorithmType(config.StaticConfigs.CryptoConf.SignerType),
	}

	//get password
	if len(cli.Parameters.Password) == 0 {
		fmt.Println("Please input password: ")
		bytepw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("input password error: ", err)
		}
		//TODO check password
		cli.Parameters.Password = string(bytepw)
	}

	//load self keys
	acc, err := account.FromStore(config.StaticConfigs.WalletConf.WalletPath, cli.Parameters.Password)
	if err != nil {
		fmt.Println("restore account error:", err)
		return
	}
	//TODO check validation of signertype; check signertype with cryptoTools
	selfSigner := acc.Signer

	//build consensus config
	bhtConfig := hotStuff.Config{
		MemberOnlineCheckInterval: time.Millisecond * time.Duration(config.StaticConfigs.ConsensusConf.MemberOnlineCheckInterval),
		ConsensusTimeout:          time.Millisecond * time.Duration(config.StaticConfigs.ConsensusConf.ConsensusTimeOut),
		SingerGenerator:           cryptoTools.SignerGenerator,
		HashCalc:                  cryptoTools.HashCalculator,
		SelfSigner:                selfSigner,
		ConsensusInterval:         time.Millisecond * time.Duration(config.StaticConfigs.ConsensusConf.ConsensusInterval),
	}

	//build consensus member list
	for _, member := range config.StaticConfigs.ConsensusConf.Members {
		newMember := hotStuff.Member{}
		mk, _ := hex.DecodeString(member)
		newMember.Signer, _ = secp256k1.SignerGenerator.FromRawPublicKey(mk)
		bhtConfig.Members = append(bhtConfig.Members, newMember)
	}

	engineCfg := engineStartup.Config{}
	engineCfg.ConsensusNetwork.ID = selfSigner.PublicKeyString()
	engineCfg.ConsensusNetwork.ServiceAddress = config.StaticConfigs.ConsensusConf.ConsensusServiceAddress
	engineCfg.ConsensusNetwork.ServiceProtocol = config.StaticConfigs.ConsensusConf.ConsensusServiceProtocol
	engineCfg.ConsensusNetwork.P2PSeeds = config.StaticConfigs.ConsensusConf.ConsensusMember
	//TODO from config file
	engineCfg.ConsensusNetwork.Topology = fullyConnect.NewTopology()

	//config consensus
	engineCfg.ConsensusDisabled = config.StaticConfigs.ConsensusConf.ConsensusDisabled
	engineCfg.Consensus = bhtConfig

	//load basic assets application
	utxoCfg := basicAssets.Config{
		Config: applicationCommonConfig.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: config.StaticConfigs.UTXOAppConf.UTXOLedgerDB,
			},
			EnableSQLDB: config.StaticConfigs.MySQLConf.EnableSQLStorage,
			SQLStorage:  sqlStorage,
		},
	}
	basicAssets.Load()
	utxoService, _ := basicAssets.NewBasicAssetsApplication(utxoCfg)

	//load memo application
	memoCfg := memo.Config{
		Config: applicationCommonConfig.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: config.StaticConfigs.MemoAppConf.MemoDB,
			},
		},
	}

	memo.Load()
	memoCfg.SQLStorage = sqlStorage
	memoCfg.EnableSQLDB = config.StaticConfigs.MySQLConf.EnableSQLStorage
	memoService, _ := memo.NewMemoApplication(memoCfg, cryptoTools)

	SealEVM.Load()

	saCfg := &smartAssets.Config{
		Config: applicationCommonConfig.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: config.StaticConfigs.SmartAssetsAppConf.SmartAssetsDB,
			},

			EnableSQLDB: config.StaticConfigs.MySQLConf.EnableSQLStorage,
			SQLStorage:  sqlStorage,
			CryptoTools: cryptoTools,
		},

		BaseAssets: smartAssetsLedger.BaseAssetsData{
			Name:        config.StaticConfigs.SmartAssetsAppConf.SmartAssetsName,
			Symbol:      config.StaticConfigs.SmartAssetsAppConf.SmartAssetsSymbol,
			Supply:      config.StaticConfigs.SmartAssetsAppConf.SmartAssetsSupply,
			Precision:   config.StaticConfigs.SmartAssetsAppConf.SmartAssetsPrecision,
			Increasable: config.StaticConfigs.SmartAssetsAppConf.SmartAssetsIncreasable,
			Owner:       config.StaticConfigs.SmartAssetsAppConf.SmartAssetsOwner,
		},
		TxPoolLimit:   config.StaticConfigs.SmartAssetsAppConf.TxPoolLimit,
		ClientTxLimit: config.StaticConfigs.SmartAssetsAppConf.ClientTxLimit,
	}

	smartAssets.Load()
	smartAssetsApplication, err := smartAssets.NewSmartAssetsApplication(saCfg)
	if err != nil {
		fmt.Println("start smart assets application failed: ", err.Error())
		return
	}

	//set application to chain service
	systemService := system.Config{}

	//config system service
	systemService.Chain = blockchain.Config{}

	systemService.Chain.Blockchain.Signer = selfSigner
	systemService.Chain.Blockchain.CryptoTools = cryptoTools
	systemService.Chain.Blockchain.NewWhenGenesis = true

	systemService.Chain.ExternalExecutors = []chainStructure.IBlockchainExternalApplication{
		utxoService,
		memoService,
		smartAssetsApplication,
	}

	//start load chain
	systemService.Chain.Api.HttpJSON = config.StaticConfigs.BlockChainConf.BlockchainApiConfig

	//config blockchain system service network
	systemService.Chain.Network.ID = selfSigner.PublicKeyString()
	systemService.Chain.Network.ServiceAddress = config.StaticConfigs.BlockChainConf.BlockchainServiceAddress
	systemService.Chain.Network.ServiceProtocol = config.StaticConfigs.BlockChainConf.BlockchainServiceProtocol
	systemService.Chain.Network.P2PSeeds = config.StaticConfigs.BlockChainConf.BlockchainServiceSeeds
	systemService.Chain.Network.Topology = fullyConnect.NewTopology()

	engineCfg.Log.LogFile = config.StaticConfigs.LogConf.LogFile
	engineCfg.Log.Level = logrus.Level(config.StaticConfigs.LogConf.LogLevel)

	engineCfg.Api.HttpJSON = config.StaticConfigs.EngineConf.EngineApiConfig

	systemService.Chain.EnableSQLDB = config.StaticConfigs.MySQLConf.EnableSQLStorage
	systemService.Chain.SQLStorage = sqlStorage
	systemService.Chain.Blockchain.StorageDriver, _ = db.NewKVDatabaseDriver(dbInterface.LevelDB, levelDB.Config{
		DBFilePath: config.StaticConfigs.BlockChainConf.ChainDB,
	})
	engine.Startup(engineCfg)

	system.NewBlockchainService(systemService.Chain)

	for {
		time.Sleep(time.Second)
	}

}

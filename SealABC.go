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
	"syscall"

	"github.com/SealSC/SealABC/account"
	"github.com/SealSC/SealABC/cli"
	"github.com/SealSC/SealABC/common/utility"
	"github.com/SealSC/SealABC/common/utility/serializer/structSerializer"
	"github.com/SealSC/SealABC/config"
	"github.com/SealSC/SealABC/consensus/hotStuff"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/crypto/hashes/sha3"
	"github.com/SealSC/SealABC/crypto/signers/ecdsa/secp256k1"
	"github.com/SealSC/SealABC/engine"
	"github.com/SealSC/SealABC/engine/engineStartup"
	"github.com/SealSC/SealABC/log"
	"github.com/SealSC/SealABC/metadata/block"
	"github.com/SealSC/SealABC/network/topology/p2p/fullyConnect"
	"golang.org/x/term"

	//"github.com/SealSC/SealABC/service/application/basicAssets"
	"github.com/SealSC/SealABC/service/application/memo"
	"github.com/SealSC/SealABC/service/application/smartAssets"
	"github.com/SealSC/SealABC/service/application/smartAssets/smartAssetsLedger"
	"github.com/SealSC/SealABC/service/application/traceableStorage"

	//"github.com/SealSC/SealABC/service/application/universalIdentification"
	"net/http"
	"runtime"
	"time"

	"github.com/SealSC/SealABC/service/system"
	"github.com/SealSC/SealABC/service/system/blockchain"
	"github.com/SealSC/SealABC/service/system/blockchain/chainStructure"
	"github.com/SealSC/SealABC/storage/db"
	"github.com/SealSC/SealABC/storage/db/dbDrivers/levelDB"
	"github.com/SealSC/SealABC/storage/db/dbDrivers/simpleMysql"
	"github.com/SealSC/SealABC/storage/db/dbInterface"
	"github.com/sirupsen/logrus"

	_ "net/http/pprof"

	appCommonCfg "github.com/SealSC/SealABC/metadata/applicationCommonConfig"
	"github.com/SealSC/SealEVM"
)

//func getPassword() string {
//    fmt.Print("enter password: ")
//    pwdBytes, err := terminal.ReadPassword(syscall.Stdin)
//    if err != nil {
//        fmt.Println("get password failed: ", err.Error())
//    }
//    pwdStr := string(pwdBytes)
//
//    fmt.Println(environment.Block{})
//    fmt.Println(evmInt256.New(0))
//    return strings.TrimSpace(pwdStr)
//}

func testHeap(prvBytes []byte) (ret []byte) {
	ret = append(ret, prvBytes...)
	blk := block.Entity{}
	blk.Header.Version = "just for test!!!!!!"
	fakePrev := make([]byte, len(prvBytes), len(prvBytes))
	copy(fakePrev, prvBytes)
	blk.Header.PrevBlock = fakePrev

	//tools := crypto.Tools{SignerGenerator: ed25519.SignerGenerator, HashCalculator: sha3.Sha256}
	tools := crypto.Tools{SignerGenerator: secp256k1.SignerGenerator, HashCalculator: sha3.Keccak256}
	_ = blk.Sign(tools, prvBytes)

	blkBytes, _ := structSerializer.ToMFBytes(blk)
	_, _ = blk.Seal.Verify(blkBytes, tools.HashCalculator)

	return
}

func main() {
	cli.Run()

	SealEVM.Load()

	//runtime.GOMAXPROCS(1)
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	//runtime.SetCPUProfileRate(1)

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	utility.Load()
	crypto.Load()
	_ = config.Configs.Load()

	log.SetUpLogger(log.Config{
		Level:   logrus.Level(config.Configs.LogLevel),
		LogFile: config.Configs.LogFile,
	})

	//load sql db
	sqlStorage, _ := db.NewSimpleSQLDatabaseDriver(dbInterface.MySQL, simpleMysql.Config{
		User:          config.Configs.MySQLUser,
		Password:      config.Configs.MySQLPwd,
		DBName:        config.Configs.MySQLDBName,
		Charset:       simpleMysql.Charsets.UTF8.String(),
		MaxConnection: 0,
	})

	//common crypto tools
	cryptoTools := crypto.Tools{
		//HashCalculator: sha3.Sha256,
		//SignerGenerator: ed25519.SignerGenerator,
		HashCalculator:  sha3.Keccak256,
		SignerGenerator: secp256k1.SignerGenerator,
	}

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
	acc, err := account.FromStore(config.Configs.WalletPath, cli.Parameters.Password)
	if err != nil {
		fmt.Println("restore account error:", err)
		return
	}
	//TODO check validation of signertype
	selfSigner := acc.Signer

	//build consensus config
	bhtConfig := hotStuff.Config{
		MemberOnlineCheckInterval: time.Millisecond * 1000,
		ConsensusTimeout:          time.Millisecond * 10000,
		//SingerGenerator:           ed25519.SignerGenerator,
		//HashCalc:          sha3.Sha256,
		SingerGenerator:   secp256k1.SignerGenerator,
		HashCalc:          sha3.Keccak256,
		SelfSigner:        selfSigner,
		ConsensusInterval: time.Millisecond * 3000,
	}

	//build consensus member list
	for _, member := range config.Configs.Members {
		newMember := hotStuff.Member{}
		mk, _ := hex.DecodeString(member)
		//newMember.Signer, _ = ed25519.SignerGenerator.FromRawPublicKey(mk)
		newMember.Signer, _ = secp256k1.SignerGenerator.FromRawPublicKey(mk)
		bhtConfig.Members = append(bhtConfig.Members, newMember)
	}

	engineCfg := engineStartup.Config{}

	////set self signer
	//engineCfg.SelfSigner = selfSigner
	//
	////set crypto tools
	//engineCfg.CryptoTools = cryptoTools

	//config network
	engineCfg.ConsensusNetwork.ID = selfSigner.PublicKeyString()
	engineCfg.ConsensusNetwork.ServiceAddress = config.Configs.ConsensusServiceAddress
	engineCfg.ConsensusNetwork.ServiceProtocol = "tcp"
	engineCfg.ConsensusNetwork.P2PSeeds = config.Configs.ConsensusMember
	engineCfg.ConsensusNetwork.Topology = fullyConnect.NewTopology()

	//config consensus
	engineCfg.ConsensusDisabled = config.Configs.ConsensusDisabled
	engineCfg.Consensus = bhtConfig
	//engineCfg.ConsensusType = consensus.BasicHotStuff

	//config db
	//engineCfg.StorageConfig = levelDB.Config{
	//    DBFilePath: Configs.ChainDB,
	//}

	//load basic assets application
	//utxoCfg := basicAssets.Config {
	//    appCommonCfg.Config{
	//        KVDBName: dbInterface.LevelDB,
	//        KVDBConfig: levelDB.Config{
	//            DBFilePath: Configs.LedgerDB,
	//        },
	//        EnableSQLDB: Configs.EnableSQLStorage,
	//        SQLStorage:  sqlStorage,
	//    },
	//}
	//basicAssets.Load()
	//utxoService, _ := basicAssets.NewBasicAssetsApplication(utxoCfg)

	//load memo application
	memoCfg := memo.Config{
		Config: appCommonCfg.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: config.Configs.MemoDB,
			},
		},
	}

	memo.Load()
	memoCfg.SQLStorage = sqlStorage
	memoCfg.EnableSQLDB = config.Configs.EnableSQLStorage
	memoService, _ := memo.NewMemoApplication(memoCfg, cryptoTools)

	saCfg := &smartAssets.Config{
		Config: appCommonCfg.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: config.Configs.SmartAssetsDB,
			},

			EnableSQLDB: config.Configs.EnableSQLStorage,
			SQLStorage:  sqlStorage,
			CryptoTools: cryptoTools,
		},

		BaseAssets: smartAssetsLedger.BaseAssetsData{
			Name:        "TheOneArt",
			Symbol:      "TOA",
			Supply:      "1000000000",
			Precision:   18,
			Increasable: true,
			Owner:       "3d468299df9391e62b5e45531169585ffde27fef",
		},
	}

	smartAssets.Load()
	smartAssetsApplication, err := smartAssets.NewSmartAssetsApplication(saCfg)
	if err != nil {
		fmt.Println("start smart assets application failed: ", err.Error())
		return
	}

	tsCfg := traceableStorage.Config{
		Config: appCommonCfg.Config{
			KVDBName: dbInterface.LevelDB,
			KVDBConfig: levelDB.Config{
				DBFilePath: config.Configs.TraceableStorageDB,
			},
			EnableSQLDB: false,
			SQLStorage:  nil,
			CryptoTools: crypto.Tools{},
		},
	}
	traceableStorage.Load()
	traceableStorageApplication, err := traceableStorage.NewTraceableStorageApplication(&tsCfg)
	if err != nil {
		fmt.Println("start traceable storage application failed: ", err.Error())
		return
	}

	//uidCfg := universalIdentification.Config  {
	//   Config:      appCommonCfg.Config{
	//       KVDBName: dbInterface.LevelDB,
	//       KVDBConfig: levelDB.Config{
	//           DBFilePath: Configs.TraceableStorageDB,
	//       },
	//       EnableSQLDB: false,
	//       SQLStorage:  nil,
	//       CryptoTools: crypto.Tools{},
	//   },
	//}
	//universalIdentification.Load()
	//uidApp, err := universalIdentification.NewUniversalIdentificationApplication(uidCfg)
	//if err != nil {
	//   fmt.Println("start traceable storage application failed: ", err.Error())
	//   return
	//}

	//set application to chain service
	systemService := system.Config{}

	//config system service
	systemService.Chain = blockchain.Config{}

	systemService.Chain.Blockchain.Signer = selfSigner
	systemService.Chain.Blockchain.CryptoTools = cryptoTools
	systemService.Chain.Blockchain.NewWhenGenesis = true

	systemService.Chain.ExternalExecutors = []chainStructure.IBlockchainExternalApplication{
		//utxoService,
		memoService,
		smartAssetsApplication,
		traceableStorageApplication,
		//uidApp,
	}

	//start load chain
	systemService.Chain.Api.HttpJSON = config.Configs.BlockchainApiConfig

	//config blockchain system service network
	systemService.Chain.Network.ID = selfSigner.PublicKeyString()
	systemService.Chain.Network.ServiceAddress = config.Configs.BlockchainServiceAddress
	systemService.Chain.Network.ServiceProtocol = "tcp"
	systemService.Chain.Network.P2PSeeds = config.Configs.BlockchainServiceSeeds
	systemService.Chain.Network.Topology = fullyConnect.NewTopology()

	engineCfg.Log.LogFile = config.Configs.LogFile
	engineCfg.Log.Level = logrus.Level(config.Configs.LogLevel)

	engineCfg.Api.HttpJSON = config.Configs.EngineApiConfig

	systemService.Chain.EnableSQLDB = config.Configs.EnableSQLStorage
	systemService.Chain.SQLStorage = sqlStorage
	systemService.Chain.Blockchain.StorageDriver, _ = db.NewKVDatabaseDriver(dbInterface.LevelDB, levelDB.Config{
		DBFilePath: config.Configs.ChainDB,
	})
	engine.Startup(engineCfg)

	system.NewBlockchainService(systemService.Chain)

	for {
		time.Sleep(time.Second)
	}
}

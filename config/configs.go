/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
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

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/SealSC/SealABC/cli"
	"github.com/SealSC/SealABC/network/http"
)

//TODO add hotstuff/pbft in ConsensusConf
type Config struct {
	ConsensusConf struct {
		ConsensusDisabled         bool     `json:"consensus_disabled"`
		ConsensusServiceProtocol  string   `json:"consensus_service_protocol"`
		ConsensusServiceAddress   string   `json:"consensus_service_address"`
		ConsensusMember           []string `json:"consensus_member"`
		Members                   []string `json:"members"`
		MemberOnlineCheckInterval uint64   `json:"online_check_interval"`
		ConsensusTimeOut          uint64   `json:"consensus_timeout"`
		ConsensusInterval         uint64   `json:"consensus_Interval"`
		ConsensusTopology         string   `json:"consensus_topology"`
	} `json:"consensus_conf"`
	BlockChainConf struct {
		BlockchainServiceAddress  string      `json:"blockchain_service_address"`
		BlockchainServiceSeeds    []string    `json:"blockchain_service_seeds"`
		BlockchainApiConfig       http.Config `json:"blockchain_api_config"`
		ChainDB                   string      `json:"chain_db"`
		BlockchainServiceProtocol string      `json:"blockchain_service_protocol"`
	} `json:"block_chain_conf"`
	DebugConf struct {
		PProfPort string `json:"pprof_port"`
	} `json:"debug_conf"`
	LogConf struct {
		LogFile  string `json:"log_file"`
		LogLevel uint32 `json:"log_level"`
	} `json:"log_conf"`
	WalletConf struct {
		WalletPath string `json:"wallet_path"`
	} `json:"wallet_conf"`
	UTXOAppConf struct {
		UTXOAppEnable bool   `json:"utxo_app_enable"`
		UTXOLedgerDB  string `json:"utxo_ledger_db"`
	} `json:"utxo_app_conf"`
	MemoAppConf struct {
		MemoAppEnable bool   `json:"memo_app_enable"`
		MemoDB        string `json:"memo_db"`
	} `json:"memo_app_conf"`
	SmartAssetsAppConf struct {
		SmartAssetsEnable bool   `json:"smart_assets_enable"`
		SmartAssetsDB     string `json:"smart_assets_db"`
		TxPoolLimit       int    `json:"tx_pool_limit"`
		ClientTxLimit     int    `json:"client_tx_limit"`
	} `json:"smart_assets_app_conf"`
	MySQLConf struct {
		EnableSQLStorage bool   `json:"enable_sql_storage"`
		MySQLDBName      string `json:"mysql_db_name"`
		MySQLUser        string `json:"mysql_user"`
		MySQLPwd         string `json:"mysql_pwd"`
	} `json:"mysql_conf"`
	EngineConf struct {
		EngineApiConfig http.Config `json:"engine_api_config"`
	} `json:"engine_conf"`
	CryptoConf struct {
		HashType   string `json:"hash_type"`
		CipherType string `json:"cipher_type"`
		SignerType string `json:"signer_type"`
	} `json:"crypto_conf"`
}

func (c *Config) Load() (err error) {
	cfgStr, err := ioutil.ReadFile(cli.Parameters.ConfigFile)

	if nil != err {
		fmt.Println("load config file error: ", err)
		return
	}

	cfgStr = bytes.TrimPrefix(cfgStr, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(cfgStr, c)
	if nil != err {
		fmt.Println("unmarshal config file error: ", err)
		return
	}

	return
}

var StaticConfigs = Config{}

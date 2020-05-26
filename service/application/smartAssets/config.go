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

package smartAssets

import (
	"SealABC/crypto"
	"SealABC/crypto/hashes/sha3"
	"SealABC/crypto/signers/ed25519"
	"SealABC/service/application/smartAssets/smartAssetsLedger"
	"SealABC/storage/db/dbDrivers/levelDB"
	"SealABC/storage/db/dbInterface"
	"SealABC/storage/db/dbInterface/simpleSQLDatabase"
)

type Config struct {
	KVDBName   string
	KVDBConfig levelDB.Config

	EnableSQLDB bool
	SQLStorage  simpleSQLDatabase.IDriver

	CryptoTools crypto.Tools

	BaseAssets smartAssetsLedger.BaseAssetsData
}

func DefaultConfig() *Config {
	return &Config {
		KVDBName: dbInterface.LevelDB,
		KVDBConfig: levelDB.Config{
			DBFilePath: "./smartAssets",
		},

		EnableSQLDB:  false,
		SQLStorage:   nil,

		CryptoTools:  crypto.Tools{
			HashCalculator:  sha3.Sha256,
			SignerGenerator: ed25519.SignerGenerator,
		},

		BaseAssets: smartAssetsLedger.BaseAssetsData{
			Name:        "Seal Smart Token",
			Symbol:      "SST",
			Supply:      "1000000000", //one billion
			Increasable: false,
			Creator:     "",
		},
	}
}
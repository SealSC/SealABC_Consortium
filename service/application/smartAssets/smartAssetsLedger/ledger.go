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

package smartAssetsLedger

import (
	"SealABC/common/utility/serializer/structSerializer"
	"SealABC/crypto"
	"SealABC/crypto/signers/signerCommon"
	"SealABC/dataStructure/enum"
	"SealABC/metadata/block"
	"SealABC/metadata/blockchainRequest"
	"SealABC/metadata/seal"
	"SealABC/storage/db/dbInterface/kvDatabase"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

type Ledger struct {
	txPool        [] *Transaction
	txHashRecord  map[string] bool
	txPoolLimit   int
	clientTxCount map[string] int
	clientTxLimit int

	operateLock sync.RWMutex
	poolLock    sync.Mutex

	genesisAssets BaseAssets
	genesisSigner signerCommon.ISigner

	preActuators map[string] txPreActuator
	actuators    map[string] batchTxActuator

	CryptoTools   crypto.Tools
	Storage       kvDatabase.IDriver
}

func Load() {
	enum.SimpleBuild(&StoragePrefixes)
	enum.SimpleBuild(&TxType)
	enum.BuildErrorEnum(&Errors, 1000)
}

func (l *Ledger) LoadGenesisAssets(creatorKey interface{}, assets BaseAssetsData) error  {

	supply := assets.Supply
	assets.Supply = "0"

	if creatorKey == nil {
		return errors.New("no creator key")
	}

	signer, err := l.CryptoTools.SignerGenerator.FromRawPrivateKey(creatorKey)
	if err != nil {
		return err
	}

	metaBytes, _ := structSerializer.ToMFBytes(assets)
	metaSeal := seal.Entity{}
	err = metaSeal.Sign(metaBytes, l.CryptoTools, creatorKey)
	if err != nil {
		return err
	}

	gAssets, exists, _ := l.getAssetsByHash(metaSeal.Hash)

	if !exists {
		assets.Supply = supply
		issuedBytes, _ := structSerializer.ToMFBytes(assets)
		issuedSeal := seal.Entity{}
		err = issuedSeal.Sign(issuedBytes, l.CryptoTools, creatorKey)
		if err != nil {
			return err
		}

		gAssets = &BaseAssets{
			BaseAssetsData: assets,
			IssuedSeal:     issuedSeal,
			MetaSeal:       metaSeal,
		}

		err = l.storeAssets(*gAssets)
		if err != nil {
			return err
		}

		balance, valid := big.NewInt(0).SetString(assets.Supply, 10)
		if !valid {
			return errors.New("invalid assets supply")
		}

		if balance.Sign() <= 0 {
			return errors.New("supply is zero or negative")
		}

		err = l.Storage.Put(kvDatabase.KVItem{
			Key:    StoragePrefixes.Balance.buildKey([]byte(signer.PublicKeyString())),
			Data:   balance.Bytes(),
			Exists: false,
		})

		return err
	}

	return nil
}

func (l *Ledger) AddTx(req blockchainRequest.Entity) error {
	tx := Transaction{}
	err := json.Unmarshal(req.Data, &tx)
	if err != nil {
		return err
	}

	if tx.Type != req.RequestAction {
		return errors.New("transaction type is not equal to block request action")
	}
	
	valid, err := tx.verify(l.CryptoTools.HashCalculator)
	if !valid {
		return err
	}

	//todo: check balance to ensure client has enough base-assets to pay transaction basic fee

	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	client := string(tx.DataSeal.SignerPublicKey)
	clientTxCount := l.clientTxCount[client]
	if clientTxCount >= l.clientTxLimit {
		return errors.New("reach transaction count limit")
	}

	if len(l.txPool) >= l.txPoolLimit {
		return errors.New("reach transaction pool limit")
	}

	_, exists, _ := l.getTxFromStorage(tx.DataSeal.Hash)
	if exists {
		return errors.New("duplicate history transaction")
	}

	txHash := string(tx.DataSeal.Hash)
	if l.txHashRecord[txHash] {
		return errors.New("duplicate pending transaction")
	}

	l.txPool = append(l.txPool, &tx)
	l.clientTxCount[client] = clientTxCount + 1

	return nil
}

func (l Ledger) txResultCheck(orgResult TransactionResult, execResult TransactionResult, txHash []byte) error {
	if orgResult.Success != execResult.Success {
		return errors.New(fmt.Sprintf("transaction %x verify failed", txHash))
	}

	if orgResult.ErrorCode != execResult.ErrorCode {
		return errors.New(fmt.Sprintf("transaction %x has different error code", txHash))
	}


	if orgResult.ErrorCode != execResult.ErrorCode {
		return errors.New(fmt.Sprintf("transaction %x has different error code",txHash))
	}

	if len(orgResult.NewState) != len(execResult.NewState) {
		return errors.New(fmt.Sprintf("transaction %x has different count state to change", txHash))
	}

	for i, s := range orgResult.NewState {
		if !bytes.Equal(s.Key, execResult.NewState[i].Key) || !bytes.Equal(s.Val, execResult.NewState[i].Val) {
			return errors.New(fmt.Sprintf("transaction %x has different state to change", txHash))
		}
	}

	return nil
}

func (l Ledger) PreExecute(txList TransactionList, blk block.Entity) (result []byte, err error) {
	//only check signature & duplicate
	txHash := map[string] bool{}
	resultCache := txResultCache{}

	for _, tx := range txList.Transactions {
		hash := string(tx.DataSeal.Hash)
		if _, exists := txHash[hash]; !exists {
			txHash[hash] = true
		} else {
			err = errors.New("duplicate transaction")
			break
		}

		_, err = tx.verify(l.CryptoTools.HashCalculator)
		if err != nil {
			break
		}

		if preExec, exists := l.preActuators[tx.Type]; exists {
			newState, _, execErr := preExec(tx, resultCache, blk)
			txForCheck := Transaction{}
			l.setTxResult(execErr, newState, &txForCheck)

			checkErr := l.txResultCheck(tx.TransactionResult, txForCheck.TransactionResult, tx.getHash())
			if checkErr != nil {
				err = checkErr
				break
			}
		}
	}

	return
}

func (l Ledger) Execute(txList []Transaction, blockHeader block.Header) (result []byte, err error) {
	return
}

func (l Ledger) setTxResult(err error, newState []StateData, tx *Transaction) {
	if err != nil {
		errEl := err.(enum.ErrorElement)
		tx.TransactionResult.Success = false
		tx.TransactionResult.ErrorCode = errEl.Code()
	} else {
		tx.TransactionResult.Success = true
		tx.TransactionResult.NewState = newState
	}
}

func (l Ledger) GetTransactionsFromPool(blk block.Entity) (txList TransactionList, count uint32) {
	l.poolLock.Lock()
	defer l.poolLock.Unlock()

	count = uint32(len(l.txPool))
	if count == 0 {
		return
	}

	resultCache := txResultCache{}
	for _, tx := range l.txPool {

		if preExec, exists := l.preActuators[tx.Type]; exists {
			newState, _, err := preExec(*tx, resultCache, blk)
			l.setTxResult(err, newState, tx)
		}

		txList.Transactions = append(txList.Transactions, *tx)
	}

	return
}

func NewLedger(tools crypto.Tools, driver kvDatabase.IDriver, genesisAssetsCreator signerCommon.ISigner) *Ledger {
	l := &Ledger{
		txPool:        [] *Transaction{},
		txHashRecord:  map[string] bool{},
		txPoolLimit:   1000,
		clientTxCount: map[string] int{},
		clientTxLimit: 4,
		operateLock:   sync.RWMutex{},
		poolLock:      sync.Mutex{},
		genesisAssets: BaseAssets{},
		genesisSigner: genesisAssetsCreator,
		CryptoTools:   tools,
		Storage:       driver,
	}

	l.preActuators = map[string]txPreActuator{
		TxType.Transfer.String(): l.preTransfer,
	}

	l.actuators = map[string]batchTxActuator{
		TxType.Transfer.String(): l.batchTransferActuator,
	}

	return l
}
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

package account

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/SealSC/SealABC/crypto/ciphers"
	"github.com/SealSC/SealABC/crypto/ciphers/cipherCommon"
	"github.com/SealSC/SealABC/crypto/ciphers/sm4"
	"github.com/SealSC/SealABC/crypto/kdf/pbkdf2"
	"github.com/SealSC/SealABC/crypto/signers"
)

func (s SealAccount) Store(filename string, password string, cipher ciphers.ICipher) (encrypted Encrypted, err error) {
	dataForEnc := accountDataForEncrypt{}
	dataForEnc.SignerType = s.SingerType
	dataForEnc.KeyData = s.Signer.PrivateKeyBytes()

	dataBytes, _ := json.Marshal(dataForEnc)

	//TODO temp modify
	keylen := encryptedKeyLen
	if cipher.Type() == sm4.Sm4.Type() {
		keylen = encryptedKeyLenSm4
	}

	key, keySalt, kdfParam, err := pbkdf2.Generator.NewKey([]byte(password), keylen)
	if err != nil {
		return
	}

	encMode := cipherCommon.CBC
	encData, err := cipher.Encrypt(dataBytes, key, encMode)
	if err != nil {
		return
	}

	encrypted.Address = s.Address
	encrypted.PublicKey = s.Signer.PublicKeyString()
	encrypted.SignerType = s.SingerType
	encrypted.Data = encData
	encrypted.Config = StoreConfig{
		CipherType:  cipher.Type(),
		CipherParam: []byte(encMode),
		KDFType:     pbkdf2.Generator.Name(),
		KDFSalt:     keySalt,
		KDFParam:    kdfParam,
		KeyLength:   keylen,
	}

	fileData, err := json.MarshalIndent(encrypted, "", "    ")
	if err != nil {
		return
	}

	err = ioutil.WriteFile(filename, fileData, 0666)
	return
}

func FromStore(filename string, password string) (sa SealAccount, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	encAccount := Encrypted{}
	err = json.Unmarshal(data, &encAccount)
	if err != nil {
		return
	}

	config := encAccount.Config

	if pbkdf2.Generator.Name() != config.KDFType {
		err = errors.New("not supported kdf type: " + encAccount.Config.KDFType)
		return
	}

	key, err := pbkdf2.Generator.RebuildKey([]byte(password), config.KeyLength, config.KDFSalt, config.KDFParam)
	if err != nil {
		return
	}

	var cipher = ciphers.CipherByAlgorithmType(encAccount.Config.CipherType)
	saData, err := cipher.Decrypt(encAccount.Data, key, string(config.CipherParam))
	if err != nil {
		return
	}

	saForEnc := accountDataForEncrypt{}

	err = json.Unmarshal(saData, &saForEnc)
	if err != nil {
		return
	}

	sg := signers.SignerGeneratorByAlgorithmType(saForEnc.SignerType)
	signer, err := sg.FromRawPrivateKey(saForEnc.KeyData)
	if err != nil {
		return
	}

	if signer.ToAddress() != encAccount.Address {
		err = errors.New("address not equal")
		return
	}

	sa.Address = encAccount.Address
	sa.SingerType = saForEnc.SignerType
	sa.Signer = signer
	return
}

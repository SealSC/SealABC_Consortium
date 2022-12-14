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
	"github.com/SealSC/SealABC/crypto/ciphers/cipherCommon"
	"github.com/SealSC/SealABC/crypto/signers"
	"github.com/SealSC/SealABC/crypto/signers/signerCommon"
)

const encryptedKeyLen = 32
const encryptedKeyLenSm4 = 16

type StoreConfig struct {
	SignerType  string
	CipherType  string
	CipherParam []byte

	KDFType  string
	KDFSalt  []byte
	KDFParam []byte

	KeyLength int
}

type Encrypted struct {
	Address    string
	PublicKey  string
	SignerType string
	Data       cipherCommon.EncryptedData
	Config     StoreConfig
}

type accountDataForEncrypt struct {
	SignerType string
	KeyData    []byte
}

type SealAccount struct {
	Address    string
	SingerType string
	Signer     signerCommon.ISigner
}

func NewAccount(privateKey []byte, sg signers.ISignerGenerator) (sa SealAccount, err error) {
	var signer signerCommon.ISigner
	if privateKey != nil {
		signer, err = sg.FromRawPrivateKey(privateKey)
	} else {
		signer, err = sg.NewSigner(nil)
	}

	if err != nil {
		return
	}

	sa.Address = signer.ToAddress()
	sa.SingerType = signer.Type()
	sa.Signer = signer
	return
}

func NewAccountForVerify(publicKey []byte, sg signers.ISignerGenerator) (sa SealAccount, err error) {
	signer, err := sg.FromRawPublicKey(publicKey)
	if err != nil {
		return
	}

	sa.Address = signer.ToAddress()
	sa.SingerType = signer.Type()
	sa.Signer = signer

	return
}

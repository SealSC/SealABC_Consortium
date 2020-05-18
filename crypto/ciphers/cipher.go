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

package ciphers

type EncryptedData struct {
    CipherText      []byte
    ExternalData    []byte
}

const (
    CBC = "CBC"
    CFB = "CFB"
    OFB = "OFB"
    CTR = "CTR"
)

type ICipher interface {
    Name() string
    Encrypt(plainText []byte, key []byte, param interface{}) (result EncryptedData, err error)
    Decrypt(cipherText EncryptedData, key []byte, param interface{}) (plaintext []byte, err error)
}

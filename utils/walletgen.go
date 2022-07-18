package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/SealSC/SealABC/account"
	"github.com/SealSC/SealABC/crypto"
	"github.com/SealSC/SealABC/crypto/ciphers"
	"github.com/SealSC/SealABC/crypto/signers"
	"golang.org/x/term"
)

//ED25519 or secp256k1 or sm2
//TODO from config file
//const SignerType = "sm2"
//const SignerType = "ED25519"
const SignerType = "secp256k1"

// AES or sm4
//TODO from config file
//const CipherType = "sm4"
const CipherType = "AES"

func walletgen(path string, password string, privateKey []byte) error {
	//init crypto
	crypto.Load()

	//get signer type
	sg := signers.SignerGeneratorByAlgorithmType(SignerType)
	if sg == nil {
		fmt.Println("unsupported signature algorithm,", SignerType)
		return errors.New("unsupported signature algorithm")
	}

	//get cipher type
	cipher := ciphers.CipherByAlgorithmType(CipherType)
	if cipher == nil {
		fmt.Println("unsupported cipher algorithm,", CipherType)
		return errors.New("unsupported cipher algorithm")
	}

	sa, err := account.NewAccount(privateKey, sg)
	if err != nil {
		fmt.Println("create new account error:", err)
		return err
	}

	_, err = sa.Store(path, password, cipher)
	if err != nil {
		fmt.Println("store account error:", err)
		return err
	}

	return nil
}

func main() {
	//get params from command-line
	argc := len(os.Args)
	if argc > 3 || argc < 2 {
		fmt.Println("usage:\n walletgen <out_path> <option: privatekey>")
		return
	}

	//TODO check validation of path
	path := os.Args[1]

	var err error
	var privateKey []byte
	if argc > 2 {
		//TODO check validation of private key
		privateKeyStr := os.Args[2]
		privateKey, err = hex.DecodeString(privateKeyStr)
		if err != nil {
			fmt.Println("convert privateKey error: ", err)
			return
		}
	}

	fmt.Println("Please input password: ")
	bytepw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("input password error: ", err)
	}
	//TODO check password
	pass := string(bytepw)

	err = walletgen(path, pass, privateKey)
	if err != nil {
		fmt.Println("Generate wallet failed: ", err)
		return
	}

	fmt.Println("Generate wallet successfully!")
}

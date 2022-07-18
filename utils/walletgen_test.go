package main

import (
	"encoding/hex"
	"testing"

	"github.com/SealSC/SealABC/account"
)

func TestWalletGen(t *testing.T) {
	var privateKeyHex string
	if SignerType == "ED25519" {
		privateKeyHex = "e8e8570fffa55501635b97dcad55c951f86725a26d7c7c1bf2616268b152b42d0104d9c3abfed18c5c0fc0e3b89b2aad5c42ac5392512c3fc9ce0d3235ad92a0"
	} else {
		privateKeyHex = "3307ebe2c1d0ab1d95e62c9920abbf6a91938c4b86e34cff6683e879014f5e58"
	}
	password := "123456"
	walletPath := "./wallet.json"

	privateKey, _ := hex.DecodeString(privateKeyHex)
	err := walletgen(walletPath, password, privateKey)
	if err != nil {
		t.Fatal("walletgen error", err)
	}

	acc, err := account.FromStore(walletPath, password)
	if err != nil {
		t.Fatal("restore account error:", err)
	}

	if ss := acc.Signer.PrivateKeyString(); ss != privateKeyHex {
		t.Fatal("privatekey restore error: ", ss)
	}

	return
}

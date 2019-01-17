package test

import (
	"fmt"
	"github.com/Infnote/infnotechain/blockchain/crypto"
	"github.com/mr-tron/base58"
	"log"
	"testing"
)

var key *crypto.Key = nil

func KeyInit() {
	if key == nil {
		key = crypto.NewKey()
	}
}

func TestNewKey(t *testing.T) {
	KeyInit()

	if key == nil {
		t.Fail()
	}
}

func TestRaw(t *testing.T) {
	KeyInit()

	raw1 := base58.Encode(key.Raw())
	key, err := crypto.FromBytes(key.Raw())
	raw2 := base58.Encode(key.Raw())
	if err != nil || raw1 != raw2 {
		t.Fail()
	}

	fmt.Printf("Private Key Raw: %v\n", string(raw1))
}

func TestWIF(t *testing.T) {
	KeyInit()

	addr1 := "1FuSPXH9MDy2qMpwMNqthGRyLpMpxLZo1r"
	wif1 := "L53XkZ7tWUcuXHqzpBdsPfbYvibTxNGTy5mWk4hTth2vbxujrDMg"
	key2, err := crypto.FromWIF(wif1)
	if err != nil {
		log.Fatal(err)
	}

	wif2 := key2.ToWIF()

	fmt.Println(key2.ToAddress())
	fmt.Println(key2.ToWIF())

	if err != nil || wif1 != wif2 || addr1 != key2.ToAddress() {
		t.Fail()
	}
}

func TestPublicKey(t *testing.T) {
	KeyInit()

	b58 := base58.Encode(key.PublicKey())
	fmt.Printf("Public Key Raw: %v\n", string(b58))
}

func TestToAddress(t *testing.T) {
	KeyInit()

	fmt.Printf("Address: %v\n", key.ToAddress())
}

func TestSignAndVerify(t *testing.T) {
	KeyInit()

	msg := []byte("Test Sign and Verify")
	sig := key.Sign(msg)
	addr := key.ToAddress()
	if !crypto.Verify(addr, sig, msg) {
		t.Fail()
	}

	fmt.Printf("Signature: %v\n", base58.Encode(sig))
}

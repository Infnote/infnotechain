package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/Infnote/infnotechain/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"log"
)

type Key struct {
	privateKey *ecdsa.PrivateKey
}

func NewKey() *Key {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	return &Key{privateKey}
}

func FromBytes(privateKey []byte) (*Key, error) {
	priv, err := crypto.ToECDSA(privateKey)
	return &Key{priv}, err
}

func FromWIF(WIF string) (*Key, error) {
	data, err := base58.Decode(WIF)
	if err != nil {
		return nil, err
	}

	hashValue := utils.SHA256(utils.SHA256(data[:len(data)-4]))
	checksum := data[len(data)-4:]

	if !bytes.Equal(hashValue[:4], checksum) {
		return nil, fmt.Errorf("WIF checksum is not matched")
	}

	priv, err := crypto.ToECDSA(data[1 : len(data)-5])
	if err != nil {
		return nil, err
	}
	return &Key{priv}, nil
}

func (k Key) Raw() []byte {
	return crypto.FromECDSA(k.privateKey)
}

func (k Key) PublicKey() []byte {
	return crypto.CompressPubkey(&k.privateKey.PublicKey)
}

func (k Key) ToAddress() string {
	return publicKeyToAddress(k.PublicKey())
}

// 0x80 + raw private key + 0x01 + checksum 4 bytes
func (k Key) ToWIF() string {
	payload := append([]byte{0x80}, k.Raw()...)
	payload = append(payload, 0x01)
	checksum := utils.SHA256(utils.SHA256(payload))[:4]
	return base58.Encode(append(payload, checksum...))
}

// Add prefix contain recid for recovery.
// 31 + recid
// means recovery for compressed public key.
func (k Key) Sign(msg []byte) []byte {
	sig, err := crypto.Sign(utils.SHA256(msg), k.privateKey)
	if err != nil {
		log.Fatal(err)
	}
	return append([]byte{sig[len(sig)-1] + 31}, sig[:len(sig)-1]...)
}

// secp256k1 standard signature format put recid back of signature
// but bitcoin put it front and add 27 or 31 for uncompressed or compressed public key
func RecoverAddress(sig []byte, msg []byte) (string, error) {
	pub, err := crypto.SigToPub(utils.SHA256(msg), append(sig[1:], sig[0]-31))
	if err != nil {
		return "", err
	}
	return publicKeyToAddress(crypto.CompressPubkey(pub)), nil
}

func Verify(address string, sig []byte, msg []byte) bool {
	addr, err := RecoverAddress(sig, msg)
	if err != nil || addr != address {
		return false
	}
	return true
}


// 0x00 + hash160(sha256(compressed public key)) + checksum 4 bytes
func publicKeyToAddress(publicKey []byte) string {
	pre := append([]byte{0}, utils.RIPEMD160(utils.SHA256(publicKey))...)
	sum := utils.SHA256(utils.SHA256(pre))[:4]
	return base58.Encode(append(pre, sum...))
}

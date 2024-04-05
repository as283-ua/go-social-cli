package global

import (
	"crypto/rsa"
	"fmt"
	"util"
)

var defaultPub *rsa.PublicKey
var defaultPriv *rsa.PrivateKey

func LoadKeys(username string) error {
	publicKeyBytes := util.ReadPublicKeyBytesFromFile(fmt.Sprintf("%s.pub", username))
	defaultPub = util.ParsePublicKey(publicKeyBytes)

	var err error
	defaultPriv, err = util.ReadRSAKeyFromFile(fmt.Sprintf("%s.key", username))

	return err
}

func SetPriv(key *rsa.PrivateKey) {
	defaultPriv = key
}

func SetPub(key *rsa.PublicKey) {
	defaultPub = key
}

// logout
func ClearKeys() {
	defaultPriv = nil
	defaultPub = nil
}

func GetPublicKey() *rsa.PublicKey {
	return defaultPub
}

func GetPrivateKey() *rsa.PrivateKey {
	return defaultPriv
}

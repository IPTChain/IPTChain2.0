package cli

import (
	"math/rand"
	"time"

	"IPT/common/config"
	"IPT/common/log"
	"IPT/crypto"
)

func init() {
	log.Init()
	crypto.SetAlg(config.Parameters.EncryptAlg)
	//seed transaction nonce
	rand.Seed(time.Now().UnixNano())
}

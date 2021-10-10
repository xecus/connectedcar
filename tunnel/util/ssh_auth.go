package util

import (
	"encoding/base64"
	"io/ioutil"
	"log"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func GeneratePublicKeyHandler() ssh.PublicKeyHandler {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		user := ctx.User()
		sessionId := ctx.SessionID()
		clientVersion := ctx.ClientVersion()
		serverVersion := ctx.ServerVersion()
		remoteAddr := ctx.RemoteAddr()
		localAddr := ctx.LocalAddr()
		publicKeyString := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())
		log.Println("user =", user)
		log.Println("sessionID =", sessionId)
		log.Println("clientVersion =", clientVersion)
		log.Println("serverVersion =", serverVersion)
		log.Println("remoteAddr =", remoteAddr)
		log.Println("localAddr =", localAddr)
		log.Println("publicKeyString=", publicKeyString)

		data, err := ioutil.ReadFile("/home/hiroyuki/.ssh/id_rsa.pub")
		if err != nil {
			log.Fatal("Error: ioutil.ReadFile", err)
		}

		allowedKey, _, _, _, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			log.Fatal("Error: ssh.ParseAuthorizedKey", err)
		}

		log.Println("GivenKey", gossh.FingerprintLegacyMD5(key))
		log.Println("AllowedKey", gossh.FingerprintLegacyMD5(allowedKey))

		return user == "admin+xxx" && ssh.KeysEqual(key, allowedKey)
	}
}

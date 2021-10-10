package util

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gliderlabs/ssh"
	"github.com/xecus/connectedcar/config"
	gossh "golang.org/x/crypto/ssh"
)

func GeneratePublicKeyHandler(globalConfig *config.GlobalConfig) ssh.PublicKeyHandler {
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

		// Find PubKey undeer $HOME/.ssh
		homeDirPath, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		pubKeyPath := filepath.Join(homeDirPath, ".ssh", "id_rsa.pub")

		//FIXME
		data, err := ioutil.ReadFile(pubKeyPath)
		if err != nil {
			log.Fatal("Error: ioutil.ReadFile", err)
			return false
		}

		allowedKey, _, _, _, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			log.Fatal("Error: ssh.ParseAuthorizedKey", err)
			return false
		}

		log.Println("GivenKey", gossh.FingerprintLegacyMD5(key))
		log.Println("AllowedKey", gossh.FingerprintLegacyMD5(allowedKey))

		return user == "admin+xxx" && ssh.KeysEqual(key, allowedKey)
	}
}

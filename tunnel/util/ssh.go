package util

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func GenerateSessionHandler() func(ssh.Session) {
	return func(s ssh.Session) {
		cmd := exec.Command("/bin/bash")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			fmt.Println("ptyReq.Term=", ptyReq.Term)
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			ptmx, err := pty.Start(cmd)
			if err != nil {
				panic(err)
			}
			// Make sure to close the pty at the end.
			defer func() { ptmx.Close() }() // Best effort.
			// Handle pty size.
			go func() {
				for win := range winCh {
					setWinsize(ptmx, win.Width, win.Height)
				}
			}()
			go func() { io.Copy(ptmx, s) }()
			io.Copy(s, ptmx) // stdout
			cmd.Wait()
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	}
}

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

		log.Println("givenKey", gossh.FingerprintLegacyMD5(key))
		log.Println("allowedKey", gossh.FingerprintLegacyMD5(allowedKey))

		return user == "admin+xxx" && ssh.KeysEqual(key, allowedKey)
	}
}

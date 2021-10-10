package util

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
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

package main

import (
	"encoding/base64"
	"github.com/xpensia/sshgate"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

type GitApp struct {
	sshgate.BaseApp
}

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// the server meeds a private key
	// generate one with `ssh-keygen -t rsa`
	pemBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		log.Panicf("Failed to load private key: %#v\n", err)
	}

	// pass the private key and an authentication function
	server, err := sshgate.NewServer(pemBytes, Authenticate)
	if err != nil {
		log.Panicf("NewServer error: %#v\n", err)
	}

	// listen on specific port and address
	// "" is equivalent to "0.0.0.0" or “all interfaces”
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Panicf("Can't parse port: %#v\n", err)
	}
	if err := server.Listen("", port); err != nil {
		log.Panicf("Listen error: %#v\n", err)
	}
}

// check authorization and return appropriate sshgate.App
func Authenticate(c sshgate.Connection, user, algo string, pubkey []byte) (bool, sshgate.App) {
	// here we return GitApp for user git
	if user == "git" {
		// we should check the content of ~/.ssh/authorized_keys
		// or search the key in a database
		log.Printf("Allow connection for user git with his %s key\n", algo)
		log.Printf("Key: %s\n", base64.StdEncoding.EncodeToString(pubkey))
		return true, GitApp{}
	} else {
		log.Printf("Refuse connection for user %s\n", user)
		return false, nil
	}
}

// Implement sshgate.Executable
func (a GitApp) CanExec(cmd string, args []string, env map[string]string) bool {
	// TODO : check repo authorization
	log.Println("User want to exec: ", cmd, args)
	return cmd == "git-receive-pack" || cmd == "git-upload-pack"
}

func (a GitApp) Exec(cmd string, args []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) int {
	log.Println("Exec: ", cmd, args)
	// see http://godoc.org/os/exec
	git := exec.Command(cmd, args...)
	git.Stdin = stdin
	git.Stdout = stdout
	git.Stderr = stderr
	git.Env = sshgate.MapToEnviron(env)
	if err := git.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			if status, ok := exit.Sys().(syscall.WaitStatus); ok {
				// see :
				// - https://groups.google.com/forum/#!topic/golang-nuts/dKbL1oOiCIY
				// - https://groups.google.com/forum/#!topic/golang-nuts/8XIlxWgpdJw
				// for Unix
				return int(status.ExitStatus())
				// for Windows
				// return int(status.ExitCode)
			} else {
				return 1
			}
		} else {
			log.Printf("IO error: %#v\n", err)
			return 1
		}
	} else {
		// all good
		return 0
	}
}

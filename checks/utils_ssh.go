package checks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHAuthenticator struct {
	passwordAuth  bool
	agentAuth     bool
	publicKeyAuth bool
	auths         []ssh.AuthMethod
	agent         net.Conn
}

type SSHAuthOptions struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Key      string `json:"key"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

func NewSSHAuthenticator(logger *log.Logger, options SSHAuthOptions) *SSHAuthenticator {

	authenticator := &SSHAuthenticator{
		auths: []ssh.AuthMethod{},
	}

	if options.Password != "" {
		logger.Println("ssh via password is an enabled option")
		authenticator.auths = append(authenticator.auths, ssh.Password(options.Password))
		authenticator.passwordAuth = true
	}

	if os.Getenv("SSH_AUTH_SOCK") != "" {
		if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
			logger.Println("ssh via ssh-agent is an enabled option")
			authenticator.auths = append(authenticator.auths, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
			authenticator.agentAuth = true
			authenticator.agent = sshAgent
		}
	}

	if options.Key != "" {
		if pubkey, err := getKey(options.Key); err == nil {
			logger.Println("ssh via public key is an enabled option")
			authenticator.auths = append(authenticator.auths, ssh.PublicKeys(pubkey))
			authenticator.publicKeyAuth = true
		}
	}

	return authenticator
}

func (s *SSHAuthenticator) Cleanup() error {
	if s.agentAuth {
		return s.agent.Close()
	}
	return nil
}

func getKey(filename string) (ssh.Signer, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	pubkey, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}
	return pubkey, nil
}

func runCommand(client *ssh.Client, cmd string) ([]byte, error) {
	var out []byte
	session, err := client.NewSession()
	if err != nil {
		return out, nil
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run(cmd)
	out = b.Bytes()
	return out, err
}

func runCommandStrOutput(client *ssh.Client, cmd string) (string, error) {
	b, err := runCommand(client, cmd)
	if err != nil {
		return "", fmt.Errorf("Error running command %s", cmd)
	}
	return string(b), nil
}

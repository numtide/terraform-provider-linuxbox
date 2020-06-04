package sshsession

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type Session struct {
	cl *ssh.Client
}

func IsExecError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "Process exited with status")
}

func IsConnectTimeout(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "timed out while connecting to ssh")

}

func (s *Session) RunInSession(cmd string) ([]byte, []byte, error) {
	session, err := s.cl.NewSession()
	if err != nil {
		return nil, nil, errors.Wrap(err, "while open ssh session")
	}
	defer session.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	session.Stdout = stdout
	session.Stderr = stderr
	err = session.Run(cmd)
	return stdout.Bytes(), stderr.Bytes(), err
}

func (s *Session) Close() error {
	return s.cl.Close()
}

func Open(d *schema.ResourceData) (*Session, error) {
	privateKeyBytes := d.Get("ssh_key").(string)

	signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(privateKeyBytes), []byte{})

	if err != nil {
		return nil, errors.Wrap(err, "while parsing private ssh_key")
	}

	sshUser := d.Get("ssh_user").(string)

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	server := d.Get("host_address").(string)

	addr := fmt.Sprintf("%s:22", server)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	for {
		c, err := (&net.Dialer{
			Timeout: 15 * time.Second,
		}).DialContext(ctx, "tcp", addr)
		if err == nil {
			c.Close()
			break
		}
		if ctx.Err() != nil {
			return nil, errors.Wrap(err, "timed out while connecting to ssh")
		}
		time.Sleep(1 * time.Second)
	}
	cancel()

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, errors.Wrapf(err, "while ssh dialing %s", server)
	}

	return &Session{
		cl: client,
	}, nil

}

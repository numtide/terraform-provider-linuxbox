package sshsession

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	serrors "errors"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

var SessionLimit = 5

type sshClient struct {
	*ssh.Client
	sessionsInUse int
	mu            *sync.Mutex
	cond          *sync.Cond
}

func newSSHClient(sc *ssh.Client) *sshClient {
	mu := new(sync.Mutex)
	return &sshClient{
		Client:        sc,
		sessionsInUse: 0,
		mu:            mu,
		cond:          sync.NewCond(mu),
	}
}

type sshSession struct {
	*ssh.Session
	cl *sshClient
}

func (s *sshClient) NewSession() (*sshSession, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	for s.sessionsInUse >= SessionLimit {
		s.cond.Wait()
	}

	cs, err := s.Client.NewSession()
	if err != nil {
		return nil, err
	}

	s.sessionsInUse++
	return &sshSession{
		Session: cs,
		cl:      s,
	}, nil
}

func (s *sshSession) Close() error {
	defer func() {
		s.cl.mu.Lock()
		s.cl.sessionsInUse--
		s.cl.cond.Broadcast()
		s.cl.mu.Unlock()
	}()
	return s.Session.Close()
}

// ErrTimeout is returned when there was a timeout when connecting to SSH daemon.
var ErrTimeout = serrors.New("timed out connecting to ssh daemon")

func newClientFuture() *clientFuture {
	mu := new(sync.Mutex)
	return &clientFuture{
		mu:  mu,
		cnd: sync.NewCond(mu),
	}
}

type clientFuture struct {
	client *sshClient
	err    error
	mu     *sync.Mutex
	cnd    *sync.Cond
}

func (cf *clientFuture) getClient() (*sshClient, error) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	for cf.client == nil && cf.err == nil {
		cf.cnd.Wait()
	}
	return cf.client, cf.err
}

func (cf *clientFuture) createClientInternal(cp clientParams) (*sshClient, error) {
	signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(cp.privateKey), []byte{})

	if err != nil {
		return nil, errors.Wrap(err, "while parsing private ssh_key")
	}

	server := cp.hostAddress

	addr := fmt.Sprintf("%s:22", server)

	config := &ssh.ClientConfig{
		User: cp.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	var client *ssh.Client
	deadline := time.Now().Add(time.Minute)
	for {
		client, err = ssh.Dial("tcp", addr, config)
		if err == nil {
			break
		}

		if IsConnectTimeout(err) {
			return nil, ErrTimeout
		}

		if time.Now().Before(deadline) {
			time.Sleep(1 * time.Second)
			continue
		}

		return nil, err
	}

	return newSSHClient(client), nil

}

func IsConnectTimeout(err error) bool {
	if err == nil {
		return false
	}

	if err == ErrTimeout {
		return true
	}

	msg := err.Error()
	return strings.Contains(msg, "timed out while connecting to ssh")

}

func (cf *clientFuture) createClient(cp clientParams) {
	cl, err := cf.createClientInternal(cp)
	cf.mu.Lock()
	cf.client = cl
	cf.err = err
	cf.cnd.Broadcast()
	cf.mu.Unlock()
}

type clientParams struct {
	privateKey  string
	user        string
	hostAddress string
}

var clientPool = map[clientParams]*clientFuture{}
var clientPoolMu = new(sync.Mutex)

func getClient(d *schema.ResourceData) (*sshClient, error) {
	cp := clientParams{
		privateKey:  d.Get("ssh_key").(string),
		user:        d.Get("ssh_user").(string),
		hostAddress: d.Get("host_address").(string),
	}

	didCreateFuture := false

	clientPoolMu.Lock()
	cf := clientPool[cp]
	if cf == nil {
		didCreateFuture = true
		cf = newClientFuture()
		clientPool[cp] = cf
	}
	clientPoolMu.Unlock()

	if didCreateFuture {
		cf.createClient(cp)
	}

	return cf.getClient()
}

func Run(d *schema.ResourceData, cmd string) ([]byte, []byte, error) {
	cl, err := getClient(d)
	if err != nil {
		return nil, nil, err
	}
	session, err := cl.NewSession()
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

func RunWithStdin(d *schema.ResourceData, cmd string, stdin io.Reader) ([]byte, []byte, error) {
	cl, err := getClient(d)
	if err != nil {
		return nil, nil, err
	}
	session, err := cl.NewSession()
	if err != nil {
		return nil, nil, errors.Wrap(err, "while open ssh session")
	}
	defer session.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr
	err = session.Run(cmd)
	return stdout.Bytes(), stderr.Bytes(), err

}

func Check(d *schema.ResourceData) error {
	_, err := getClient(d)
	return err
}

func IsExecError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "Process exited with status")
}

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,

		Schema: map[string]*schema.Schema{
			"ssh_key": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"ssh_user": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Default:  "root",
				Optional: true,
			},

			"host_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"registry_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"username": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	client, err := createSSHClient(d)
	if err != nil {
		return errors.Wrap(err, "while creating ssh client")
	}

	defer client.Close()

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	registryAddress := d.Get("registry_address").(string)

	cmd := []string{
		"docker",
		"login",
		"-u",
		username,
		"-p",
		password,
		registryAddress,
	}

	line := strings.Join(cmd, " ")

	output, err := runInSession(client, line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`: %s", line, string(output))
	}

	d.SetId(fmt.Sprintf("%s@%s", username, registryAddress))

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] host address not changed")

	client, err := createSSHClient(d)
	if err != nil {
		if isConnectTimeout(err) {
			log.Println("[DEBUG] connecting to SSH timed out, resource is dead")
			d.SetId("")
			return nil
		}
		return errors.Wrap(err, "while creating ssh client")
	}

	log.Println("[DEBUG] got ssh client")

	output, err := runInSession(client, "cat ~/.docker/config.json")
	if err != nil {
		if isExecError(err) {
			d.SetId("")
			return nil
		}
		return errors.Wrap(err, "while getting docker auth config")
	}

	cfg := &dockerConfig{}

	err = json.Unmarshal(output, cfg)

	if err != nil {
		return errors.Wrap(err, "while parsing docker config file")
	}

	registryAddress := d.Get("registry_address").(string)

	authStruct, found := cfg.Auths[registryAddress]

	if !found {
		d.SetId("")
		return nil
	}

	authString, err := base64.StdEncoding.DecodeString(authStruct.Auth)
	if err != nil {
		return errors.Wrapf(err, "while base64 decoding auth for %s", registryAddress)
	}

	parts := strings.SplitN(string(authString), ":", 2)
	if len(parts) != 2 {
		return errors.Errorf("malformed auth for %s", registryAddress)
	}

	d.Set("username", parts[0])
	d.Set("password", parts[1])

	d.SetId(fmt.Sprintf("%s@%s", parts[0], registryAddress))

	return nil
}

type dockerConfig struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceCreate(d, m)
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	client, err := createSSHClient(d)
	if err != nil {
		return errors.Wrap(err, "while creating ssh client")
	}

	defer client.Close()

	registryAddress := d.Get("registry_address").(string)

	cmd := []string{
		"docker",
		"logout",
		registryAddress,
	}

	line := strings.Join(cmd, " ")

	output, err := runInSession(client, line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`: %s", line, string(output))
	}

	return nil
}

func isExecError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "Process exited with status")
}

func isConnectTimeout(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "timed out while connecting to ssh")

}

func runInSession(c *ssh.Client, command string) ([]byte, error) {
	session, err := c.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "while open ssh session")
	}
	defer session.Close()
	return session.Output(command)
}

func createSSHClient(d *schema.ResourceData) (*ssh.Client, error) {
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

	return client, nil

}

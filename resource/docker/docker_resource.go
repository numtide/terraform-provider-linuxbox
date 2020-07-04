package docker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Create:             resourceCreate,
		Read:               resourceRead,
		Update:             resourceUpdate,
		Delete:             resourceDelete,
		DeprecationMessage: "This resource is deprecated, please use linuxbox_run_setup instead",

		Schema: map[string]*schema.Schema{
			"host_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	privateKeyBytes := d.Get("ssh_key").(string)
	sshUser := d.Get("ssh_key").(string)

	signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(privateKeyBytes), []byte{})

	if err != nil {
		return errors.Wrap(err, "while parsing private ssh_key")
	}

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	server := d.Get("host_address").(string)

	addr := fmt.Sprintf("%s:22", server)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	for {
		c, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
		if err == nil {
			c.Close()
			break
		}
		time.Sleep(1 * time.Second)
	}
	cancel()

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return errors.Wrapf(err, "while ssh dialing %s", server)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return errors.Wrap(err, "while open ssh session")
	}

	output, err := runInSession(client, "which docker || true")
	if err != nil {
		return errors.Wrap(err, "while running `which docker`")
	}

	if string(output) == "" {
		commands := []string{
			"apt update",
			"apt install -y apt-transport-https ca-certificates curl software-properties-common",
			"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
			"add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable\"",
			"apt update",
			"apt install -y docker-ce",
		}

		for _, cmd := range commands {
			output, err = runInSession(client, cmd)
			if err != nil {
				return errors.Wrapf(err, "while running `%s`, output: %s", cmd, string(output))
			}
			session.Stdout = nil
		}

	}

	// TODO add docker version computed value
	d.SetId("docker")

	// return errors.Errorf("output %q", string(output))

	return resourceRead(d, m)
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceRead(d, m)
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	// TODO add removing docker
	return nil
}

func runInSession(c *ssh.Client, command string) ([]byte, error) {
	session, err := c.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "while open ssh session")
	}
	defer session.Close()
	return session.Output(command)
}

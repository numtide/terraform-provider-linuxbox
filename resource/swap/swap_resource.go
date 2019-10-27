package swap

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
			"host_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"swap_size": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	privateKeyBytes := d.Get("ssh_key").(string)

	swapSize := d.Get("swap_size").(string)

	signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(privateKeyBytes), []byte{})

	if err != nil {
		return errors.Wrap(err, "while parsing private ssh_key")
	}

	config := &ssh.ClientConfig{
		User: "root",
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

	output, err := runInSession(client, "swapon -s")
	if err != nil {
		return errors.Wrap(err, "while running `swapon -s`")
	}

	if string(output) == "" {
		commands := []string{
			fmt.Sprintf("fallocate -l %s /swapfile", swapSize),
			"chmod 0600 /swapfile",
			"mkswap /swapfile",
			"swapon /swapfile",
			"echo /swapfile none swap defaults 0 0 >> /etc/fstab",
		}

		for _, cmd := range commands {
			output, err = runInSession(client, cmd)
			if err != nil {
				return errors.Wrapf(err, "while running `%s`, output: %s", cmd, string(output))
			}
			session.Stdout = nil
		}

	}
	d.SetId(server + ":" + swapSize)

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

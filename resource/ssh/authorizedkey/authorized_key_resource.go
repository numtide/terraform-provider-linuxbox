package authorizedkey

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/alessio/shellescape"
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
			"key_to_add": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	privateKeyBytes := d.Get("ssh_key").(string)
	sshUser := d.Get("ssh_user").(string)

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

	keyToAdd := d.Get("key_to_add").(string)

	script := fmt.Sprintf("([ ! -d ~/.ssh ] && (mkdir ~/.ssh && chmod 700 ~/.ssh)) ; echo %s >> ~/.ssh/authorized_keys && chmod 700 ~/.ssh/authorized_keys", shellescape.Quote(keyToAdd))

	output, err := runInSession(client, script)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`: %s", script, string(output))
	}

	d.SetId("key")

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

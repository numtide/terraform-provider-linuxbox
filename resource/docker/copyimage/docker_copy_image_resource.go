package copyimage

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/docker/docker/client"
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
			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	dc, err := client.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "while creating docker client")
	}

	imageID := d.Get("image_id").(string)

	rc, err := dc.ImageSave(context.Background(), []string{imageID})
	if err != nil {
		return errors.Wrapf(err, "while saving image %s", imageID)
	}

	defer rc.Close()

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

	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return errors.Wrapf(err, "while ssh dialing %s", server)
	}

	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return errors.Wrap(err, "while creating ssh session")
	}

	defer session.Close()

	session.Stdin = rc
	output, err := session.Output("docker load")
	if err != nil {
		return errors.Wrapf(err, "error while executing `docker load` via ssh: %s", string(output))
	}

	d.SetId(imageID)

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

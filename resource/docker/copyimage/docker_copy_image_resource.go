package copyimage

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
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
			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"image_path": &schema.Schema{
				Type:     schema.TypeString,
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

	imageID := d.Get("image_id")
	imagePath := d.Get("image_path")

	if imageID == nil && imagePath == nil {
		return errors.New("argument error: one of image_id or image_path must be passed")
	}

	if imageID != nil && imagePath != nil {
		return errors.New("argument error: only one of image_id or image_path must be passed")
	}

	var (
		r io.ReadCloser
		resourceID string
	)

	if imageID != nil {
		resourceID = imageID.(string)
		r, err = dc.ImageSave(context.Background(), []string{resourceID})
		if err != nil {
			return errors.Wrapf(err, "image_id: while saving image %s", resourceID)
		}
	} else {
		resourceID = imagePath.(string)
		r, err = os.Open(resourceID)
		if err != nil {
			return errors.Wrapf(err, "image_path: while saving image %s", resourceID)
		}
	}
	defer r.Close()

	privateKeyBytes := d.Get("ssh_key").(string)

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

	session.Stdin = r
	output, err := session.Output("docker load")
	if err != nil {
		return errors.Wrapf(err, "error while executing `docker load` via ssh: %s", string(output))
	}

	d.SetId(resourceID)

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

package swap

import (
	"encoding/base64"
	"fmt"

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
				Type:     schema.TypeString,
				Required: true,
			},
			"remote": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"swap_size": &schema.Schema{
				Type:     schema.TypeString,
				Computed: false,
				ForceNew: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	privateKeyBytes, err := base64.StdEncoding.DecodeString(d.Get("ssh_key").(string))

	if err != nil {
		return errors.Wrap(err, "while base64 decoding ssh_key")
	}

	signer, err := ssh.ParsePrivateKey(privateKeyBytes)

	if err != nil {
		return errors.Wrap(err, "while parsing private ssh_key")
	}

	config := &ssh.ClientConfig{
		User: "username",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	server := d.Get("remote").(string)

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", server), config)
	if err != nil {
		return errors.Wrapf(err, "while ssh dialing %s", server)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return errors.Wrap(err, "while open ssh session")
	}

	output, err := session.Output("swapon -s")
	if err != nil {
		return errors.Wrap(err, "while running `swapon -s`")
	}

	return errors.New(string(output))

	// return resourceRead(d, m)
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

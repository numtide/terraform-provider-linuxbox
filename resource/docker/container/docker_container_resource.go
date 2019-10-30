package container

import (
	"context"
	"fmt"
	"net"
	"strings"
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
				Required: true,
				ForceNew: true,
			},

			"ports": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	portsSet := d.Get("ports").(*schema.Set)
	ports := []string{}

	for _, p := range portsSet.List() {
		ports = append(ports, p.(string))
	}

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

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return errors.Wrapf(err, "while ssh dialing %s", server)
	}

	defer client.Close()

	imageID := d.Get("image_id").(string)

	// cmd := "docker run " + shellescape.Quote(imageID)

	cmd := []string{
		"docker",
		"run",
		"-d",
	}

	if len(ports) > 0 {
		for _, p := range ports {
			cmd = append(cmd, "-p", shellescape.Quote(p))
		}
	}

	cmd = append(cmd, shellescape.Quote(imageID))

	line := strings.Join(cmd, " ")

	output, err := runInSession(client, line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`: %s", line, string(output))
	}

	outputLines := strings.Split(string(output), "\n")

	if len(outputLines) < 2 {
		return errors.New("remote docker didn't return container id")
	}

	containerID := outputLines[len(outputLines)-2]

	d.Set("container_id", containerID)
	d.SetId(containerID)

	return resourceRead(d, m)
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceRead(d, m)
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
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

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return errors.Wrapf(err, "while ssh dialing %s", server)
	}

	defer client.Close()

	cmd := []string{
		"docker",
		"rm",
		"-fv",
		shellescape.Quote(d.Id()),
	}

	line := strings.Join(cmd, " ")

	output, err := runInSession(client, line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`: %s", line, string(output))
	}

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

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
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},

			"volumes": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},

			"labels": &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},

			"env": &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},

			"restart": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"args": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},

			"container_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

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

	cmd := []string{
		"docker",
		"run",
		"-d",
	}

	restart, restartSet := d.GetOkExists("restart")

	if restartSet {
		cmd = append(cmd, "--restart", shellescape.Quote(restart.(string)))
	}

	labelsMap := d.Get("labels").(map[string]interface{})

	for k, v := range labelsMap {
		cmd = append(cmd, "-l", fmt.Sprintf("%s=%s", shellescape.Quote(k), shellescape.Quote(v.(string))))
	}

	envMap := d.Get("env").(map[string]interface{})

	for k, v := range envMap {
		cmd = append(cmd, "-e", fmt.Sprintf("%s=%s", shellescape.Quote(k), shellescape.Quote(v.(string))))
	}

	portsSet := d.Get("ports").(*schema.Set)
	ports := []string{}

	for _, p := range portsSet.List() {
		ports = append(ports, p.(string))
	}

	if len(ports) > 0 {
		for _, p := range ports {
			cmd = append(cmd, "-p", shellescape.Quote(p))
		}
	}

	volumesSet := d.Get("volumes").(*schema.Set)
	volumes := []string{}

	for _, p := range volumesSet.List() {
		volumes = append(volumes, p.(string))
	}

	if len(volumes) > 0 {
		for _, v := range volumes {
			cmd = append(cmd, "-v", shellescape.Quote(v))
		}
	}

	cmd = append(cmd, shellescape.Quote(imageID))

	for _, a := range d.Get("args").([]interface{}) {
		cmd = append(cmd, shellescape.Quote(a.(string)))
	}

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

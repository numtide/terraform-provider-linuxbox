package network

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/numtide/terraform-provider-linuxbox/sshsession"
	"github.com/pkg/errors"
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

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	name := d.Get("name").(string)

	cmd := []string{
		"docker",
		"network",
		"create",
		shellescape.Quote(name),
	}

	line := strings.Join(cmd, " ")

	stdout, stderr, err := sshsession.Run(d, line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`: %s", line, string(stderr))
	}

	id := strings.TrimSuffix(string(stdout), "\n")

	d.SetId(id)

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {

	stdout, _, err := sshsession.Run(d, fmt.Sprintf("docker network inspect %s", d.Id()))
	if sshsession.IsExecError(err) {
		d.SetId("")
		return nil
	}

	type network struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	}

	networks := []network{}

	err = json.Unmarshal(stdout, &networks)

	if err != nil {
		return errors.Wrap(err, "while parsing docker network json")
	}

	if len(networks) != 1 {
		return errors.Errorf("expected one network with id %s, found %d", d.Id(), len(networks))
	}

	d.Set("name", networks[0].Name)

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	return errors.New("update is not supported")
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {

	cmd := fmt.Sprintf("docker network rm %s", d.Id())

	stdout, stderr, err := sshsession.Run(d, cmd)

	if err != nil {
		errors.Wrapf(err, "error while executing `%s` via ssh STDOUT:\n%s\nSTDERR:%s\n", cmd, string(stdout), string(stderr))
	}

	return err
}

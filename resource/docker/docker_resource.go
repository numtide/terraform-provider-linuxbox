package docker

import (
	"github.com/numtide/terraform-provider-linuxbox/sshsession"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
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
			"host_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Default:  "root",
				Optional: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	line := "which docker || true"

	stdout, stderr, err := sshsession.Run(d, line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`\nstdout:\n%s\nstderr:\n%s\n", line, string(stdout), string(stderr))
	}

	if string(stdout) == "" {
		commands := []string{
			"apt update",
			"apt install -y apt-transport-https ca-certificates curl software-properties-common",
			"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
			"add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable\"",
			"apt update",
			"apt install -y docker-ce",
		}

		for _, cmd := range commands {
			stdout, stderr, err := sshsession.Run(d, cmd)
			if err != nil {
				return errors.Wrapf(err, "while running `%s`\nstdout:\n%s\nstderr:\n%s\n", cmd, string(stdout), string(stderr))
			}
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

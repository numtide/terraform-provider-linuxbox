package swap

import (
	"fmt"

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
			"ssh_user": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Default:  "root",
				Optional: true,
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

	cmd := "swapon -s"
	stdout, stderr, err := sshsession.Run(d, cmd)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`:\nSTDOUT:\n%s\nSTDERR:\n%s\n", cmd, string(stdout), string(stderr))
	}

	swapSize := d.Get("swap_size").(string)

	if string(stdout) == "" {
		commands := []string{
			fmt.Sprintf("fallocate -l %s /swapfile", swapSize),
			"chmod 0600 /swapfile",
			"mkswap /swapfile",
			"swapon /swapfile",
			"echo /swapfile none swap defaults 0 0 >> /etc/fstab",
		}

		for _, cmd := range commands {
			stdout, stderr, err := sshsession.Run(d, cmd)
			if err != nil {
				return errors.Wrapf(err, "while running `%s`:\nSTDOUT:\n%s\nSTDERR:\n%s\n", cmd, string(stdout), string(stderr))
			}
		}

	}

	server := d.Get("host_address").(string)

	d.SetId(server + ":" + swapSize)

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

package textfile

import (
	"fmt"

	"github.com/alessio/shellescape"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/numtide/terraform-provider-linuxbox/sshsession"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Read: resourceRead,

		Schema: map[string]*schema.Schema{
			"ssh_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"ssh_user": {
				Type:     schema.TypeString,
				Required: false,
				Default:  "root",
				Optional: true,
			},

			"host_address": {
				Type:     schema.TypeString,
				Required: true,
			},

			"sudo": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"path": {
				Type:     schema.TypeString,
				Required: true,
			},

			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)

	cmd := fmt.Sprintf("cat %s", shellescape.Quote(path))
	if d.Get("sudo").(bool) {
		cmd = fmt.Sprintf("sudo %s", cmd)
	}

	stdout, _, err := sshsession.Run(d, cmd)
	if err != nil {
		return fmt.Errorf("while getting content of %s: %w", path, err)
	}

	stdoutString := string(stdout)

	d.Set("content", stdoutString)

	d.SetId("-")

	return nil

}

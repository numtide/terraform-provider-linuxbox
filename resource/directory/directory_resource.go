package directory

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/numtide/terraform-provider-linuxbox/sshsession"
	"github.com/pkg/errors"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Create: resourceUpdateAndCreate,
		Read:   resourceRead,
		Update: resourceUpdateAndCreate,
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

			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"owner": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"group": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "755",
			},
		},
	}
}

func resourceUpdateAndCreate(d *schema.ResourceData, m interface{}) error {

	path := d.Get("path").(string)

	owner := d.Get("owner").(int)
	group := d.Get("group").(int)

	mode := d.Get("mode").(string)

	cmd := fmt.Sprintf("mkdir -p %s && chmod %s %s && chown %d:%d %s",
		shellescape.Quote(path),
		shellescape.Quote(mode),
		shellescape.Quote(path),
		owner,
		group,
		shellescape.Quote(path),
	)

	stdout, stderr, err := sshsession.Run(d, cmd)
	if err != nil {
		return errors.Wrapf(err, "error while creating dir %q:\nSTDOUT:\n%s\nSTDERR:\n%s\n", path, string(stdout), string(stderr))
	}

	sh := sha256.New()

	sh.Write([]byte(path))
	sum := sh.Sum(nil)

	d.SetId(hex.EncodeToString(sum[:]))

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)

	{
		cmd := fmt.Sprintf("stat -c '%%u %%g %%a' %s", shellescape.Quote(path))

		stdout, _, err := sshsession.Run(d, cmd)
		if err != nil {
			d.SetId("")
			return nil
		}

		stdoutString := string(stdout)

		stdoutString = strings.TrimSuffix(stdoutString, "\n")

		parts := strings.Split(stdoutString, " ")
		if len(parts) != 3 {
			return errors.Errorf("malformed output of %q: %q", cmd, stdoutString)
		}

		owner, err := strconv.Atoi(parts[0])
		if err != nil {
			return errors.Wrapf(err, "while parsing owner id %q", parts[0])
		}

		d.Set("owner", owner)

		group, err := strconv.Atoi(parts[1])
		if err != nil {
			return errors.Wrapf(err, "while parsing group id %q", parts[1])
		}

		d.Set("group", group)

		d.Set("mode", parts[2])
	}

	return nil

}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)

	cmd := fmt.Sprintf("rm -rf %s", shellescape.Quote(path))

	stdout, stderr, err := sshsession.Run(d, cmd)
	if err != nil {
		return errors.Wrapf(err, "error while deleting dir %q:\nSTDOUT:\n%s\nSTDERR:\n%s\n", path, string(stdout), string(stderr))
	}

	return nil
}

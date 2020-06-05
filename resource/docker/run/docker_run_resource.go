package run

import (
	"fmt"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/draganm/terraform-provider-linuxbox/sshsession"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

			"caps": &schema.Schema{
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

			"network": &schema.Schema{
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

			"stdout": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"stderr": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	ss, err := sshsession.Open(d)
	if err != nil {
		return errors.Wrap(err, "while opening ssh session")
	}

	defer ss.Close()

	imageID := d.Get("image_id").(string)

	cmd := []string{
		"docker",
		"run",
		"--rm",
	}

	network, networkSet := d.GetOkExists("network")

	if networkSet {
		cmd = append(cmd, "--network", shellescape.Quote(network.(string)))
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

	capsSet := d.Get("caps").(*schema.Set)
	caps := []string{}

	for _, p := range capsSet.List() {
		caps = append(caps, p.(string))
	}

	if len(caps) > 0 {
		for _, c := range caps {
			cmd = append(cmd, fmt.Sprintf("--cap-add=%s", shellescape.Quote(c)))
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

	stdout, stderr, err := ss.RunInSession(line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`\nstdout:\n%s\nstderr:\n%s\n", line, string(stdout), string(stderr))
	}

	d.SetId("success")
	d.Set("stdout", string(stdout))
	d.Set("stderr", string(stderr))

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {

	return nil
}

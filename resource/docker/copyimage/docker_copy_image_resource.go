package copyimage

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/numtide/terraform-provider-linuxbox/sshsession"
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
			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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

	imageID := d.Get("image_id").(string)

	rc, err := dc.ImageSave(context.Background(), []string{imageID})
	if err != nil {
		return errors.Wrapf(err, "while saving image %s", imageID)
	}

	defer rc.Close()

	output, stderr, err := sshsession.RunWithStdin(d, "docker load", rc)
	if err != nil {
		return errors.Wrapf(err, "error while executing `docker load` via ssh STDOUT:\n%s\nSTDERR:%s\n", string(output), string(stderr))
	}

	d.SetId(imageID)

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

package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

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

			"registry_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"username": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	registryAddress := d.Get("registry_address").(string)

	cmd := []string{
		"docker",
		"login",
		"-u",
		username,
		"-p",
		password,
		registryAddress,
	}

	line := strings.Join(cmd, " ")

	stdout, stderr, err := sshsession.Run(d, line)
	if err != nil {
		return errors.Wrapf(err, "while running `%s`\nSTDOUT:\n%s\nSTDERR\n%s\n", line, string(stdout), string(stderr))
	}

	d.SetId(fmt.Sprintf("%s@%s", username, registryAddress))

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {

	stdout, _, err := sshsession.Run(d, "cat ~/.docker/config.json")
	if err != nil {

		if sshsession.IsConnectTimeout(err) {
			// can't reach the host. Possibly the IP has changed, so assume everything is ok.
			return nil
		}

		if sshsession.IsExecError(err) {
			d.SetId("")
			return nil
		}
		return errors.Wrap(err, "while getting docker auth config")
	}

	cfg := &dockerConfig{}

	err = json.Unmarshal(stdout, cfg)

	if err != nil {
		return errors.Wrap(err, "while parsing docker config file")
	}

	registryAddress := d.Get("registry_address").(string)

	authStruct, found := cfg.Auths[registryAddress]

	if !found {
		d.SetId("")
		return nil
	}

	authString, err := base64.StdEncoding.DecodeString(authStruct.Auth)
	if err != nil {
		return errors.Wrapf(err, "while base64 decoding auth for %s", registryAddress)
	}

	parts := strings.SplitN(string(authString), ":", 2)
	if len(parts) != 2 {
		return errors.Errorf("malformed auth for %s", registryAddress)
	}

	d.Set("username", parts[0])
	d.Set("password", parts[1])

	d.SetId(fmt.Sprintf("%s@%s", parts[0], registryAddress))

	return nil
}

type dockerConfig struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceCreate(d, m)
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {

	registryAddress := d.Get("registry_address").(string)

	cmd := []string{
		"docker",
		"logout",
		registryAddress,
	}

	line := strings.Join(cmd, " ")

	stdout, stderr, err := sshsession.Run(d, line)
	if err != nil {

		if sshsession.IsConnectTimeout(err) {
			// if we can't reach the host, probably it's destroyed
			return nil
		}

		return errors.Wrapf(err, "while running `%s`\nSTDOUT:\n%s\nSTDERR\n%s\n", line, string(stdout), string(stderr))
	}

	return nil
}

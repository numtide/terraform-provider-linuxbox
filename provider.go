package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/numtide/terraform-provider-linuxbox/datasource/sourcehash"
	"github.com/numtide/terraform-provider-linuxbox/resource/docker"
	"github.com/numtide/terraform-provider-linuxbox/resource/docker/auth"
	"github.com/numtide/terraform-provider-linuxbox/resource/docker/build"
	"github.com/numtide/terraform-provider-linuxbox/resource/docker/container"
	"github.com/numtide/terraform-provider-linuxbox/resource/docker/copyimage"
	"github.com/numtide/terraform-provider-linuxbox/resource/docker/network"
	"github.com/numtide/terraform-provider-linuxbox/resource/docker/run"
	"github.com/numtide/terraform-provider-linuxbox/resource/runsetup"
	"github.com/numtide/terraform-provider-linuxbox/resource/ssh/authorizedkey"
	"github.com/numtide/terraform-provider-linuxbox/resource/swap"
	"github.com/numtide/terraform-provider-linuxbox/sshsession"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ssh_session_limit": &schema.Schema{
				Type:     schema.TypeInt,
				Default:  5,
				Optional: true,
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"linuxbox_source_hash": sourcehash.Resource(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"linuxbox_swap":               swap.Resource(),
			"linuxbox_ssh_authorized_key": authorizedkey.Resource(),
			"linuxbox_docker":             docker.Resource(),
			"linuxbox_docker_copy_image":  copyimage.Resource(),
			"linuxbox_docker_build":       build.Resource(),
			"linuxbox_docker_container":   container.Resource(),
			"linuxbox_docker_auth":        auth.Resource(),
			"linuxbox_docker_network":     network.Resource(),
			"linuxbox_docker_run":         run.Resource(),
			"linuxbox_run_setup":          runsetup.Resource(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			sshsession.SessionLimit = d.Get("ssh_session_limit").(int)
			return nil, nil
		},
	}
}

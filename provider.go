package main

import (
	"github.com/draganm/terraform-provider-linuxbox/datasource/sourcehash"
	"github.com/draganm/terraform-provider-linuxbox/resource/docker"
	"github.com/draganm/terraform-provider-linuxbox/resource/docker/build"
	"github.com/draganm/terraform-provider-linuxbox/resource/docker/copyimage"
	"github.com/draganm/terraform-provider-linuxbox/resource/swap"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"linuxbox_source_hash": sourcehash.Resource(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"linuxbox_swap":              swap.Resource(),
			"linuxbox_docker":            docker.Resource(),
			"linuxbox_docker_copy_image": copyimage.Resource(),
			"linuxbox_docker_build":      build.Resource(),
		},
	}
}

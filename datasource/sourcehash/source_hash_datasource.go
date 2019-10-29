package sourcehash

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Read: resourceRead,

		Schema: map[string]*schema.Schema{
			"source_dirs": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"hash": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRead(d *schema.ResourceData, m interface{}) error {

	schemaSet := d.Get("source_dirs").(*schema.Set)
	sourceDirs := []string{}

	id := ""

	for _, dir := range schemaSet.List() {
		sourceDirs = append(sourceDirs, dir.(string))
		id = id + "|" + dir.(string)
	}

	hash, err := hashOfDirs(sourceDirs)
	if err != nil {
		return errors.Wrap(err, "while calculating hash of source dirs")
	}

	d.Set("hash", hex.EncodeToString(hash))

	d.SetId(id)

	return nil
}

func hashOfDirs(dirs []string) ([]byte, error) {

	dataFiles := map[string]struct{}{}

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.Mode().IsRegular() {
				dataFiles[path] = struct{}{}
			}
			return nil
		})

		if err != nil {
			return nil, errors.Wrapf(err, "while reading dir %s", dir)
		}
	}

	fileNames := []string{}
	for df := range dataFiles {
		fileNames = append(fileNames, df)
	}

	sort.Strings(fileNames)

	sha := sha3.New256()
	for _, df := range fileNames {
		f, err := os.Open(df)
		if err != nil {
			return nil, errors.Wrapf(err, "while opening file %s", df)
		}
		_, err = io.Copy(sha, f)

		if err != nil {
			f.Close()
			return nil, errors.Wrapf(err, "while reading file %s", df)
		}

		err = f.Close()
		if err != nil {
			return nil, errors.Wrapf(err, "while closing file %s", df)
		}
	}
	return sha.Sum(nil), nil
}

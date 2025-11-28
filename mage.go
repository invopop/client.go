//go:build mage

package main

import (
	"io/ioutil"
	"path"

	"github.com/magefile/mage/sh"
)

const (
	name     = "provider"
	runImage = "gcr.io/invopop/golang:1.22.3-alpine"
)

// Protocol takes the protobuf source and converts them into go code.
func Protocol() error {
	p := "./gateway"
	args := []string{
		"protoc",
		"--go_out=" + p,
		"--proto_path=" + p,
	}
	files, _ := ioutil.ReadDir(p)
	for _, file := range files {
		if path.Ext(file.Name()) == ".proto" {
			args = append(args, path.Join(p, file.Name()))
		}
	}
	return dockerRunCmd(name+"-proto", "", args...)
}

func dockerRunCmd(name, publicPort string, cmd ...string) error {
	args := []string{
		"run",
		"--rm",
		"--name", name,
		"--network", "invopop-local",
		"-v", "$PWD:/src",
		"-w", "/src",
		"-it", // interactive
	}
	if publicPort != "" {
		args = append(args,
			"--label", "traefik.enable=true",
			"--label", "traefik.http.routers."+name+".rule=Host(`"+name+".invopop.dev`)",
			"--label", "traefik.http.routers."+name+".tls=true",
			"--expose", publicPort,
		)
	}
	args = append(args, runImage)
	args = append(args, cmd...)
	return sh.RunV("docker", args...)
}

// Shell runs an interactive shell within a docker container.
func Shell() error {
	return dockerRunCmd(name+"-shell", "", "ash")
}

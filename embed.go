package main

import "embed"

//go:embed lib/*.joy
var embeddedLibs embed.FS

func readEmbeddedLib(name string) ([]byte, error) {
	return embeddedLibs.ReadFile("lib/" + name)
}

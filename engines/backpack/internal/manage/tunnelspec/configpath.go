package tunnelspec

import (
	"github.com/backpack/backpack/internal/app"
	"path/filepath"
)

// Private test seam; production paths remain fixed and cannot be changed by API input.
var configRoot = app.ConfigDir

func configPath(name string) string { return filepath.Join(configRoot, name+".toml") }

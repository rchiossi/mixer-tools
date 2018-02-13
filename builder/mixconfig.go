package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
)

type MixConfig struct {
	// [Builder]
	BundleDir  string
	Cert       string
	StateDir   string
	VersionDir string
	YumConf    string

	// [swupd]
	Bundle     string
	ContentURL string
	Format     string
	VersionURL string

	// [Server]
	DebugInfoBanned string
	DebugInfoLib    string
	DebugInfoSrc    string

	// [Mixer]
	LocalBundleDir string
	RepoDir        string
	RPMDir         string
}

func (config *MixConfig) LoadDefaults() error {
	pwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	// [Builder]
	config.BundleDir = filepath.Join(pwd, "mix-bundles")
	config.Cert = filepath.Join(pwd, "Swupd_Root.pem")
	config.StateDir = filepath.Join(pwd, "update")
	config.VersionDir = pwd
	config.YumConf = filepath.Join(pwd, ".yum-mix.conf")

	// [Swupd]
	config.Bundle = "os-core-update"
	config.ContentURL = "<URL where the content will be hosted>"
	config.Format = "1"
	config.VersionURL = "<URL where the version of the mix will be hosted>"

	// [Server]
	config.DebugInfoBanned = "true"
	config.DebugInfoLib = "/usr/lib/debug"
	config.DebugInfoSrc = "/usr/src/debug"

	// [Mixer]
	config.LocalBundleDir = filepath.Join(pwd, "local-bundles")
	config.RPMDir = ""
	config.RepoDir = ""

	return nil
}

func (config *MixConfig) mapToIni() (*ini.File, error) {
	cfg := ini.Empty()

	// [Builder]
	section, err := cfg.NewSection("Builder")
	if err != nil {
		return nil, err
	}
	if _, err = section.NewKey("BUNDLE_DIR", config.BundleDir); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("CERT", config.Cert); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("SERVER_STATE_DIR", config.StateDir); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("VERSIONS_PATH", config.VersionDir); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("YUM_CONF", config.YumConf); err != nil {
		return nil, err
	}

	// [Swupd]
	section, err = cfg.NewSection("swupd")
	if err != nil {
		return nil, err
	}
	if _, err = section.NewKey("BUNDLE", config.Bundle); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("CONTENTURL", config.ContentURL); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("FORMAT", config.Format); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("VERSIONURL", config.VersionURL); err != nil {
		return nil, err
	}

	// [Server]
	section, err = cfg.NewSection("Server")
	if err != nil {
		return nil, err
	}
	if _, err = section.NewKey("debuginfo_banned", config.DebugInfoBanned); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("debuginfo_lib", config.DebugInfoLib); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("debuginfo_src", config.DebugInfoSrc); err != nil {
		return nil, err
	}

	// [Mixer]
	section, err = cfg.NewSection("Mixer")
	if err != nil {
		return nil, err
	}
	if _, err = section.NewKey("LOCAL_BUNDLE_DIR", config.LocalBundleDir); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("LOCAL_REPO_DIR", config.RepoDir); err != nil {
		return nil, err
	}
	if _, err = section.NewKey("LOCAL_RPM_DIR", config.RPMDir); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (config *MixConfig) mapFromIni(cfg *ini.File) error {
	// [Builder]
	section, err := cfg.GetSection("Builder")
	if err == nil {
		key, err := section.GetKey("BUNDLE_DIR")
		if err == nil {
			config.BundleDir = key.String()
		}
		key, err = section.GetKey("CERT")
		if err == nil {
			config.Cert = key.String()
		}
		key, err = section.GetKey("SERVER_STATE_DIR")
		if err == nil {
			config.StateDir = key.String()
		}
		key, err = section.GetKey("VERSIONS_PATH")
		if err == nil {
			config.VersionDir = key.String()
		}
		key, err = section.GetKey("YUM_CONF")
		if err == nil {
			config.YumConf = key.String()
		}
	}

	// [swupd]
	section, err = cfg.GetSection("swupd")
	if err == nil {
		key, err := section.GetKey("BUNDLE")
		if err == nil {
			config.Bundle = key.String()
		}
		key, err = section.GetKey("CONTENTURL")
		if err == nil {
			config.ContentURL = key.String()
		}
		key, err = section.GetKey("FORMAT")
		if err == nil {
			config.Format = key.String()
		}
		key, err = section.GetKey("VERSIONURL")
		if err == nil {
			config.VersionURL = key.String()
		}
	}

	// [Server]
	section, err = cfg.GetSection("Server")
	if err == nil {
		key, err := section.GetKey("debuginfo_banned")
		if err == nil {
			config.DebugInfoBanned = key.String()
		}
		key, err = section.GetKey("debuginfo_lib")
		if err == nil {
			config.DebugInfoLib = key.String()
		}
		key, err = section.GetKey("debuginfo_src")
		if err == nil {
			config.DebugInfoSrc = key.String()
		}
	}

	// [Mixer]
	section, err = cfg.GetSection("Mixer")
	if err == nil {
		key, err := section.GetKey("LOCAL_BUNDLE_DIR")
		if err == nil {
			config.LocalBundleDir = key.String()
		}
		key, err = section.GetKey("LOCAL_REPO_DIR")
		if err == nil {
			config.RepoDir = key.String()
		}
		key, err = section.GetKey("LOCAL_RPM_DIR")
		if err == nil {
			config.RPMDir = key.String()
		}
	}

	return nil
}

// Replace environment variables in the config file with the actual value of that variable
func (config *MixConfig) expandEnvVars() error {
	return nil
}

// CreateDefaultConfig creates a default builder.conf using the active
// directory as base path for the variables values.
func (config *MixConfig) CreateDefaultConfig(localrpms bool) error {
	config.LoadDefaults()

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	builderconf := filepath.Join(pwd, "builder.conf")

	if localrpms {
		config.RPMDir = filepath.Join(pwd, "local-rpms")
		config.RepoDir = filepath.Join(pwd, "local-yum")
	}

	fmt.Println("Creating new builder.conf configuration file...")

	cfg, err := config.mapToIni()
	if err != nil {
		return err
	}

	return cfg.SaveTo(builderconf)
}

// LoadConfig loads a configuration file from a provided path or from local directory
// is none is provided
func (config *MixConfig) LoadConfig(filename string) error {
	cfg, err := ini.InsensitiveLoad(filename)
	if err != nil {
		return err
	}

	return config.mapFromIni(cfg)
}

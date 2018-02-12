// Copyright © 2018 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// loadDefaults set default values for the properties in builder.conf
func loadDefaults() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	viper.SetDefault("Builder.SERVER_STATE_DIR", filepath.Join(pwd, "update"))
	viper.SetDefault("Builder.BUNDLE_DIR", filepath.Join(pwd, "mix-bundles"))
	viper.SetDefault("Builder.YUM_CONF", filepath.Join(pwd, ".yum-mix.conf"))
	viper.SetDefault("Builder.CERT", filepath.Join(pwd, "Swupd_Root.pem"))
	viper.SetDefault("Builder.VERSIONS_PATH", pwd)

	viper.SetDefault("swupd.BUNDLE", "os-core-update")
	viper.SetDefault("swupd.CONTENTURL", "<URL where the content will be hosted>")
	viper.SetDefault("swupd.VERSIONURL", "<URL where the version of the mix will be hosted>")
	viper.SetDefault("swupd.FORMAT", "1")

	viper.SetDefault("Server.debuginfo_banned", "true")
	viper.SetDefault("Server.debuginfo_lib", "/usr/lib/debug")
	viper.SetDefault("Server.debuginfo_src", "/usr/src/debug")

	viper.SetDefault("Mixer.LOCAL_BUNDLE_DIR", filepath.Join(pwd, "local-bundles"))
	viper.SetDefault("Mixer.LOCAL_RPM_DIR", "")
	viper.SetDefault("Mixer.LOCAL_REPO_DIR", "")

	viper.SetConfigName("builder")
	viper.AddConfigPath(pwd)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	return nil
}

var defaultsLoaded = false

// checkDefaults Check and initialize defaults if neede
func checkDefaults() {
	if !defaultsLoaded {
		loadDefaults()
	}
}

// CreateDefaultConfig creates a default builder.conf using the active
// directory as base path for the variables values.
func CreateDefaultConfig(localrpms bool) error {
	checkDefaults()

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if localrpms {
		viper.Set("Mixer.RPMDIR", filepath.Join(pwd, "rpms"))
		viper.Set("Mixer.REPODIR", filepath.Join(pwd, "local"))
	}

	fmt.Println("Creating new builder.conf configuration file...")

	if err = viper.WriteConfigAs(filepath.Join(pwd, "builder.toml")); err != nil {
		return err
	}

	// For WriteConfig, Viper defines config type based on file extension, so we need
	// to rename the file after creation
	return os.Rename(filepath.Join(pwd, "builder.toml"), filepath.Join(pwd, "builder.conf"))
}

// LoadBuilderConf will read the builder configuration from the command line if
// it was provided, otherwise it will fall back to reading the configuration from
// the local builder.conf file.
func LoadBuilderConf(builderconf string) (string, error) {
	checkDefaults()

	var config string
	// If builderconf is set via cmd line, use that one
	if len(builderconf) > 0 {
		config = builderconf
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		config = filepath.Join(pwd, "builder.conf")
	}

	reader, err := os.Open(config)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = reader.Close()
	}()

	viper.SetConfigType("toml")

	if err := viper.ReadConfig(reader); err != nil {
		return "", err
	}

	return config, nil
}

func GetStringProperty(property string) string {
	checkDefaults()

	return viper.GetString(property)
}

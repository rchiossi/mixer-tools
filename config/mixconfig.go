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

package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// MixConfig represents the config parameters found in the builder config file.
type MixConfig struct {
	Builder builderConf
	Swupd   swupdConf
	Server  serverConf
	Mixer   mixerConf

	/* hidden properties */
	filename string
	version  string

	// Format moved into mixer.state file. This variable is set
	// if the value is still present in the parsed config to
	// print a warning for the user.
	hasFormatField bool
}

type builderConf struct {
	Cert           string `required:"true" mount:"true" toml:"CERT"`
	ServerStateDir string `required:"true" mount:"true" toml:"SERVER_STATE_DIR"`
	VersionPath    string `required:"true" mount:"true" toml:"VERSIONS_PATH"`
	DNFConf        string `required:"true" mount:"true" toml:"YUM_CONF"`
}

type swupdConf struct {
	Bundle     string `required:"false" toml:"BUNDLE"`
	ContentURL string `required:"false" toml:"CONTENTURL"`
	VersionURL string `required:"false" toml:"VERSIONURL"`
}

type serverConf struct {
	DebugInfoBanned string `required:"false" toml:"DEBUG_INFO_BANNED"`
	DebugInfoLib    string `required:"false" toml:"DEBUG_INFO_LIB"`
	DebugInfoSrc    string `required:"false" toml:"DEBUG_INFO_SRC"`
}

type mixerConf struct {
	LocalBundleDir string `required:"false" mount:"true" toml:"LOCAL_BUNDLE_DIR"`
	LocalRepoDir   string `required:"false" mount:"true" toml:"LOCAL_REPO_DIR"`
	LocalRPMDir    string `required:"false" mount:"true" toml:"LOCAL_RPM_DIR"`
	DockerImgPath  string `required:"false" toml:"DOCKER_IMAGE_PATH"`
}

// LoadDefaults sets sane values for the config properties
func (config *MixConfig) LoadDefaults() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	config.LoadDefaultsForPath(pwd)
	return nil
}

// LoadDefaultsForPath sets sane values for config properties using `path` as base directory
func (config *MixConfig) LoadDefaultsForPath(path string) {

	// [Builder]
	config.Builder.Cert = filepath.Join(path, "Swupd_Root.pem")
	config.Builder.ServerStateDir = filepath.Join(path, "update")
	config.Builder.VersionPath = path
	config.Builder.DNFConf = filepath.Join(path, ".yum-mix.conf")

	// [Swupd]
	config.Swupd.Bundle = "os-core-update"
	config.Swupd.ContentURL = "<URL where the content will be hosted>"
	config.Swupd.VersionURL = "<URL where the version of the mix will be hosted>"

	// [Server]
	config.Server.DebugInfoBanned = "true"
	config.Server.DebugInfoLib = "/usr/lib/debug"
	config.Server.DebugInfoSrc = "/usr/src/debug"

	// [Mixer]
	config.Mixer.LocalBundleDir = filepath.Join(path, "local-bundles")
	config.Mixer.DockerImgPath = "clearlinux/mixer"

	config.Mixer.LocalRPMDir = filepath.Join(path, "local-rpms")
	config.Mixer.LocalRepoDir = filepath.Join(path, "local-yum")

	config.version = CurrentConfigVersion
	config.filename = filepath.Join(path, "builder.conf")

	config.hasFormatField = false
}

// CreateDefaultConfig creates a default builder.conf using the active
// directory as base path for the variables values.
func (config *MixConfig) CreateDefaultConfig() error {
	if err := config.LoadDefaults(); err != nil {
		return err
	}

	err := config.InitConfigPath("")
	if err != nil {
		return err
	}

	return config.Save()
}

// Save saves the properties in MixConfig to a TOML config file
func (config *MixConfig) Save() error {
	var buffer bytes.Buffer
	buffer.Write([]byte("#VERSION " + config.version + "\n\n"))

	enc := toml.NewEncoder(&buffer)

	if err := enc.Encode(config); err != nil {
		return err
	}

	w, err := os.OpenFile(config.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = w.Close()
	}()

	_, err = buffer.WriteTo(w)

	return err
}

// Load loads a configuration file from a provided path or from local directory
// is none is provided
func (config *MixConfig) Load(filename string) error {
	if err := config.InitConfigPath(filename); err != nil {
		return err
	}
	if ok, err := ParseVersion(config); err != nil {
		return err
	} else if !ok {
		if err = config.convert(); err != nil {
			return err
		}
	}
	if err := config.parse(); err != nil {
		return err
	}
	if err := config.expandEnv(); err != nil {
		return err
	}

	return config.validate()
}

func (config *MixConfig) parse() error {
	_, err := toml.DecodeFile(config.filename, &config)
	return err
}

func (config *MixConfig) expandEnv() error {
	re := regexp.MustCompile(`\$\{?([[:word:]]+)\}?`)
	rv := reflect.ValueOf(config).Elem()

	for i := 0; i < rv.NumField(); i++ {
		sectionV := rv.Field(i)

		/* ignore unexported fields */
		if !sectionV.CanSet() {
			continue
		}

		for j := 0; j < sectionV.NumField(); j++ {
			val := sectionV.Field(j).String()
			matches := re.FindAllStringSubmatch(val, -1)

			for _, s := range matches {
				if _, ok := os.LookupEnv(s[1]); !ok {
					return errors.Errorf("buildconf contains an undefined environment variable: %s\n", s[1])
				}
			}

			sectionV.Field(j).SetString(os.ExpandEnv(val))
		}

	}

	return nil
}

func (config *MixConfig) validate() error {
	rv := reflect.ValueOf(config).Elem()

	for i := 0; i < rv.NumField(); i++ {
		sectionV := rv.Field(i)
		/* ignore unexported fields */
		if !sectionV.CanSet() {
			continue
		}

		sectionT := reflect.TypeOf(rv.Field(i).Interface())

		for j := 0; j < sectionT.NumField(); j++ {
			tag, ok := sectionT.Field(j).Tag.Lookup("required")

			if ok && tag == "true" && sectionV.Field(j).String() == "" {
				name, ok := sectionT.Field(j).Tag.Lookup("toml")
				if !ok || name == "" {
					// Default back to variable name if no TOML tag is defined
					name = sectionT.Field(j).Name
				}

				return errors.Errorf("Missing required field in config file: %s", name)
			}
		}
	}

	if config.hasFormatField {
		log.Println("Warning: FORMAT value was transferred to mixer.state file")
	}

	return nil
}

// Convert parses an old config file and converts it to TOML format
func (config *MixConfig) Convert(filename string) error {
	if err := config.InitConfigPath(filename); err != nil {
		return err
	}

	if err := config.LoadDefaults(); err != nil {
		return err
	}

	if ok, err := ParseVersion(config); err != nil {
		return err
	} else if ok {
		// Already on latest version
		return nil
	}

	return config.convert()
}

// Print print variables and values of a MixConfig struct
func (config *MixConfig) Print() error {
	sb := bytes.NewBufferString("")

	enc := toml.NewEncoder(sb)
	if err := enc.Encode(config); err != nil {
		return err
	}

	fmt.Println(sb.String())

	return nil
}

// InitConfigPath sets the main config name to what was passed in,
// or defaults to the current working directory + builder.conf
func (config *MixConfig) InitConfigPath(fullpath string) error {
	if fullpath != "" {
		config.filename = fullpath
		return nil
	}
	// Create a builder.conf in the current directory if none is passed in
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	config.filename = filepath.Join(pwd, "builder.conf")

	return nil
}

// SetFilename receives a filename and sets it as the config file. It is used for
// loading and saving the file.
func (config *MixConfig) SetFilename(filename string) {
	config.filename = filename
}

// GetFilename returns the file name of current config
func (config *MixConfig) GetFilename() string {
	/* This variable cannot be public or else it will be added to the config file */
	return config.filename
}

//SetVersion sets the version number for the config
func (config *MixConfig) SetVersion(version string) {
	config.version = version
}

//GetVersion returns the current version of the config
func (config *MixConfig) GetVersion() string {
	return config.version
}

//GetLatestVersion returts the latest version know to mixer for this config file
func (config *MixConfig) GetLatestVersion() string {
	return CurrentConfigVersion
}

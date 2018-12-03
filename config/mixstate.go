// Copyright 2018 Intel Corporation
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
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
)

type mixSection struct {
	Format  string `toml:"FORMAT"`
	Offline string `toml:"OFFLINE"`
}

// MixState holds the current state of the mix
type MixState struct {
	Mix mixSection

	/* hidden properties */
	filename string
	version  string

	/* Inform the user where the source of mix format*/
	formatSource string
}

// CurrentStateVersion is the current revision for the state file structure
var CurrentStateVersion = "1.1"

// DefaultFormatPath is the default path for the format file specified by swupd
const DefaultFormatPath = "/usr/share/defaults/swupd/format"

// LoadDefaults initialize the state object with sane values
func (state *MixState) LoadDefaults() {
	state.loadDefaultFormat()

	state.Mix.Offline = "false"

	state.filename = "mixer.state"
	state.version = CurrentStateVersion
}

func (state *MixState) loadDefaultFormat() {
	/* Get format from legacy config file */
	format, err := state.getFormatFromConfig()
	if err == nil && format != "" {
		state.Mix.Format = format
		state.formatSource = "builder.conf"
		return
	}

	/* Get format from system */
	formatBytes, err := ioutil.ReadFile(DefaultFormatPath)
	if err == nil {
		state.Mix.Format = string(formatBytes)
		state.formatSource = DefaultFormatPath
		return
	}

	state.Mix.Format = "1"
	state.formatSource = "Mixer internal value"
}

func (state *MixState) getFormatFromConfig() (string, error) {
	confBytes, err := ioutil.ReadFile("builder.conf")
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(`FORMAT[\s"=]*([0-9]+)[\s"]*\n`)
	match := r.FindStringSubmatch(string(confBytes))
	if len(match) == 2 {
		return match[1], nil
	}

	return "", nil
}

// Save creates or overwrites the mixer.state file
func (state *MixState) Save() error {
	var buffer bytes.Buffer
	buffer.Write([]byte("#VERSION " + state.version + "\n\n"))

	enc := toml.NewEncoder(&buffer)

	if err := enc.Encode(state); err != nil {
		return err
	}

	w, err := os.OpenFile(state.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = w.Close()
	}()

	_, err = buffer.WriteTo(w)

	return err
}

// Load the mixer.state file
func (state *MixState) Load(filename string) error {
	state.LoadDefaults()

	if filename != "" {
		state.filename = filename
	}

	f, err := os.Open(state.filename)
	if err != nil {
		// If state does not exists, create a default state
		log.Println("Warning: Using FORMAT value from " + state.formatSource)
		return state.Save()
	}
	defer func() {
		_ = f.Close()
	}()

	var ok bool
	ok, err = ParseVersion(state)
	if err != nil {
		return err
	} else if !ok {
		fmt.Printf("Converting state to version %s\n", CurrentStateVersion)
		state.version = CurrentStateVersion
		return state.Save()
	}

	_, err = toml.DecodeFile(state.filename, &state)
	return err
}

// SetFilename receives a filename and sets it as the state file. It is used for
// loading and saving the file.
func (state *MixState) SetFilename(filename string) {
	state.filename = filename
}

// GetFilename returns the file name of current state
func (state *MixState) GetFilename() string {
	/* This variable cannot be public or else it will be added to the state file */
	return state.filename
}

//SetVersion sets the version number for the state
func (state *MixState) SetVersion(version string) {
	state.version = version
}

//GetVersion returns the current version of the state
func (state *MixState) GetVersion() string {
	return state.version
}

//GetLatestVersion returts the latest version know to mixer for this state file
func (state *MixState) GetLatestVersion() string {
	return CurrentStateVersion
}

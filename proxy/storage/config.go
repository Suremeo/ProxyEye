/*
 *
 *           ____                        ______
 *          / __ \_________  _  ____  __/ ____/_  _____
 *         / /_/ / ___/ __ \| |/_/ / / / __/ / / / / _ \
 *        / ____/ /  / /_/ />  </ /_/ / /___/ /_/ /  __/
 *       /_/   /_/   \____/_/|_|\__, /_____/\__, /\___/
 *                                /_/         /_/
 *       ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
 *
 *                     Author: Suremeo (github.com/Suremeo)
 *
 *
 */

package storage

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/Suremeo/ProxyEye/proxy/console"
	"github.com/Suremeo/ProxyEye/proxy/storage/structures"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

var Config *structures.Config

func init() {
	b, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		console.Info("Config doesn't exist, creating it")
		c := defaultConfig()
		createConfig(c)
		Config = c
	} else {
		c, err := parseConfig(b)
		if err != nil {
			console.Panic("Invalid configuration file", err)
		}
		Config = c
	}
	setLoggingLevel(Config)
}

func setLoggingLevel(config *structures.Config) {
	if strings.ToUpper(config.Logging.Level) == "DEVELOPMENT" {
		console.SetLevel(true)
	} else {
		console.SetLevel(false)
	}
}

func parseConfig(b []byte) (*structures.Config, error) {
	c := &structures.Config{}
	return c, yaml.Unmarshal(b, c)
}

func createConfig(config *structures.Config) {
	setLoggingLevel(config)
	d, err := yaml.Marshal(config)
	if err != nil {
		console.Warn("Config failed to marshal while creating...: %v", err.Error())
	} else {
		err = ioutil.WriteFile("./config.yml", d, os.ModePerm)
		if err != nil {
			console.Warn("Failed to write the config to config.yml: %v", err.Error())
		}
	}
}

func defaultConfig() *structures.Config {
	return &structures.Config{
		Listener: &structures.ConfigListener{
			Address: "0.0.0.0:19132",
			Auth:    true,
			Motd:    "<bold><dark-purple>Proxy</dark-purple><purple>Eye</purple></bold>",
			Secret: func() string { // Generate a cryptographically secure secret for the configuration.
				b := make([]byte, 32)
				_, err := rand.Read(b)
				console.Fatal("Error generating secret", err)
				return hex.EncodeToString(b)
			}(),
		},
		Logging: &structures.ConfigLogging{
			Level: "PRODUCTION",
		},
	}
}

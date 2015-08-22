/*
 * Minio Client (C) 2014, 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/minio/mc/internal/github.com/minio/cli"
	"github.com/minio/mc/internal/github.com/minio/minio/pkg/probe"
	"github.com/minio/mc/internal/github.com/minio/minio/pkg/quick"
)

//   Configure minio client
//
//   ----
//   NOTE: that the configure command only writes values to the config file.
//   It does not use any configuration values from the environment variables.
//
//   One needs to edit configuration file manually, this is purposefully done
//   so to avoid taking credentials over cli arguments. It is a security precaution
//   ----
//
var configCmd = cli.Command{
	Name:   "config",
	Usage:  "Modify, add alias, oauth into default configuration file [~/.mc/config.json]",
	Action: mainConfig,
	CustomHelpTemplate: `NAME:
   mc {{.Name}} - {{.Usage}}

USAGE:
   mc {{.Name}}{{if .Flags}} [ARGS...]{{end}} alias NAME HOSTURL

EXAMPLES:
   1. Add alias URLs.
      $ mc config alias zek https://s3.amazonaws.com/

`,
}

// mainConfig is the handle for "mc config" sub-command. writes configuration data in json format to config file.
func mainConfig(ctx *cli.Context) {
	// show help if nothing is set
	if !ctx.Args().Present() || ctx.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(ctx, "config", 1) // last argument is exit code
	}
	if strings.TrimSpace(ctx.Args().First()) == "" {
		cli.ShowCommandHelpAndExit(ctx, "config", 1) // last argument is exit code
	}
	arg := ctx.Args().First()
	tailArgs := ctx.Args().Tail()
	if len(tailArgs) > 2 {
		fatalIf(probe.NewError(errors.New("")), "Incorrect number of arguments to config command. Please read ‘mc config help’")
	}

	switch arg {
	case "alias":
		if len(tailArgs) < 2 {
			cli.ShowCommandHelpAndExit(ctx, "config", 1) // last argument is exit code
		}
		alias := tailArgs[0]
		url := tailArgs[1]
		addAlias(alias, url)
	default:
		cli.ShowCommandHelpAndExit(ctx, "config", 1) // last argument is exit code
	}
}

// addAlias - add new aliases
func addAlias(alias, url string) {
	if alias == "" || url == "" {
		fatalIf(probe.NewError(errors.New("")), "Alias or URL cannot be empty.")
	}
	conf := newConfigV1()
	config, err := quick.New(conf)
	fatalIf(err.Trace(conf.Version), "Failed to initialize ‘quick’ configuration data structure.")

	err = config.Load(mustGetMcConfigPath())
	fatalIf(err.Trace(), "Unable to load config path")

	url = strings.TrimSuffix(url, "/")
	if !strings.HasPrefix(url, "http") {
		fatalIf(probe.NewError(errors.New("")), fmt.Sprintf("Invalid alias URL ‘%s’. Valid examples are: http://s3.amazonaws.com, https://yourbucket.example.com.", url))
	}
	if isAliasReserved(alias) {
		fatalIf(probe.NewError(errors.New("")), fmt.Sprintf("Cannot use a reserved name ‘%s’ as an alias. Following are reserved names: [help, private, readonly, public, authenticated].", alias))
	}
	if !isValidAliasName(alias) {
		fatalIf(probe.NewError(errors.New("")), fmt.Sprintf("Alias name ‘%s’ is invalid, valid examples are: mybucket, Area51, Grand-Nagus", alias))
	}
	// convert interface{} back to its original struct
	newConf := config.Data().(*configV1)
	if oldURL, ok := newConf.Aliases[alias]; ok {
		fatalIf(probe.NewError(errors.New("")), fmt.Sprintf("Alias ‘%s’ already exists for ‘%s’.", alias, oldURL))
	}
	newConf.Aliases[alias] = url
	newConfig, err := quick.New(newConf)
	fatalIf(err.Trace(conf.Version), "Failed to initialize ‘quick’ configuration data structure.")

	err = writeConfig(newConfig)
	fatalIf(err.Trace(alias, url), "Unable to save alias ‘"+alias+"’.")
}

/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cliFlags

import (
	cliV2 "github.com/urfave/cli/v2"
)

func newConfigFlag() cliV2.Flag {
	return &(cliV2.StringFlag{
		Name:     Config,
		Usage:    "set config file name",
		Aliases:  []string{"c"},
		Hidden:   false,
		Required: true,
	})
}

func newPasswordFlag() cliV2.Flag {
	return &(cliV2.StringFlag{
		Name:     Password,
		Usage:    "set password",
		Aliases:  []string{"p"},
		Hidden:   false,
		Required: false,
	})
}

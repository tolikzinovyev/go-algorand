// Copyright (C) 2019-2021 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package fuzzer

import (
	"encoding/json"
	"io/ioutil"
	"os"
	//ossignal "os/signal"
	"path/filepath"
	//"runtime/pprof"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/algorand/go-algorand/logging"
)

type FuzzerTestFile struct {
	FuzzerName string
	NodesCount int
	Filters    []interface{}
	Validator  ValidatorConfig
	LogLevel   int
}

func TestFuzzer(t *testing.T) {
	jsonFiles := make(map[string]string) // map json test to full json file name.
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".json") {
			jsonFiles[info.Name()] = path
		}
		return nil
	})
	require.NoError(t, err)
	for testName := range jsonFiles {
		t.Run(testName, func(t *testing.T) {
			jsonFilename := jsonFiles[testName]
			jsonBytes, err := ioutil.ReadFile(jsonFilename)
			require.NoError(t, err)
			var fuzzerTest FuzzerTestFile
			err = json.Unmarshal(jsonBytes, &fuzzerTest)
			if err != nil {
				t.Skip()
			}

			filters := []NetworkFilterFactory{}
			// generate a list of concrete filters
			for _, fuzzerFilterData := range fuzzerTest.Filters {
				// convert the interface into a byte-stream.
				filterConfig, err := json.Marshal(fuzzerFilterData)
				require.NoError(t, err)
				var filterFactory NetworkFilterFactory
				for _, regFactory := range registeredFilterFactories {
					filterFactory = regFactory.Unmarshal(filterConfig)
					if filterFactory != nil {
						// we found a filter factory!
						break
					}
				}
				if filterFactory == nil {
					t.Skip()
				}
				filters = append(filters, filterFactory)
			}
			config := &FuzzerConfig{
				FuzzerName: fuzzerTest.FuzzerName,
				NodesCount: fuzzerTest.NodesCount,
				Filters:    filters,
				LogLevel:   logging.Level(fuzzerTest.LogLevel),
			}

			validator := MakeValidator(&fuzzerTest.Validator, t)
			validator.Go(config)
		})
	}
}

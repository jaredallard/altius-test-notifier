// Copyright 2020 Jared Allard
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
package altius

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type TestResult string

const (
	Processing TestResult = "processing"

	// Changed denotes a test isn't processing.
	// This is only a valid state if all other
	// states aren't implemented or if it's unknown
	Changed TestResult = "changed"

	// Not implemented
	Positive TestResult = "positive"
	Negative TestResult = "negative"
)

// GetTestResult returns a enum of a test's status
func GetTestResult(retrievalCode, dateOfBirth string) (TestResult, error) {
	retrievalCode = strings.ReplaceAll(retrievalCode, "-", "")

	resp, err := http.Get(
		fmt.Sprintf("https://covid19.altius.org/nabu/test-report/%s/%s", retrievalCode, dateOfBirth),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read body")
	}

	status := string(b)
	switch status {
	case "Test processing.":
		return Processing, nil
	case "Cannot read property 'email' of undefined":
		return "", fmt.Errorf("invalid retrieval code or date of birth")
	}

	return Changed, nil
}

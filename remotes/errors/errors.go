/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/containerd/log"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
)

var _ error = ErrUnexpectedStatus{}

// ErrUnexpectedStatus is returned if a registry API request returned with unexpected HTTP status
type ErrUnexpectedStatus struct {
	Status                    string
	StatusCode                int
	Body                      []byte
	RequestURL, RequestMethod string
	Message                   string
}

func (e ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status(%s) from %s request to %s: %s", e.Status, e.RequestMethod, e.RequestURL, e.Message)
}

// NewUnexpectedStatusErr creates an ErrUnexpectedStatus from HTTP response
func NewUnexpectedStatusErr(resp *http.Response) error {
	var b []byte
	var message string
	if resp.Body != nil {
		b, _ = io.ReadAll(io.LimitReader(resp.Body, 64000)) // 64KB
		if len(b) > 0 {
			var r dtspecv1.ErrorResponse
			err := json.Unmarshal(b, &r)
			if err != nil {
				log.G(context.Background()).Debugf("Unmarshal response body failed: %v", err)
			}
			for _, e := range r.Errors {
				if len(e.Message) > 0 {
					message = e.Message
				}
			}
		}
	}
	err := ErrUnexpectedStatus{
		Body:          b,
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		RequestMethod: resp.Request.Method,
		Message:       message,
	}
	if resp.Request.URL != nil {
		err.RequestURL = resp.Request.URL.String()
	}
	return err
}

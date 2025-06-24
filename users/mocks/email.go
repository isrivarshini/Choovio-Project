// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/users"
)

type emailerMock struct {
}

// NewEmailer provides emailer instance for  the test
func NewEmailer() users.Emailer {
	return &emailerMock{}
}

func (e *emailerMock) SendPasswordReset([]string, string, string) errors.Error {
	return nil
}

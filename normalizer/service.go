// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package normalizer

import "github.com/mainflux/mainflux"

// Service specifies API for normalizing messages.
type Service interface {
	// Normalizes raw message to array of standard SenML messages.
	Normalize(mainflux.RawMessage) ([]mainflux.Message, error)
}

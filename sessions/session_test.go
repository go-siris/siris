// Copyright 2017 Go-SIRIS Author. All Rights Reserved.

package sessions

import (
	"testing"

	"github.com/go-siris/siris/sessions/utils"
)

func TestSession_RandomCreateBytes(t *testing.T) {
	_ = utils.RandomCreateBytes(32)
	_ = utils.RandomCreateBytes(64)
	_ = utils.RandomCreateBytes(128)
	_ = utils.RandomCreateBytes(256)
	_ = utils.RandomCreateBytes(512)
	_ = utils.RandomCreateBytes(1024)
}

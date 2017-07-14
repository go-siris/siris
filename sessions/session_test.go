// Copyright 2017 Go-SIRIS Author. All Rights Reserved.

package sessions

import (
	"testing"
)

func TestSession_RandomCreateBytes(t *testing.T) {
	_ = RandomCreateBytes(32)
	_ = RandomCreateBytes(64)
	_ = RandomCreateBytes(128)
	_ = RandomCreateBytes(256)
	_ = RandomCreateBytes(512)
	_ = RandomCreateBytes(1024)
}

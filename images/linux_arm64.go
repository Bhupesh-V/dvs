//go:build arm64
// +build arm64

package images

import _ "embed"

//go:embed busybox_arm64.tar
var Busybox []byte

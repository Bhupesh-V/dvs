//go:build amd64
// +build amd64

package images

import _ "embed"

//go:embed busybox_amd64.tar
var Busybox []byte

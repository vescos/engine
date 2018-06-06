//TODO: incomplete
package decoders

import (
	"github.com/vescos/assets"
)

type Decoder interface {
	Decode(assets.Asset) []int16
}

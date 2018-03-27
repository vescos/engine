//TODO: incomplete
package decoders

import (
	"graphs/assets"
)

type Decoder interface {
	Decode(assets.Asset) []int16
}

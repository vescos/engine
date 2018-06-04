# LabyrinthEngine

Experimental 2D/3D engine in Go(Golang) and C.

### Supported platforms:
Linux(X11) - amd64.

Android - arm(armeabi, armeabi-v7a), arm64(armv8a), 386(x86), amd64(x86_64).

### Third party software in this library
Some parts of glue and input are derived from [gomobile project](https://github.com/golang/mobile)

gles2 package is from [gl package](https://github.com/goxjs/gl)

Ogg Vorbis decoder [stb_vorbis](http://nothings.org/stb_vorbis/) (Open Domain license)

### Dependencies 
Standard Golang library.

### Usage

```go
import (
	"graphs/engine/glue"
	"graphs/engine/input/keys"
	"graphs/engine/input/size"
	"graphs/engine/input/touch"
)

type State struct {
	*glue.Glue
	// App flags and props
	// ...
}

```

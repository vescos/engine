// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package size defines an event for the dimensions, physical resolution and
// orientation of the app's window.
//
// See the golang.org/x/mobile/app package for details on the event model.
package size 

import (
//"image"
)

// Event holds the dimensions, physical resolution and orientation of the app's
// window.
type Event struct {
	// WidthPx and HeightPx are the window's dimensions in pixels.
	WidthPx, HeightPx int

	// WidthPt and HeightPt are the window's dimensions in points (1/72 of an
	// inch).
	WidthPt, HeightPt float32

	// PixelsPerPt is the window's physical resolution. It is the number of
	// pixels in a single geom.Pt, from the golang.org/x/mobile/geom package.
	//
	// There are a wide variety of pixel densities in existing phones and
	// tablets, so apps should be written to expect various non-integer
	// PixelsPerPt values. In general, work in geom.Pt.
	PixelsPerPt float32

	// Orientation is the orientation of the device screen.
	Orientation Orientation
}

// Bounds returns the window's bounds in pixels, at the time this size event
// was sent.
//
// The top-left pixel is always (0, 0). The bottom-right pixel is given by the
// width and height.
//func (e *Event) Bounds() image.Rectangle {
//return image.Rectangle{Max: image.Point{e.WidthPx, e.HeightPx}}
//}

// Orientation is the orientation of the device screen.
type Orientation int

const (
	// OrientationUnknown means device orientation cannot be determined.
	//
	// Equivalent on Android to Configuration.ORIENTATION_UNKNOWN
	// and on iOS to:
	//	UIDeviceOrientationUnknown
	//	UIDeviceOrientationFaceUp
	//	UIDeviceOrientationFaceDown
	OrientationUnknown Orientation = iota

	// OrientationPortrait is a device oriented so it is tall and thin.
	//
	// Equivalent on Android to Configuration.ORIENTATION_PORTRAIT
	// and on iOS to:
	//	UIDeviceOrientationPortrait
	//	UIDeviceOrientationPortraitUpsideDown
	OrientationPortrait

	// OrientationLandscape is a device oriented so it is short and wide.
	//
	// Equivalent on Android to Configuration.ORIENTATION_LANDSCAPE
	// and on iOS to:
	//	UIDeviceOrientationLandscapeLeft
	//	UIDeviceOrientationLandscapeRight
	OrientationLandscape
)

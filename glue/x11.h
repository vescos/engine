// +build linux,!android

#include <EGL/egl.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <X11/Xlib.h>
#include <X11/extensions/Xrandr.h>

typedef struct {
	EGLContext	eglContext;
	EGLConfig	eglConfig;
	EGLSurface	eglSurface;
	EGLDisplay	eglDisplay;
	Display		*xDisplay;
	Atom wmDeleteWindow;
} cRefs;

void * cRefsPtr();
short createWindow(int w, int h, cRefs * crefs);

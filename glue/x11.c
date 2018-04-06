// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modified from golang.org/x/mobile

// +build linux,!android

#include "_cgo_export.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <X11/Xlib.h>
#include <X11/extensions/Xrandr.h>

static Window new_window(int w, int h, Display *xDpy, EGLDisplay eDpy, EGLSurface *eSurf, EGLContext *ctx) {
	static const EGLint attribs[] = {
		EGL_RENDERABLE_TYPE, EGL_OPENGL_ES2_BIT,
		EGL_SURFACE_TYPE, EGL_WINDOW_BIT,
		EGL_BLUE_SIZE, 8,
		EGL_GREEN_SIZE, 8,
		EGL_RED_SIZE, 8,
		EGL_DEPTH_SIZE, 24,
		EGL_CONFIG_CAVEAT, EGL_NONE,
		EGL_NONE
	};
	EGLConfig config;
	EGLint num_configs;
	if (!eglChooseConfig(eDpy, attribs, &config, 1, &num_configs)) {
		fprintf(stderr, "eglChooseConfig failed\n");
		exit(1);
	}
	EGLint vid;
	if (!eglGetConfigAttrib(eDpy, config, EGL_NATIVE_VISUAL_ID, &vid)) {
		fprintf(stderr, "eglGetConfigAttrib failed\n");
		exit(1);
	}

	XVisualInfo visTemplate;
	visTemplate.visualid = vid;
	int num_visuals;
	XVisualInfo *visInfo = XGetVisualInfo(xDpy, VisualIDMask, &visTemplate, &num_visuals);
	if (!visInfo) {
		fprintf(stderr, "XGetVisualInfo failed\n");
		exit(1);
	}

	Window root = RootWindow(xDpy, DefaultScreen(xDpy));
	XSetWindowAttributes attr;

	attr.colormap = XCreateColormap(xDpy, root, visInfo->visual, AllocNone);
	if (!attr.colormap) {
		fprintf(stderr, "XCreateColormap failed\n");
		exit(1);
	}

	attr.event_mask = StructureNotifyMask | ExposureMask |
		ButtonPressMask | ButtonReleaseMask | ButtonMotionMask |
        KeyPressMask | KeyReleaseMask;
	Window win = XCreateWindow(
		xDpy, root, 0, 0, w, h, 0, visInfo->depth, InputOutput,
		visInfo->visual, CWColormap | CWEventMask, &attr);
	XFree(visInfo);

	XSizeHints sizehints;
	sizehints.width  = w;
	sizehints.height = h;
	sizehints.flags = USSize;
	XSetNormalHints(xDpy, win, &sizehints);
	XSetStandardProperties(xDpy, win, "App", "App", None, (char **)NULL, 0, &sizehints);

	static const EGLint ctx_attribs[] = {
		EGL_CONTEXT_CLIENT_VERSION, 2,
		EGL_NONE
	};
	*ctx = eglCreateContext(eDpy, config, EGL_NO_CONTEXT, ctx_attribs);
	if (!*ctx) {
		fprintf(stderr, "eglCreateContext failed\n");
		exit(1);
	}
	*eSurf = eglCreateWindowSurface(eDpy, config, win, NULL);
	if (!*eSurf) {
		fprintf(stderr, "eglCreateWindowSurface failed\n");
		exit(1);
	}
	return win;
}

short createWindow(int w, int h, Display *xDpy, EGLDisplay eDpy, EGLSurface eSurf) {
	Window win;
	EGLContext eCtx;
	Atom wm_delete_window;

	XInitThreads();
	xDpy = XOpenDisplay(NULL);
	if (!xDpy) {
		fprintf(stderr, "XOpenDisplay failed\n");
		exit(1);
	}
	eDpy = eglGetDisplay(xDpy);
	if (!eDpy) {
		fprintf(stderr, "eglGetDisplay failed\n");
		exit(1);
	}
	EGLint e_major, e_minor;
	if (!eglInitialize(eDpy, &e_major, &e_minor)) {
		fprintf(stderr, "eglInitialize failed\n");
		exit(1);
	}
	eglBindAPI(EGL_OPENGL_ES_API);
	win = new_window(w, h, xDpy, eDpy, &eSurf, &eCtx);

	wm_delete_window = XInternAtom(xDpy, "WM_DELETE_WINDOW", True);
	if (wm_delete_window != None) {
		XSetWMProtocols(xDpy, win, &wm_delete_window, 1);
	}

	XMapWindow(xDpy, win);
	if (!eglMakeCurrent(eDpy, eSurf, eSurf, eCtx)) {
		fprintf(stderr, "eglMakeCurrent failed\n");
		exit(1);
	}

	// Window size and DPI should be initialized before starting app.
	XEvent ev;
	while (1) {
		if (XCheckMaskEvent(xDpy, StructureNotifyMask, &ev) == False) {
			sleep(1);
			continue;
		}
		if (ev.type == ConfigureNotify) {
			break;
		}
	}

	XRRScreenConfiguration *xrr_conf = XRRGetScreenInfo(xDpy, win);
	return XRRConfigCurrentRate(xrr_conf);
}

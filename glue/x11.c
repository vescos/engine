// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modified from golang.org/x/mobile

// +build linux,!android

#include "x11.h"

void * cRefsPtr() {
	cRefs *p;
	p = (cRefs *) malloc(sizeof(cRefs));
	if (p != NULL) {
		p->eglDisplay = NULL;
		p->eglConfig = NULL;
		p->eglContext = NULL;
		p->eglSurface = NULL;
		p->xDisplay = NULL;
		p->wmDeleteWindow = 0;
		return (void *) p;
	}
	return NULL;
} 

static Window newWindow(int w, int h, cRefs *crefs) {
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
	EGLint num_configs;
	if (!eglChooseConfig(crefs->eglDisplay, attribs, &crefs->eglConfig, 1, &num_configs)) {
		fprintf(stderr, "eglChooseConfig failed\n");
		exit(1);
	}
	EGLint vid;
	if (!eglGetConfigAttrib(crefs->eglDisplay, crefs->eglConfig, EGL_NATIVE_VISUAL_ID, &vid)) {
		fprintf(stderr, "eglGetConfigAttrib failed\n");
		exit(1);
	}

	XVisualInfo visTemplate;
	visTemplate.visualid = vid;
	int num_visuals;
	XVisualInfo *visInfo = XGetVisualInfo(crefs->xDisplay, VisualIDMask, &visTemplate, &num_visuals);
	if (!visInfo) {
		fprintf(stderr, "XGetVisualInfo failed\n");
		exit(1);
	}

	Window root = RootWindow(crefs->xDisplay, DefaultScreen(crefs->xDisplay));
	XSetWindowAttributes attr;

	attr.colormap = XCreateColormap(crefs->xDisplay, root, visInfo->visual, AllocNone);
	if (!attr.colormap) {
		fprintf(stderr, "XCreateColormap failed\n");
		exit(1);
	}

	attr.event_mask = StructureNotifyMask | ExposureMask |
		ButtonPressMask | ButtonReleaseMask | ButtonMotionMask |
        KeyPressMask | KeyReleaseMask;
	Window win = XCreateWindow(
		crefs->xDisplay, root, 0, 0, w, h, 0, visInfo->depth, InputOutput,
		visInfo->visual, CWColormap | CWEventMask, &attr);
	XFree(visInfo);

	XSizeHints sizehints;
	sizehints.width  = w;
	sizehints.height = h;
	sizehints.flags = USSize;
	XSetNormalHints(crefs->xDisplay, win, &sizehints);
	XSetStandardProperties(crefs->xDisplay, win, "App", "App", None, (char **)NULL, 0, &sizehints);

	static const EGLint ctx_attribs[] = {
		EGL_CONTEXT_CLIENT_VERSION, 2,
		EGL_NONE
	};
	crefs->eglContext = eglCreateContext(crefs->eglDisplay, crefs->eglConfig, EGL_NO_CONTEXT, ctx_attribs);
	if (!crefs->eglContext) {
		fprintf(stderr, "eglCreateContext failed\n");
		exit(1);
	}
	crefs->eglSurface = eglCreateWindowSurface(crefs->eglDisplay, crefs->eglConfig, win, NULL);
	if (!crefs->eglSurface) {
		fprintf(stderr, "eglCreateWindowSurface failed\n");
		exit(1);
	}
	return win;
}

short createWindow(int w, int h, cRefs * crefs) {
	Window win;

	//XInitThreads();
	//crefs->xDisplay = XOpenDisplay(NULL);
	//if (!crefs->xDisplay) {
		//fprintf(stderr, "XOpenDisplay failed\n");
		//exit(1);
	//}
	crefs->eglDisplay = eglGetDisplay(crefs->xDisplay);
	if (!crefs->eglDisplay) {
		fprintf(stderr, "eglGetDisplay failed\n");
		exit(1);
	}
	EGLint e_major, e_minor;
	if (!eglInitialize(crefs->eglDisplay, &e_major, &e_minor)) {
		fprintf(stderr, "eglInitialize failed\n");
		exit(1);
	}
	eglBindAPI(EGL_OPENGL_ES_API);
	win = newWindow(w, h, crefs);

	crefs->wmDeleteWindow = XInternAtom(crefs->xDisplay, "WM_DELETE_WINDOW", True);
	if (crefs->wmDeleteWindow != None) {
		XSetWMProtocols(crefs->xDisplay, win, &crefs->wmDeleteWindow, 1);
	}

	XMapWindow(crefs->xDisplay, win);
	if (!eglMakeCurrent(crefs->eglDisplay, crefs->eglSurface, crefs->eglSurface, crefs->eglContext)) {
		fprintf(stderr, "eglMakeCurrent failed\n");
		exit(1);
	}

	// Window size and DPI should be initialized before starting app.
	XEvent ev;
	while (1) {
		if (XCheckMaskEvent(crefs->xDisplay, StructureNotifyMask, &ev) == False) {
			sleep(1);
			continue;
		}
		if (ev.type == ConfigureNotify) {
			break;
		}
	}

	XRRScreenConfiguration *xrr_conf = XRRGetScreenInfo(crefs->xDisplay, win);
	return XRRConfigCurrentRate(xrr_conf);
}

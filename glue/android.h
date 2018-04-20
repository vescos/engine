// +build android

#include <stdlib.h>
#include <time.h>

#include <EGL/egl.h>

#include <android/configuration.h>
#include <android/log.h>
#include <android/native_activity.h>
#include <android/native_window.h>

void * cRefsPtr();

typedef struct {
	EGLContext	eglContext;
	EGLConfig	eglConfig;
	EGLSurface	eglSurface;
	EGLDisplay	eglDisplay;
	ANativeActivity *aActivity;
	ANativeWindow *aWindow;
	AConfiguration *aConfig;
} cRefs;

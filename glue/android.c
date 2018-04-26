// +build android

#include "_cgo_export.h"

void * cRefsPtr() {
	cRefs *p;
	p = (cRefs *) malloc(sizeof(cRefs));
	if (p != NULL) {
		p->eglDisplay = NULL;
		p->eglConfig = NULL;
		p->eglContext = NULL;
		p->eglSurface = NULL;
		p->aActivity = NULL;
		p->aWindow = NULL;
		p->aConfig = NULL;
		return (void *) p;
	}
	return NULL;
} 


// derived from gomobile
jint JNI_OnLoad(JavaVM* vm, void* reserved) {
    JNIEnv* env;
    if ((*vm)->GetEnv(vm, (void**)&env, JNI_VERSION_1_6) != JNI_OK) {
        return -1;
    }

    // Get jclass with env->FindClass.
    // Register methods with env->RegisterNatives.

    return JNI_VERSION_1_6;
}

// Entry point for app
// called when java.NativeActivity.onCreate is called
// register callback functions here
// derived from gomobile
void ANativeActivity_onCreate(ANativeActivity* activity, void* savedState, size_t savedStateSize) {
	// callbacks
	activity->callbacks->onStart = onStart;
	activity->callbacks->onResume = onResume;
	activity->callbacks->onDestroy = onDestroy;
	activity->callbacks->onPause = onPause;
	activity->callbacks->onStop = onStop;
	activity->callbacks->onWindowFocusChanged = onWindowFocusChanged;
	activity->callbacks->onNativeWindowCreated = onNativeWindowCreated;
	activity->callbacks->onConfigurationChanged = onConfigurationChanged;
	activity->callbacks->onNativeWindowDestroyed = onNativeWindowDestroyed;
	activity->callbacks->onNativeWindowRedrawNeeded = onNativeWindowRedrawNeeded;
	activity->callbacks->onSaveInstanceState = onSaveInstanceState;
	activity->callbacks->onLowMemory = onLowMemory;
	activity->callbacks->onInputQueueCreated = onInputQueueCreated;
	activity->callbacks->onInputQueueDestroyed = onInputQueueDestroyed;

	// Call the Go main.main.
	uintptr_t mainPC = (uintptr_t)dlsym(RTLD_DEFAULT, "main.main");
	if (!mainPC) {
		LOG_FATAL("missing main.main");
	}
	callMain(activity, savedState, savedStateSize, mainPC);
}


// EGL
// main egl config
const EGLint RGB_888[] = {
    EGL_SAMPLES, 4,
    EGL_BUFFER_SIZE, 32,
    EGL_BLUE_SIZE, 8,
    EGL_GREEN_SIZE, 8,
    EGL_RED_SIZE, 8,
    EGL_ALPHA_SIZE, 8,
    EGL_DEPTH_SIZE, 24,
    EGL_RENDERABLE_TYPE, EGL_OPENGL_ES2_BIT,
    EGL_SURFACE_TYPE, EGL_WINDOW_BIT,
    EGL_CONFIG_CAVEAT, EGL_NONE,
    EGL_NONE
};
// fallback egl config
const EGLint RGB_565[] = {
    EGL_BUFFER_SIZE, 16,
    EGL_BLUE_SIZE, 5,
    EGL_GREEN_SIZE, 6,
    EGL_RED_SIZE, 5,
    EGL_ALPHA_SIZE, 0,
    EGL_DEPTH_SIZE, 16,
    EGL_NONE
};

int getDisplay(cRefs * p) {
	if (p == NULL) {
		return 1;
	}
    p->eglDisplay = eglGetDisplay(EGL_DEFAULT_DISPLAY);
    if (p->eglDisplay == EGL_NO_DISPLAY) {
		return 1;
	}
    if (!eglInitialize(p->eglDisplay, NULL, NULL)) {
        return 1;
    }
    return 0;
}

int setEGLConfig(cRefs * p) {
	if (p == NULL) {
		return 1;
	}
	EGLint numConfigs = 0;

	if (!eglChooseConfig(p->eglDisplay, RGB_888, &(p->eglConfig), 1, &numConfigs)) {
		return 1;
	}
	if (numConfigs <= 0) {
		if (!eglChooseConfig(p->eglDisplay, RGB_565, &(p->eglConfig), 1, &numConfigs)) {
			return 1;
		}
		if (numConfigs <= 0) {
			return 1;
		} else {
			LOG_INFO(">>>>> EGL: choose RGB_565 config.");
		}
	} else {
		LOG_INFO(">>>>> EGL: choose RGB_8888 config.");
	}

	return 0;
}

int createGLContext(cRefs *p) {
	if (p == NULL) {
		return 1;
	} 
	if (p->eglContext != NULL) {
		return 0;
	}
	const EGLint contextAttribs[] = { EGL_CONTEXT_CLIENT_VERSION, 2, EGL_NONE };
	p->eglContext = eglCreateContext(p->eglDisplay, p->eglConfig, EGL_NO_CONTEXT, contextAttribs);
	if (p->eglContext == EGL_NO_CONTEXT) {
		return 1;
	}
	return 0;
}

int bindGLContext(cRefs *p, int w, int h) {
	if (p == NULL) {
		return 1;
	}

	EGLint format;

	eglGetConfigAttrib(p->eglDisplay, p->eglConfig, EGL_NATIVE_VISUAL_ID, &format);
	if (ANativeWindow_setBuffersGeometry(p->aWindow, w, h, format) != 0) {
		return 1;
	}

	p->eglSurface = eglCreateWindowSurface(p->eglDisplay, p->eglConfig, p->aWindow, NULL);
	if (p->eglSurface == EGL_NO_SURFACE) {
		return 1;
	}
	if (eglMakeCurrent(p->eglDisplay, p->eglSurface, p->eglSurface, p->eglContext) == EGL_FALSE) {
		return 1;
	}
	return 0;
}

// +build android

#include <stdlib.h>
#include <time.h>
#include <dlfcn.h>
#include <unistd.h>

#include <EGL/egl.h>

#include <android/configuration.h>
#include <android/log.h>
#include <android/native_activity.h>
#include <android/native_window.h>

#define LOG_INFO(...) __android_log_print(ANDROID_LOG_INFO, "LabyrinthEngine", __VA_ARGS__)
#define LOG_ERROR(...) __android_log_print(ANDROID_LOG_ERROR, "LabyrinthEngine", __VA_ARGS__)

typedef struct {
	EGLContext	eglContext;
	EGLConfig	eglConfig;
	EGLSurface	eglSurface;
	EGLDisplay	eglDisplay;
	ANativeActivity *aActivity;
	ANativeWindow *aWindow;
	AConfiguration *aConfig;
	AInputQueue *aInputQueue;
} cRefs;

void * cRefsPtr();
int getDisplay(cRefs * p);
int setEGLConfig(cRefs * p);
int createGLContext(cRefs *p);
int bindGLContext(cRefs *p, int w, int h);
float getRefreshRate(ANativeActivity* activity);
char * getPackageName(ANativeActivity* activity);
char * getIntentExtras(ANativeActivity* activity);
char * getSharedPrefs(ANativeActivity* activity, char * prefsName);
char * getNextEnv(int i);


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

	callMain(activity, savedState, savedStateSize);
}

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
		p->aInputQueue = NULL;
		return (void *) p;
	}
	return NULL;
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
		LOG_ERROR("missing main.main");
	} else {
		callMain(activity, savedState, savedStateSize, mainPC);
	}
}

float getRefreshRate(ANativeActivity* activity) {
	JNIEnv* env;
	JavaVM* vm = activity->vm;
	
	(*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6);
	(*vm)->AttachCurrentThread(vm, &env, NULL);

	jclass  activityClass = (*env)->GetObjectClass(env, activity->clazz);
	jmethodID getWindowManager = (*env)->GetMethodID(env, activityClass, "getWindowManager", "()Landroid/view/WindowManager;");

	jobject wm = (jobject)(*env)->CallObjectMethod(env, activity->clazz, getWindowManager);
	jclass wmClass = (*env)->FindClass(env, "android/view/WindowManager");
	jmethodID getDefaultDisplay = (*env)->GetMethodID(env, wmClass, "getDefaultDisplay", "()Landroid/view/Display;");

	jobject display = (jobject)(*env)->CallObjectMethod(env, wm, getDefaultDisplay);
	jclass displayClass = (*env)->FindClass(env, "android/view/Display");
	jmethodID getRefreshRateM = (*env)->GetMethodID(env, displayClass, "getRefreshRate", "()F");

	float refreshRate = (float)(*env)->CallFloatMethod(env, display, getRefreshRateM);
	
	(*vm)->DetachCurrentThread(vm);
	return refreshRate;
}

// caller must free return 
char *getPackageName(ANativeActivity* activity) {
	JNIEnv* env;
	JavaVM* vm = activity->vm;
	
	(*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6);
	(*vm)->AttachCurrentThread(vm, &env, NULL);

	jclass  activityClass = (*env)->GetObjectClass(env, activity->clazz);
	jmethodID getPkgName = (*env)->GetMethodID(env, activityClass, "getPackageName", "()Ljava/lang/String;");
	jstring jAppName = (jstring)(*env)->CallObjectMethod(env, activity->clazz, getPkgName);
	const char* str = (*env)->GetStringUTFChars(env, jAppName, NULL);
	char * cstr = strdup(str);
	
	(*env)->ReleaseStringUTFChars(env, jAppName, str);
	(*vm)->DetachCurrentThread(vm);
	return cstr;
}

// caller must free return 
char *getIntentExtras(ANativeActivity* activity) {
	JNIEnv* env;
	JavaVM* vm = activity->vm;
	
	(*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6);
	(*vm)->AttachCurrentThread(vm, &env, NULL);
	
	jclass  activityClass = (*env)->GetObjectClass(env, activity->clazz);
	jmethodID getIntent = (*env)->GetMethodID(env, activityClass, "getIntent", "()Landroid/content/Intent;");
	
	jobject intent = (jobject)(*env)->CallObjectMethod(env, activity->clazz, getIntent);
	jclass intentClass = (*env)->FindClass(env, "android/content/Intent");
	jmethodID getExtras = (*env)->GetMethodID(env, intentClass, "getExtras", "()Landroid/os/Bundle;");
	
	jobject bundle = (jobject)(*env)->CallObjectMethod(env, intent, getExtras);
	if (bundle == NULL) {
		(*vm)->DetachCurrentThread(vm);
		return NULL;
	}
	jclass bundleClass = (*env)->FindClass(env, "android/os/Bundle");
	jmethodID keySet = (*env)->GetMethodID(env, bundleClass, "keySet", "()Ljava/util/Set;");
	
	jobject set = (jobject)(*env)->CallObjectMethod(env, bundle, keySet);
	jclass setClass = (*env)->FindClass(env, "java/util/Set");
	jmethodID iteratorM = (*env)->GetMethodID(env, setClass, "iterator", "()Ljava/util/Iterator;");

	jobject iterator = (jobject)(*env)->CallObjectMethod(env, set, iteratorM);
	jclass iteratorClass = (*env)->FindClass(env, "java/util/Iterator");
 	jmethodID hasNext = (*env)->GetMethodID(env, iteratorClass, "hasNext", "()Z");
	jmethodID next = (*env)->GetMethodID(env, iteratorClass, "next", "()Ljava/lang/Object;");
	jmethodID bundleGet = (*env)->GetMethodID(env, bundleClass, "get", "(Ljava/lang/String;)Ljava/lang/Object;");
	
	char * cstr = (char *) malloc(10);
	cstr[0] = '\0'; 
	jboolean has;
	while (has = (*env)->CallBooleanMethod(env, iterator, hasNext)) {
		jstring strObj = (jstring)(*env)->CallObjectMethod(env, iterator, next);
		jobject val = (jobject)(*env)->CallObjectMethod(env, bundle, bundleGet, strObj);
		jclass  unknownClass = (*env)->GetObjectClass(env, val);
		jmethodID toStr = (*env)->GetMethodID(env, unknownClass, "toString", "()Ljava/lang/String;");
		jstring valStrObj = (jstring)(*env)->CallObjectMethod(env, val, toStr);
		
		const char* key = (*env)->GetStringUTFChars(env, strObj, NULL);
		const char* valStr = (*env)->GetStringUTFChars(env, valStrObj, NULL);
		
		if (strchr(key, '\n') != NULL || strchr(key, '=') != NULL) {
			LOG_INFO("Intent key contains newline or equal sign: %s, skipping", key);
			continue;
		}
		if (strchr(valStr, '\n') != NULL) {
			LOG_INFO("Intent value contains newline: %s, skipping", valStr);
			continue;
		}
		int len = strlen(cstr) + strlen(key) + strlen(valStr) + 3;
		cstr = (char *) realloc(cstr, sizeof(char) * len);
		if (strlen(cstr)) {
			strcat(cstr, "\n");
		}
		strcat(cstr, key);
		strcat(cstr, "=");
		strcat(cstr, valStr);
		
		(*env)->ReleaseStringUTFChars(env, valStrObj, valStr);
		(*env)->ReleaseStringUTFChars(env, strObj, key);
	}
	
	(*vm)->DetachCurrentThread(vm);
	if (strlen(cstr) == 0) {
		free(cstr);
		return NULL;
	}
	return cstr;
}

// caller must free return
char *getSharedPrefs(ANativeActivity* activity, char * prefsName) {
	JNIEnv* env;
	JavaVM* vm = activity->vm;
	
	(*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6);
	(*vm)->AttachCurrentThread(vm, &env, NULL);
	
	jclass  activityClass = (*env)->GetObjectClass(env, activity->clazz);
	jmethodID prefsM = (*env)->GetMethodID(env, activityClass, 
				"getSharedPreferences", "(Ljava/lang/String;I)Landroid/content/SharedPreferences;");
	jstring jPrefsName = (*env)->NewStringUTF(env, prefsName);
	// 0 - MODE_PRIVATE
	jobject prefsObj = (jobject)(*env)->CallObjectMethod(env, activity->clazz, prefsM, jPrefsName, 0);
	jclass prefsClass = (*env)->FindClass(env, "android/content/SharedPreferences");
	jmethodID getAllM = (*env)->GetMethodID(env, prefsClass, "getAll", "()Ljava/util/Map;");

	jobject mapObj = (jobject)(*env)->CallObjectMethod(env, prefsObj, getAllM);
	jclass mapClass = (*env)->FindClass(env, "java/util/Map");
	jmethodID keySetM = (*env)->GetMethodID(env, mapClass, "keySet", "()Ljava/util/Set;");
	
	jobject setObj = (jobject)(*env)->CallObjectMethod(env, mapObj, keySetM);
	jclass setClass = (*env)->FindClass(env, "java/util/Set");
	jmethodID iteratorM = (*env)->GetMethodID(env, setClass, "iterator", "()Ljava/util/Iterator;");

	jobject iterator = (jobject)(*env)->CallObjectMethod(env, setObj, iteratorM);
	jclass iteratorClass = (*env)->FindClass(env, "java/util/Iterator");
 	jmethodID hasNext = (*env)->GetMethodID(env, iteratorClass, "hasNext", "()Z");
	jmethodID next = (*env)->GetMethodID(env, iteratorClass, "next", "()Ljava/lang/Object;");
	jmethodID mapGet = (*env)->GetMethodID(env, mapClass, "get", "(Ljava/lang/Object;)Ljava/lang/Object;");
	
	char * cstr = (char *) malloc(10);
	cstr[0] = '\0';
	jboolean has;
	while (has = (*env)->CallBooleanMethod(env, iterator, hasNext)) {
		jstring strObj = (jstring)(*env)->CallObjectMethod(env, iterator, next);
		jobject val = (jobject)(*env)->CallObjectMethod(env, mapObj, mapGet, strObj);
		jclass  unknownClass = (*env)->GetObjectClass(env, val);
		jmethodID toStr = (*env)->GetMethodID(env, unknownClass, "toString", "()Ljava/lang/String;");
		jstring valStrObj = (jstring)(*env)->CallObjectMethod(env, val, toStr);
		
		const char* key = (*env)->GetStringUTFChars(env, strObj, NULL);
		const char* valStr = (*env)->GetStringUTFChars(env, valStrObj, NULL);
		
		if (strchr(key, '\n') != NULL || strchr(key, '=') != NULL) {
			LOG_INFO("Intent key contains newline or equal sign: %s, skipping", key);
			continue;
		}
		if (strchr(valStr, '\n') != NULL) {
			LOG_INFO("Intent value contains newline: %s, skipping", valStr);
			continue;
		}
		int len = strlen(cstr) + strlen(key) + strlen(valStr) + 3;
		cstr = (char *) realloc(cstr, sizeof(char) * len);
		if (strlen(cstr)) {
			strcat(cstr, "\n");
		}
		strcat(cstr, key);
		strcat(cstr, "=");
		strcat(cstr, valStr);
		
		(*env)->ReleaseStringUTFChars(env, valStrObj, valStr);
		(*env)->ReleaseStringUTFChars(env, strObj, key);
	}
	(*vm)->DetachCurrentThread(vm);
	if (strlen(cstr) == 0) {
		free(cstr);
		return NULL;
	}
	return cstr;
}

// return 1 on succes 0 on failure
void saveSharedPrefs(ANativeActivity* activity, char * prefsName, char * prefs) {
	JNIEnv* env;
	JavaVM* vm = activity->vm;
	
	(*vm)->GetEnv(vm, (void **)&env, JNI_VERSION_1_6);
	(*vm)->AttachCurrentThread(vm, &env, NULL);
	
	jclass  activityClass = (*env)->GetObjectClass(env, activity->clazz);
	jmethodID prefsM = (*env)->GetMethodID(env, activityClass, 
			"getSharedPreferences", "(Ljava/lang/String;I)Landroid/content/SharedPreferences;");
	jstring jPrefsName = (*env)->NewStringUTF(env, prefsName);
	// 0 - MODE_PRIVATE
	jobject prefsObj = (jobject)(*env)->CallObjectMethod(env, activity->clazz, prefsM, jPrefsName, 0);
	jclass prefsClass = (*env)->FindClass(env, "android/content/SharedPreferences");
	jmethodID editorM = (*env)->GetMethodID(env, prefsClass, "edit", "()Landroid/content/SharedPreferences$Editor;");
	
	jobject editorObj = (jobject)(*env)->CallObjectMethod(env, prefsObj, editorM);
	jclass editorClass = (*env)->GetObjectClass(env, editorObj);
	jmethodID putStringM = (*env)->GetMethodID(env, editorClass, 
			"putString", "(Ljava/lang/String;Ljava/lang/String;)Landroid/content/SharedPreferences$Editor;");
	// expected string format is 
	// key=val\n....key=val\n\0
	char *key, *val, *end;
	key = prefs;
	// There is no check that key with same name exist but with different type(eg. created from java side)!!!
	// Behavior is undefined(crash) in this case.
	while (end = strstr(key, "\n")) {
		val = strstr(key, "=");
		char * k = strndup(key, val - key);
		val++;
		// if val is empty string end - val will be 0?
		char * v = strndup(val, end - val);
		jstring jk = (*env)->NewStringUTF(env, k);
		jstring jv = (*env)->NewStringUTF(env, v);
		
		(*env)->CallObjectMethod(env, editorObj, putStringM, jk, jv);
		
		(*env)->DeleteLocalRef(env, jk);
		(*env)->DeleteLocalRef(env, jv);
		free(v);
		free(k);
		key = end + 1;
		if (key == '\0') {
			break;
		}
	} 
	
	jmethodID applyM = (*env)->GetMethodID(env, editorClass, "apply", "()V");
	(*env)->CallVoidMethod(env, editorObj, applyM);
	
	(*vm)->DetachCurrentThread(vm);
}

char *getNextEnv(int i) {
	// environ is defined in unistd.h
	return *(environ+i);
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

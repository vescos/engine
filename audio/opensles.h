
#include <SLES/OpenSLES.h>
#include <SLES/OpenSLES_Android.h>

#include <stdlib.h>
#include <stdio.h>




//OpenSL ES objects/interfaces
typedef struct {
	SLObjectItf engineObj;
	SLEngineItf engineEng;
	SLObjectItf outputMixObj;
	SLObjectItf playerObj;
	SLPlayItf	playerPlay;
	SLAndroidSimpleBufferQueueItf playerBuffQueue;
	uint buffSizeCnt;
	uint singleBuffSize;
	void * ringBuff;
	uint ringBuffNextAdd;
} oslHandle;

extern void availWriteCallback(oslHandle * handle);

//keep slES pointers in C memory to avoid bad pointers and heap corruption in GO runtime
static inline oslHandle* getOslHandle(uint buffSizeCnt, uint buffSize) {
	oslHandle *p;
	void * ringBuff;
	int i;
	
	ringBuff = (void *) malloc(buffSize);
	if (ringBuff == NULL) {
		return NULL;
	}
	p = (oslHandle *) malloc(sizeof(oslHandle));
	if (p == NULL) {
		return NULL;
	}
	if (p != NULL) {
		p->engineObj = NULL;
		p->engineEng = NULL;
		p->outputMixObj = NULL;
		p->playerObj = NULL;
		p->playerPlay = NULL;
		p->playerBuffQueue = NULL;
		p->buffSizeCnt = buffSizeCnt;
		p->singleBuffSize = buffSize / buffSizeCnt;
		p->ringBuff = ringBuff;
		p->ringBuffNextAdd = 0;
	}
	
	return (void *) p;
}

static inline int createEngine(oslHandle * handle) {
	SLresult result;
	result = slCreateEngine(&(handle->engineObj), 0, NULL, 0, NULL, NULL);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	return 0;
}

//initialize osl interface and outputMix
static inline int initOsl (oslHandle *handle) {
	SLresult result;
	//get sl object
	result = (*(handle->engineObj))->Realize(handle->engineObj, SL_BOOLEAN_FALSE);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	//get sl obj itf
	result = (*(handle->engineObj))->GetInterface(handle->engineObj, SL_IID_ENGINE, &(handle->engineEng));
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	//create/realize outputMix 
	result = (*(handle->engineEng))->CreateOutputMix(handle->engineEng, &(handle->outputMixObj), 0, NULL, NULL);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	result = (*(handle->outputMixObj))->Realize(handle->outputMixObj, SL_BOOLEAN_FALSE);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	return 0;
}

//this callback is called every time buffer is finished
//and is called when there is no buffer at all too(called on interval)
static inline void playerCallback (SLAndroidSimpleBufferQueueItf bq, void *context) {
	oslHandle* handle = (oslHandle *) context;
	availWriteCallback(handle);
	return;
}

//create audioPlayer 
static inline int createPlayer(oslHandle *handle, uint rate, uint channels, uint sample_size) {
	SLresult result;
	//bellow is copy/paste/modify from native-audio-jni.c
	 // configure audio source
	SLDataLocator_AndroidSimpleBufferQueue loc_bufq = {SL_DATALOCATOR_ANDROIDSIMPLEBUFFERQUEUE, handle->buffSizeCnt};
	SLDataFormat_PCM format_pcm = {SL_DATAFORMAT_PCM, channels, rate * 1000,
		sample_size * 8, sample_size * 8,
		SL_SPEAKER_FRONT_LEFT | SL_SPEAKER_FRONT_RIGHT, SL_BYTEORDER_LITTLEENDIAN};
	SLDataSource audioSrc = {&loc_bufq, &format_pcm};
	// configure audio sink
	SLDataLocator_OutputMix loc_outmix = {SL_DATALOCATOR_OUTPUTMIX, handle->outputMixObj};
	SLDataSink audioSnk = {&loc_outmix, NULL};
	// create audio player
	const SLInterfaceID ids[1] = {SL_IID_BUFFERQUEUE};
	const SLboolean req[1] = {SL_BOOLEAN_TRUE};
	result = (*(handle->engineEng))->CreateAudioPlayer(handle->engineEng, &(handle->playerObj), &audioSrc, &audioSnk,
			1, ids, req);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	// realize the player
    result = (*(handle->playerObj))->Realize(handle->playerObj, SL_BOOLEAN_FALSE);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	// get the play interface
	result = (*(handle->playerObj))->GetInterface(handle->playerObj, SL_IID_PLAY, &(handle->playerPlay));
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	// get the buffer queue interface
    result = (*(handle->playerObj))->GetInterface(handle->playerObj, SL_IID_BUFFERQUEUE,
            &(handle->playerBuffQueue));
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	// register callback on the buffer queue
    result = (*(handle->playerBuffQueue))->RegisterCallback(handle->playerBuffQueue, playerCallback, handle);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	// set the player's state to playing
    result =  (*(handle->playerPlay))->SetPlayState(handle->playerPlay, SL_PLAYSTATE_PLAYING);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	return 0;
}

//get encueued buffers count
static inline int getQueuedBuffsCount (oslHandle *handle) {
	SLresult result;
	SLAndroidSimpleBufferQueueState state;
	result =  (*(handle->playerBuffQueue))->GetState(handle->playerBuffQueue, &state);
	if (result != SL_RESULT_SUCCESS) {
		return -1;
	}
	return (int) state.count;
} 

//send new buffer to player queue
static inline int enqueueBuff (oslHandle *handle, const void *buff, uint buffSize) {
	SLresult result;
	int i;
	
	//go pointers to go memory must not be stored in c memory - copy buffer
	void *p = handle->ringBuff + (handle->ringBuffNextAdd * handle->singleBuffSize);
	
	memcpy(p, buff, buffSize);

	result =  (*(handle->playerBuffQueue))->Enqueue(handle->playerBuffQueue, p, buffSize);
	if (result != SL_RESULT_SUCCESS) {
		return result;
	}
	
	handle->ringBuffNextAdd += 1;
	if (handle->ringBuffNextAdd >= handle->buffSizeCnt) {
		handle->ringBuffNextAdd = 0;
	}
	return 0;
}

//cleanup osl 
static inline void closeDevice (oslHandle *handle) {
	int i;
	if (handle == NULL) {
		return;
	}
	if (handle->playerObj != NULL) {
		(*(handle->playerBuffQueue))->Clear(handle->playerBuffQueue);
		(*(handle->playerObj))->Destroy(handle->playerObj);
		handle->playerPlay = NULL;
		handle->playerBuffQueue = NULL;
	}
	if (handle->outputMixObj != NULL) {
		(*(handle->outputMixObj))->Destroy(handle->outputMixObj);
		handle->outputMixObj = NULL;
	}
	if (handle->engineObj != NULL) {
		(*(handle->engineObj))->Destroy(handle->engineObj);
		handle->engineObj = NULL;
		handle->engineEng = NULL;
	}
	if (handle->ringBuff != NULL) {
		free(handle->ringBuff);
	}
	free(handle);
}

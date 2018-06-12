# LabyrinthEngine

Experimental, cross platform  2D/3D engine in Go(Golang) and C.

### License
BSD 3-Clause License

### Supported platforms:
#### Build
Linux(X11) - amd64.  
Other platforms that support golang tools probably will work but this is not tested.
#### Target
Linux(X11) - amd64.  
Android - arm(armeabi, armeabi-v7a), arm64(armv8a), 386(x86), amd64(x86_64(build but untested)).

### Third party software in this library
Some parts of glue and input are derived from [gomobile project](https://github.com/golang/mobile)  
gles2/gl package is from [gl package](https://github.com/goxjs/gl)  
Ogg Vorbis decoder [stb_vorbis](http://nothings.org/stb_vorbis/) (Open Domain license)

### Dependencies 
#### To build
Standard Golang library/tools.

Linux(X11) - C compiler(GCC), X11, randr, EGL, and header files.

Android - Standalone Android NDK toolchains.
1. Download Android Ndk [https://developer.android.com/ndk/downloads/](https://developer.android.com/ndk/downloads/)
2. Extract Ndk archive.
3. $ cd PATHTO/ndk-bundle/build/tools
4. Build standalone toolchains(edit install-dir and ndk version by will). 
armeabi and armeabi-v7a toolchain is most used platform(binaries will run on arm64 also)  
// armeabi and armeabi-v7a toolchain  
$ ./make_standalone_toolchain.py --arch arm --api 16 --install-dir ${HOME}/android/ndk14b-19-arm  
// x86  
$ ./make_standalone_toolchain.py --arch x86 --api 16 --install-dir ${HOME}/android/ndk14b-19-x86  
//x86_64  
$ ./make_standalone_toolchain.py --arch x86_64 --api 21 --install-dir ${HOME}/android/ndk14b-21-x86_64  
//arm64  
$ ./make_standalone_toolchain.py --arch arm64 --api 21 --install-dir ${HOME}/android/ndk14b-21-arm64  
5. Build shared C library. In this case for armeabi-v7a (assuming that $NDK_ROOT_ARM is properly set to pathto/toolchain)  
$ GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 \  
CC=$(NDK_ROOT_ARM)/bin/arm-linux-androideabi-gcc \  
CXX=$(NDK_ROOT_ARM)/bin/arm-linux-androideabi-g++ \  
go build -ldflags '-X glue.goarm=7' -buildmode=c-shared  -o /tmp/libexample.so github.com/vescos/engine/sample  
6. Build more C chared libs if needed.
Minimum tested api is 16 but probably will work with 10+

#### To run

Linux(X11) - mesa with gles2.

Android - use Your prefered method to build and install Android packages.   
No java code is needed to run engine. Mixing with java and jni code is possible.  
Use glue.State.hackEngine() to obtain C struct that holds references to nativeActivity object.  
See definition of this structure in glue/android.h file. 

### Install
go get github.com/vescos/engine

### Usage
See sample folder(sample is incomplete)
// TODO:

### History
Engine was part of [Labyrinth Lost Gems](https://play.google.com/store/apps/details?id=xyz.live3dgraphs.labyrinth) game(include ads)
 

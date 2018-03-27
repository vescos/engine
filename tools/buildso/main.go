//not in use(replaced by Makefile)

package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	tmp := "/tmp/gosharedlibfoo.so"
	pkg := os.Args[1]
	out := os.Args[2]

	cmd := exec.Command("go")
	cmd.Args = append(cmd.Args, "build")
	cmd.Args = append(cmd.Args, "-buildmode=c-shared")
	//cmd.Args = append(cmd.Args, "-x")
	//cmd.Args = append(cmd.Args, "-v")
	//cmd.Args = append(cmd.Args, "-tags", "gldebug")

	cmd.Args = append(cmd.Args, "-ldflags", `"-s"`)
	cmd.Args = append(cmd.Args, "-o", tmp)
	cmd.Args = append(cmd.Args, pkg)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=android")
	cmd.Env = append(cmd.Env, "GOARCH=arm")
	cmd.Env = append(cmd.Env, "GOARM=6")

	cmd.Env = append(cmd.Env, "CC="+os.Getenv("NDK_ROOT")+"/bin/arm-linux-androideabi-gcc")
	cmd.Env = append(cmd.Env, "CXX="+os.Getenv("NDK_ROOT")+"/bin/arm-linux-androideabi-g++")

	cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Print("go build:", err)
		os.Exit(1)
	}

	cmd = exec.Command("mv")
	cmd.Args = append(cmd.Args, tmp)
	cmd.Args = append(cmd.Args, out)
	err = cmd.Run()
	if err != nil {
		log.Print("mv: ", err)
		os.Exit(1)
	}
	log.Print("Succes EABI6")

	os.Exit(0)

}

//aapt package -M AndroidManifest.xml -A assets -S res
//android create project --name labyrinth --target android-16 --path ./labyrinth --package xyz.live3dgraphs.labyrinth --activity LabyrinthNativeActivity
//mkdir -p labyrinth/src/main/java/xyz/live3dgrphics/labyrinth
//mkdir -p labyrinth/src/main/jniLibs/armeabi
//adb install -r build/outputs/apk/labyrinth-debug.apk
//keytool -genkey -v -keystore live3dgraphsrelease.keystore -alias live3dgraphsXYZ -keyalg RSA -keysize 2048 -validity 15000

//adb shell screencap -p /sdcard/screencap.png
//adb pull /sdcard/screencap.png

//android list targets
//android list avds
//android create avd -n a19 -t "android-19"
//android create avd -n a15_QVGA -t "android-16" -s QVGA
//android delete avd -n name
//emulator -avd a19 -gpu on //gpu acceleration available on platform >= android-15
//available avds a10, a15, a15_QVGA, a16, a17, a18, a19 (10 - supports GLES 1.0 only)

//rebuild go android_arm toolchain
//CGO_ENABLED=1 CC=$NDK_ROOT/bin/arm-linux-androideabi-gcc CXX=$NDK_ROOT/bin/arm-linux-androideabi-g++ GOOS=android GOARCH=arm GOARM=6 go install std
//buildso graphs/labyrinth $HOME/android/projects/labyrinth/src/main/jniLibs/armeabi/liblabyrinth.so && gradle assembleDebug && adb install -r $HOME/android/projects/labyrinth/build/outputs/apk/labyrinth-debug.apk
//buildso jni/jnitest $HOME/android/projects/jnitest/src/main/jniLibs/armeabi/libjnitest.so && gradle assembleDebug && adb install -r $HOME/android/projects/jnitest/build/outputs/apk/jnitest-debug.apk

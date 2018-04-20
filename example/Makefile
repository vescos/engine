
help:
	@echo 'armv7a 		build armv7a'
	@echo 'start_x11 	install and run x11'

armv7a:
	GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 \
	CC=$(NDK_ROOT_ARM)/bin/arm-linux-androideabi-gcc \
	CXX=$(NDK_ROOT_ARM)/bin/arm-linux-androideabi-g++ \
	go build -ldflags '-X main.goarm=7' -buildmode=c-shared  -o /tmp/libexample.so graphs/engine/example

x11:
	go install -race graphs/engine/example
	GODEBUG="cgocheck=2" \
	GORACE="history_size=7 strip_path_prefix="$(HOME) \
	example  
#	GOTRACEBACK=crash \
#	GODEBUG="cgocheck=2,gctrace=1" \


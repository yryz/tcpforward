build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o tcpforward-linux_x64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags "-s -w" -trimpath -o tcpforward-linux_arm
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o tcpforward-win64.exe
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -w" -trimpath -o tcpforward-win32.exe
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o tcpforward-mac64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -trimpath -o tcpforward-mac_arm64

clean:
	rm tcpforward-*
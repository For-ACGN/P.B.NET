set GOOS=windows
go build -v -trimpath -ldflags "-s -w" -o injector.exe
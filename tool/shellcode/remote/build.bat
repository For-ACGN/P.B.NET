set GOARCH=386
go build -v -trimpath -ldflags "-s -w" -o remote-32.exe
set GOARCH=amd64
go build -v -trimpath -ldflags "-s -w" -o remote-64.exe
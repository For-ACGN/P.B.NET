set GOARCH=386
go build -v -trimpath -ldflags "-s -w" -o local-32.exe
set GOARCH=amd64
go build -v -trimpath -ldflags "-s -w" -o local-64.exe

# Need to update go-sdl2 to static link, still is wip
#$env:CGO_ENABLED = "1"
#$env:CC = "gcc"
#$env:GOOS = "windows"
#$env:GOARCH = "amd64"
#go build -tags static -ldflags "-s -w -H windowsgui"
#go build -ldflags "-s -w -H windowsgui"

go build -ldflags "-H windowsgui"

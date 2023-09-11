package xversion

// GitCommit inject by -ldflags
// GIT_COMMIT=git rev-list -1 HEAD && go build -ldflags "-X main.GitCommit=$GIT_COMMIT"
var GitCommit string

// Version of module from go.mod
var Version string

// BuildDate time of build
var BuildDate string

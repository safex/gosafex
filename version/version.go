package version

import (
	"fmt"
	"runtime"
)

// GitCommit is the git commit that was compiled. This will be filled in by the compiler
var GitCommit string

// Version is the main version number of the package
const Version = "0.1.0"

// BuildDate is the build datetime of the last compile. This will be filled by the compiler
var BuildDate = ""

// GoVersion is the version of the go runtime
var GoVersion = runtime.Version()

// OsArch is the OS and architecture of the runtime
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

//go:build windows
// +build windows

package daemon

const dockerExe = "C:\\bin\\docker.exe"
const dockerdExe = ""
const dockerHome = "C:\\ProgramData\\docker\\"

func startDaemon(daemon Daemon) {
	// this is a no-op on windows
}

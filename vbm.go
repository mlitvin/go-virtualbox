package virtualbox

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	VBM     string           // Path to VBoxManage utility.
	Verbose bool             // Verbose mode.
	Log     func(msg string) // function to be used for logging
	LogOut  func(msg string) // function to be used for logging stdout
	LogErr  func(msg string) // function to be used for logging stderr
)

func init() {
	VBM = "VBoxManage"
	if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" && runtime.GOOS == "windows" {
		VBM = filepath.Join(p, "VBoxManage.exe")
	}
}

var (
	reVMNameUUID      = regexp.MustCompile(`"(.+)" {([0-9a-f-]+)}`)
	reVMInfoLine      = regexp.MustCompile(`(?:"(.+)"|(.+))=(?:"(.*)"|(.*))`)
	reColonLine       = regexp.MustCompile(`(.+):\s+(.*)`)
	reMachineNotFound = regexp.MustCompile(`Could not find a registered machine named '(.+)'`)
	reErrorMessage    = regexp.MustCompile(`VBoxManage: error: (.*)`)
)

var (
	ErrMachineExist    = errors.New("machine already exists")
	ErrMachineNotExist = errors.New("machine does not exist")
	ErrVBMNotFound     = errors.New("VBoxManage not found")
)

func vbm(args ...string) error {
	_, _, err := vbmOutErr(args...)
	return err
}

func vbmOut(args ...string) (string, error) {
	out, _, err := vbmOutErr(args...)
	return out, err
}

func vbmOutErr(args ...string) (string, string, error) {
	cmd := exec.Command(VBM, args...)
	LogMessage("executing: %v %v", VBM, strings.Join(args, " "))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		LogMessage("Error: %s", err)
	}
	logOut(stdout.String())
	logErr(stderr.String())
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrVBMNotFound
		}
		if _, ok := err.(*exec.ExitError); ok {
			if res := reErrorMessage.FindStringSubmatch(stderr.String()); res != nil {
				err = errors.New(res[1])
			}
		}
	}
	return stdout.String(), stderr.String(), err
}

func LogMessage(format string, args ...interface{}) {
	switch {
	case Log != nil:
		Log(fmt.Sprintf(format, args...))
	case Verbose:
		log.Printf(format, args...)
	}
}

func logOut(msg string) {
	switch {
	case msg == "":
		return
	case LogOut != nil:
		LogOut(msg)
	case Log != nil:
		Log("StdOut: " + msg)
	case Verbose:
		os.Stdout.WriteString(msg)
	}
}

func logErr(msg string) {
	switch {
	case msg == "":
		return
	case LogOut != nil:
		LogErr(msg)
	case Log != nil:
		Log("StdErr: " + msg)
	case Verbose:
		os.Stderr.WriteString(msg)
	}
}

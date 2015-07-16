package virtualbox

import (
	"fmt"
	"regexp"
	"strings"
)

// SharedFolder represent a single shared folder
type SharedFolder struct {
	Name string // Name of the shared folder
	Path string // Path (on the host machine)
}

type SharedFolders struct {
	folders map[string]*SharedFolder
}

var (
	reSharedFolder = regexp.MustCompile(`SharedFolder(Name|Path)(Machine|Transient)Mapping([0-9]+)$`)
)

// parse property key value as parsed from the output of VBoxManage showvminfo --machinereadable
//
// If it is a SharedFolder property, it parses (returning error if it is a bad one)
// and store the value
//
// Otherwise it ignores it and return nil
func (s *SharedFolders) parseProperty(key, value string) error {
	if !strings.HasPrefix(key, "SharedFolder") {
		return nil
	}

	match := reSharedFolder.FindStringSubmatch(key)
	if match == nil {
		return fmt.Errorf("Bad property: %s", key)
	}

	if s.folders == nil {
		s.folders = make(map[string]*SharedFolder)
	}

	action := match[1]
	id := match[2] + match[3]

	sf := s.folders[id]
	if sf == nil {
		sf = &SharedFolder{}
		s.folders[id] = sf
	}

	switch action {
	case "Name":
		sf.Name = value
	case "Path":
		sf.Path = value
	default:
		return fmt.Errorf("Bad property: %s", key)
	}
	return nil
}

func (s *SharedFolders) List() []SharedFolder {
	if s == nil || len(s.folders) == 0 {
		return nil
	}

	list := make([]SharedFolder, len(s.folders))
	i := 0
	for _, f := range s.folders {
		list[i] = *f
		i++
	}
	return list
}

func (m *Machine) SharedFolderAddMachine(name, path string) error {
	return vbmParseErr([]string{"sharedfolder", "add", m.Name, "--name", name, "--hostpath", path}...)
}

func (m *Machine) SharedFolderAddTransient(name, path string) error {
	return vbmParseErr([]string{"sharedfolder", "add", m.Name, "--name", name, "--hostpath", path, "--transient"}...)
}

func (m *Machine) SharedFolderAdd(name, path string) error {
	if m.State == Poweroff || m.State == Aborted {
		return m.SharedFolderAddMachine(name, path)
	}
	return m.SharedFolderAddTransient(name, path)
}

func (m *Machine) SharedFolderRemove(name string) error {
	params := []string{"sharedfolder", "remove", m.Name, "--name", name, "--transient"}
	if m.State == Poweroff || m.State == Aborted {
		params = params[:len(params)-1]
	}
	return vbmParseErr(params...)
}

package fileaccess

import (
	"fmt"

	"github.com/aquasecurity/libbpfgo"
	"github.com/mrtc0/bouheki/pkg/config"
	log "github.com/mrtc0/bouheki/pkg/log"
)

type Manager struct {
	mod    *libbpfgo.Module
	config *config.Config
	pb     *libbpfgo.PerfBuffer
}

func (m *Manager) Start(eventChannel chan []byte) error {
	pb, err := m.mod.InitPerfBuf("fileopen_events", eventChannel, nil, 1024)
	if err != nil {
		return err
	}

	pb.Start()
	m.pb = pb

	return nil
}

func (m *Manager) Close() {
	m.pb.Close()
}

func (m *Manager) Attach() error {
	prog, err := m.mod.GetProgram(BPF_PROGRAM_NAME)
	if err != nil {
		return err
	}

	_, err = prog.AttachLSM()
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("%s attached.", BPF_PROGRAM_NAME))
	return nil
}

func (m *Manager) SetConfigToMap() error {
	map_allowed_files, err := m.mod.GetMap(ALLOWED_FILES_MAP_NAME)
	if err != nil {
		return err
	}
	map_denied_files, err := m.mod.GetMap(DENIED_FILES_MAP_NAME)
	if err != nil {
		return err
	}

	allowed_paths := m.config.RestrictedFileAccess.Allow

	for i, path := range allowed_paths {
		err = map_allowed_files.Update(uint8(i), []byte(path))
		if err != nil {
			return err
		}
	}

	denied_paths := m.config.RestrictedFileAccess.Deny

	for i, path := range denied_paths {
		err = map_denied_files.Update(uint8(i), []byte(path))
		if err != nil {
			return err
		}
	}

	return nil
}

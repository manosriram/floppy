package config

import (
	"os"
	"strings"
)

type Mountpoints struct {
	Path        string
	MountPoints map[string]string
}

func NewMountPoints(path string) Mountpoints {
	return Mountpoints{
		Path:        path,
		MountPoints: make(map[string]string),
	}
}

func (m *Mountpoints) ReadMountPointsFromConfig() error {
	mountPointsFileData, err := os.ReadFile(m.Path)
	if err != nil {
		return err
	}

	mountPoints := strings.Split(string(mountPointsFileData), "\n")

	var mountPointsInReq []string
	for _, mountPoint := range mountPoints {
		mountPointEntry := strings.Split(mountPoint, ":")
		if len(mountPointEntry) >= 2 {
			mountPointsInReq = append(mountPointsInReq, mountPointEntry[0])
			m.MountPoints[mountPointEntry[0]] = mountPointEntry[1]
		}
	}
	// m.MountPoints = mountPointsInReq
	return nil
}

func (m *Mountpoints) ListMountPoints() map[string]string {
	return m.MountPoints
}

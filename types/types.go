package types

import ctypes "gitlab.ricebook.net/platform/core/types"

type Container struct {
	ID         string
	AppName    string
	Entrypoint string
	Memory     int64
	CPU        ctypes.CPUMap
}

type Entrypoint struct {
	Count int
	Mem   int64
}

type App struct {
	Entrypoints map[string]*Entrypoint
	MemTotal    int64
	Mem         string
	CPUTotal    int64
	Count       int
}

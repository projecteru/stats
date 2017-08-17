package types

import ctypes "github.com/projecteru2/core/types"

type Container struct {
	ID         string
	AppName    string
	Entrypoint string
	Memory     int64
	CPU        ctypes.CPUMap
	Pod        string
	Node       string
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

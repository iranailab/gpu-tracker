package types

import "time"

// GPU describes a single GPU device snapshot.
type GPU struct {
	Index        int
	Name         string
	UUID         string
	UtilGPU      float64 // percent 0..100
	UtilMem      float64 // percent 0..100
	MemUsedMB    float64
	MemTotalMB   float64
	TempC        float64
	PowerDrawW   float64
	PowerLimitW  float64
}

// GPUProcess describes a compute process using a GPU.
type GPUProcess struct {
	PID         int
	ProcessName string
	UsedMemMB   float64
	GPUUUID     string
	User        string // resolved from UID
}

// Snapshot is a full capture of system GPUs and processes at a moment.
type Snapshot struct {
	ID       int64
	TS       time.Time
	GPUs     []GPU
	Procs    []GPUProcess
}

// Aggregated view per user (in MB).
type UserAgg struct {
	User      string
	MemUsedMB float64
}

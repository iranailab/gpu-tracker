package sampler

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gpuwatch/internal/types"
	"gpuwatch/internal/util"
)

var ErrNoNvidiaSMI = errors.New("nvidia-smi not found or not working")

// Sample queries nvidia-smi for GPU and per-process data and maps PIDs to usernames.
func Sample() (types.Snapshot, error) {
	if err := checkNvidiaSMI(); err != nil {
		return types.Snapshot{}, err
	}

	gpus, err := queryGPUs()
	if err != nil {
		return types.Snapshot{}, err
	}
	procs, err := queryProcs()
	if err != nil {
		// Not fatal: some systems may have no compute apps; keep GPUs only.
		procs = nil
	}
	// resolve PIDs -> users
	uidMap := util.BuildUIDMap()
	for i := range procs {
		if uid, ok := util.ReadProcUID(procs[i].PID); ok {
			if u, ok := uidMap[uid]; ok {
				procs[i].User = u
			} else {
				procs[i].User = fmt.Sprintf("uid:%s", uid)
			}
		} else {
			procs[i].User = "?"
		}
	}

	return types.Snapshot{
		TS:    time.Now(),
		GPUs:  gpus,
		Procs: procs,
	}, nil
}

func checkNvidiaSMI() error {
	cmd := exec.Command("nvidia-smi", "-L")
	if err := cmd.Run(); err != nil {
		return ErrNoNvidiaSMI
	}
	return nil
}

func queryGPUs() ([]types.GPU, error) {
	args := []string{"--query-gpu=index,name,uuid,utilization.gpu,utilization.memory,memory.used,memory.total,temperature.gpu,power.draw,power.limit", "--format=csv,noheader,nounits"}
	out, err := exec.Command("nvidia-smi", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi gpu query: %w", err)
	}
	reader := csv.NewReader(strings.NewReader(string(out)))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	var res []types.GPU
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(rec) < 10 {
			continue
		}
		idx := atoi(rec[0])
		name := strings.TrimSpace(rec[1])
		uuid := strings.TrimSpace(rec[2])
		utilGPU := atof(rec[3])
		utilMem := atof(rec[4])
		memUsed := atof(rec[5])
		memTot := atof(rec[6])
		temp := atof(rec[7])
		pwr := atof(rec[8])
		pwrLim := atof(rec[9])
		res = append(res, types.GPU{
			Index: idx, Name: name, UUID: uuid,
			UtilGPU: utilGPU, UtilMem: utilMem,
			MemUsedMB: memUsed, MemTotalMB: memTot,
			TempC: temp, PowerDrawW: pwr, PowerLimitW: pwrLim,
		})
	}
	return res, nil
}

func queryProcs() ([]types.GPUProcess, error) {
	args := []string{"--query-compute-apps=pid,process_name,used_memory,gpu_uuid", "--format=csv,noheader,nounits"}
	cmd := exec.Command("nvidia-smi", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	s := bufio.NewScanner(stdout)
	var res []types.GPUProcess
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		// CSV without quotes, split by comma
		parts := splitCSVLine(line)
		if len(parts) < 4 {
			continue
		}
		pid := atoi(parts[0])
		pname := strings.TrimSpace(parts[1])
		mem := atof(parts[2])
		uuid := strings.TrimSpace(parts[3])
		res = append(res, types.GPUProcess{PID: pid, ProcessName: pname, UsedMemMB: mem, GPUUUID: uuid})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	_ = cmd.Wait()
	return res, nil
}

func atoi(s string) int { v, _ := strconv.Atoi(strings.TrimSpace(s)); return v }
func atof(s string) float64 { v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64); return v }

// splitCSVLine handles simple comma-separated values that may include extra spaces.
func splitCSVLine(s string) []string {
	reader := csv.NewReader(strings.NewReader(s))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	rec, err := reader.Read()
	if err != nil {
		// fallback
		return strings.Split(s, ",")
	}
	return rec
}

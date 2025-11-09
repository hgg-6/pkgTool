package gopsutilx

import (
	"context"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"time"
)

type SystemLoad struct {
}

func NewSystemLoad() *SystemLoad {
	return &SystemLoad{}
}

// DiskTotals 获取总磁盘信息
func (s *SystemLoad) DiskTotals() ([]string, error) {
	partitions, err := disk.Partitions(false) // false 表示不包括只读文件系统
	if err != nil {
		return nil, err
	}
	var total []string
	for _, p := range partitions {
		total = append(total, p.String())
	}
	return total, nil
}

// Usage 空间结构体使用情况
//   - name: 磁盘/内存等名称/其他描述
//   - total: 总空间，默认单位为字节
//   - used: 已使用空间，默认单位为字节
//   - usable: 可用空间，默认单位为字节
//   - usedPercent: 使用百分比,使用率，%0-100, 类型float64
type Usage struct {
	Name        string  // 磁盘名称
	Total       uint64  // 总空间
	Used        uint64  // 已使用空间
	Usable      uint64  // 可用空间
	UsedPercent float64 // 使用百分比,使用率
}

// DiskUsage 获取磁盘使用情况
//   - name: 磁盘名称
func (s *SystemLoad) DiskUsage(name []string) ([]Usage, error) {
	var diskUsage []Usage
	for k, v := range name {
		usage, err := disk.Usage(v)
		if err != nil {
			return diskUsage, err
		}
		disks := Usage{Name: "磁盘" + name[k], Total: usage.Total, Used: usage.Used, Usable: usage.Free, UsedPercent: usage.UsedPercent}
		diskUsage = append(diskUsage, disks)
	}
	return diskUsage, nil
}

// MemUsage 获取内存使用情况
func (s *SystemLoad) MemUsage() (Usage, error) {
	var usage Usage
	v, err := mem.VirtualMemory()
	if err != nil {
		return usage, err
	}
	usage = Usage{Name: "内存", Total: v.Total, Used: v.Used, Usable: v.Available, UsedPercent: v.UsedPercent}
	return usage, nil
}

// CpuUsage 获取CPU使用情况
func (s *SystemLoad) CpuUsage() ([]float64, error) {
	percent, err := cpu.Percent(0, true) // 第二个参数为 true 表示 per-CPU
	if err != nil {
		return nil, err
	}

	return percent, nil
}

// CpuAllUsage 获取整体 CPU 使用率（阻塞约 1 秒）
//   - interval: 获取 CPU 使用率的间隔时间，默认为 0，表示阻塞约 1 秒
func (s *SystemLoad) CpuAllUsage(interval ...time.Duration) (float64, error) {
	var it time.Duration = 0
	if len(interval) > 0 {
		it = interval[0]
	}
	percent, err := cpu.Percent(it, false)
	if err != nil {
		return 0, err
	}
	return percent[0], nil
}

// CpuInfo 获取 CPU 信息
func (s *SystemLoad) CpuInfo() ([]cpu.InfoStat, error) {
	infos, err := cpu.Info()
	if err != nil {
		var info []cpu.InfoStat
		return info, err
	}
	return infos, nil
}

// SystemLoad 获取系统负载【根据cpu、内存使用情况，给出综合评分结果】
//   - 0未获取、1系统良好负载、2系统警戒负载、3系统危险负载
func (s *SystemLoad) SystemLoad() (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	l, err := load.AvgWithContext(ctx)
	if err != nil {
		return 0, err
	}
	c, err := cpu.Counts(true)
	if err != nil {
		return 0, err
	}

	var nowLoadCpu uint
	if l.Load1 < float64(c) {
		nowLoadCpu = uint(1)
	} else if l.Load1 > float64(c) || l.Load1 < float64(c)*2 {
		nowLoadCpu = uint(2)
	} else {
		// 系统危险负载
		nowLoadCpu = uint(3)
		return nowLoadCpu, nil
	}

	m, err := s.MemUsage()
	if err != nil {
		return 0, err
	}
	var nowLoadMem uint
	if m.UsedPercent < 70 {
		nowLoadMem = uint(1)
	} else if m.UsedPercent > 70 || m.UsedPercent < 90 {
		nowLoadMem = uint(2)
	} else {
		// 系统危险负载
		nowLoadMem = uint(3)
		return nowLoadMem, nil
	}

	if nowLoadCpu == 1 || nowLoadMem == 1 {
		return uint(1), nil
	} else if nowLoadCpu == 3 || nowLoadMem == 3 {
		return uint(3), nil
	} else {
		return uint(2), nil
	}
}

type HostInfo struct {
	Hostname             string `json:"hostname"`
	Uptime               uint64 `json:"uptime"`
	BootTime             uint64 `json:"bootTime"`
	Procs                uint64 `json:"procs"`
	OS                   string `json:"os"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platformFamily"`
	PlatformVersion      string `json:"platformVersion"`
	KernelVersion        string `json:"kernelVersion"`
	KernelArch           string `json:"kernelArch"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole"`
	HostID               string `json:"hostId"`
}

// HostInfo 获取系统信息【hostname、hostId、os等】
func (s *SystemLoad) HostInfo(ctx context.Context) (HostInfo, error) {
	info, err := host.InfoWithContext(ctx)
	if info != nil {
		return HostInfo{
			Hostname:             info.Hostname,
			Uptime:               info.Uptime,
			BootTime:             info.BootTime,
			Procs:                info.Procs,
			OS:                   info.OS,
			Platform:             info.Platform,
			PlatformFamily:       info.PlatformFamily,
			PlatformVersion:      info.PlatformVersion,
			KernelVersion:        info.KernelVersion,
			KernelArch:           info.KernelArch,
			VirtualizationSystem: info.VirtualizationSystem,
			VirtualizationRole:   info.VirtualizationRole,
			HostID:               info.HostID,
		}, err
	}
	return HostInfo{}, err
}

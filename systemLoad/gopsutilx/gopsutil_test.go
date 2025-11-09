package gopsutilx

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestGetProcessInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 获取负载信息
	//loadAvg, err := load.Avg()
	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		panic(err)
	}

	// 打印结果（Load1/Load5/Load15）
	fmt.Printf("1分钟负载: %.2f, 5分钟负载: %.2f, 15分钟负载: %.2f\n", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)

	// 获取系统信息
	systemInfo, err := host.Info()
	if err != nil {
		panic(err)
	}
	fmt.Println("system info: ", systemInfo)

	// 获取CPU信息
	cpuInfo, err := cpu.Info()
	if err != nil {
		panic(err)
	}
	fmt.Println("cpu info num: ", len(cpuInfo))
	for k, v := range cpuInfo {
		fmt.Printf("cpu%d info%v:\n ", k, v)
	}
}

func TestLocal(t *testing.T) {
	// 获取磁盘信息
	s := NewSystemLoad()
	total, err := s.DiskTotals()
	assert.NoError(t, err)
	log.Println(total)

	// 获取磁盘使用情况
	usage, err := s.DiskUsage([]string{"C:"})
	assert.NoError(t, err)
	log.Println("===================")
	for _, v := range usage {
		log.Printf("disk%s info%v:\n ", v.Name, v)
	}

	// 获取内存使用情况
	us, err := s.MemUsage()
	assert.NoError(t, err)
	log.Println("===================")
	log.Println("mem info: ", us)

	// 获取CPU使用情况
	c, err := s.CpuUsage()
	assert.NoError(t, err)
	log.Println("===================")
	for i, f := range c {
		log.Printf("CPU %d: %.2f%%\n", i, f)
	}

	// 获取整体 CPU 使用率
	cAll, err := s.CpuAllUsage()
	assert.NoError(t, err)
	log.Println("===================")
	log.Printf("CPU: %.2f%%\n", cAll)

	// 获取cpu信息
	cInfo, err := s.CpuInfo()
	if err == nil {
		log.Println("===================")
		for _, info := range cInfo {
			fmt.Printf("CPU 型号: %s\n", info.ModelName)
			fmt.Printf("核心数: %d\n", info.Cores)
		}
	}

	// 获取系统负载
	i, err := s.SystemLoad()
	if err == nil {
		switch i {
		case 0:
			log.Println("未获取到系统负载")
		case 1:
			log.Println("系统负载良好")
		case 2:
			log.Println("系统负载警戒")
		case 3:
			log.Println("系统负载危险")
		}
	}
	log.Println("===================")

	info, err := s.HostInfo(context.Background())
	if err == nil {
		log.Println("Host Info: ", info)
	}
}

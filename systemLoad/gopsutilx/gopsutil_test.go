package gopsutilx

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
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

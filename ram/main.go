package ram

import (
	"fmt"

	"github.com/shirou/gopsutil/mem"
)

func Main() {

	v, _ := mem.VirtualMemory()
	fmt.Println("Total:", v.Total/1024/1024, "MiB")
	fmt.Println("Available:", v.Available/1024/1024, "MiB")
	fmt.Println("Used MiB:", v.Used/1024/1024, "MiB")
	fmt.Printf("Used: %.2f%%\n", v.UsedPercent)
	fmt.Printf("Available: %.2f%%\n", 100-v.UsedPercent)
}

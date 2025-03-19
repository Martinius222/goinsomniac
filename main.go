package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/distatus/battery"
)

const barLength = 100
const percentageWidthConst = 20 

func preventSleep() *exec.Cmd {
	cmd := exec.Command("caffeinate", "-dims")
	cmd.Start()
	return cmd
}

func usageBar(usage float64) string {
	filled := int(usage / 100 * float64(barLength))
	return "‖" + strings.Repeat("■", filled) + strings.Repeat("-", barLength-filled) + "‖"
}

func formatUsage(label string, percentageWidth int, value float64) string {
	return fmt.Sprintf("%s %*.*f%%  %s", label+":", percentageWidth, 2, value, usageBar(value))
}

func getNetworkUsage() (float64, float64) {
	interfaces, _ := net.IOCounters(false)
	if len(interfaces) > 0 {
		uploadSpeed := float64(interfaces[0].BytesSent) / 1024
		downloadSpeed := float64(interfaces[0].BytesRecv) / 1024
		return uploadSpeed, downloadSpeed
	}
	return 0, 0
}

func getBatteryLevel() (float64, string) {
	batteries, err := battery.GetAll()
	if err == nil && len(batteries) > 0 {
		level := batteries[0].Current / batteries[0].Full * 100
		status := "🔋 Discharging"
		if batteries[0].State.String() == battery.Charging.String() {
			status = "⚡ Charging   "
		} else if batteries[0].State.String() == battery.Full.String() {
			status = "✅ Full    "
		}
		return level, status
	}
	return 0, "N/A"
}

func getDiskUsage() float64 {
	diskStat, _ := disk.Usage("/")
	return diskStat.UsedPercent
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	
	clearTerminal()

	startTime := time.Now()

	fmt.Println(`
 ██████╗  ██████╗     ██╗███╗   ██╗███████╗ ██████╗ ███╗   ███╗███╗   ██╗██╗ █████╗  ██████╗ 
██╔════╝ ██╔═══██╗    ██║████╗  ██║██╔════╝██╔═══██╗████╗ ████║████╗  ██║██║██╔══██╗██╔════╝ 
██║  ███╗██║   ██║    ██║██╔██╗ ██║███████╗██║   ██║██╔████╔██║██╔██╗ ██║██║███████║██║      
██║   ██║██║   ██║    ██║██║╚██╗██║╚════██║██║   ██║██║╚██╔╝██║██║╚██╗██║██║██╔══██║██║      
╚██████╔╝╚██████╔╝    ██║██║ ╚████║███████║╚██████╔╝██║ ╚═╝ ██║██║ ╚████║██║██║  ██║╚██████╗ 
 ╚═════╝  ╚═════╝     ╚═╝╚═╝  ╚═══╝╚══════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝  ╚═══╝╚═╝╚═╝  ╚═╝ ╚═════╝ 
	`)

	fmt.Println("\nGoInsomniac is running... preventing sleep mode!")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	sleepProcess := preventSleep()
	defer sleepProcess.Process.Kill()

	for {
		select {
		case <-stop:
			fmt.Println("\nGoInsomniac is shutting down...")
			return
		default:
			cpuPercent, _ := cpu.Percent(0, false)
			vmStat, _ := mem.VirtualMemory()
			uptime := time.Since(startTime)
			currentTime := time.Now().Format("15:04:05")
			uploadSpeed, downloadSpeed := getNetworkUsage()
			batteryLevel, batteryStatus := getBatteryLevel()
			diskUsage := getDiskUsage()
			
			fmt.Print("\033[H\033[10B") 

			fmt.Println("\n⏳ Running Time:", uptime.Round(time.Second))
			fmt.Println("⏰ Current Time:", currentTime)
			fmt.Println()
			fmt.Println(formatUsage("💻 CPU", 20, cpuPercent[0]))
			fmt.Println(formatUsage("🖥  RAM", 20, vmStat.UsedPercent))
			fmt.Println(formatUsage("💾 Disk", 19, diskUsage))
			fmt.Printf("\n🌐 Upload:   %*.*f KB/s\n", 19, 2, uploadSpeed)
			fmt.Printf("🌐 Download: %*.*f KB/s\n", 19, 2, downloadSpeed)
			fmt.Printf("\n🔋 Battery:  %*.*f%%  %s\n", 15, 2, batteryLevel, batteryStatus)

			time.Sleep(1 * time.Second)
		}
	}
}

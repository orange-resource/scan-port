package ip

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type PortInfo struct {
	Ip        string
	Port      int
	Available bool
}

func ScanPort(result chan PortInfo, ip string, startPort int, endPort int) {
	for i := startPort; i <= endPort; i++ {
		go scan(result, ip, i)
	}
}

func scan(result chan PortInfo, ip string, port int) {
	con, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(port), 60*time.Second)
	var available bool
	if err != nil {
		available = true
	} else {
		defer con.Close()
		available = false
	}
	result <- PortInfo{
		Ip:        ip,
		Port:      port,
		Available: available,
	}
}

func FindProcess(port int) []string {
	goos := runtime.GOOS
	if goos == "windows" {
		cmdStr := fmt.Sprintf("netstat -ano -p tcp | findstr %d", port)
		cmd := exec.Command("cmd", "/c", cmdStr)
		output, err := cmd.Output()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(output))
			return []string{}
		}
		fmt.Println(string(output))
		return windowsOutputHandle(string(output))
	} else {
		cmd := exec.Command("lsof", "-i", "tcp:"+strconv.Itoa(port))
		output, err := cmd.Output()
		if err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + string(output))
			return []string{}
		}
		fmt.Println(string(output))
		return macOutputHandle(string(output))
	}
}

func KillProcess(pid string) {
	goos := runtime.GOOS
	if goos == "windows" {
		cmdStr := fmt.Sprintf("taskkill /pid %s -f", pid)
		cmd := exec.Command("cmd", "/c", cmdStr)
		cmd.Run()
	} else {
		cmd := exec.Command("kill", "-9", pid)
		cmd.Run()
	}
}

func macOutputHandle(output string) []string {
	outputArr := strings.Split(output, "\n")
	pidArr := []string{}
	for i := 0; i < len(outputArr); i++ {
		if i != 0 {
			outputText := outputArr[i]
			if strings.Contains(outputText, "(LISTEN)") {
				for {
					if strings.Contains(outputText, "  ") {
						outputText = strings.ReplaceAll(outputText, "  ", " ")
					} else {
						break
					}
				}
				processTextArr := strings.Split(outputText, " ")
				pidArr = append(pidArr, processTextArr[1])
			}
		}
	}
	return pidArr
}

func windowsOutputHandle(output string) []string {
	outputArr := strings.Split(output, "\n")
	pidArr := []string{}
	for i := 0; i < len(outputArr); i++ {
		outputText := outputArr[i]
		if strings.Contains(outputText, "LISTENING") {
			for {
				if strings.Contains(outputText, "  ") {
					outputText = strings.ReplaceAll(outputText, "  ", " ")
				} else {
					break
				}
			}
			processTextArr := strings.Split(outputText, " ")
			pidArr = append(pidArr, processTextArr[len(processTextArr)-1])
		}
	}
	return pidArr
}

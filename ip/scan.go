package ip

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var workCount int64

type PortInfo struct {
	Ip        string
	Port      int
	Available bool
}

func ScanPort(result chan PortInfo, ip string, startPort int, endPort int) {
	println("scan port: ", ip, startPort, endPort)

	//for i := startPort; i <= endPort; i++ {
	//	go scan(result, ip, i)
	//}

	var limit int64 = 1000
	if runtime.GOOS != "windows" {
		limit = 50
	}

	println("limit: ", limit)

	workCount = 0
	i := startPort
	for {
		wc := atomic.LoadInt64(&workCount)
		if wc < limit {
			atomic.AddInt64(&workCount, 1)
			go scan(result, ip, i)
			i += 1
		} else {
			time.Sleep(10)
		}
		if i > endPort {
			break
		}
	}
}

func scan(result chan PortInfo, ip string, port int) {
	con, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(port), 15*time.Second)
	var available bool
	if err != nil {
		fmt.Println(err)
		available = true
	} else {
		con.Close()
		available = false
	}
	result <- PortInfo{
		Ip:        ip,
		Port:      port,
		Available: available,
	}
	atomic.AddInt64(&workCount, -1)
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

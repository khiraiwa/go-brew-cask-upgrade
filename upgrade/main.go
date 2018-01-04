package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/blang/semver"
)

// GREEN 緑色の制御文字
const GREEN = "\033[92m"

// YELLOW 黄色の制御文字
const YELLOW = "\033[93m"

// RED 赤色の制御文字
const RED = "\033[91m"

// ENDC 制御文字の終端
const ENDC = "\033[0m"

func printOK(msg string) {
	fmt.Printf("%s%s%s\n", GREEN, msg, ENDC)
}

func printInfo(msg string) {
	fmt.Printf("%s%s%s\n", YELLOW, msg, ENDC)
}

func printNG(msg string) {
	fmt.Printf("%s%s%s\n", RED, msg, ENDC)
}

type Info struct {
	Name                    string
	LatestVersion           string
	LocalInstalledDirs      []string
	LocalLatestInstalledDir string
	LocalLatestVersion      string
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	fmt.Printf("NumCPU=%d\n", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("=== START brew cask upgrade task ===")
	out, err := exec.Command("brew", "cask", "list").Output()
	if err != nil {
		printNG(err.Error())
		return
	}

	caskListStr := string(out)
	printOK(caskListStr)
	caskListArr := strings.Split(caskListStr, "\n")

	var wg sync.WaitGroup
	ch := make(chan string, 1)
	retInfo := make(chan Info, len(caskListArr))

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				s, ok := <-ch
				//;fmt.Println(ok)
				//fmt.Println(s)
				if !ok {
					break
				}
				out, err := exec.Command("brew", "cask", "info", s).Output()
				log.Println(string(out))
				if err != nil {
					printNG(err.Error())
					wg.Done()
					return
				}

				lines := strings.Split(string(out), "\n")
				nameVersion := strings.Split(lines[0], ": ")
				name := nameVersion[0]
				latestVersion := strings.Trim(nameVersion[1], " ")

				// TODO 複数あったときにどうなるかまだ不明。わかったらちゃんと書く
				localInstalledDirs := []string{}
				for i := 2; ; i++ {
					localInstalledDirs = append(localInstalledDirs, lines[i])
					if strings.Contains(lines[i+1], "From: ") || strings.Contains(lines[i+1], "==>") {
						break
					}
				}
				localLatestInstalledDir := strings.Split(localInstalledDirs[len(localInstalledDirs)-1], " ")[0]
				localLatestVersions := strings.Split(localLatestInstalledDir, "/")
				localLatestVersion := strings.Trim(strings.Split(localLatestVersions[len(localLatestVersions)-1], " ")[0], " ")

				retInfo <- Info{
					Name:                    name,
					LatestVersion:           latestVersion,
					LocalInstalledDirs:      localInstalledDirs,
					LocalLatestInstalledDir: localLatestInstalledDir,
					LocalLatestVersion:      localLatestVersion,
				}
				wg.Done()
			}
		}()
	}
	for i, r := range caskListArr {
		_ = i
		if r == "\n" {
			return
		}
		wg.Add(1)
		ch <- r
	}

	wg.Wait()

	retInfos := []Info{}
	i := 0
	for i < len(retInfo) {
		ret, ok := <-retInfo
		_ = ok
		retInfos = append(retInfos, ret)
		// logger.Println(ret)
		if ret.LocalLatestVersion == "latest" {
			//logger.Println("latest!")
			// force installしたほうがいいかは相談かな。
			continue
		}
		// 不完全なsemantic versioning
		localLatestVersion := ret.LocalLatestVersion
		localVer, err := semver.Make(localLatestVersion)
		localForth := 0
		if localVer.String() == "0.0.0" {
			dotNum := strings.Count(localLatestVersion, ".")
			if dotNum == 0 {
				localLatestVersion += ".0.0"
			} else if dotNum == 1 {
				localLatestVersion += ".0"
			}
			if strings.Count(localLatestVersion, ",") > 0 {
				localLatestVersion = strings.Split(localLatestVersion, ",")[0]
			}
			if strings.Count(localLatestVersion, "-") > 0 {
				localLatestVersion = strings.Split(localLatestVersion, "-")[0]
			}
			if strings.Count(localLatestVersion, "_") > 0 {
				localLatestVersion = strings.Split(localLatestVersion, "_")[0]
			}

			//logger.Printf("%s\n ", localLatestVersion)
			localLatestVersions := strings.Split(localLatestVersion, ".")
			for i, r := range localLatestVersions {
				_ = i
				a, b := strconv.ParseInt(r, 10, 0)
				_ = b
				localLatestVersions[i] = strconv.FormatInt(a, 10)
			}

			if len(localLatestVersions) == 4 {
				localLatestVersions = localLatestVersions[0:2]
				a, b := strconv.Atoi(localLatestVersions[3])
				_ = b
				localForth = a
			}

			localLatestVersion = strings.Join(localLatestVersions, ".")
			localVer, err = semver.Make(localLatestVersion)
		}
		_ = err
		latestForth := 0
		latestVersion := ret.LatestVersion
		latestVer, err := semver.Make(latestVersion)
		if latestVer.String() == "0.0.0" {
			dotNum := strings.Count(latestVersion, ".")
			if dotNum == 0 {
				latestVersion += ".0.0"
			} else if dotNum == 1 {
				latestVersion += ".0"
			}
			if strings.Count(latestVersion, ",") > 0 {
				latestVersion = strings.Split(latestVersion, ",")[0]
			}
			if strings.Count(latestVersion, "-") > 0 {
				latestVersion = strings.Split(latestVersion, "-")[0]
			}
			if strings.Count(latestVersion, "_") > 0 {
				latestVersion = strings.Split(latestVersion, "_")[0]
			}
			latestVersions := strings.Split(latestVersion, ".")
			for i, r := range latestVersions {
				_ = i
				a, b := strconv.ParseInt(r, 10, 0)
				_ = b
				latestVersions[i] = strconv.FormatInt(a, 10)
			}

			if len(latestVersions) == 4 {
				latestVersions = latestVersions[0:2]
				latestForth, _ = strconv.Atoi(latestVersions[3])
			}

			latestVersion = strings.Join(latestVersions, ".")
			latestVer, err = semver.Make(latestVersion)
		}

		_ = err
		if latestVer.String() == "0.0.0" {
			logger.Println(ret.Name + ":" + ret.LatestVersion)
			continue
		}
		//logger.Printf("%s : %s\n ", localVer.String(), latestVer.String())
		//logger.Printf("%s : %s\n ", ret.LocalLatestVersion, ret.LatestVersion)
		if localVer.LT(latestVer) {
			logger.Println("less!")
			out, err = exec.Command("brew", "cask", "install", ret.Name, "--force").Output()
			logger.Printf("%s\n", out)
		} else if localVer.EQ(latestVer) {
			if localForth < latestForth {
				logger.Println("less2!")
				out, err = exec.Command("brew", "cask", "install", ret.Name, "--force").Output()
				logger.Printf("%s\n", out)
			}
		}

	}
	close(retInfo)
	close(ch)

	out, err = exec.Command("brew", "cask", "cleanup").Output()
	_ = err
	logger.Printf("%s\n", out)

	return
}

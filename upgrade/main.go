package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"

	shellwords "github.com/mattn/go-shellwords"
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

func init() {
	logger.Printf("NumCPU=%d\n", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func runCmdStr(cmdstr string) error {
	logger.Println(cmdstr)
	c, err := shellwords.Parse(cmdstr)
	if err != nil {
		return err
	}

	var out = []byte{}
	switch len(c) {
	case 0:
		return nil
	case 1:
		out, err = exec.Command(c[0]).Output()
	default:
		out, err = exec.Command(c[0], c[1:]...).Output()
	}
	logger.Printf("%s\n", out)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	logger.Println("=== START brew cask upgrade task ===")
	err := runCmdStr("brew update")
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = runCmdStr("brew upgrade")
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = runCmdStr("brew cleanup")
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = runCmdStr("brew cask outdated")
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = runCmdStr("brew cask upgrade")
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = runCmdStr("brew cask cleanup")
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Println("=== FINISH brew cask upgrade task ===")
}

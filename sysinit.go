package docker

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Setup networking
func setupNetworking(gw string) {
	if gw == "" {
		return
	}
	if _, err := ip("route", "add", "default", "via", gw); err != nil {
		log.Fatalf("Unable to set up networking: %v", err)
	}
}

// Takes care of dropping privileges to the desired user
func changeUser(u string) {
	if u == "" {
		return
	}
	//userent, err := user.LookupId(u)
	//if err != nil {
	//	userent, err = user.Lookup(u)
	//}
	//if err != nil {
	//	log.Fatalf("Unable to find user %v: %v", u, err)
	//}

	uid,_ := strconv.Atoi(u)
	//if err != nil {
	//	log.Fatalf("Invalid uid: %v", userent.Uid)
	//}
	gid,_ := strconv.Atoi("1001")
	//if err != nil {
	//	log.Fatalf("Invalid gid: %v", userent.Gid)
	//}

	if err := syscall.Setgid(gid); err != nil {
		log.Fatalf("setgid failed: %v", err)
	}
	if err := syscall.Setuid(uid); err != nil {
		log.Fatalf("setuid failed: %v", err)
	}
}

// Clear environment pollution introduced by lxc-start
func cleanupEnv(env ListOpts) {
	os.Clearenv()
	for _, kv := range env {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 1 {
			parts = append(parts, "")
		}
		if parts[0] == "container" {
			continue
		}
		os.Setenv(parts[0], parts[1])
	}
}

func executeProgram(name string, args []string) {
	path, err := exec.LookPath(name)
	if err != nil {
		log.Printf("Unable to locate %v", name)
		os.Exit(127)
	}

        rlim := new(syscall.Rlimit);
        syscall.Getrlimit(6, rlim);
        (*rlim).Max=512;
        (*rlim).Cur=512;
        if err:=syscall.Setrlimit(6, rlim); err!=nil {
          panic(err)
        }
        cpuLim := new (syscall.Rlimit);
        (*cpuLim).Max=2;
        (*cpuLim).Cur=1;
        if err:=syscall.Setrlimit(syscall.RLIMIT_CPU, cpuLim); err!=nil {
          panic(err)
        }
        //startpid, _, err := syscall.StartProcess(path,args,procAttr);
	if err := syscall.Exec(path, args, os.Environ()); err != nil {
        //if err!=nil {
		panic(err)
	}
}

// Sys Init code
// This code is run INSIDE the container and is responsible for setting
// up the environment before running the actual process
func SysInit() {
	if len(os.Args) <= 1 {
		fmt.Println("You should not invoke docker-init manually")
		os.Exit(1)
	}
	var u = flag.String("u", "", "username or uid")
	var gw = flag.String("g", "", "gateway address")

	var flEnv ListOpts
	flag.Var(&flEnv, "e", "Set environment variables")

	flag.Parse()

	cleanupEnv(flEnv)
	setupNetworking(*gw)
	changeUser(*u)
	executeProgram(flag.Arg(0), flag.Args())
}

package client

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// K8STestSetup is instantiated given the pods and services we must wait for
type K8STestSetup struct {
	WaitForPods         []string
	PortForwardServices map[string]int

	forwards map[string]int

	cmd *exec.Cmd

	in  io.WriteCloser
	out io.ReadCloser
	er  io.ReadCloser
}

var setupCMD = flag.String("setup", "", "the path to the it setup command")

// TestMain will execute the provided setup command, wait for configured pods and services to be
// ready, and then forwards service logs to test output.  On completion, signals to the it_sidecar
// to teardown the test namespace
func (s *K8STestSetup) TestMain(m *testing.M) {
	s.forwards = make(map[string]int)
	wg := new(sync.WaitGroup)
	wg.Add(2) // there will be 2 goroutines, one reading stdout and one reading stdin
	os.Exit(func() int {
		flag.Parse()
		// Defer sidecar process tear-down.
		defer func() {
			//Closing standard-in pipe signals sidecar process to exit,
			s.in.Close()
			wg.Wait() // Wait for reader goroutines to actually finish
			if err := s.cmd.Wait(); err != nil {
				log.Fatal(err)
			}
		}()
		s.before(wg)
		// Run tests.
		return m.Run()
	}())
}

func (s *K8STestSetup) GetServiceLocalPort(serviceName string) int {
	return s.forwards[serviceName]
}

func (s *K8STestSetup) before(wg *sync.WaitGroup) {
	log.Printf("setup command: %s\n", *setupCMD)

	args := make([]string, 0)
	for _, app := range s.WaitForPods {
		args = append(args, fmt.Sprintf("-waitforapp=%s", app))
	}
	for service, port := range s.PortForwardServices {
		args = append(args, fmt.Sprintf("-portforward=%s:%d", service, port))
	}

	s.cmd = exec.Command(*setupCMD, args...)

	var err error
	// Open and start reading stderr in a new goroutine. StderrPipe will be closed automatically by the call to Wait
	// so we do not need to close this ourselves.  We must also guarantee that all reads on this pipe are completed
	// before calling wait, so the goroutines below must be canceled before the defered teardown above
	if s.er, err = s.cmd.StderrPipe(); err != nil {
		log.Fatal(err)
	}
	go func() {
		rd := bufio.NewReader(s.er)
		for {
			str, err := rd.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			log.Print(str)
		}
		wg.Done()
	}()

	//Open stdin and stdout
	if s.out, err = s.cmd.StdoutPipe(); err != nil {
		log.Fatal(err)
	}
	if s.in, err = s.cmd.StdinPipe(); err != nil {
		log.Fatal(err)
	}

	//Start the sidecar process
	if err := s.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	//Wait for all pods to be ready
	rd := bufio.NewReader(s.out)
waitForReady:
	for {
		str, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(str)
		if strings.HasPrefix(str, "FORWARD") {
			// remove the "FORWARD " prefix, and any trailing space, split on ":"
			parts := strings.Split(strings.TrimSpace(str[8:]), ":")
			localPort, _ := strconv.Atoi(parts[2])
			s.forwards[parts[0]] = localPort
		}
		if "READY\n" == str {
			break waitForReady
		}
	}
	//Start reading stdout in a new goroutine
	go func() {
		for {
			str, err := rd.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			log.Print(str)
		}
		wg.Done()
	}()

}


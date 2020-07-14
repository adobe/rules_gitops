package client

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"testing"
)

// K8STestSetup is instantiated given the pods and services we must wait for
type K8STestSetup struct {
	WaitForPods         []string
	PortForwardServices map[string]int

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
	os.Exit(func() int {
		flag.Parse()

		// Defer sidecar process tear-down.
		defer func() {
			//Closing standard-in pipe signals sidecar process to exit,
			s.in.Close()

			if err := s.cmd.Wait(); err != nil {
				log.Fatal(err)
			}

			s.out.Close()
			s.er.Close()
		}()

		s.before()
		// Run tests.
		return m.Run()
	}())
}

func (s *K8STestSetup) before() {
	fmt.Printf("setup command: %s\n", *setupCMD)

	args := make([]string, 0)
	for _, app := range s.WaitForPods {
		args = append(args, fmt.Sprintf("-waitforapp=%s", app))
	}
	for service, port := range s.PortForwardServices {
		args = append(args, fmt.Sprintf("-portforward=%s:%d", service, port))
	}

	s.cmd = exec.Command(*setupCMD, args...)

	var err error
	//Open and start reading stderr in a new goroutine
	if s.er, err = s.cmd.StderrPipe(); err != nil {
		log.Fatal(err)
	}
	go func() {
		rd := bufio.NewReader(s.er)
		str, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		log.Println(str)
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
		if "READY\n" == str {
			break waitForReady
		} else {
			fmt.Print(str)
		}
	}
	//Start reading stdout in a new goroutine
	go func() {
		str, err := rd.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		log.Println(str)
	}()

}

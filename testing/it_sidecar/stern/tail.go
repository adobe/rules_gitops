//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package stern

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type Tail struct {
	Namespace     string
	PodName       string
	ContainerName string
	req           *rest.Request
	closed        chan struct{}
}

// NewTail returns a new tail for a Kubernetes container inside a pod
func NewTail(namespace, podName, containerName string) *Tail {
	return &Tail{
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		closed:        make(chan struct{}),
	}
}

// Start starts tailing
func (t *Tail) Start(ctx context.Context, i v1.PodInterface) {

	go func() {
		fmt.Fprintf(os.Stderr, "+ %s/%s\n", t.PodName, t.ContainerName)

		req := i.GetLogs(t.PodName, &corev1.PodLogOptions{
			Follow:     true,
			Timestamps: true,
			Container:  t.ContainerName,
		})

		stream, err := req.Stream()
		if err != nil {
			log.Printf("Error opening stream to %s/%s/%s: %s", t.Namespace, t.PodName, t.ContainerName, err)
			return
		}
		defer stream.Close()

		go func() {
			<-t.closed
			stream.Close()
		}()

		reader := bufio.NewReader(stream)

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}

			str := string(line)
			t.Print(str)
		}
	}()

	go func() {
		<-ctx.Done()
		close(t.closed)
	}()
}

// Close stops tailing
func (t *Tail) Close() {
	fmt.Fprintf(os.Stderr, "Log finished %s\n", t.PodName)
	close(t.closed)
}

// Print prints a color coded log message with the pod and container names
func (t *Tail) Print(msg string) {
	fmt.Fprintf(os.Stderr, "[%s/%s]: %s", t.PodName, t.ContainerName, msg)
}

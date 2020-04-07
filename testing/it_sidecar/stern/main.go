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
	"context"
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// Run starts the main run loop
func Run(ctx context.Context, namespace string, clientset *kubernetes.Clientset) error {

	added, removed, err := Watch(ctx, clientset.CoreV1().Pods(namespace), regexp.MustCompile(".*"), regexp.MustCompile(".*"), RUNNING, labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to set up watch: %v", err)
	}

	tails := make(map[string]*Tail)

	go func() {
		for p := range added {
			id := p.GetID()
			if tails[id] != nil {
				continue
			}

			tail := NewTail(p.Namespace, p.Pod, p.Container)
			tails[id] = tail

			tail.Start(ctx, clientset.CoreV1().Pods(p.Namespace))
		}
	}()

	go func() {
		for p := range removed {
			id := p.GetID()
			if tails[id] == nil {
				continue
			}
			tails[id].Close()
			delete(tails, id)
		}
	}()

	<-ctx.Done()

	return nil
}

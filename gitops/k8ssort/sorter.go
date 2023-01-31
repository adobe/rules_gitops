// Package k8ssort sorts a YAML stream and produces output in a sorted format.
// Inspired by https://github.com/grafana/tanka/blob/1bf5e549f35f9d00b11cf34786e4b67c32e72e4f/pkg/process/sort.go#L52
package k8ssort

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sort"

	"gopkg.in/yaml.v2"
)

var kindOrder = []string{
	"Namespace",
	"NetworkPolicy",
	"ResourceQuota",
	"LimitRange",
	"PodSecurityPolicy",
	"PodDisruptionBudget",
	"ServiceAccount",
	"Secret",
	"ConfigMap",
	"StorageClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"CustomResourceDefinition",
	"ClusterRole",
	"ClusterRoleList",
	"ClusterRoleBinding",
	"ClusterRoleBindingList",
	"GatewayController",
	"Role",
	"RoleList",
	"RoleBinding",
	"RoleBindingList",
	"Service",
	"DaemonSet",
	"Pod",
	"ReplicationController",
	"ReplicaSet",
	"Deployment",
	"Gateway",
	"HorizontalPodAutoscaler",
	"StatefulSet",
	"Job",
	"CronJob",
	"Ingress",
	"HTTPRoute",
	"APIService",
}

var lookup map[string]int = map[string]int{}

func Sort(stream io.Reader, filter string, reverse bool) ([]byte, error) {
	type kindRes struct {
		Kind string
		Rest interface{}
	}
	vals := []kindRes{}
	d := yaml.NewDecoder(stream)
	for {
		var value map[string]interface{}
		err := d.Decode(&value)
		if err != nil {
			if err != io.EOF {
				log.Panicf("error: %v", err)
			}
			break
		}
		rawKind, ok := value["kind"]
		if !ok {
			return nil, fmt.Errorf("invalid k8s resource: missing kind: %v", value)
		}
		kind, ok := rawKind.(string)
		if !ok {
			return nil, fmt.Errorf("invalid k8s resource: kind not a string: %v", value)
		}
		if filter != "" && kind != filter {
			continue
		}
		vals = append(vals, kindRes{Kind: kind, Rest: value})
	}
	sort.SliceStable(vals, func(i, j int) bool {
		akind, aok := lookup[vals[i].Kind]
		bkind, bok := lookup[vals[j].Kind]

		if aok && bok {
			return akind < bkind
		} else if aok {
			return true
		} else if bok {
			return false
		} else {
			return vals[i].Kind < vals[j].Kind
		}
	})
	if reverse {
		for i := 0; i < len(vals)/2; i++ {
			vals[i], vals[len(vals)-i-1] = vals[len(vals)-i-1], vals[i]
		}
	}
	buf := bytes.Buffer{}
	e := yaml.NewEncoder(&buf)
	for _, v := range vals {
		if err := e.Encode(v.Rest); err != nil {
			return nil, fmt.Errorf("error encoding yaml: %v", v.Rest)
		}
	}
	return buf.Bytes(), nil
}

func init() {
	for i, v := range kindOrder {
		lookup[v] = i
	}
}

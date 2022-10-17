package filter

import (
	"fmt"
	"io"
	"log"
	"strings"

	yamlenc "github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/util/jsonpath"
)

// ReplacePDWithEmptyDirs reads yaml or json stream from in, deserialize it and replace references to all PVC volumes with EmptyDir
// remove all PersistemtVolumeClaim objects
// then serialize it back into out stream.
func ReplacePDWithEmptyDirs(in io.Reader, out io.Writer) error {
	decoder := yaml.NewYAMLOrJSONDecoder(in, 1024)
	var err error
	firstObj := true
	for err == nil || isEmptyYamlError(err) {
		var obj unstructured.Unstructured
		err = decoder.Decode(&obj)
		if err != nil {
			continue
		}
		if obj.GetName() == "" {
			return fmt.Errorf("Missing metadata.name in object %v", obj)
		}
		if obj.GetKind() == "" {
			return fmt.Errorf("Missing kind in object %v", obj)
		}
		if obj.GetKind() == "PersistentVolumeClaim" {
			continue // skip all PVCs
		}
		if obj.GetKind() == "Ingress" {
			continue // skip all Ingress objects
		}
		if obj.GetKind() == "StatefulSet" && obj.GetAPIVersion() == "apps/v1" {
			var statefulset appsv1.StatefulSet
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &statefulset)
			if err != nil {
				return fmt.Errorf("Unable to decode statefulset object %s", obj.GetName())
			}
			processStatefulSet(&statefulset)
			obj.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&statefulset)
			if err != nil {
				return fmt.Errorf("Unable to convert statefulset to unstructured: %v", err)
			}
			delete(obj.Object, "status")
		}
		if obj.GetKind() == "Certificate" {
			findAndReplaceIssuerName(obj.Object)
		}

		findAndReplacePVC(obj.Object)
		buf, err := yamlenc.Marshal(obj.Object)
		if err != nil {
			return fmt.Errorf("Unable to marshal object %v", obj.Object)
		}
		if firstObj {
			firstObj = false
		} else {
			_, err = out.Write([]byte("---\n"))
			if err != nil {
				return err
			}
		}
		_, err = out.Write(buf)
		if err != nil {
			return err
		}

	}
	if err != io.EOF {
		return err
	}
	return nil
}

func isEmptyYamlError(err error) bool {
	return strings.Contains(err.Error(), "is missing in 'null'")
}

var emptydirSpec = make(map[string]interface{})

/*
 findAndReplaceTag replaces the image tags inside one object
 It searches the object for container session
 then loops though all images inside containers session, finds matched ones and update the tag name
*/
func findAndReplacePVC(obj map[string]interface{}) {
	found := false
	_, found = obj["persistentVolumeClaim"]
	if found {
		delete(obj, "persistentVolumeClaim")
		obj["emptyDir"] = emptydirSpec
	}
	if !found {
		findPVC(obj)
	}
}

func findPVC(obj map[string]interface{}) {
	for key := range obj {
		switch typedV := obj[key].(type) {
		case map[string]interface{}:
			findAndReplacePVC(typedV)
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if ok {
					findAndReplacePVC(typedItem)
				}
			}
		}
	}
}

func processStatefulSet(obj *appsv1.StatefulSet) {
	if len(obj.Spec.VolumeClaimTemplates) == 0 {
		return
	}
	//collect existing volumes
	existingVolumes := make(map[string]int)
	for i, v := range obj.Spec.Template.Spec.Volumes {
		existingVolumes[v.Name] = i
	}
	for _, vct := range obj.Spec.VolumeClaimTemplates {
		name := vct.GetObjectMeta().GetName()
		vol := v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		}
		if storage, ok := vct.Spec.Resources.Requests["storage"]; ok {
			vol.VolumeSource.EmptyDir.SizeLimit = &storage
		}
		if i, ok := existingVolumes[name]; ok {
			obj.Spec.Template.Spec.Volumes[i] = vol
		} else {
			obj.Spec.Template.Spec.Volumes = append(obj.Spec.Template.Spec.Volumes, vol)
		}
	}
	obj.Spec.VolumeClaimTemplates = nil
}

func findAndReplaceIssuerName(obj map[string]interface{}) {
	j := jsonpath.New("cert_issuer_name")
	err := j.Parse(`{.spec.issuerRef}`)
	if err != nil {
		log.Fatalln("Unable to parse jsonpath: ", err)
	}
	res, err := j.FindResults(obj)
	if err != nil {
		log.Println("Unable to find jsonpath: ", err)
		return
	}
	issuerRef := res[0][0].Interface().(map[string]interface{})
	if issuerRef["name"] == "letsencrypt-prod" {
		issuerRef["name"] = "letsencrypt-staging"
	}
}

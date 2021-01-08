package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	machyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type filterPath struct {
	keep bool
	path []string
}

var defaultFilters = mustParseFilters("metadata.managedFields,metadata.selfLink,metadata.annotations.kubectl\\.kubernetes\\.io/last-applied-configuration", false)

func main() {
	// parse the rest of the command to kubectl and run it
	output, outputType, filters := parseAndRunCommand()

	obj, err := unmarshalUnstructured(output)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	filter(obj, filters)

	filteredOutput, err := marshal(outputType, obj)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "%s\n", string(filteredOutput))
}

func parseAndRunCommand() ([]byte, string, []filterPath) {
	args, outputType, filterExpr := getKubectlArgs()

	filters, err := parseFilters(filterExpr, true)
	if err != nil {
		fmt.Fprint(os.Stdout, err.Error())
		os.Exit(1)
	}

	// check that the global output type was set, if it's not set we can not decode the secret
	if outputType == "" {
		fmt.Fprintf(os.Stdout, "please set -o flag to json or yaml\n")
		os.Exit(1)
	}

	cmd := exec.Command("kubectl", args...)

	var output, errb bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stdout, errb.String())
		os.Exit(1)
	}

	return output.Bytes(), outputType, append(defaultFilters, filters...)
}

func getKubectlArgs() ([]string, string, string) {
	skip := 1
	var outputType, filterExpression string
	lastOutputOption := false
	lastFilterOption := false
	for i, arg := range os.Args[1:] {
		if arg == "-o" {
			lastOutputOption = true
		} else if lastOutputOption {
			lastOutputOption = false
			if arg == "json" || arg == "yaml" {
				outputType = arg
				break
			}
		} else if i == 0 && arg == "--filter" {
			lastFilterOption = true
			skip = 2
		} else if lastFilterOption {
			lastFilterOption = false
			filterExpression = arg
			skip = 3
		} else if arg == "-o=json" || arg == "-ojson" {
			outputType = "json"
			break
		} else if arg == "-o=yaml" || arg == "-oyaml" {
			outputType = "yaml"
			break
		}
	}

	return os.Args[skip:], outputType, filterExpression
}

func marshal(outputType string, obj interface{}) ([]byte, error) {
	if isJSON(outputType) {
		return json.MarshalIndent(obj, "", "    ")
	}
	return yaml.Marshal(obj)
}

func unmarshalUnstructured(output []byte) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	decoder := machyaml.NewYAMLOrJSONDecoder(bytes.NewReader(output), 100)
	err := decoder.Decode(obj)
	if err != nil {
		return nil, err
	}
	return obj.Object, nil
}

func isJSON(o string) bool {
	return o == "json"
}

func filter(obj map[string]interface{}, filters []filterPath) {
	if obj == nil {
		return
	}
	if obj["kind"] == "List" {
		items := obj["items"].([]interface{})
		for _, item := range items {
			if itemObj, ok := item.(map[string]interface{}); ok {
				filterObject(itemObj, filters)
			}
		}
		return
	}
	filterObject(obj, filters)
}

func filterObject(obj map[string]interface{}, filters []filterPath) {
	var current map[string]interface{}
	workingSetObjects := map[string]map[string]interface{}{}
	workingSetKeepFields := map[string]map[string]struct{}{}

	keepField := func(path []string, k int, key string) {
		fullkey := strings.Join(path[:k], ".")
		workingSetObjects[fullkey] = current
		keepFields := workingSetKeepFields[fullkey]
		if keepFields == nil {
			keepFields = map[string]struct{}{}
			workingSetKeepFields[fullkey] = keepFields
		}
		keepFields[key] = struct{}{}
	}
	for _, fp := range filters {
		current = obj
		for i, key := range fp.path {
			if i < len(fp.path)-1 {
				sub, ok := current[key]
				if !ok || sub == nil {
					break
				}
				if fp.keep {
					keepField(fp.path, i, key)
				}
				current, ok = sub.(map[string]interface{})
				if !ok {
					break
				}
			} else {
				if !fp.keep {
					delete(current, key)
				} else {
					keepField(fp.path, len(fp.path)-1, key)
				}
			}
		}
	}

	for fullkey, obj := range workingSetObjects {
		keepFields := workingSetKeepFields[fullkey]
		for key := range obj {
			if _, ok := keepFields[key]; !ok {
				delete(obj, key)
			}
		}
	}
}

func mustParseFilters(filterExpression string, defaultKeep bool) []filterPath {
	array, err := parseFilters(filterExpression, defaultKeep)
	if err != nil {
		panic(fmt.Sprintf("Invalid filter expression: %s", err))
	}
	return array
}

func parseFilters(filterExpression string, defaultKeep bool) ([]filterPath, error) {
	if len(filterExpression) == 0 {
		return []filterPath{}, nil
	}

	parts := strings.Split(filterExpression, ",")

	var filters []filterPath
	for _, part := range parts {
		keep := defaultKeep
		if strings.HasPrefix(part, "+") {
			keep = true
			part = part[1:]
		} else if strings.HasPrefix(part, "-") {
			keep = false
			part = part[1:]
		}
		modpart := strings.ReplaceAll(part, "\\.", "~~~")
		path := strings.Split(modpart, ".")
		for i, s := range path {
			path[i] = strings.ReplaceAll(s, "~~~", ".")
		}
		filters = append(filters, filterPath{keep: keep, path: path})
	}
	return filters, nil
}

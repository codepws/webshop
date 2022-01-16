/*
Package yaml2json provides functionality to transform YAML to JSON.

See the README at https://github.com/peter-edge/go-yaml2json/blob/master/README.md for more details.
*/
package yaml2json

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v2"
)

const (
	defaultIndent = "\t"
)

// TransformOptions are the options to pass to Transform
type TransformOptions struct {
	// Pretty says to output the JSON with json.MarshalIndent.
	Pretty bool
	// Indent is the string to use for indenting. This only applies
	// if Pretty is set. The default is "\t".
	Indent string
}

// Transform transforms an YAML input and transforms it to JSON.
func Transform(p []byte, options TransformOptions) ([]byte, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(p, &yamlData); err != nil {
		return nil, err
	}
	jsonData, err := transform(yamlData)
	if err != nil {
		return nil, err
	}
	if options.Pretty {
		if options.Indent != "" {
			return json.MarshalIndent(jsonData, "", options.Indent)
		}
		return json.MarshalIndent(jsonData, "", defaultIndent)
	}
	return json.Marshal(jsonData)
}

func transform(yamlData interface{}) (interface{}, error) {
	switch yamlData.(type) {
	case map[interface{}]interface{}:
		jsonData := make(map[string]interface{})
		for key, yamlValue := range yamlData.(map[interface{}]interface{}) {
			jsonValue, err := transform(yamlValue)
			if err != nil {
				return nil, err
			}
			switch key.(type) {
			case string:
				jsonData[key.(string)] = jsonValue
			case int:
				jsonData[strconv.Itoa(key.(int))] = jsonValue
			default:
				return nil, fmt.Errorf("yaml2json: unexpected key type %T for %v", key, key)
			}
		}
		return jsonData, nil
	case []interface{}:
		yamlDataSlice := yamlData.([]interface{})
		jsonDataSlice := make([]interface{}, len(yamlDataSlice))
		for i, yamlValue := range yamlDataSlice {
			jsonValue, err := transform(yamlValue)
			if err != nil {
				return nil, err
			}
			jsonDataSlice[i] = jsonValue
		}
		return jsonDataSlice, nil
	default:
		return yamlData, nil
	}
}

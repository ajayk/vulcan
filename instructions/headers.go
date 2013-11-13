package instructions

import (
	"fmt"
	"net/http"
)

func HeadersFromObj(in interface{}) (http.Header, error) {
	inHeaders, ok := in.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Headers should be a dictionary")
	}
	headers := make(http.Header)
	for key, inHeader := range inHeaders {
		switch values := inHeader.(type) {
		case string:
			headers.Add(key, values)
		case []interface{}:
			for _, valueI := range values {
				value, ok := valueI.(string)
				if !ok {
					return nil, fmt.Errorf("Header value should be a string got unknown type: %#v in %v", valueI, in)
				}
				headers.Add(key, value)
			}
		default:
			return nil, fmt.Errorf("Unsupported header type: %T in %#v", values, in)
		}
	}
	return headers, nil
}

func AddRemoveHeadersFromDict(in map[string]interface{}) (http.Header, http.Header, error) {
	addHeadersI, exists := in["add-headers"]
	var addHeaders http.Header
	if exists {
		addHeaders, err := HeadersFromObj(addHeadersI)
		if err != nil {
			return nil, nil, err
		}
	}

	removeHeadersI, exists := in["remove-headers"]
	var removeHeaders http.Header
	if exists {
		removeHeaders, err := HeadersFromObj(removeHeadersI)
		if err != nil {
			return nil, nil, err
		}
	}
	return addHeaders, removeHeaders
}

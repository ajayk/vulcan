package instructions

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Reply struct {
	//
	Code int
	//
	Message interface{}
}

// On every request proxy asks control server what to do
// with the request, control server replies with this structure
// or rejects the request.
type Forward struct {
	// Allows proxy to fall back to the next upstream
	// if the selected upstream failed
	Failover *Failover
	// Tokens uniquely identify the requester. E.g. token can be account id or
	// combination of ip and account id. Tokens can be throttled as well.
	// The reply can have 0 or several tokens
	Tokens []*Token
	// List of upstreams that can accept this request. Load balancer will
	// choose an upstream based on the algo, e.g. random, round robin,
	// or least connections. At least one upstream is required.
	Upstreams []*Upstream
	// If supplied, headers will be added to the proxied request.
	AddHeaders    http.Header
	RemoveHeaders http.Header
}

func NewForward(
	failover *Failover,
	tokens []*Token,
	upstreams []*Upstream,
	addHeaders http.Header,
	removeHeaders http.Header) (*Forward, error) {

	if len(upstreams) <= 0 {
		return nil, fmt.Errorf("At least one upstream is required")
	}

	return &Forward{
		Failover:      failover,
		Tokens:        tokens,
		Upstreams:     upstreams,
		AddHeaders:    addHeaders,
		RemoveHeaders: removeHeaders,
	}, nil
}

func NewCommandFromObj(in interface{}) (interface{}, error) {
	obj, ok := in.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected dictionary, got %T", in)
	}
	code, exists := obj["code"]
	if exists {
		return NewReplyFromDict(obj)
	} else {
		return NewForwardFromDict(obj)
	}
}

func NewReplyFromDict(in map[string]interface{}) (interface{}, error) {
	codeI, exists := in["code"]
	if !exists {
		return nil, fmt.Errorf("Expected code")
	}
	codeF, ok := codeI.(float64)
	if !ok || codeF != float64(int(codeF)) || codeF < 0 {
		return nil, fmt.Errorf("Proper HTTP code is required")
	}

	messageI, exists := in["message"]
	if !exists {
		return nil, fmt.Errorf("Expected message")
	}
	_, err := json.Marshal(messageI)
	if err != nil {
		return nil, fmt.Errorf("Message property should be json encodeable")
	}
	return &Reply{Code: int(codeF), Message: messageI}, nil
}

func NewForwardFromDict(in map[string]interface{}) (interface{}, error) {
	upstreamsI, exists := in["upstreams"]
	if !exists {
		return nil, fmt.Errorf("Upstreams are required")
	}
	upstreamsS, ok := upstreamsI.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Upstreams: expected array, got %T", upstreamsI)
	}
	if len(upstreamsS) == 0 {
		return nil, fmt.Errorf("Upstreams: at least one is required")
	}
	upstreams := make([]*Upstream, len(upstreamsS))
	for i, upstreamI := range upstreamsS {
		u, err := NewUpstreamFromObj(upstreamI)
		if err != nil {
			return nil, err
		}
		upstreams[i] = u
	}

	failoverI, exists := in["failover"]
	var failover *Failover
	var err error
	if exists {
		failover, err = NewFailoverFromObj(failoverI)
		if err != nil {
			return nil, err
		}
	}

	addHeaders, removeHeaders, err := AddRemoveHeadersFromDict(in)
	if err != nil {
		return nil, err
	}

	return &Forward{
		Failover:      failover,
		Upstreams:     upstreams,
		AddHeaders:    addHeaders,
		RemoveHeaders: removeHeaders,
	}, nil
}

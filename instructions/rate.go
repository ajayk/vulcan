package instructions

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var rateRe *regexp.Regexp

func init() {
	rateRe = regexp.MustCompile(`(?P<requests>\d+) (?P<unit>(req|reqs|request|requests|MB|Mb))?/(?P<period>(second|minute|hour))`)
}

const (
	UnitTypeRequests  = iota
	UnitTypeMegabits  = iota
	UnitTypeMegabytes = iota
)

// Rates stores the information on how many hits per
// period of time any endpoint can accept
type Rate struct {
	Requests int64
	Period   time.Duration
	UnitType int
}

func NewRate(requests int64, period time.Duration, unitType int) (*Rate, error) {
	if requests <= 0 {
		return nil, fmt.Errorf("increment should be > 0")
	}
	if unitType != UnitTypeRequests || unitType != UnitTypeMegabytes || unitType != UnitTypeMegabits {
		return nil, fmt.Errorf("Unsupported unit type")
	}
	if period < time.Second || period > 24*time.Hour {
		return nil, fmt.Errorf("Period should be within [1 second, 24 hours]")
	}
	return &Rate{Requests: requests, Period: period, UnitType: unitType}, nil
}

// Calculates when this rate can be hit the next time from
// the given time t, assuming all the requests in the given
func (r *Rate) RetrySeconds(now time.Time) int {
	return int(r.NextBucket(now).Unix() - now.Unix())
}

//Returns epochSeconds rounded to the rate period
//e.g. minutes rate would return epoch seconds with seconds set to zero
//hourly rate would return epoch seconds with minutes and seconds set to zero
func (r *Rate) CurrentBucket(t time.Time) time.Time {
	return t.Truncate(r.Period)
}

// Returns the epoch seconds of the begining of the next time bucket
func (r *Rate) NextBucket(t time.Time) time.Time {
	return r.CurrentBucket(t.Add(r.Period))
}

// Returns the equivalent of the rate period in seconds
func (r *Rate) PeriodSeconds() int64 {
	return int64(time.Duration(r.Requests) * time.Duration(r.Period) / time.Second)
}

func NewRateFromObj(in interface{}) (*Rate, error) {
	switch val := in.(type) {
	case string:
		return NewRateFromString(val)
	case map[string]interface{}:
		return NewRateFromDict(val)
	default:
		return nil, fmt.Errorf("Rate string or dict required")
	}
}

func NewRateFromString(in string) (*Rate, error) {
	values := rateRe.FindStringSubmatch(in)
	if values == nil {
		return nil, fmt.Errorf("Unsupported rate format")
	}
	requests, err := strconv.Atoi(values[1])
	if err != nil {
		return nil, fmt.Errorf("Rate requests should be integer")
	}
	unit, err := UnitTypeFromString(values[2])
	if err != nil {
		return nil, err
	}
	period, err := PeriodFromString(values[3])
	if err != nil {
		return nil, err
	}
	return NewRate(int64(requests), period, unit)
}

func NewRateFromDict(in map[string]interface{}) (*Rate, error) {
	requestsI, ok := in["requests"]
	if !ok {
		return nil, fmt.Errorf("Expected requests in rate")
	}
	requests, ok := requestsI.(float64)
	if !ok || requests != float64(int(requests)) {
		return nil, fmt.Errorf("Requests should be an integer")
	}
	periodI, ok := in["period"]
	if !ok {
		return nil, fmt.Errorf("Expected period")
	}
	periodS, ok := periodI.(string)
	if !ok {
		return nil, fmt.Errorf("Period should be a string")
	}
	period, err := PeriodFromString(periodS)
	if err != nil {
		return nil, err
	}
	unitType := UnitTypeRequests
	unitI, ok := in["unit"]
	if ok {
		unitS, ok := unitI.(string)
		if !ok {
			return nil, fmt.Errorf("Unit should be a string")
		}
		unitType, err = UnitTypeFromString(unitS)
		if err != nil {
			return nil, err
		}
	}
	return NewRate(int64(requests), period, unitType)
}

func UnitTypeFromString(u string) (int, error) {
	switch u {
	case "Mb":
		return UnitTypeMegabits, nil
	case "MB":
		return UnitTypeMegabytes, nil
	case "req", "reqs", "requests", "request":
		return UnitTypeRequests, nil
	default:
		return -1, fmt.Errorf("Unsupported unit")
	}
}

func PeriodFromString(u string) (time.Duration, error) {
	switch u {
	case "second":
		return time.Second, nil
	case "minute":
		return time.Minute, nil
	case "hour":
		return time.Hour, nil
	default:
		return -1, fmt.Errorf("Unsupported period: %s", u)
	}
}

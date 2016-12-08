package main

import "time"

type timeoutmapdata struct {
	timeout time.Time
	data    interface{}
}

type TimeOutMap struct {
	data    map[string]*timeoutmapdata
	timeout time.Duration
	ticker  *time.Ticker
}

func NewTimeOutMap(dt time.Duration) *TimeOutMap {
	return &TimeOutMap{
		data:    make(map[string]*timeoutmapdata),
		timeout: dt,
		ticker:  time.NewTicker(dt),
	}
}

func (tom *TimeOutMap) Start() {
	go func() {
		for {
			<-tom.ticker.C
			now := time.Now()
			for k, v := range tom.data {
				if now.Sub(v.timeout) > tom.timeout {
					delete(tom.data, k)
				}
			}
		}
	}()
}

func (tom *TimeOutMap) Set(k string, v interface{}) {
	tom.data[k] = &timeoutmapdata{
		timeout: time.Now(),
		data:    v,
	}
}

func (tom *TimeOutMap) Get(k string) (interface{}, bool) {
	if v, ok := tom.data[k]; ok {
		v.timeout = time.Now()
		return v.data, true
	}
	return nil, false
}

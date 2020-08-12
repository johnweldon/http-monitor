package main

import (
	"fmt"
	"net/http"
)

type check struct {
	URL      string
	Method   string
	Headers  http.Header
	Validate validator
}

func newCheck(opts ...option) check {
	c := &check{Method: http.MethodGet, Validate: generallySucceeds}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return *c
}

func (c check) makeRequest() (*http.Request, error) {
	r, err := http.NewRequest(c.Method, c.URL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.Headers {
		for _, vi := range v {
			r.Header.Add(k, vi)
		}
	}
	return r, err
}

type validator func(*http.Response) (string, bool)

type option func(*check)

func optMethod(method string) option {
	return func(c *check) {
		c.Method = method
	}
}

func optURL(u string) option {
	return func(c *check) {
		c.URL = u
	}
}

func optHdr(h http.Header) option {
	return func(c *check) {
		c.Headers = h
	}
}

func optSetHdr(k string, v []string) option {
	return func(c *check) {
		if c.Headers == nil {
			c.Headers = http.Header{}
		}
		c.Headers[k] = v
	}
}

func optAddHdr(k string, v string) option {
	return func(c *check) {
		if c.Headers == nil {
			c.Headers = http.Header{}
		}
		c.Headers.Add(k, v)
	}
}

func optValidate(v validator) option {
	return func(c *check) {
		c.Validate = v
	}
}

func optGet() option     { return optMethod(http.MethodGet) }
func optPut() option     { return optMethod(http.MethodPut) }
func optPost() option    { return optMethod(http.MethodPost) }
func optDelete() option  { return optMethod(http.MethodDelete) }
func optPatch() option   { return optMethod(http.MethodPatch) }
func optOptions() option { return optMethod(http.MethodOptions) }

func optExpectCode(codes ...int) option { return optValidate(expectResponseCode(codes...)) }

func optSuccess() option   { return optValidate(generallySucceeds) }
func optFailure() option   { return optValidate(generallyFails) }
func optForbidden() option { return optExpectCode(http.StatusForbidden) }

func generallySucceeds(r *http.Response) (string, bool) {
	if r == nil {
		return "missing response", false
	}
	if http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusBadRequest {
		return "", true
	}
	return fmt.Sprintf("unexpected response code %d (expected success)", r.StatusCode), false
}

func generallyFails(r *http.Response) (string, bool) {
	if r == nil {
		return "missing response", false
	}
	if http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusBadRequest {
		return fmt.Sprintf("unexpected response code %d (expected non-success)", r.StatusCode), false
	}
	return "", true
}

func expectResponseCode(codes ...int) validator {
	return func(r *http.Response) (string, bool) {
		if r == nil {
			return "missing response", false
		}
		for _, code := range codes {
			if code == r.StatusCode {
				return "", true
			}
		}
		return fmt.Sprintf("unexpected response code %d (expected %+v)", r.StatusCode, codes), false
	}
}

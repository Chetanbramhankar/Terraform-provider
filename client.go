/*
Copyright (c) 2015 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.

resty source code and usage is governed by a MIT style
license that can be found in the LICENSE file.
*/
package resty

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
)

var (
	hdrUserAgentKey     = http.CanonicalHeaderKey("User-Agent")
	hdrAcceptKey        = http.CanonicalHeaderKey("Accept")
	hdrContentTypeKey   = http.CanonicalHeaderKey("Content-Type")
	hdrContentLengthKey = http.CanonicalHeaderKey("Content-Length")
	hdrAuthorizationKey = http.CanonicalHeaderKey("Authorization")

	plainTextType   = "text/plain; charset=utf-8"
	jsonContentType = "application/json; charset=utf-8"
	formContentType = "application/x-www-form-urlencoded"

	plainTextCheck = regexp.MustCompile("(?i:text/plain)")
	jsonCheck      = regexp.MustCompile("(?i:[application|text]/json)")
	xmlCheck       = regexp.MustCompile("(?i:[application|text]/xml)")

	hdrUserAgentValue = "go-resty v%s - https://github.com/go-resty/resty"
)

// Type Client is used for HTTP/RESTful global values
// for all request raised from the client
type Client struct {
	HostUrl    string
	QueryParam url.Values
	FormData   url.Values
	Header     http.Header
	UserInfo   *User
	Token      string
	Cookies    []*http.Cookie
	Error      interface{}
	Debug      bool
	Log        *log.Logger

	httpClient       *http.Client
	transport        *http.Transport
	setContentLength bool
	isHTTPMode       bool
	beforeRequest    []func(*Client, *Request) error
	afterResponse    []func(*Client, *Response) error
}

// Type User is hold a username and password information
type User struct {
	Username, Password string
}

// SetHeader method sets a single header field and its value in the client instance.
// These headers will be applied to all requests raised from this client instance.
// Also it can be overridden in the request level header option, see `resty.R().SetHeader`.
//
// For Example: To set `Content-Type` and `Accept` as `application/json`
//
// 		resty.
//      	SetHeader("Content-Type", "application/json").
//			SetHeader("Accept", "application/json")
//
func (c *Client) SetHeader(header, value string) *Client {
	c.Header.Set(header, value)
	return c
}

// SetHeaders method sets multiple headers field and its values at one go in the client instance.
// These headers will be applied to all requests raised from this client instance. Also it can be
// overridden in the request level headers option, see `resty.R().SetHeaders`.
//
// For Example: To set `Content-Type` and `Accept` as `application/json`
//
// 		resty.SetHeaders(map[string]string{
//				"Content-Type": "application/json",
//				"Accept": "application/json",
//			})
//
func (c *Client) SetHeaders(headers map[string]string) *Client {
	for h, v := range headers {
		c.Header.Set(h, v)
	}

	return c
}

// SetCookie method sets a single cookie in the client instance.
// These cookies will be added to all the request raised from this client instance.
// 		resty.SetCookie(&http.Cookie{
// 					Name:"go-resty",
//					Value:"This is cookie value",
//					Path: "/",
// 					Domain: "sample.com",
// 					MaxAge: 36000,
// 					HttpOnly: true,
//					Secure: false,  // baseds on https or http
// 				})
//
func (c *Client) SetCookie(hc *http.Cookie) *Client {
	c.Cookies = append(c.Cookies, hc)
	return c
}

// SetCookies method sets an array of cookies in the client instance.
// These cookies will be added to all the request raised from this client instance.
// 		cookies := make([]*http.Cookie, 0)
//
//		cookies = append(cookies, &http.Cookie{
// 					Name:"go-resty-1",
//					Value:"This is cookie 1 value",
//					Path: "/",
// 					Domain: "sample.com",
// 					MaxAge: 36000,
// 					HttpOnly: true,
//					Secure: false,  // baseds on https or http
// 				})
//
//		cookies = append(cookies, &http.Cookie{
// 					Name:"go-resty-2",
//					Value:"This is cookie 2 value",
//					Path: "/",
// 					Domain: "sample.com",
// 					MaxAge: 36000,
// 					HttpOnly: true,
//					Secure: false,  // baseds on https or http
// 				})
//
// 		resty.SetCookies(cookies)
//
func (c *Client) SetCookies(cs []*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cs...)
	return c
}

// SetQueryParam method sets single paramater and its value in the client instance.
// It will be formed as query string for the request. For example: `search=kitchen%20papers&size=large`
// in the URL after `?` mark. These query params will be added to all the request raised from
// this client instance. Also it can be overridden in the request level Query Param option,
// see `resty.R().SetQueryParam`.
// 		resty.
//			SetQueryParam("search", "kitchen papers").
//			SetQueryParam("size", "large")
//
func (c *Client) SetQueryParam(param, value string) *Client {
	c.QueryParam.Add(param, value)
	return c
}

// SetQueryParams method sets multiple paramaters and its values at one go in the client instance.
// It will be formed as query string for the request. For example: `search=kitchen%20papers` in the URL after `?` mark.
// These query params will be added to all the request raised from this client instance.
// Also it can be overridden in the request level Query Param option, see `resty.R().SetQueryParams`.
// 		resty.SetQueryParams(map[string]string{
//				"search": "kitchen papers",
//				"size": "large",
//			})
//
func (c *Client) SetQueryParams(params map[string]string) *Client {
	for p, v := range params {
		c.QueryParam.Add(p, v)
	}

	return c
}

// SetFormData method sets Form parameters and its values in the client instance.
// It's applicable only HTTP method `POST` and `PUT` and requets content type would be
// `application/x-www-form-urlencoded`. These form data will be added to all the request raised from
// this client instance. Also it can be overridden in the request level form data, see `resty.R().SetFormData`.
// 		resty.SetFormData(map[string]string{
//				"access_token": "BC594900-518B-4F7E-AC75-BD37F019E08F",
//				"user_id": "3455454545",
//			})
//
func (c *Client) SetFormData(data map[string]string) *Client {
	for k, v := range data {
		c.FormData.Add(k, v)
	}

	return c
}

// SetBasicAuth method sets the basic authentication header in the HTTP request. For example -
// `Authorization: Basic <base64-encoded-value>`
//
// For example: To set the header for username "go-resty" and password "welcome"
// 		resty.SetBasicAuth("go-resty", "welcome")
//
// This basic auth information gets added to all the request rasied from this client instance.
// Also it can be overriden or set one at the request level is supported, see `resty.R().SetBasicAuth`.
//
func (c *Client) SetBasicAuth(username, password string) *Client {
	c.UserInfo = &User{Username: username, Password: password}
	return c
}

// SetAuthToken method sets bearer auth token header in the HTTP request. For exmaple -
// `Authorization: Bearer <auth-token-value-comes-here>`
//
// For example: To set auth token BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F
//
// 		resty.SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F")
//
// This bearer auth token gets added to all the request rasied from this client instance.
// Also it can be overriden or set one at the request level is supported, see `resty.R().SetAuthToken`.
//
func (c *Client) SetAuthToken(token string) *Client {
	c.Token = token
	return c
}

// R method creates a request instance, its used for Get, Post, Put, Delete, Patch, Head and Options.
func (c *Client) R() *Request {
	r := &Request{
		Url:        "",
		Method:     "",
		QueryParam: url.Values{},
		FormData:   url.Values{},
		Header:     http.Header{},
		Body:       nil,
		Result:     nil,
		Error:      nil,
		RawRequest: nil,
		client:     c,
		bodyBuf:    nil,
	}

	return r
}

// OnBeforeRequest method sets request middleware into the before request chain.
// Its gets applied after default `go-resty` request middlewares and before request
// been sent from `go-resty` to host server.
// 		resty.OnBeforeRequest(func(c *Client, r *Request) error {
//				// Now you have access to Client and Request instance
//				// manipulate it as per your need
//			})
//
func (c *Client) OnBeforeRequest(m func(*Client, *Request) error) *Client {
	c.beforeRequest[len(c.beforeRequest)-1] = m
	c.beforeRequest = append(c.beforeRequest, requestLogger)

	return c
}

// OnAfterResponse method sets response middleware into the after response chain.
// Once we receive response from host server, default `go-resty` response middleware
// gets applied and then user assigened response middlewares applied.
// 		resty.OnAfterResponse(func(c *Client, r *Response) error {
//				// Now you have access to Client and Response instance
//				// manipulate it as per your need
//			})
//
func (c *Client) OnAfterResponse(m func(*Client, *Response) error) *Client {
	c.afterResponse = append(c.afterResponse, m)
	return c
}

// SetDebug method enables the debug mode on `go-resty` client. Client logs details of every request and response.
// For `Request` it logs information such as HTTP verb, Relative URL path, Host, Headers, Body if it has one.
// For `Response` it logs information such as Status, Response Time, Headers, Body if it has one.
//
func (c *Client) SetDebug(d bool) *Client {
	c.Debug = d
	return c
}

func (c *Client) SetLogger(w io.Writer) *Client {
	c.Log = getLogger(w)
	return c
}

func (c *Client) SetContentLength(l bool) *Client {
	c.setContentLength = l
	return c
}

func (c *Client) SetError(err interface{}) *Client {
	c.Error = err
	return c
}

func (c *Client) SetRedirectPolicy(policy func(*http.Request, []*http.Request) error) *Client {
	c.httpClient.CheckRedirect = policy
	return c
}

// SetHTTPMode sets go-resty mode into HTTP
func (c *Client) SetHTTPMode() *Client {
	return c.SetMode("http")
}

// SetRESTMode sets go-resty mode into RESTful
func (c *Client) SetRESTMode() *Client {
	return c.SetMode("rest")
}

// SetMode sets go-resty client mode to given value such as 'http' & 'rest'.
// 	RESTful:
//		- No Redirect
//		- Automatic response unmarshal if it is JSON or XML
//	HTML:
//		- Up to 10 Redirects
//		- No automatic unmarshall. Response will be treated as `response.String()`
//
// If you want more redirects, use FlexibleRedirectPolicy
//		resty.SetRedirectPolicy(FlexibleRedirectPolicy(20))
//
func (c *Client) SetMode(mode string) *Client {
	if mode == "http" {
		c.isHTTPMode = true
		c.httpClient.CheckRedirect = FlexibleRedirectPolicy(10)
		c.afterResponse = []func(*Client, *Response) error{
			responseLogger,
		}
	} else { // RESTful
		c.isHTTPMode = false
		c.httpClient.CheckRedirect = NoRedirectPolicy
		c.afterResponse = []func(*Client, *Response) error{
			responseLogger,
			parseResponseBody,
		}
	}

	return c
}

func (c *Client) SetTLSClientConfig(config *tls.Config) *Client {
	c.transport.TLSClientConfig = config
	return c
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.transport.Dial = func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, timeout)
		if err != nil {
			c.Log.Printf("Error: %v", err)
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))

		return conn, nil
	}

	return c
}

func (c *Client) execute(req *Request) (*Response, error) {
	// Apply Request middleware
	var err error
	for _, f := range c.beforeRequest {
		err = f(c, req)
		if err != nil {
			return nil, err
		}
	}

	req.Time = time.Now()
	c.httpClient.Transport = c.transport
	resp, err := c.httpClient.Do(req.RawRequest)
	if err != nil {
		return nil, err
	}

	response := &Response{
		Request:     req,
		ReceivedAt:  time.Now(),
		RawResponse: resp,
	}

	defer resp.Body.Close()
	response.Body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Apply Response middleware
	for _, f := range c.afterResponse {
		err = f(c, response)
		if err != nil {
			break
		}
	}

	return response, err
}

func (c *Client) enableLogPrefix() {
	c.Log.SetFlags(log.LstdFlags)
	c.Log.SetPrefix("RESTY ")
}

func (c *Client) disableLogPrefix() {
	c.Log.SetFlags(0)
	c.Log.SetPrefix("")
}

//
// Request
//

// Type Request
type Request struct {
	Url        string
	Method     string
	QueryParam url.Values
	FormData   url.Values
	Header     http.Header
	UserInfo   *User
	Token      string
	Body       interface{}
	Result     interface{}
	Error      interface{}
	Time       time.Time
	RawRequest *http.Request

	client           *Client
	bodyBuf          *bytes.Buffer
	isMultiPart      bool
	isFormData       bool
	setContentLength bool
}

func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParam.Add(param, value)
	return r
}

func (r *Request) SetQueryParams(params map[string]string) *Request {
	for p, v := range params {
		r.QueryParam.Add(p, v)
	}

	return r
}

func (r *Request) SetHeader(header, value string) *Request {
	r.Header.Set(header, value)
	return r
}

func (r *Request) SetHeaders(headers map[string]string) *Request {
	for h, v := range headers {
		r.Header.Set(h, v)
	}

	return r
}

func (r *Request) SetFormData(data map[string]string) *Request {
	for k, v := range data {
		r.FormData.Add(k, v)
	}

	return r
}

func (r *Request) SetBody(body interface{}) *Request {
	r.Body = body
	return r
}

func (r *Request) SetResult(res interface{}) *Request {
	r.Result = res
	return r
}

func (r *Request) SetError(err interface{}) *Request {
	r.Error = err
	return r
}

func (r *Request) SetFile(param, filePath string) *Request {
	r.FormData.Set("@"+param, filePath)
	r.isMultiPart = true

	return r
}

func (r *Request) SetFiles(files map[string]string) *Request {
	for f, fp := range files {
		r.FormData.Set("@"+f, fp)
	}
	r.isMultiPart = true

	return r
}

func (r *Request) SetContentLength(l bool) *Request {
	r.setContentLength = true

	return r
}

func (r *Request) SetBasicAuth(username, password string) *Request {
	r.UserInfo = &User{Username: username, Password: password}
	return r
}

func (r *Request) SetAuthToken(token string) *Request {
	r.Token = token
	return r
}

//
// HTTP verb method starts here
//

func (r *Request) Get(url string) (*Response, error) {
	return r.execute(GET, url)
}

func (r *Request) Post(url string) (*Response, error) {
	return r.execute(POST, url)
}

func (r *Request) Put(url string) (*Response, error) {
	return r.execute(PUT, url)
}

func (r *Request) Delete(url string) (*Response, error) {
	return r.execute(DELETE, url)
}

func (r *Request) Patch(url string) (*Response, error) {
	return r.execute(PATCH, url)
}

func (r *Request) Head(url string) (*Response, error) {
	return r.execute(HEAD, url)
}

func (r *Request) Options(url string) (*Response, error) {
	return r.execute(OPTIONS, url)
}

func (r *Request) execute(method, url string) (*Response, error) {
	if r.isMultiPart && !(method == POST || method == PUT) {
		return nil, fmt.Errorf("File upload is not allowed in HTTP verb [%v]", method)
	}

	r.Method = method
	r.Url = url

	return r.client.execute(r)
}

//
// Response
//

// Type Response
type Response struct {
	Body        []byte
	ReceivedAt  time.Time
	Request     *Request
	RawResponse *http.Response
}

func (r *Response) Status() string {
	return r.RawResponse.Status
}

func (r *Response) StatusCode() int {
	return r.RawResponse.StatusCode
}

func (r *Response) Result() interface{} {
	return r.Request.Result
}

func (r *Response) Error() interface{} {
	return r.Request.Error
}

func (r *Response) Header() http.Header {
	return r.RawResponse.Header
}

func (r *Response) Cookies() []*http.Cookie {
	return r.RawResponse.Cookies()
}

func (r *Response) String() string {
	if r.Body == nil {
		return ""
	}

	return string(r.Body)
}

func (r *Response) Time() time.Duration {
	return r.ReceivedAt.Sub(r.Request.Time)
}

//
// Resty's handy redirect polices
//

func NoRedirectPolicy(req *http.Request, via []*http.Request) error {
	return errors.New("Auto redirect is disabled")
}

func FlexibleRedirectPolicy(noOfRedirect int) func(*http.Request, []*http.Request) error {
	fn := func(req *http.Request, via []*http.Request) error {
		if len(via) >= noOfRedirect {
			return fmt.Errorf("Stopped after %d redirects", noOfRedirect)
		}
		return nil
	}

	return fn
}

//
// Helper methods
//

func IsStringEmpty(str string) bool {
	return (len(strings.TrimSpace(str)) == 0)
}

func IsMarshalRequired(body interface{}) bool {
	kind := reflect.ValueOf(body).Kind()
	return (kind == reflect.Struct || kind == reflect.Map)
}

func DetectContentType(body interface{}) string {
	contentType := plainTextType
	kind := reflect.ValueOf(body).Kind()

	switch kind {
	case reflect.Struct, reflect.Map:
		contentType = jsonContentType
	case reflect.String:
		contentType = plainTextType
	default:
		contentType = http.DetectContentType(body.([]byte))
	}

	return contentType
}

func IsJsonType(ct string) bool {
	return jsonCheck.MatchString(ct)
}

func IsXmlType(ct string) bool {
	return xmlCheck.MatchString(ct)
}

func Unmarshal(ct string, b []byte, d interface{}) (err error) {
	if IsJsonType(ct) {
		err = json.Unmarshal(b, d)
	} else if IsXmlType(ct) {
		err = xml.Unmarshal(b, d)
	}

	return
}

func getLogger(w io.Writer) *log.Logger {
	var l *log.Logger
	if w == nil {
		l = log.New(os.Stderr, "RESTY ", log.LstdFlags)
	} else {
		l = log.New(w, "RESTY ", log.LstdFlags)
	}

	return l
}

func addFile(w *multipart.Writer, fieldName, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := w.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	return err
}

func getRequestBodyString(r *Request) (body string) {
	body = "***** NO CONTENT *****"
	if r.Method == POST || r.Method == PUT || r.Method == PATCH {
		// multipart/form-data OR form data
		if r.isMultiPart || r.isFormData {
			body = string(r.bodyBuf.Bytes())

			return
		}

		// request body data
		if r.Body != nil {
			contentType := r.Header.Get(hdrContentTypeKey)
			var prtBodyBytes []byte
			var err error
			isMarshal := IsMarshalRequired(r.Body)
			if IsJsonType(contentType) && isMarshal {
				prtBodyBytes, err = json.MarshalIndent(&r.Body, "", "   ")
			} else if IsXmlType(contentType) && isMarshal {
				prtBodyBytes, err = xml.MarshalIndent(&r.Body, "", "   ")
			} else if b, ok := r.Body.(string); ok {
				if IsJsonType(contentType) {
					bodyBytes := []byte(b)
					var out bytes.Buffer
					if err = json.Indent(&out, bodyBytes, "", "   "); err == nil {
						prtBodyBytes = out.Bytes()
					}
				}
			} else if b, ok := r.Body.([]byte); ok {
				body = base64.StdEncoding.EncodeToString(b)
			}

			if prtBodyBytes != nil {
				body = string(prtBodyBytes)
			}
		}

	}

	return
}

func getResponseBodyString(res *Response) string {
	bodyStr := "***** NO CONTENT *****"
	if res.Body != nil {
		ct := res.Header().Get(hdrContentTypeKey)
		if IsJsonType(ct) {
			var out bytes.Buffer
			if err := json.Indent(&out, res.Body, "", "   "); err == nil {
				bodyStr = string(out.Bytes())
			}
		} else {
			str := res.String()
			if !IsStringEmpty(str) {
				bodyStr = str
			}
		}
	}
	return bodyStr
}

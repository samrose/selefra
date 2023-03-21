package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/cmd/version"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"reflect"
)

const DefaultMaxTryTimes = 3

// ------------------------------------------------- --------------------------------------------------------------------

func GetYaml[Response any](ctx context.Context, targetUrl string, options ...*Options[any, Response]) (Response, error) {

	if len(options) == 0 {
		options = append(options, NewOptions[any, Response](targetUrl, YamlResponseHandler[Response]()))
	}

	options[0] = options[0].WithTargetURL(targetUrl).WithYamlResponseHandler()

	return SendRequest[any, Response](ctx, options[0])
}

// ------------------------------------------------- --------------------------------------------------------------------

func GetJson[Response any](ctx context.Context, targetUrl string, options ...*Options[any, Response]) (Response, error) {

	if len(options) == 0 {
		options = append(options, NewOptions[any, Response](targetUrl, JsonResponseHandler[Response]()))
	}

	options[0] = options[0].WithTargetURL(targetUrl).
		WithJsonResponseHandler().
		AppendRequestSetting(func(httpRequest *http.Request) error {
			httpRequest.Header.Set("Content-Type", "application/json")
			return nil
		})

	return SendRequest[any, Response](ctx, options[0])
}

func PostJson[Request any, Response any](ctx context.Context, targetUrl string, request Request, options ...*Options[Request, Response]) (Response, error) {

	if len(options) == 0 {
		options = append(options, NewOptions[Request, Response](targetUrl, JsonResponseHandler[Response]()))
	}

	marshal, err := json.Marshal(request)
	if err != nil {
		var zero Response
		return zero, fmt.Errorf("PostJson json marshal request error: %s, typer = %s", err.Error(), reflect.TypeOf(request).String())
	}

	options[0] = options[0].WithTargetURL(targetUrl).
		WithJsonResponseHandler().
		AppendRequestSetting(func(httpRequest *http.Request) error {
			httpRequest.Header.Set("Content-Type", "application/json")
			return nil
		}).
		WithBody(marshal)

	return SendRequest[Request, Response](ctx, options[0])
}

// ------------------------------------------------- --------------------------------------------------------------------

func GetBytes(ctx context.Context, targetUrl string, options ...*Options[any, []byte]) ([]byte, error) {

	if len(options) == 0 {
		options = append(options, NewOptions[any, []byte](targetUrl, BytesResponseHandler()))
	}

	options[0] = options[0].WithTargetURL(targetUrl).
		WithResponseHandler(BytesResponseHandler())

	return SendRequest[any, []byte](ctx, options[0])
}

// ------------------------------------------------- --------------------------------------------------------------------

func GetString(ctx context.Context, targetUrl string, options ...*Options[any, []byte]) (string, error) {
	responseBytes, err := GetBytes(ctx, targetUrl, options...)
	if err != nil {
		return "", err
	}
	return string(responseBytes), nil
}

// ------------------------------------------------- --------------------------------------------------------------------

// ResponseHandler A component used to process http responses
type ResponseHandler[Response any] func(httpResponse *http.Response) (Response, error)

// BytesResponseHandler The default response handler, automatically reads the response body when the response code is a given value
func BytesResponseHandler(readResponseOnStatusCodeIn ...int) ResponseHandler[[]byte] {

	// By default, the response body is read only when the status code is 200
	if len(readResponseOnStatusCodeIn) == 0 {
		readResponseOnStatusCodeIn = append(readResponseOnStatusCodeIn, http.StatusOK)
	}

	return func(httpResponse *http.Response) ([]byte, error) {
		for _, status := range readResponseOnStatusCodeIn {
			if status == httpResponse.StatusCode {
				responseBodyBytes, err := io.ReadAll(httpResponse.Body)
				if err != nil {
					return nil, fmt.Errorf("response statuc code: %d, read body error: %s", httpResponse.StatusCode, err.Error())
				}
				return responseBodyBytes, nil
			}
		}
		return nil, fmt.Errorf("response status code: %d", httpResponse.StatusCode)
	}
}

func StringResponseHandler(readResponseOnStatusCodeIn ...int) ResponseHandler[string] {
	return func(httpResponse *http.Response) (string, error) {
		responseBytes, err := BytesResponseHandler(readResponseOnStatusCodeIn...)(httpResponse)
		if err != nil {
			return "", err
		}
		return string(responseBytes), nil
	}
}

func YamlResponseHandler[Response any](readResponseOnStatusCodeIn ...int) ResponseHandler[Response] {
	return func(httpResponse *http.Response) (Response, error) {

		var r Response

		responseBytes, err := BytesResponseHandler(readResponseOnStatusCodeIn...)(httpResponse)
		if err != nil {
			return r, err
		}

		err = yaml.Unmarshal(responseBytes, &r)
		if err != nil {
			return r, fmt.Errorf("response body yaml unmarshal error: %s, type: %s, response body: %s", err.Error(), reflect.TypeOf(r).String(), string(responseBytes))
		}
		return r, nil
	}
}

func JsonResponseHandler[Response any](readResponseOnStatusCodeIn ...int) ResponseHandler[Response] {
	return func(httpResponse *http.Response) (Response, error) {

		var r Response

		responseBytes, err := BytesResponseHandler(readResponseOnStatusCodeIn...)(httpResponse)
		if err != nil {
			return r, err
		}

		err = json.Unmarshal(responseBytes, &r)
		if err != nil {
			return r, fmt.Errorf("response body json unmarshal error: %s, type: %s, response body: %s", err.Error(), reflect.TypeOf(r).String(), string(responseBytes))
		}
		return r, nil
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

type RequestSetting func(httpRequest *http.Request) error

const DefaultUserAgent = "selefra-cli"

func DefaultUserAgentRequestSetting() RequestSetting {
	return func(httpRequest *http.Request) error {
		httpRequest.Header.Set("User-Agent", MyUserAgent())
		return nil
	}
}

func MyUserAgent() string {
	return fmt.Sprintf("%s/%s", DefaultUserAgent, version.Version)
}

// ------------------------------------------------- --------------------------------------------------------------------

const DefaultMethod = http.MethodGet

type Options[Request any, Response any] struct {
	MaxTryTimes         int
	TargetURL           string
	Method              string
	Body                []byte
	RequestSettingSlice []RequestSetting
	ResponseHandler     ResponseHandler[Response]
	MessageChannel      chan *schema.Diagnostics
}

func NewOptions[Request any, Response any](targetUrl string, responseHandler ResponseHandler[Response]) *Options[Request, Response] {
	return &Options[Request, Response]{
		MaxTryTimes:         DefaultMaxTryTimes,
		TargetURL:           targetUrl,
		Method:              DefaultMethod,
		Body:                []byte{},
		RequestSettingSlice: nil,
		ResponseHandler:     responseHandler,
	}
}

func (x *Options[Request, Response]) WithMaxTryTimes(maxTryTimes int) *Options[Request, Response] {
	x.MaxTryTimes = maxTryTimes
	return x
}

func (x *Options[Request, Response]) WithTargetURL(targetURL string) *Options[Request, Response] {
	x.TargetURL = targetURL
	return x
}

func (x *Options[Request, Response]) WithMethod(method string) *Options[Request, Response] {
	x.Method = method
	return x
}

func (x *Options[Request, Response]) WithBody(body []byte) *Options[Request, Response] {
	if x.Method == DefaultMethod {
		x.Method = http.MethodPost
	}
	x.Body = body
	return x
}

func (x *Options[Request, Response]) WithRequestSettingSlice(requestSettingSlice []RequestSetting) *Options[Request, Response] {
	x.RequestSettingSlice = requestSettingSlice
	return x
}

func (x *Options[Request, Response]) AppendRequestSetting(requestSetting RequestSetting) *Options[Request, Response] {
	x.RequestSettingSlice = append(x.RequestSettingSlice, requestSetting)
	return x
}

func (x *Options[Request, Response]) WithResponseHandler(responseHandler ResponseHandler[Response]) *Options[Request, Response] {
	x.ResponseHandler = responseHandler
	return x
}

func (x *Options[Request, Response]) WithYamlResponseHandler() *Options[Request, Response] {
	x.ResponseHandler = YamlResponseHandler[Response]()
	return x
}

func (x *Options[Request, Response]) WithJsonResponseHandler() *Options[Request, Response] {
	x.ResponseHandler = JsonResponseHandler[Response]()
	return x
}

func (x *Options[Request, Response]) WithMessageChannel(messageChannel chan *schema.Diagnostics) *Options[Request, Response] {
	x.MessageChannel = messageChannel
	return x
}

func (x *Options[Request, Response]) SendMessage(message *schema.Diagnostics) *Options[Request, Response] {
	if x.MessageChannel != nil {
		x.MessageChannel <- message
	}
	return x
}

// ------------------------------------------------- --------------------------------------------------------------------

// SendRequest Sending requests is a low-level API
func SendRequest[Request any, Response any](ctx context.Context, options *Options[Request, Response]) (Response, error) {

	// TODO set default params

	var lastErr error
	for tryTimes := 0; tryTimes < options.MaxTryTimes; tryTimes++ {
		var client http.Client
		httpRequest, err := http.NewRequest(options.Method, options.TargetURL, bytes.NewReader(options.Body))
		if err != nil {
			lastErr = err
			continue
		}

		httpRequest = httpRequest.WithContext(ctx)

		for _, requestSettingFunc := range options.RequestSettingSlice {
			if err := requestSettingFunc(httpRequest); err != nil {
				lastErr = err
				continue
			}
		}

		httpResponse, err := client.Do(httpRequest)
		if err != nil {
			lastErr = err
			continue
		}
		defer httpResponse.Body.Close()

		response, err := options.ResponseHandler(httpResponse)
		if err != nil {
			lastErr = err
			continue
		}
		return response, nil
	}

	var zero Response
	return zero, lastErr
}

// ------------------------------------------------- --------------------------------------------------------------------

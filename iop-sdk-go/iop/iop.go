package iop

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	neturl "net/url"
	"sort"
	"strings"
	"time"
)

const (
	// Version
	Version = "lazop-sdk-go-20230910"

	// APIGatewaySG endpoint
	APIGatewaySG = "https://api.lazada.sg/rest"
	// APIGatewayMY endpoint
	APIGatewayMY = "https://api.lazada.com.my/rest"
	// APIGatewayVN endpoint
	APIGatewayVN = "https://api.lazada.vn/rest"
	// APIGatewayTH endpoint
	APIGatewayTH = "https://api.lazada.co.th/rest"
	// APIGatewayPH endpoint
	APIGatewayPH = "https://api.lazada.com.ph/rest"
	// APIGatewayID endpoint
	APIGatewayID = "https://api.lazada.co.id/rest"

	AuthURL = "https://auth.lazada.com/oauth/authorize"
)

// ClientOptions params
type ClientOptions struct {
	APIKey    string
	APISecret string
	Region    string
}

// IopClient represents a client to Lazada
type IopClient struct {
	APIKey      string
	APISecret   string
	Region      string
	CallbackURL string

	Method     string
	SysParams  map[string]string
	APIParams  map[string]string
	FileParams map[string][]byte
}

// NewClient init
func NewClient(opts *ClientOptions) *IopClient {
	return &IopClient{
		Region:    opts.Region,
		APIKey:    opts.APIKey,
		APISecret: opts.APISecret,
		SysParams: map[string]string{
			"app_key":     opts.APIKey,
			"sign_method": "sha256",
			"timestamp":   fmt.Sprintf("%d000", time.Now().Unix()),
			"partner_id":  Version,
		},
		APIParams:  map[string]string{},
		FileParams: map[string][]byte{},
	}
}

// Debug setter
func (lc *IopClient) Debug(enableDebug bool) *IopClient {
	if enableDebug {
		lc.SysParams["debug"] = "true"
	} else {
		lc.SysParams["debug"] = "false"
	}
	return lc
}

// Generate auth url
func (me *IopClient) MakeAuthURL() string {
	params := neturl.Values{}
	params.Add("response_type", "code")
	params.Add("force_auth", "true")
	params.Add("country", me.Region)
	params.Add("redirect_uri", me.CallbackURL)
	params.Add("client_id", me.APIKey)
	return AuthURL + `?` + params.Encode()
}

// Set callback url
func (me *IopClient) SetCallbackUrl(url string) {
	me.CallbackURL = url
}

// SetAccessToken setter
func (lc *IopClient) SetAccessToken(accessToken string) *IopClient {
	lc.SysParams["access_token"] = accessToken
	return lc
}

// ChangeRegion setter
func (lc *IopClient) ChangeRegion(region string) *IopClient {
	lc.Region = region
	return lc
}

// AddAPIParam setter
func (lc *IopClient) AddAPIParam(key string, val string) *IopClient {
	lc.APIParams[key] = val
	return lc
}

// AddFileParam setter
func (lc *IopClient) AddFileParam(key string, val []byte) *IopClient {
	lc.FileParams[key] = val
	return lc
}

// Create sign from system params and api params
func (lc *IopClient) sign(url string) string {
	keys := []string{}
	union := map[string]string{}
	for key, val := range lc.SysParams {
		union[key] = val
		keys = append(keys, key)
	}
	for key, val := range lc.APIParams {
		union[key] = val
		keys = append(keys, key)
	}

	// sort sys params and api params by key
	sort.Strings(keys)

	var message bytes.Buffer
	message.WriteString(fmt.Sprintf("%s", url))
	for _, key := range keys {
		message.WriteString(fmt.Sprintf("%s%s", key, union[key]))
	}

	hash := hmac.New(sha256.New, []byte(lc.APISecret))
	hash.Write(message.Bytes())
	return strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
}

// Response success
type Response struct {
	Code      string          `json:"code"`
	Type      string          `json:"type"`
	Message   string          `json:"message"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

// ResponseError defines a error response
type ResponseError struct {
	Code      string `json:"code"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func (lc *IopClient) getServerURL() string {
	switch lc.Region {
	case "SG":
		return APIGatewaySG
	case "MY":
		return APIGatewayMY
	case "VN":
		return APIGatewayVN
	case "TH":
		return APIGatewayTH
	case "PH":
		return APIGatewayPH
	case "ID":
		return APIGatewayID
	}
	return ""
}

// Execute sends the request though http.request and collect the response
func (lc *IopClient) Execute(apiPath string, apiMethod string, bodyParams map[string]string) (*Response, error) {
	var req *http.Request
	var err error
	var contentType string

	// add query params
	values := url.Values{}
	for key, val := range lc.SysParams {
		values.Add(key, val)
	}

	// POST handle
	body := &bytes.Buffer{}
	if apiMethod == http.MethodPost {
		writer := multipart.NewWriter(body)
		contentType = writer.FormDataContentType()
		if len(lc.FileParams) > 0 {
			// add formfile to handle file upload
			for key, val := range lc.FileParams {
				part, err := writer.CreateFormFile("image", key)
				if err != nil {
					return nil, err
				}
				_, err = part.Write(val)
				if err != nil {
					return nil, err
				}
			}
		}

		if len(bodyParams) != 0 {
			for k, v := range bodyParams {
				lc.AddAPIParam(k, v)
				_ = writer.WriteField(k, v)
			}

		}

		if err = writer.Close(); err != nil {
			return nil, err
		}
	}

	// GET handle
	if apiMethod == http.MethodGet {
		for key, val := range lc.APIParams {
			values.Add(key, val)
		}
	}

	apiServerURL := lc.getServerURL()

	values.Add("sign", lc.sign(apiPath))
	fullURL := fmt.Sprintf("%s%s?%s", apiServerURL, apiPath, values.Encode())
	req, err = http.NewRequest(apiMethod, fullURL, body)

	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	log.Println(req)
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	resp := &Response{}
	err = json.Unmarshal(respBody, resp)

	lc.APIParams = nil
	lc.FileParams = nil

	return resp, err
}

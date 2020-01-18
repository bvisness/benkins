package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

const SlackBaseUrl = "https://slack.com/api/"

type SlackMessageRequest struct {
	Channel string        `json:"channel"`
	Text    string        `json:"text"`
	Blocks  []*SlackBlock `json:"blocks"`
}

type SlackBlock struct {
	Type string          `json:"type"`
	Text SlackTextObject `json:"text"`
}

type SlackTextObject struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Emoji    bool   `json:"emoji,omitempty"`
	Verbatim bool   `json:"verbatim,omitempty"`
}

func TextBlock(format string, a ...interface{}) *SlackBlock {
	if format == "" {
		return nil
	}

	return &SlackBlock{
		Type: "section",
		Text: SlackTextObject{
			Type: "mrkdwn",
			Text: fmt.Sprintf(format, a...),
		},
	}
}

type SlackClient struct {
	httpClient *http.Client
	token      string
}

func NewSlackClient(token string) *SlackClient {
	return &SlackClient{
		httpClient: &http.Client{},
		token:      token,
	}
}

func (s *SlackClient) SlackPostMessage(r SlackMessageRequest) (*http.Response, error) {
	var nonNilBlocks []*SlackBlock

	// Remove nil blocks
	for _, block := range r.Blocks {
		if block != nil {
			nonNilBlocks = append(nonNilBlocks, block)
		}
	}
	r.Blocks = nonNilBlocks

	js, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", SlackBaseUrl+"chat.postMessage", bytes.NewBuffer(js))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return res, err
	}

	dumpBytes, _ := httputil.DumpResponse(res, true)
	dump := string(dumpBytes)

	if res.StatusCode < 200 || 299 < res.StatusCode {
		return res, fmt.Errorf("Got non-success status code %v from Slack:\n%v", res.StatusCode, dump)
	}

	body, _ := ioutil.ReadAll(res.Body)
	var responseBody struct {
		Ok bool `json:"ok"`
	}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return res, fmt.Errorf("Failed to parse JSON response from Slack:\n%v", dump)
	}

	if !responseBody.Ok {
		return res, fmt.Errorf("Got non-ok response from Slack:\n%v", dump)
	}

	return res, nil
}

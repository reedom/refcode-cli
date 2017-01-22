package uniqid

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/reedom/refcode-cli/log"
)

type remoteStore struct {
	url string
}

func NewRemoteStore(url string) remoteStore {
	return remoteStore{url}
}

type counterReq struct {
	AppID    string `json:"appID"`
	Category string `json:"category"`
	N        int64  `json:"n"`
}

type counterRes struct {
	After int64 `json:"after"`
}

func (s remoteStore) Generate(ctx context.Context, key, sub []byte, n int64) ([][]byte, error) {
	param := counterReq{
		AppID:    string(key),
		Category: string(sub),
		N:        n,
	}
	body, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	log.Verbose.Print("url", s.url)
	req, err := http.NewRequest("POST", s.url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		log.ErrorLog.Printf("remote returns error: %v, %s", resp.Status, string(body))
	}

	var res counterRes
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	ids := make([][]byte, n)
	for i := range ids {
		ids[i] = strconv.AppendInt(nil, res.After-n+int64(i)+1, 10)
	}
	return ids, nil
}

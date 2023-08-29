package src

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// 根据传入的 key 选择相应节点 PeerGetter。
type PeerPicker interface {
	PickPeer(key string) (peer PeerPicker, ok bool)
}

// 从对应 group 查找缓存值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

// for example:  "http://10.0.0.2:8008"
type httpGetter struct {
	baseURL string
}

// 从远端获取 返回值
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)

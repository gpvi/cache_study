package src

import (
	"GeeCache/src/consistenthash"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

const defaultBasePath = "/cache/"
const defaultReplicas = 50

type HTTPPool struct {
	self     string
	basePath string
	mu       sync.Mutex
	//"类型为Map",一致性哈希实现
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

// 获取HTTPPool 对象
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// 打印 服务端
func (p HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// 实现ServeHttp,参数
func (p HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 检查 字符穿之间的是否有共同前缀来判断baseURL 是否正确
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpect path" + r.URL.Path)
	}
	// 路径切割 n= 2: 键值对类型
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	// 	非键值对类型
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// 得到组名和键名
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/octet--stream")
	w.Write(view.ByteSlice())
}

// peer : ip
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 初始化一致性哈希
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	//  正确找到节点
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	// 只有自己一个节点
	return nil, false
}

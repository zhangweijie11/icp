package crawler

import (
	"crypto/tls"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"net"
	"net/http"
	"time"
)

// CollyScraper colly 抓取结构
type CollyScraper struct {
	Collector *colly.Collector
	Transport *http.Transport
	Response  *http.Response
}

type GoWapTransport struct {
	*http.Transport
	respCallBack func(resp *http.Response)
}

func NewGoWapTransport(t *http.Transport, f func(resp *http.Response)) *GoWapTransport {
	return &GoWapTransport{t, f}
}

// NewCollyScraper 初始化 Colly 爬虫
func NewCollyScraper() *CollyScraper {
	collyScraper := &CollyScraper{}
	collyScraper.Transport = &http.Transport{
		//用于创建未加密的TCP连接
		DialContext: (&net.Dialer{
			Timeout: time.Second * time.Duration(5),
		}).DialContext,
		//控制所有主机的最大空闲（保持活动）连接数。零表示没有限制。
		MaxIdleConns: 100,
		//空闲（保持活动状态）连接在关闭自身之前保持空闲的最长时间。零表示没有限制。
		IdleConnTimeout: 90 * time.Second,
		//等待 TLS 握手的最长时间。零表示无超时。
		TLSHandshakeTimeout: 5 * time.Second,
		//如果非零，则指定在完全写入请求标头后等待服务器的第一个响应标头的时间量（如果请求具有“预期：100-continue”标头）。零表示没有超时，并导致正文立即发送，而无需等待服务器批准。此时间不包括发送请求标头的时间。
		ExpectContinueTimeout: time.Duration(5) * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	collyScraper.Collector = colly.NewCollector()
	setResp := func(r *http.Response) {
		collyScraper.Response = r
	}

	//自定义传输
	collyScraper.Collector.WithTransport(NewGoWapTransport(collyScraper.Transport, setResp))

	// 为请求设置有效的 Referer HTTP 标头。警告：仅当使用 Request.Visit from 回调而不是 Collector.Visit 时，此扩展才有效。
	extensions.Referer(collyScraper.Collector)

	return collyScraper
}

func (gt *GoWapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rsp, err := gt.Transport.RoundTrip(req)
	gt.respCallBack(rsp)
	return rsp, err
}

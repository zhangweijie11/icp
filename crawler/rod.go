package crawler

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/temoto/robotstxt"
	"icp/esicp"
	"sync"
)

type RodScraper struct {
	Browser        *rod.Browser
	pagePool       rod.PagePool
	protoUserAgent *proto.NetworkSetUserAgentOverride
	lock           *sync.RWMutex
	robotsMap      map[string]*robotstxt.RobotsData
}

func NewRodScraper() *RodScraper {
	rodScraper := &RodScraper{}
	// 寻找可执行程序的路径
	path, _ := launcher.LookPath()
	u := launcher.New().Bin(path).Headless(false).NoSandbox(true).MustLaunch()
	rodScraper.lock = &sync.RWMutex{}
	rodScraper.robotsMap = make(map[string]*robotstxt.RobotsData)
	// 如果 ControlURL 未设置， MustConnect 将自动运行 launcher.New().MustLaunch()。 默认情况下，launcher 将自动下载并使用固定版本的浏览器，以保证浏览器的行为一致性。
	// MustIgnoreCertErrors 忽略证书错误
	rodScraper.Browser = rod.New().ControlURL(u).MustConnect().MustIgnoreCertErrors(true)
	rodScraper.pagePool = createPagePool(10)

	return rodScraper
}

// GetBrowser 获得浏览器对象
func getBrowser(l *launcher.Launcher) *rod.Browser {
	u := l.MustLaunch()
	return rod.New().ControlURL(u).MustConnect().MustIgnoreCertErrors(true)

}

// createPage 生成一个page对象
func (s *RodScraper) createPage() (page *rod.Page) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.Browser == nil {
		// 寻找可执行程序的路径
		path, _ := launcher.LookPath()
		u := launcher.New().Bin(path).NoSandbox(true)
		s.Browser = getBrowser(u)
	}
	page = stealth.MustPage(s.Browser)
	return
}

// CreatePagePool 内部pagePool大小
func createPagePool(pageSize int) rod.PagePool {
	return rod.NewPagePool(pageSize)
}

func (s *RodScraper) GetPage() *rod.Page {
	s.lock.Lock()
	if s.pagePool == nil {
		s.pagePool = createPagePool(10)
	}
	s.lock.Unlock()
	return s.pagePool.Get(s.createPage)
}

// PutPage 回收page
func (s *RodScraper) PutPage(pageInterface interface{}) {
	page := pageInterface.(*rod.Page)
	err := page.Navigate("about:blank")
	if err != nil {
		esicp.Logger.Println("回收页面出现问题")
	} else {
		s.pagePool.Put(page)
	}
}

// Close 关闭浏览器
func (s *RodScraper) Close() {
	if s.Browser != nil {
		pages, _ := s.Browser.Pages()
		for _, page := range pages {
			err := page.Close()
			if err != nil {
				esicp.Logger.Println("关闭页面出现问题")
				continue
			}
		}
		err := s.Browser.Close()
		if err != nil {
			esicp.Logger.Println("关闭浏览器出现错误", err)
		}
	} else {
		esicp.Logger.Println("关闭浏览器出现错误，浏览器实例为 nil")
	}
}

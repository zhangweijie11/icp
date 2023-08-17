package icpapi

import (
	"github.com/go-rod/rod"
	"icp/crawler"
	"icp/esicp"
	"strings"
	"time"
)

// GetIcpapiIcpInfo 获取 ICP 备案查询网的 ICP 数据
func GetIcpapiIcpInfo(domain string, rodScraper *crawler.RodScraper) *esicp.ESIcpInfo {
	var esIcpInfo *esicp.ESIcpInfo
	paramURL := "https://www.icpapi.com/"
	if rodScraper == nil {
		rodScraper = crawler.NewRodScraper()
	}
	// 创建一个新的页面
	page := rodScraper.GetPage()
	// 当页面上弹出警告框、确认框、提示框等对话框时，使用 MustHandleDialog 方法可以自动处理这些对话框
	go page.MustHandleDialog()
	err := rod.Try(func() {
		page.Timeout(esicp.Timeout * time.Second).MustNavigate(paramURL).MustElement("#searchInput").MustInput(domain)
		page.Timeout(esicp.Timeout * time.Second).MustElement("#search-input-box > span > button").MustClick()
	})
	if err == nil {
		err = rod.Try(func() {
			time.Sleep(3 * time.Second)
			// ICP 备案网站信息-备案号
			status, serviceLicence, err := page.Has("body > div > div:nth-child(3) > div > div > div > table > tbody > tr:nth-child(5) > td:nth-child(2)")
			if err == nil && status {
				service, err := serviceLicence.Text()
				if err == nil {
					esIcpInfo.ServiceLicence = service
					// ICP 备案主体信息-备案号
					esIcpInfo.MainLicence = strings.Split(service, "-")[0]
				}
			}
			// ICP 备案网站信息-网站域名
			status, domains, err := page.Has("body > div > div:nth-child(3) > div > div > div > table > tbody > tr:nth-child(2) > td:nth-child(2) > a")
			if err == nil && status {
				do, err := domains.Text()
				if err == nil {
					esIcpInfo.Domain = do
				}
			}
			// ICP 备案网站信息-网站名称
			status, websiteName, err := page.Has("body > div > div:nth-child(3) > div > div > div > table > tbody > tr:nth-child(1) > td:nth-child(2)")
			if err == nil && status {
				webName, err := websiteName.Text()
				if err == nil && webName != "-" {
					esIcpInfo.WebsiteName = webName
				}
			}

			// ICP 备案主体信息-主办单位性质
			status, natureName, err := page.Has("body > div > div:nth-child(3) > div > div > div > table > tbody > tr:nth-child(3) > td:nth-child(2)")
			if err == nil && status {
				nature, err := natureName.Text()
				if err == nil {
					esIcpInfo.NatureName = nature

				}

			}
			// ICP 备案主体信息-主办单位名称
			status, unitName, err := page.Has("body > div > div:nth-child(3) > div > div > div > table > tbody > tr:nth-child(4) > td:nth-child(2) > a")
			if err == nil && status {
				unit, err := unitName.Text()
				if err == nil {
					esIcpInfo.UnitName = unit
				}

			}
			// ICP 备案主体信息-审核通过日期
			status, updateRecordTime, err := page.Has("body > div > div:nth-child(3) > div > div > div > table > tbody > tr:nth-child(6) > td:nth-child(2)")
			if err == nil && status {
				updateTime, err := updateRecordTime.Text()
				if err == nil {
					parsedTime, err := time.Parse(esicp.TimeFormatSecond, updateTime)
					var validTime string
					if err == nil {
						// 转换为所需的输出格式
						outputLayout := "2006-01-02T15:04:05" // 所需输出日期时间的格式
						validTime = parsedTime.Format(outputLayout)
					}
					esIcpInfo.UpdateRecordTime = validTime
				}
			}
			if esIcpInfo.ServiceLicence != "" {
				// ICP 备案主体信息-是否限制接入
				esIcpInfo.LimitAccess = "否"
			}
		})
	}

	rodScraper.PutPage(page)

	return esIcpInfo
}

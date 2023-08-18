package main

import (
	"errors"
	"fmt"
	"icp/beian"
	"icp/chinaz"
	"icp/crawler"
	"icp/esicp"
	"icp/icpapi"
)

// GetIcpInfo 获取 ICP 数据
func GetIcpInfo(domain string, rodScraper *crawler.RodScraper) *esicp.ESIcpInfo {
	esIcpInfo, err := beian.GetBeianIcpInfo(domain)
	if esIcpInfo == nil && errors.Is(err, esicp.ErrDomainNotRegistered) {
		if rodScraper == nil {
			rodScraper = crawler.NewRodScraper()
		}
		icpSources := []string{esicp.SourceIcpApi, esicp.SourceChinaz}
		for _, icpSource := range icpSources {
			if icpSource == esicp.SourceIcpApi {
				esIcpInfo = icpapi.GetIcpapiIcpInfo(domain, rodScraper)
				if esIcpInfo != nil {
					break
				}
			}

			if icpSource == esicp.SourceChinaz {
				esIcpInfo = chinaz.GetChinazIcpInfo(domain, rodScraper)
				if esIcpInfo != nil {
					break
				}
			}
		}

		if rodScraper != nil {
			rodScraper.Close()
		}
	}

	return esIcpInfo
}

func main() {
	icpInfo := GetIcpInfo("examole.com", nil)
	fmt.Println("------------>", icpInfo)
}

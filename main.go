package main

import (
	"fmt"
	"icp/beian"
	"icp/esicp"
)

// GetIcpInfo 获取 ICP 数据
func GetIcpInfo(domain string) *esicp.ESIcpInfo {
	esIcpInfo, _ := beian.GetBeianIcpInfo(domain)

	return esIcpInfo
}

func main() {
	icpInfo := GetIcpInfo("socmap.net")
	fmt.Println("------------>", icpInfo)
}

package esicp

import (
	"errors"
	"log"
)

const (
	Timeout          = 10
	TimeFormatSecond = "2006-01-02 15:04:05" // 固定format时间，2006-12345
	SourceIcpApi     = "icpapi"
	SourceChinaz     = "chinaz"
)

var Logger log.Logger

type ESIcpInfo struct {
	ContentTypeName  string `json:"content_type_name"`
	Domain           string `json:"domain"`
	DomainId         string `json:"domain_id"`
	LeaderName       string `json:"leader_name"`
	LimitAccess      string `json:"limit_access"`
	MainId           string `json:"main_id"`
	MainLicence      string `json:"main_licence"`
	NatureName       string `json:"nature_name"`
	ServiceId        string `json:"service_id"`
	ServiceLicence   string `json:"service_licence"`
	WebsiteName      string `json:"website_name"`
	Website          string `json:"website"`
	UnitName         string `json:"unit_name"`
	UpdateRecordTime string `json:"update_record_time"`
}

// ErrDomainNotRegistered 创建一个特定类型的错误变量
var ErrDomainNotRegistered = errors.New("域名未备案")

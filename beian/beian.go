package beian

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"icp/crawler"
	"icp/esicp"
	"math/rand"
	"strconv"
	"time"
)

type commonParams struct {
	token       string
	ip          string
	contentType string
}

type responseParams struct {
	EndRow           int    `json:"endRow"`
	FirstPage        int    `json:"firstPage"`
	HasNextPage      bool   `json:"hasNextPage"`
	HasPreviousPage  bool   `json:"hasPreviousPage"`
	IsFirstPage      bool   `json:"isFirstPage"`
	IsLastPage       bool   `json:"isLastPage"`
	LastPage         int    `json:"lastPage"`
	List             []*icp `json:"list"`
	NavigatePages    int    `json:"navigatePages"`
	NavigatepageNums []int  `json:"navigatepageNums"`
	NextPage         int    `json:"nextPage"`
	PageNum          int    `json:"pageNum"`
	PageSize         int    `json:"pageSize"`
	Pages            int    `json:"pages"`
	PrePage          int    `json:"prePage"`
	Size             int    `json:"size"`
	StartRow         int    `json:"startRow"`
	Total            int    `json:"total"`
}

type icp struct {
	ContentTypeName  string `json:"contentTypeName"`
	Domain           string `json:"domain"`
	DomainId         int    `json:"domainId"`
	LeaderName       string `json:"leaderName"`
	LimitAccess      string `json:"limitAccess"`
	MainId           int    `json:"mainId"`
	MainLicence      string `json:"mainLicence"`
	NatureName       string `json:"natureName"`
	ServiceId        int    `json:"serviceId"`
	ServiceLicence   string `json:"serviceLicence"`
	UnitName         string `json:"unitName"`
	UpdateRecordTime string `json:"updateRecordTime"`
}

type commonResponse struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Params  interface{} `json:"params"`
}

type authParams struct {
	Bussiness string `json:"bussiness"`
	Expire    int64  `json:"expire"`
	Refresh   string `json:"refresh"`
}

type commonRequest struct {
	PageNum  string `json:"pageNum"`
	PageSize string `json:"pageSize"`
	UnitName string `json:"unitName"`
}

func Md5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has) //将[]byte转成16进制
}

// 真实请求
func (cp *commonParams) post(url string, requestData []byte, responseData interface{}) error {
	postUrl := fmt.Sprintf("https://hlwicpfwc.miit.gov.cn/icpproject_query/api/%s", url)
	collyScraper := crawler.NewCollyScraper()
	extensions.RandomUserAgent(collyScraper.Collector)
	collyScraper.Collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Content-Type", cp.contentType)
		r.Headers.Set("Origin", "https://beian.miit.gov.cn/")
		r.Headers.Set("Referer", "https://beian.miit.gov.cn/")
		r.Headers.Set("token", cp.token)
		r.Headers.Set("CLIENT_IP", cp.ip)
		r.Headers.Set("X-FORWARDED-FOR", cp.ip)
	})

	collyScraper.Collector.OnResponse(func(r *colly.Response) {
		err := json.Unmarshal(r.Body, &responseData)
		if err != nil {
			esicp.Logger.Println("正确返回反序列化 ICP 数据时出现错误", err)
		}
	})
	collyScraper.Collector.OnError(func(r *colly.Response, err error) {
		esicp.Logger.Println("错误返回反序列化 ICP 数据时出现错误", err)
	})
	collyScraper.Collector.SetRequestTimeout(esicp.Timeout * time.Second)
	err := collyScraper.Collector.PostRaw(postUrl, requestData)
	if err != nil {
		return err
	}

	return nil
}

// 获取随机 IP
func (cp *commonParams) getIp() {
	if cp.ip != "" {
		return
	}
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	cp.ip = fmt.Sprintf("101.%d.%d.%d", 1+rng.Intn(254), 1+rng.Intn(254), 1+rng.Intn(254))
}

// 获取 token
func (cp *commonParams) getToken() error {
	timestamp := time.Now().Unix()
	authKey := Md5(fmt.Sprintf("testtest%d", timestamp))
	authBody := fmt.Sprintf("authKey=%s&timeStamp=%d", authKey, timestamp)
	cp.token = "0"
	cp.contentType = "application/x-www-form-urlencoded;charset=UTF-8"
	response := &commonResponse{Params: &authParams{}}
	err := cp.post("auth", []byte(authBody), response)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("获取token失败：%s", response.Msg)
	}

	auth := response.Params.(*authParams)
	cp.token = auth.Bussiness
	return nil
}

// 前期的验证等操作结束后开始获取 ICP 数据
func (cp *commonParams) getIcpInfo(domain string) (*icp, error) {
	request, _ := json.Marshal(&commonRequest{
		UnitName: domain,
	})
	cp.contentType = "application/json;charset=UTF-8"
	response := &commonResponse{Params: &responseParams{}}
	err := cp.post("icpAbbreviateInfo/queryByCondition", request, response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New(fmt.Sprintf("查询：%s", response.Msg))
	}

	queryParams := response.Params.(*responseParams)
	if len(queryParams.List) == 0 {
		return nil, esicp.ErrDomainNotRegistered
	}

	return queryParams.List[0], nil
}

// GetBeianIcpInfo 获取官网 ICP 数据
func GetBeianIcpInfo(domain string) (*esicp.ESIcpInfo, error) {
	request := &commonParams{}
	request.getIp()
	err := request.getToken()
	if err != nil {
		return nil, err
	}
	icpInfo, err := request.getIcpInfo(domain)
	if err != nil {
		return nil, err
	}

	if icpInfo != nil {
		parsedTime, err := time.Parse(esicp.TimeFormatSecond, icpInfo.UpdateRecordTime)
		var validTime string
		if err == nil {
			// 转换为所需的输出格式
			outputLayout := "2006-01-02T15:04:05" // 所需输出日期时间的格式
			validTime = parsedTime.Format(outputLayout)
		}

		esIcpInfo := &esicp.ESIcpInfo{
			ContentTypeName:  icpInfo.ContentTypeName,
			Domain:           icpInfo.Domain,
			DomainId:         strconv.Itoa(icpInfo.DomainId),
			LeaderName:       icpInfo.LeaderName,
			LimitAccess:      icpInfo.LimitAccess,
			MainId:           strconv.Itoa(icpInfo.MainId),
			MainLicence:      icpInfo.MainLicence,
			NatureName:       icpInfo.NatureName,
			ServiceId:        strconv.Itoa(icpInfo.ServiceId),
			ServiceLicence:   icpInfo.ServiceLicence,
			UnitName:         icpInfo.UnitName,
			UpdateRecordTime: validTime,
		}

		return esIcpInfo, err
	}

	return nil, err
}

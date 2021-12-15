package linstor

import (
	"context"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)


type LinstorGetter struct {
	Code int `json:"code"`
	Count int `json:"count"`
	Data []map[string]string `json:"data"`
}

func (d *LinstorGetter) List(query *query.Query) {

	if len(query.Filters) != 0 {
		for k, v := range query.Filters {
			newListData := make([]map[string]string,0)
			for _, mapData := range d.Data{
				if strings.Contains(mapData[string(k)], string(v)) {
					newListData = append(newListData,mapData)
				}
			}
			d.Data = newListData
			d.Count = len(newListData)

		}
	}

	startIndex, endIndex := query.Pagination.GetValidPagination(d.Count)
	data := d.Data[startIndex:endIndex]
	d.Data = data

}


func GetClient() (*client.Client, context.Context)  {
	ctx := context.TODO()
	u, err := url.Parse("http://10.203.1.158:3370")
	if err != nil {
		log.Fatal(err)
	}

	c, err := client.NewClient(client.BaseURL(u), client.Log(log.StandardLogger()))
	if err != nil {
		log.Fatal(err)
	}
	return c,ctx
}


func FormatSize(size int64) string{
	sizeStr := ""
	if size == 9223372036854775807{
		return sizeStr
	}

	switch {
	case size > 1024 * 1024 * 1024:
		sizeStr = strconv.FormatInt(size/(1024*1024*1024),10)+" TB"
	case size > 1024 * 1024 :
		sizeStr = strconv.FormatInt(size/(1024*1024),10)+" GB"
	case size > 1024:
		sizeStr = strconv.FormatInt(size/1024,10)+" MB"
	default:
		sizeStr = strconv.FormatInt(size,10)+" KB"
	}

	return sizeStr
}


func ParseSize(size string) (uint64,error) {
	strSize := strings.ToUpper(size)
	str := `^([0-9.]+)(K|M|G|T)(?:I?B)?$`
	r := regexp.MustCompile(str)
	matchsResult := r.FindStringSubmatch(strSize)
	if len(matchsResult) == 0 {
		return strconv.ParseUint(size,10,64)
	}
	finalSize, err := strconv.ParseUint(matchsResult[1],10,64)
	switch matchsResult[2] {
	case "K","KB","KIB":
		finalSize = finalSize
	case "M","MB","MIB":
		finalSize = finalSize * 1024
	case "G","GB","GIB":
		finalSize = finalSize * 1024 * 1024
	case "T","TB","TIB":
		finalSize = finalSize * 1024 * 1024 * 1024
	}
	return finalSize,err
}
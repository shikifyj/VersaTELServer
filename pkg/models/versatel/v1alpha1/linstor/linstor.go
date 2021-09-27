package linstor

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"net/url"
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
					fmt.Println("***",mapData)
					newListData = append(newListData,mapData)
				}
			}
			d.Data = newListData
			d.Count = len(newListData)

		}
	}

	startIndex, endIndex := query.Pagination.GetValidPagination(d.Count)
	fmt.Println("startIndex:",startIndex)
	fmt.Println("endIndex:",endIndex)
	data := d.Data[startIndex:endIndex]
	d.Data = data

}


func GetClient() (*client.Client, context.Context)  {
	ctx := context.TODO()
	u, err := url.Parse("http://10.203.1.157:3370")
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
		sizeStr = strconv.FormatInt(size/(1024*1024*1024),10)+" GB"
	case size > 1024 * 1024 :
		sizeStr = strconv.FormatInt(size/(1024*1024),10)+" MB"
	case size > 1024:
		sizeStr = strconv.FormatInt(size/1024,10)+" KB"
	}

	return sizeStr
}

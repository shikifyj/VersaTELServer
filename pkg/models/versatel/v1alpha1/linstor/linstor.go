package linstor

import (
	"fmt"
	"strings"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
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



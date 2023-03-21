package backup

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

type LinstorController struct {
	IP string `yaml:"linstor"`
}

type LinstorGetter struct {
	Code  int                 `json:"code"`
	Count int                 `json:"count"`
	Data  []map[string]string `json:"data"`
}

func (d *LinstorGetter) List(query *query.Query) {

	if len(query.Filters) != 0 {
		for k, v := range query.Filters {
			newListData := make([]map[string]string, 0)
			for _, mapData := range d.Data {
				if strings.Contains(mapData[string(k)], string(v)) {
					newListData = append(newListData, mapData)
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

func GetClient(ip string) (*client.Client, context.Context) {
	ctx := context.TODO()
	controllerIP := fmt.Sprintf("http://%v", ip)
	u, err := url.Parse(controllerIP)
	if err != nil {
		log.Fatal(err)
	}

	c, err := client.NewClient(client.BaseURL(u), client.Log(log.StandardLogger()))
	if err != nil {
		log.Fatal(err)
	}
	return c, ctx
}
func FormatSize(size int64) string {
	sizeStr := ""
	if size == 9223372036854775807 {
		return sizeStr
	}
	sizefl := float64(size)
	switch {
	case size > 1024*1024*1024:
		sizeStr = fmt.Sprintf("%.2f", sizefl/(1024*1024*1024)) + " TiB"
	case size > 1024*1024:
		sizeStr = fmt.Sprintf("%.2f", sizefl/(1024*1024)) + " GiB"
		//sizeStr = strconv.FormatFloat(size/(1024*1024), 10) + " GiB"
	case size > 1024:
		sizeStr = fmt.Sprintf("%.2f", sizefl/(1024)) + " MiB"
		//sizeStr = strconv.FormatFloat(size/1024, 10) + " MiB"
	default:
		sizeStr = fmt.Sprintf("%.2f", sizefl) + " KiB"
	}

	return sizeStr
}

func ParseSize(size string) (uint64, error) {
	strSize := strings.ToUpper(size)
	str := `^([0-9.]+)(K|M|G|T)(?:I?B)?$`
	r := regexp.MustCompile(str)
	matchsResult := r.FindStringSubmatch(strSize)
	if len(matchsResult) == 0 {
		return strconv.ParseUint(size, 10, 64)
	}
	finalSize, err := strconv.ParseUint(matchsResult[1], 10, 64)
	switch matchsResult[2] {
	case "K", "KB", "KIB":
		finalSize = finalSize
	case "M", "MB", "MIB":
		finalSize = finalSize * 1024
	case "G", "GB", "GIB":
		finalSize = finalSize * 1024 * 1024
	case "T", "TB", "TIB":
		finalSize = finalSize * 1024 * 1024 * 1024
	}
	return finalSize, err
}


func ParseSizeForLvm(size string) (string, error) {
	reg := regexp.MustCompile(`[^\s]+`)

	matchsResult := reg.FindAllStringSubmatch(size, -1)

	floatSize, err := strconv.ParseFloat(matchsResult[0][0],64)
	finalSize := uint64(floatSize)

	fmt.Println("matchsResult0:",matchsResult[0])
	fmt.Println("matchsResult1:",matchsResult[1])
	switch matchsResult[1][0] {
	case "K", "KB", "KiB":
		finalSize = finalSize
	case "M", "MB", "MiB":
		finalSize = finalSize * 1024
	case "G", "GB", "GiB":
		finalSize = finalSize * 1024 * 1024
	case "T", "TB", "TiB":
		finalSize = finalSize * 1024 * 1024 * 1024
	}
	fmt.Println("finalSizeqqq:",finalSize)
	return strconv.FormatUint(finalSize, 10), err
}




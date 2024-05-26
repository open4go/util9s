package util9s

type IPInfo struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

const (
	ipAPI = "http://ip-api.com/json/"
)

func IP2Area(ip string) IPInfo {
	// http://ip-api.com/json/182.89.35.123
	resp := IPInfo{}
	err := Get(ipAPI+ip, &resp)
	if err != nil {
		return resp
	}
	return resp
}

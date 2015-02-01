package main

type Orientation struct {
	Angle int    `json:"angle"`
	Type  string `json:"type"`
}

type Screen struct {
	Height      int         `json:"height"`
	Width       int         `json:"width"`
	AvailWidth  int         `json:"availWidth"`
	AvailHeight int         `json:"availHeight"`
	ColorDepth  int         `json:"colorDepth"`
	PixelDepth  int         `json:"pixelDepth"`
	Orientation Orientation `json:"orientation"`
}

type Navigator struct {
	AppCodeName    string `json:"appCodeName"`
	AppName        string `json:"appName"`
	AppVersion     string `json:"appVersion"`
	CookieEnabled  bool   `json:"cookieEnabled"`
	DoNotTrack     string `json:"doNotTrack"`
	Language       string `json:"language"`
	MaxTouchPoints int    `json:"maxTouchPoints"`
	Platform       string `json:"platform"`
	Product        string `json:"product"`
	ProductSub     string `json:"productSub"`
	UserAgent      string `json:"userAgent"`
	Vendor         string `json:"vendor"`
	VendorSub      string `json:"vendorSub"`
}

type Location struct {
	Hash     string `json:"hash"`
	Host     string `json:"host"`
	Hostname string `json:"hostname"`
	Href     string `json:"href"`
	Origin   string `json:"origin"`
	Pathname string `json:"pathname"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Search   string `json:"search"`
}

type Stats struct {
	Timestamp   Timestamp `bson:"timestamp"`
	Screen      Screen    `json:"screen"`
	Host        string    `json:"host"`
	Navigator   Navigator `json:"navigator"`
	Location    Location  `json:"location"`
	Fingerprint int       `json:"fingerprint"`
}

type StatsMonthResult struct {
	Id   map[string]int `bson:"_id"`
	Hits int            `bson:"hits"`
}

func (self *StatsMonthResult) Day() int {
	return self.Id["day_of_month"]
}

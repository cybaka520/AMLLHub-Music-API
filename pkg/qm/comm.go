package qm

// commCommon 是 fcg 接口的 comm 字段。
//
// 字段顺序与上游 TS 实现完全一致，且必须用 struct 固定声明顺序：
// 签名针对本端实际字节，改动顺序会让黄金向量测试失效，也可能被服务端判定参数异常。
// 不要改成 map 拼装。
type commCommon struct {
	CT           int    `json:"ct"`
	CV           any    `json:"cv"` // int 或 string（搜索用 string）
	V            any    `json:"v"`
	Chid         string `json:"chid"`
	OSVer        string `json:"os_ver"`
	PhoneType    string `json:"phonetype"`
	TMEAppID     string `json:"tmeAppID"`
	NetType      string `json:"nettype"`
	UDID         string `json:"udid"`
	OpenUDID     string `json:"OpenUDID"`
	QIMEI36      string `json:"QIMEI36"`
	Uin          string `json:"uin"`
	QQ           string `json:"qq,omitempty"`
	Authst       string `json:"authst,omitempty"`
	TMELoginType int    `json:"tmeLoginType,omitempty"`
	UID          string `json:"uid,omitempty"`
	SID          string `json:"sid,omitempty"`
	UserIP       string `json:"userip,omitempty"`
}

// commPatch 覆盖 comm 中的个别字段，nil 表示保持基础值。
type commPatch struct {
	tmeAppID *string
	cv       *string // 搜索接口要求字符串
	v        *string
}

// baseComm 返回匿名态的基础 comm。
func baseComm() commCommon {
	return commCommon{
		CT:        commCT,
		CV:        commCV,
		V:         commCV,
		Chid:      commChid,
		OSVer:     commOSVer,
		PhoneType: commPhoneType,
		TMEAppID:  commAppID,
		NetType:   commNetType,
		UDID:      commZero,
		OpenUDID:  commZero,
		QIMEI36:   commZero,
		Uin:       commZero,
	}
}

// apply 按 patch 覆盖字段。
func (c *commCommon) apply(p commPatch) {
	if p.tmeAppID != nil {
		c.TMEAppID = *p.tmeAppID
	}
	if p.cv != nil {
		c.CV = *p.cv
	}
	if p.v != nil {
		c.V = *p.v
	}
}

// vkeyComm 是直链接口的 comm，字段与 fcg 完全不同，不可复用 commCommon。
type vkeyComm struct {
	Uin         string `json:"uin"` // 未登录为空串
	Format      string `json:"format"`
	CT          int    `json:"ct"`
	CV          int    `json:"cv"`
	Platform    string `json:"platform"`
	Chid        string `json:"chid"`
	Gtk         uint32 `json:"g_tk"`
	GtkNew      uint32 `json:"g_tk_new_20200303"`
	InCharset   string `json:"inCharset"`
	OutCharset  string `json:"outCharset"`
	Notice      int    `json:"notice"`
	NeedNewCode int    `json:"needNewCode"`
}

// newVkeyComm 构造直链接口 comm。g_tk 由 musickey 推出，未登录时 hash33("",5381)=5381。
func newVkeyComm(uin, musicKey string) vkeyComm {
	gtk := hash33(musicKey, 5381)
	return vkeyComm{
		Uin:         uin,
		Format:      "json",
		CT:          24,
		CV:          4747474,
		Platform:    "yqq.json",
		Chid:        "0",
		Gtk:         gtk,
		GtkNew:      gtk,
		InCharset:   "utf-8",
		OutCharset:  "utf-8",
		Notice:      0,
		NeedNewCode: 1,
	}
}

// strPtr 返回字符串指针，便于构造 commPatch。
func strPtr(s string) *string { return &s }

package service

import (
	"net/url"
	"sort"
	"strings"

	"onehost/internal/util"
)

// Yipay 易支付（国内主流第四方支付网关）封装。
// 签名规则：参数按键名升序排列拼接为 a=b&c=d，末尾拼接商户密钥后取小写 MD5。
type Yipay struct {
	Gateway string
	PID     string
	Key     string
}

// Sign 计算签名。
func (y *Yipay) Sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || k == "sign_type" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	return util.MD5Hex(sb.String() + y.Key)
}

// BuildPayURL 生成页面跳转支付地址（mapi 接口）。
func (y *Yipay) BuildPayURL(outTradeNo, name, money, payType, notifyURL, returnURL string) string {
	params := map[string]string{
		"pid":          y.PID,
		"type":         payType, // alipay / wxpay
		"out_trade_no": outTradeNo,
		"notify_url":   notifyURL,
		"return_url":   returnURL,
		"name":         name,
		"money":        money,
	}
	params["sign"] = y.Sign(params)
	params["sign_type"] = "MD5"

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	return strings.TrimRight(y.Gateway, "/") + "/mapi.php?" + q.Encode()
}

// VerifyNotify 校验异步回调签名。
func (y *Yipay) VerifyNotify(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}
	return strings.EqualFold(y.Sign(params), sign)
}

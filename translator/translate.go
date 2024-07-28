package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const deeplBaseURL = "https://api.deepl.com/jsonrpc"

var baseHeaders = map[string]string{
	"Content-Type":       "application/json",
	"Accept":             "*/*",
	"Accept-Language":    "en-US,en;q=0.9",
	"sec-ch-ua":          `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
	"sec-ch-ua-mobile":   "?0",
	"sec-ch-ua-platform": `"Windows"`,
	"sec-fetch-dest":     "empty",
	"sec-fetch-mode":     "cors",
	"sec-fetch-site":     "same-site",
	"Referer":            "https://www.deepl.com/",
	"Referrer-Policy":    "strict-origin-when-cross-origin",
}

func getICount(translateText string) int {
	count := 0
	for _, char := range translateText {
		if char == 'i' {
			count++
		}
	}
	return count
}

func getRandomNumber() int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(100000) + 83000000000
}

func getTimestamp(iCount int) int64 {
	ts := time.Now().Unix()
	if iCount == 0 {
		return ts
	}
	iCount++
	return ts - (ts % int64(iCount)) + int64(iCount)
}

type TranslateParams struct {
	Texts     []Text     `json:"texts"`
	Splitting string     `json:"splitting"`
	Lang      LangParams `json:"lang"`
	Timestamp int64      `json:"timestamp"`
}

type Text struct {
	Text                string `json:"text"`
	RequestAlternatives int    `json:"request_alternatives"`
}

type LangParams struct {
	SourceLangUserSelected string `json:"source_lang_user_selected"`
	TargetLang             string `json:"target_lang"`
}

type TranslateRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      int64           `json:"id"`
	Params  TranslateParams `json:"params"`
}

type TranslateResponse struct {
	Result struct {
		Texts []struct {
			Text string `json:"text"`
		} `json:"texts"`
	} `json:"result"`
}

func InitTranslator() {
	initCookies()
	initProxies()
}

func Translate(text, sourceLang, targetLang string, numberAlternative, tryCount int) (string, error) {
	iCount := getICount(text)
	id := getRandomNumber()
	proxy, _ := GetNextProxy()
	cookie, err := getNextCookie()
	if err != nil {
		return "", err
	}

	maxRetries := 5
	if tryCount >= maxRetries {
		log.Println("Max retry limit reached.")
		return "", fmt.Errorf("max retry limit reached")
	}

	headers := make(map[string]string)
	for k, v := range baseHeaders {
		headers[k] = v
	}

	postData := TranslateRequest{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_texts",
		ID:      id,
		Params: TranslateParams{
			Texts:     []Text{{Text: text, RequestAlternatives: numberAlternative}},
			Splitting: "newlines",
			Lang: LangParams{
				SourceLangUserSelected: sourceLang,
				TargetLang:             targetLang,
			},
			Timestamp: getTimestamp(iCount),
		},
	}

	data, err := json.Marshal(postData)
	if err != nil {
		return "", err
	}

	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}

	if proxy != "" {
		options = append(options, tls_client.WithProxyUrl(proxy))
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// 设置cookie
	parsedURL, err := url.Parse(deeplBaseURL)
	if err != nil {
		return "", err
	}
	client.SetCookies(parsedURL, []*http.Cookie{
		{
			Name:  "dl_session",
			Value: cookie,
		},
	})

	req, err := http.NewRequest(http.MethodPost, deeplBaseURL, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		MarkProxyInvalid(proxy)
		markCookieInvalid(cookie)
		log.Println("Retrying due to proxy or cookie error...")
		return Translate(text, sourceLang, targetLang, numberAlternative, tryCount+1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: %d\n", resp.StatusCode)
		return "", fmt.Errorf("error: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var translateResp TranslateResponse
	if err := json.Unmarshal(body, &translateResp); err != nil {
		return "", err
	}

	if len(translateResp.Result.Texts) == 0 {
		return "", fmt.Errorf("translation failed")
	}

	return translateResp.Result.Texts[0].Text, nil
}

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
	Jobs            []Job      `json:"jobs"`
	Lang            LangParams `json:"lang"`
	Priority        int        `json:"priority"`
	CommonJobParams JobParams  `json:"commonJobParams"`
	Timestamp       int64      `json:"timestamp"`
}

type Job struct {
	Kind               string     `json:"kind"`
	Sentences          []Sentence `json:"sentences"`
	RawEnContextBefore []string   `json:"raw_en_context_before"`
	RawEnContextAfter  []string   `json:"raw_en_context_after"`
	PreferredNumBeams  int        `json:"preferred_num_beams"`
}

type Sentence struct {
	Text   string `json:"text"`
	ID     int    `json:"id"`
	Prefix string `json:"prefix"`
}

type JobParams struct {
	Quality         string `json:"quality"`
	RegionalVariant string `json:"regional_variant"`
	Mode            string `json:"mode"`
	BrowserType     int    `json:"browser_type"`
	TextType        string `json:"text_type"`
	AdvancedMode    bool   `json:"advanced_mode"`
}

type LangParams struct {
	TargetLang         string     `json:"target_lang"`
	Preference         Preference `json:"preference"`
	SourceLangComputed string     `json:"source_lang_computed"`
}

type Preference struct {
	Weight  map[string]interface{} `json:"weight"`
	Default string                 `json:"default"`
}

type TranslateRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      int64           `json:"id"`
	Params  TranslateParams `json:"params"`
}

type TranslateResponse struct {
	Result struct {
		Translations []struct {
			Beams []struct {
				Sentences []struct {
					Text string `json:"text"`
				} `json:"sentences"`
				NumSymbols int `json:"num_symbols"`
			} `json:"beams"`
			Quality string `json:"quality"`
		} `json:"translations"`
		TargetLang            string                 `json:"target_lang"`
		SourceLang            string                 `json:"source_lang"`
		SourceLangIsConfident bool                   `json:"source_lang_is_confident"`
		DetectedLanguages     map[string]interface{} `json:"detectedLanguages"`
	} `json:"result"`
}

func InitTranslator() {
	initCookies()
	initProxies()
}

func Translate(text, sourceLang, targetLang, quality string, tryCount int) (string, error) {
	if quality == "" {
		quality = "normal"
	}

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

	priority := 1
	advancedMode := true
	if quality == "fast" {
		priority = -1
		advancedMode = false
	}

	headers := make(map[string]string)
	for k, v := range baseHeaders {
		headers[k] = v
	}

	postData := TranslateRequest{
		Jsonrpc: "2.0",
		Method:  "LMT_handle_jobs",
		ID:      id,
		Params: TranslateParams{
			Jobs: []Job{
				{
					Kind: "default",
					Sentences: []Sentence{
						{
							Text:   text,
							ID:     1,
							Prefix: "",
						},
					},
					RawEnContextBefore: []string{},
					RawEnContextAfter:  []string{},
					PreferredNumBeams:  4,
				},
			},
			Lang: LangParams{
				TargetLang: targetLang,
				Preference: Preference{
					Weight:  map[string]interface{}{},
					Default: "default",
				},
				SourceLangComputed: sourceLang,
			},
			Priority: priority,
			CommonJobParams: JobParams{
				Quality:         quality,
				RegionalVariant: "zh-Hans",
				Mode:            "translate",
				BrowserType:     1,
				TextType:        "plaintext",
				AdvancedMode:    advancedMode,
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
		return Translate(text, sourceLang, targetLang, quality, tryCount+1)
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

	if len(translateResp.Result.Translations) == 0 || len(translateResp.Result.Translations[0].Beams) == 0 {
		return "", fmt.Errorf("translation failed")
	}

	// 提取第一个翻译结果
	translatedText := translateResp.Result.Translations[0].Beams[0].Sentences[0].Text

	return translatedText, nil
}

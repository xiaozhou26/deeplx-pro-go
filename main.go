package main

import (
  "encoding/json"
  "fmt"
  "math/rand"
  "net/http"
  "strings"
  "time"
  "os"
  "github.com/gorilla/mux"
  "github.com/valyala/fasthttp"
)

const (
  DEEPL_BASE_URL = "https://api.deepl.com/jsonrpc"
  PORT           = 9000
)


var cookieValue = os.Getenv("COOKIE_VALUE")
var headers = map[string]string{
  "Content-Type":             "application/json",
  "Accept":                   "*/*",
  "Accept-Language":          "en-US,en;q=0.9",
  "sec-ch-ua":                `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`,
  "sec-ch-ua-mobile":         "?0",
  "sec-ch-ua-platform":       `"Windows"`,
  "sec-fetch-dest":           "empty",
  "sec-fetch-mode":           "cors",
  "sec-fetch-site":           "same-site",
  "cookie":                   cookieValue,
  "Referer":                  "https://www.deepl.com/",
  "Referrer-Policy":          "strict-origin-when-cross-origin",
}

type TranslationRequest struct {
  Text       string `json:"text"`
  SourceLang string `json:"source_lang"`
  TargetLang string `json:"target_lang"`
}

type TranslationResponse struct {
  Alternatives []string `json:"alternatives"`
  Code         int      `json:"code"`
  Data         string   `json:"data"`
  ID           int      `json:"id"`
  Method       string   `json:"method"`
  SourceLang   string   `json:"source_lang"`
  TargetLang   string   `json:"target_lang"`
}

type PostData struct {
  Jsonrpc string `json:"jsonrpc"`
  Method  string `json:"method"`
  ID      int    `json:"id"`
  Params  struct {
    Texts     []struct {
      Text                 string `json:"text"`
      RequestAlternatives int    `json:"requestAlternatives"`
    } `json:"texts"`
    Splitting string `json:"splitting"`
    Lang      struct {
      SourceLangUserSelected string `json:"source_lang_user_selected"`
      TargetLang             string `json:"target_lang"`
    } `json:"lang"`
    Timestamp int64 `json:"timestamp"`
  } `json:"params"`
}

func main() {
  r := mux.NewRouter()

  r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Welcome to deeplx-pro")
  }).Methods("GET")

  r.HandleFunc("/translate", TranslateHandler).Methods("POST")

  http.Handle("/", r)
  fmt.Printf("Server is running on http://localhost:%d\n", PORT)
  http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil)
}


func TranslateHandler(w http.ResponseWriter, r *http.Request) {
  var req TranslationRequest
  err := json.NewDecoder(r.Body).Decode(&req)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  resultCh := make(chan string, 1)
  errorCh := make(chan error, 1)

  go func() {
    result, err := translate(req.Text, req.SourceLang, req.TargetLang)
    if err != nil {
      errorCh <- err
      return
    }
    resultCh <- result
  }()

  select {
  case result := <-resultCh:
    responseData := TranslationResponse{
      Alternatives: []string{},
      Code:         200,
      Data:         result,
      ID:           rand.Intn(10000000000),
      Method:       "Free",
      SourceLang:   req.SourceLang,
      TargetLang:   req.TargetLang,
    }

    json.NewEncoder(w).Encode(responseData)
  case err := <-errorCh:
    http.Error(w, "Translation failed: "+err.Error(), http.StatusInternalServerError)
  }
}


func translate(text string, sourceLang string, targetLang string) (string, error) {
  iCount := strings.Count(text, "i")
  id := getRandomNumber()

  postData := PostData{
    Jsonrpc: "2.0",
    Method:  "LMT_handle_texts",
    ID:      id,
  }
  postData.Params.Texts = append(postData.Params.Texts, struct {
    Text                 string `json:"text"`
    RequestAlternatives int    `json:"requestAlternatives"`
  }{
    Text:                 text,
    RequestAlternatives: 3,
  })
  postData.Params.Splitting = "newlines"
  postData.Params.Lang.SourceLangUserSelected = strings.ToUpper(sourceLang)
  postData.Params.Lang.TargetLang = strings.ToUpper(targetLang)
  postData.Params.Timestamp = getTimestamp(iCount)

  postDataBytes, _ := json.Marshal(postData)

  req := fasthttp.AcquireRequest()
  req.SetRequestURI(DEEPL_BASE_URL)
  req.Header.SetMethod("POST")
  req.SetBody(postDataBytes)
  for k, v := range headers {
    req.Header.Set(k, v)
  }

  resp := fasthttp.AcquireResponse()
  client := &fasthttp.Client{}
  err := client.Do(req, resp)

  if err != nil {
    return "", err
  }

  if resp.StatusCode() == 429 {
    return "", fmt.Errorf("Too many requests, your IP has been blocked by DeepL temporarily, please don't request it frequently in a short time.")
  }

  if resp.StatusCode() != 200 {
    return "", fmt.Errorf("Error %d", resp.StatusCode())
  }

  var result map[string]interface{}
  json.Unmarshal(resp.Body(), &result)

  return result["result"].(map[string]interface{})["texts"].([]interface{})[0].(map[string]interface{})["text"].(string), nil
}

func getRandomNumber() int {
  rand.Seed(time.Now().UnixNano())
  return rand.Intn(8399998-8300000+1) + 8300000
}

func getTimestamp(iCount int) int64 {
  ts := time.Now().UnixNano() / int64(time.Millisecond)
  if iCount == 0 {
    return ts
  }
  iCount++
  return ts - (ts % int64(iCount)) + int64(iCount)
}

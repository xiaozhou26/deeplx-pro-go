package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"deeplx-pro/config"
	"deeplx-pro/translator"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// 初始化配置
	config.InitConfig()

	// 初始化翻译器
	translator.InitTranslator()

	port := config.AppConfig.Port

	// 设置Gin引擎
	r := gin.Default()

	// 根路由，返回欢迎信息
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to deeplx-pro")
	})

	// GET方法不支持翻译请求
	r.GET("/translate", func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "GET method not supported for this endpoint. Please use POST.")
	})

	// POST方法处理翻译请求
	r.POST("/translate", func(c *gin.Context) {
		var reqBody struct {
			Text       string `json:"text"`
			SourceLang string `json:"source_lang"`
			TargetLang string `json:"target_lang"`
		}

		// 绑定JSON请求体
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			log.Printf("Invalid request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// 调用翻译函数
		result, err := translator.Translate(reqBody.Text, reqBody.SourceLang, reqBody.TargetLang, 0, 0)
		if err != nil || result == "" {
			log.Printf("Translation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Translation failed", "details": err.Error()})
			return
		}

		// 返回翻译结果
		response := gin.H{
			"alternatives": []string{},
			"code":         200,
			"data":         result,
			"id":           rand.Intn(10000000000),
			"source_lang":  reqBody.SourceLang,
			"target_lang":  reqBody.TargetLang,
		}
		c.JSON(http.StatusOK, response)
	})

	log.Printf("Server is running on http://localhost:%s", port)
	log.Printf("DEEPL_COOKIES: %s", os.Getenv("DEEPL_COOKIES"))
	log.Printf("PROXY_LIST: %s", os.Getenv("PROXY_LIST"))

	// 设置并启动HTTP服务器
	s := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

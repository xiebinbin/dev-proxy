package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func main() {
	var (
		port int
		host string
	)
	flag.IntVar(&port, "port", 8080, "port")
	flag.StringVar(&host, "host", "127.0.0.1", "host")
	flag.Parse()
	listenAddr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("listenAddr: %s\n", listenAddr)
	jsonFile, err := os.Open("./hostmap.json")
	if err != nil {
		log.Fatalf("Error opening JSON file: %v", err)
	}
	// 读取文件内容
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	defer jsonFile.Close()
	var hostMap map[string]string
	err = json.Unmarshal(byteValue, &hostMap)
	if err != nil {
		log.Fatalf("Error while unmarshalling JSON: %v", err)
	}

	router := gin.Default()
	router.Use(func(ctx *gin.Context) {
		host := ctx.Request.Header.Get("X-Forwarded-Host")
		hostSplit := strings.Split(host, ".")
		fmt.Println("hostSplit:", hostSplit)
		if len(hostSplit) < 2 {
			ctx.String(200, "mirror manager")
			return
		}
		hostIndex := hostSplit[0]
		if _, ok := hostMap[hostIndex]; !ok {
			ctx.String(200, "mirror manager")
			return
		}
		ctx.Request.URL.Host = hostMap[hostIndex]
		ctx.Request.URL.Scheme = "https"
		url := ctx.Request.URL.String()
		fmt.Println("url:", url)
		client := resty.New()
		req := client.R()
		for key, values := range ctx.Request.Header {
			for _, value := range values {
				req.SetHeader(key, value)
			}
		}

		req.Method = ctx.Request.Method
		req.SetBody(ctx.Request.Body)
		if strings.ToUpper(ctx.Request.Method) == "GET" {
			req.SetBody(nil)
		}
		resp, err := req.SetDoNotParseResponse(true).Execute(req.Method, url)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "转发请求失败"})
			return
		}
		defer resp.RawBody().Close()
		ctx.Writer.WriteHeader(resp.StatusCode())
		for key, values := range resp.Header() {
			for _, value := range values {
				ctx.Writer.Header().Add(key, value)
			}
		}
		_, err = io.Copy(ctx.Writer, resp.RawBody())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "传输响应失败"})
			return
		}
	})

	fmt.Println("Starting Proxy on", listenAddr)
	if err := router.Run(listenAddr); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

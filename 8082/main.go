package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// Redis client
var rdb *redis.Client

func main() {
	// 初始化 Redis 客户端
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 服務器地址
		Password: "12345",
		DB:       0, // 使用的 Redis database
	})

	// 初始化 HTTP 路由
	r := mux.NewRouter()
	r.HandleFunc("/shorten", shortenURLHandler).Methods("POST")
	r.HandleFunc("/{hash}", redirectHandler).Methods("GET")

	// 啟動 HTTP 服務
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8082",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server on :8082")
	log.Fatal(srv.ListenAndServe())
}

// 短網址生成處理函數
func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}
	fmt.Println(originalURL)
	// 生成短網址的哈希值
	hash := generateHash(originalURL)
	fmt.Println(hash)
	// 存儲原始 URL 和哈希的映射關係
	err := rdb.Set(ctx, hash, originalURL, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, "Failed to shorten URL", http.StatusInternalServerError)
		return
	}

	// 返回短網址
	shortURL := fmt.Sprintf("http://localhost:8080/%s", hash)
	w.Write([]byte(shortURL))
}

// 重定向處理函數
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	hash := mux.Vars(r)["hash"]

	// 從 Redis 中查找原始 URL
	originalURL, err := rdb.Get(ctx, hash).Result()
	if err == redis.Nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve URL", http.StatusInternalServerError)
		return
	}

	// 重定向到原始 URL
	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

// 生成哈希值
func generateHash(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))[:8] // 使用前 8 個字符作為哈希值
}

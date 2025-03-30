package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// 定义结果结构体
type BizData struct {
	Fact   string `json:"fact"`
	Length int    `json:"length"`
}

func fetchURL(ctx context.Context) (bizData *BizData, err error) {
	url := "https://catfact.ninja/fact"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var data BizData
	json.Unmarshal(body, &data)

	time.Sleep(1 * time.Second)

	fmt.Println("fetch url finish")

	return &data, nil
}

func main() {
	const numRequests = 10     // 并发请求数量
	const concurrencyLimit = 3 // 并发限制
	semaphore := make(chan struct{}, concurrencyLimit)

	// 使用 errgroup 管理并发任务
	g, ctx := errgroup.WithContext(context.Background())

	retList := make([]BizData, 0)
	mu := sync.Mutex{}
	for i := 0; i < numRequests; i++ {
		g.Go(func() error {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			data, err := fetchURL(ctx)
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			retList = append(retList, *data)
			return nil
		})
	}

	// 等待所有任务完成，并检查是否有错误
	if err := g.Wait(); err != nil {
		fmt.Printf("Error occurred in one of the goroutines: %v", err)
	}

	// 打印结果
	fmt.Println("All requests completed.")
	for _, result := range retList {
		fmt.Printf("retdata: %+v\n", result)
	}
}

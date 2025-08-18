package gobergamot_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xxnuo/gobergamot"
)

// 测试文本样本
var (
	shortText  = "Hello, World!"
	mediumText = "Computers have become an integral part of our daily lives. They have a great impact on the way we live, work, and communicate."
	longText   = "Computers have become an integral part of our daily lives. They have a great impact on the way we live, work, and communicate. Computers have opened up new possibilities. Due to the Internet, students have access to information beyond traditional textbooks. They can conduct research, collaborate with peers on projects, expanding their knowledge horizons. In today's world, being computer literate is essential for future success. By integrating computers into education, students can learn how to navigate digital tools, analyze and evaluate online information, and develop problem-solving and coding skills."
)

// 测试不同长度文本的翻译速度 (英文到俄文)
func TestTranslationSpeedEnRu(t *testing.T) {
	startCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 创建翻译器
	translator, err := gobergamot.New(startCtx, gobergamot.Config{
		FilesBundle: testBundle(t),
	})
	if err != nil {
		t.Fatalf("创建翻译器失败: %v", err)
	}
	defer translator.Close(context.Background())

	// 测试短文本
	t.Run("ShortTextEnRu", func(t *testing.T) {
		benchmarkTranslation(t, translator, shortText, "短文本(英->俄)")
	})

	// 测试中等长度文本
	t.Run("MediumTextEnRu", func(t *testing.T) {
		benchmarkTranslation(t, translator, mediumText, "中等文本(英->俄)")
	})

	// 测试长文本
	t.Run("LongTextEnRu", func(t *testing.T) {
		benchmarkTranslation(t, translator, longText, "长文本(英->俄)")
	})
}

// 测试不同长度文本的翻译速度 (英文到中文)
func TestTranslationSpeedEnZh(t *testing.T) {
	startCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 创建翻译器
	translator, err := gobergamot.New(startCtx, gobergamot.Config{
		FilesBundle: testBundleEnZh(t),
	})
	if err != nil {
		t.Fatalf("创建翻译器失败: %v", err)
	}
	defer translator.Close(context.Background())

	// 测试短文本
	t.Run("ShortTextEnZh", func(t *testing.T) {
		benchmarkTranslation(t, translator, shortText, "短文本(英->中)")
	})

	// 测试中等长度文本
	t.Run("MediumTextEnZh", func(t *testing.T) {
		benchmarkTranslation(t, translator, mediumText, "中等文本(英->中)")
	})

	// 测试长文本
	t.Run("LongTextEnZh", func(t *testing.T) {
		benchmarkTranslation(t, translator, longText, "长文本(英->中)")
	})
}

// 测试翻译池的性能
func TestTranslationPoolSpeed(t *testing.T) {
	startCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 创建翻译池
	pool, err := gobergamot.NewPool(startCtx, gobergamot.PoolConfig{
		Config: gobergamot.Config{
			FilesBundle: testBundle(t),
		},
		PoolSize: 4, // 使用4个工作线程
	})
	if err != nil {
		t.Fatalf("创建翻译池失败: %v", err)
	}
	defer pool.Close(context.Background())

	// 测试短文本
	t.Run("ShortTextPool", func(t *testing.T) {
		benchmarkPoolTranslation(t, pool, shortText, "短文本(翻译池)")
	})

	// 测试中等长度文本
	t.Run("MediumTextPool", func(t *testing.T) {
		benchmarkPoolTranslation(t, pool, mediumText, "中等文本(翻译池)")
	})

	// 测试长文本
	t.Run("LongTextPool", func(t *testing.T) {
		benchmarkPoolTranslation(t, pool, longText, "长文本(翻译池)")
	})

	// 测试并发翻译
	t.Run("ConcurrentTranslation", func(t *testing.T) {
		benchmarkConcurrentTranslation(t, pool, mediumText, 10)
	})
}

// 基准测试单个翻译器的翻译速度
func benchmarkTranslation(t *testing.T, translator *gobergamot.Translator, text, label string) {
	t.Helper()

	ctx := context.Background()

	// 预热一次翻译
	_, err := translator.Translate(ctx, gobergamot.TranslationRequest{
		Text: text,
	})
	if err != nil {
		t.Fatalf("预热翻译失败: %v", err)
	}

	// 开始计时
	start := time.Now()

	// 执行翻译
	output, err := translator.Translate(ctx, gobergamot.TranslationRequest{
		Text: text,
	})
	if err != nil {
		t.Fatalf("翻译失败: %v", err)
	}

	// 计算耗时
	duration := time.Since(start)

	// 计算每秒字符数
	charsPerSecond := float64(len(text)) / duration.Seconds()

	t.Logf("%s 翻译结果: %s", label, output)
	t.Logf("%s 翻译耗时: %v, 输入字符数: %d, 输出字符数: %d, 速度: %.2f 字符/秒",
		label, duration, len(text), len(output), charsPerSecond)
}

// 基准测试翻译池的翻译速度
func benchmarkPoolTranslation(t *testing.T, pool *gobergamot.Pool, text, label string) {
	t.Helper()

	ctx := context.Background()

	// 预热一次翻译
	_, err := pool.Translate(ctx, gobergamot.TranslationRequest{
		Text: text,
	})
	if err != nil {
		t.Fatalf("预热翻译失败: %v", err)
	}

	// 开始计时
	start := time.Now()

	// 执行翻译
	output, err := pool.Translate(ctx, gobergamot.TranslationRequest{
		Text: text,
	})
	if err != nil {
		t.Fatalf("翻译失败: %v", err)
	}

	// 计算耗时
	duration := time.Since(start)

	// 计算每秒字符数
	charsPerSecond := float64(len(text)) / duration.Seconds()

	t.Logf("%s 翻译结果: %s", label, output)
	t.Logf("%s 翻译耗时: %v, 输入字符数: %d, 输出字符数: %d, 速度: %.2f 字符/秒",
		label, duration, len(text), len(output), charsPerSecond)
}

// 测试并发翻译性能
func benchmarkConcurrentTranslation(t *testing.T, pool *gobergamot.Pool, text string, concurrency int) {
	t.Helper()

	ctx := context.Background()

	// 创建通道接收结果
	resultCh := make(chan time.Duration, concurrency)

	// 开始计时
	start := time.Now()

	// 并发执行翻译
	for i := 0; i < concurrency; i++ {
		go func(i int) {
			startTime := time.Now()

			output, err := pool.Translate(ctx, gobergamot.TranslationRequest{
				Text: fmt.Sprintf("%s [%d]", text, i),
			})

			if err != nil {
				t.Errorf("并发翻译 %d 失败: %v", i, err)
				resultCh <- 0
				return
			}

			duration := time.Since(startTime)
			t.Logf("并发翻译 %d 完成, 输出长度: %d, 耗时: %v", i, len(output), duration)
			resultCh <- duration
		}(i)
	}

	// 收集结果
	var totalDuration time.Duration
	for i := 0; i < concurrency; i++ {
		d := <-resultCh
		totalDuration += d
	}

	// 计算总耗时和平均耗时
	totalTime := time.Since(start)
	avgTime := totalDuration / time.Duration(concurrency)

	t.Logf("并发翻译测试 (%d 个并发请求) - 总耗时: %v, 平均每个请求耗时: %v",
		concurrency, totalTime, avgTime)

	// 计算吞吐量 (每秒请求数)
	throughput := float64(concurrency) / totalTime.Seconds()
	t.Logf("并发翻译吞吐量: %.2f 请求/秒", throughput)
}

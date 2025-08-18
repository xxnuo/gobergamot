package gobergamot_test

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/xxnuo/gobergamot"
	"github.com/xxnuo/gobergamot/internal/wasm"
)

// MemoryStats 存储内存统计信息
type MemoryStats struct {
	HeapAlloc    uint64 // 堆分配的字节数
	HeapSys      uint64 // 从系统获取的堆内存
	HeapObjects  uint64 // 堆对象数量
	StackSys     uint64 // 栈内存使用量
	Sys          uint64 // 从系统获取的总内存
	TotalAlloc   uint64 // 总分配的内存
	Mallocs      uint64 // 内存分配次数
	Frees        uint64 // 内存释放次数
	NumGC        uint32 // GC次数
	PauseTotalNs uint64 // GC暂停总时间
	Timestamp    int64  // 时间戳
}

// getMemoryStats 获取当前内存使用情况
func getMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapObjects:  m.HeapObjects,
		StackSys:     m.StackSys,
		Sys:          m.Sys,
		TotalAlloc:   m.TotalAlloc,
		Mallocs:      m.Mallocs,
		Frees:        m.Frees,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
		Timestamp:    time.Now().UnixNano(),
	}
}

// printMemDiff 打印内存使用变化
func printMemDiff(before, after MemoryStats) {
	fmt.Printf("内存使用变化:\n")
	fmt.Printf("堆分配: %+v KB -> %+v KB (差异: %+v KB)\n",
		before.HeapAlloc/1024, after.HeapAlloc/1024, (after.HeapAlloc-before.HeapAlloc)/1024)
	fmt.Printf("堆系统: %+v KB -> %+v KB (差异: %+v KB)\n",
		before.HeapSys/1024, after.HeapSys/1024, (after.HeapSys-before.HeapSys)/1024)
	fmt.Printf("堆对象: %+v -> %+v (差异: %+v)\n",
		before.HeapObjects, after.HeapObjects, after.HeapObjects-before.HeapObjects)
	fmt.Printf("总分配: %+v KB -> %+v KB (差异: %+v KB)\n",
		before.TotalAlloc/1024, after.TotalAlloc/1024, (after.TotalAlloc-before.TotalAlloc)/1024)
	fmt.Printf("内存分配次数: %+v -> %+v (差异: %+v)\n",
		before.Mallocs, after.Mallocs, after.Mallocs-before.Mallocs)
	fmt.Printf("内存释放次数: %+v -> %+v (差异: %+v)\n",
		before.Frees, after.Frees, after.Frees-before.Frees)
	fmt.Printf("GC次数: %+v -> %+v (差异: %+v)\n",
		before.NumGC, after.NumGC, after.NumGC-before.NumGC)
	fmt.Printf("GC暂停时间: %+v ms -> %+v ms (差异: %+v ms)\n",
		before.PauseTotalNs/1000000, after.PauseTotalNs/1000000, (after.PauseTotalNs-before.PauseTotalNs)/1000000)
	fmt.Printf("执行时间: %+v ms\n", (after.Timestamp-before.Timestamp)/1000000)
}

// TestTranslatorMemoryUsage 测试翻译过程中的内存使用情况
func TestTranslatorMemoryUsage(t *testing.T) {
	// 强制GC，获取基准内存状态
	debug.FreeOSMemory()
	runtime.GC()

	ctx := context.Background()
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	// 测试不同大小文本的翻译
	testTexts := []struct {
		name string
		text string
	}{
		{
			name: "短文本",
			text: "Hello, World!",
		},
		{
			name: "中等文本",
			text: "Computers have become an integral part of our daily lives. They have a great impact on the way we live, work, and communicate.",
		},
		{
			name: "长文本",
			text: "Computers have become an integral part of our daily lives. They have a great impact on the way we live, work, and communicate. Computers have opened up new possibilities. Due to the Internet, students have access to information beyond traditional textbooks. They can conduct research, collaborate with peers on projects, expanding their knowledge horizons. In today's world, being computer literate is essential for future success. By integrating computers into education, students can learn how to navigate digital tools, analyze and evaluate online information, and develop problem-solving and coding skills. The rapid advancement of technology has revolutionized various sectors, from healthcare to transportation. Artificial intelligence and machine learning are being integrated into everyday applications, making our lives more convenient and efficient. As we continue to embrace technological innovations, it is crucial to address concerns related to privacy, security, and ethical considerations. The digital divide remains a challenge, with disparities in access to technology and digital literacy. Efforts should be made to ensure that the benefits of technology reach all segments of society.",
		},
		{
			name: "HTML文本",
			text: "<div><h1>Hello World</h1><p>This is a test of <strong>HTML</strong> translation.</p><a href=\"https://example.com\">Link</a></div>",
		},
	}

	// 测试创建翻译器的内存使用
	fmt.Println("===== 测试创建翻译器的内存使用 =====")
	beforeInit := getMemoryStats()

	translator, err := gobergamot.New(ctx, gobergamot.Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: stderr,
			Stdout: stdout,
		},
		FilesBundle: testBundleEnZh(t),
	})

	afterInit := getMemoryStats()

	if err != nil {
		t.Fatalf("创建翻译器失败: %v", err)
	}

	printMemDiff(beforeInit, afterInit)
	fmt.Println()

	defer func() {
		if err := translator.Close(ctx); err != nil {
			t.Fatalf("关闭翻译器失败: %v", err)
		}
	}()

	// 测试不同文本的翻译内存使用
	for _, tt := range testTexts {
		fmt.Printf("===== 测试翻译 %s 的内存使用 =====\n", tt.name)

		// 强制GC，获取基准内存状态
		debug.FreeOSMemory()
		runtime.GC()
		time.Sleep(100 * time.Millisecond) // 等待GC完成

		beforeTranslate := getMemoryStats()

		request := gobergamot.TranslationRequest{
			Text: tt.text,
			Options: gobergamot.TranslationOptions{
				HTML: tt.name == "HTML文本",
			},
		}

		output, err := translator.Translate(ctx, request)

		afterTranslate := getMemoryStats()

		if err != nil {
			t.Fatalf("翻译失败: %v", err)
		}

		fmt.Printf("原文长度: %d 字节, 译文长度: %d 字节\n", len(tt.text), len(output))
		printMemDiff(beforeTranslate, afterTranslate)
		fmt.Println()
	}

	// 测试多次翻译同一文本的内存变化
	fmt.Println("===== 测试多次翻译同一文本的内存变化 =====")

	// 强制GC，获取基准内存状态
	debug.FreeOSMemory()
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // 等待GC完成

	beforeMulti := getMemoryStats()

	// 执行10次相同的翻译
	const repeatCount = 10
	sampleText := "This is a sample text for repeated translation to test memory usage patterns."
	request := gobergamot.TranslationRequest{
		Text: sampleText,
	}

	for i := 0; i < repeatCount; i++ {
		_, err := translator.Translate(ctx, request)
		if err != nil {
			t.Fatalf("第 %d 次翻译失败: %v", i+1, err)
		}
	}

	afterMulti := getMemoryStats()

	fmt.Printf("执行 %d 次相同文本翻译的内存变化:\n", repeatCount)
	printMemDiff(beforeMulti, afterMulti)
	fmt.Printf("平均每次翻译内存分配: %+v KB\n", (afterMulti.TotalAlloc-beforeMulti.TotalAlloc)/(1024*repeatCount))
}

// TestTranslatorMemoryLeakCheck 测试翻译器是否存在内存泄漏
func TestTranslatorMemoryLeakCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存泄漏测试")
	}

	ctx := context.Background()
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	// 创建翻译器
	translator, err := gobergamot.New(ctx, gobergamot.Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: stderr,
			Stdout: stdout,
		},
		FilesBundle: testBundleEnZh(t),
	})

	if err != nil {
		t.Fatalf("创建翻译器失败: %v", err)
	}

	defer func() {
		if err := translator.Close(ctx); err != nil {
			t.Fatalf("关闭翻译器失败: %v", err)
		}
	}()

	// 强制GC，获取基准内存状态
	debug.FreeOSMemory()
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // 等待GC完成

	var initialStats MemoryStats
	var finalStats MemoryStats

	// 执行一组翻译，记录初始内存状态
	sampleText := "This is a sample text for memory leak test."
	request := gobergamot.TranslationRequest{Text: sampleText}

	// 先执行几次翻译，让内存使用稳定下来
	for i := 0; i < 5; i++ {
		_, err := translator.Translate(ctx, request)
		if err != nil {
			t.Fatalf("预热翻译失败: %v", err)
		}
	}

	// 强制GC，获取初始内存状态
	debug.FreeOSMemory()
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // 等待GC完成
	initialStats = getMemoryStats()

	// 执行大量翻译，检查内存是否持续增长
	const iterations = 100
	fmt.Printf("执行 %d 次翻译检查内存泄漏...\n", iterations)

	for i := 0; i < iterations; i++ {
		_, err := translator.Translate(ctx, request)
		if err != nil {
			t.Fatalf("第 %d 次翻译失败: %v", i+1, err)
		}

		if i%10 == 0 {
			// 每10次打印一次内存状态
			var currentStats = getMemoryStats()
			fmt.Printf("迭代 %d: 堆分配 %d KB, 堆对象 %d\n",
				i, currentStats.HeapAlloc/1024, currentStats.HeapObjects)
		}
	}

	// 强制GC，获取最终内存状态
	debug.FreeOSMemory()
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // 等待GC完成
	finalStats = getMemoryStats()

	// 比较初始和最终内存状态
	fmt.Println("\n===== 内存泄漏检测结果 =====")
	fmt.Printf("初始堆分配: %d KB, 最终堆分配: %d KB, 差异: %d KB\n",
		initialStats.HeapAlloc/1024, finalStats.HeapAlloc/1024,
		(finalStats.HeapAlloc-initialStats.HeapAlloc)/1024)
	fmt.Printf("初始堆对象: %d, 最终堆对象: %d, 差异: %d\n",
		initialStats.HeapObjects, finalStats.HeapObjects,
		finalStats.HeapObjects-initialStats.HeapObjects)

	// 判断是否可能存在内存泄漏
	// 这里设置一个阈值，如果内存增长超过这个阈值，则认为可能存在内存泄漏
	const leakThresholdKB = 1000 // 1MB
	if finalStats.HeapAlloc > initialStats.HeapAlloc+leakThresholdKB*1024 {
		t.Errorf("可能存在内存泄漏: 内存增长 %d KB 超过阈值 %d KB",
			(finalStats.HeapAlloc-initialStats.HeapAlloc)/1024, leakThresholdKB)
	} else {
		fmt.Println("未检测到明显的内存泄漏")
	}
}

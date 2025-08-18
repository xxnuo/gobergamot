package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xxnuo/gobergamot"
)

func main() {
	// 定义命令行参数
	modelPath := flag.String("model", "", "模型文件路径 (必需)")
	modelPathShort := flag.String("m", "", "模型文件路径简写 (必需)")
	lexPath := flag.String("lex", "", "词典短列表文件路径 (必需)")
	lexPathShort := flag.String("l", "", "词典短列表文件路径简写 (必需)")
	vocabPath := flag.String("vocab", "", "源语言词汇表文件路径 (必需)")
	vocabPathShort := flag.String("v", "", "源语言词汇表文件路径简写 (必需)")
	vocab2Path := flag.String("vocab2", "", "目标语言词汇表文件路径")
	vocab2PathShort := flag.String("v2", "", "目标语言词汇表文件路径简写")

	// 解析命令行参数
	flag.Parse()

	// 获取实际参数值（优先使用长参数）
	model := getParam(*modelPath, *modelPathShort)
	lex := getParam(*lexPath, *lexPathShort)
	vocab := getParam(*vocabPath, *vocabPathShort)
	vocab2 := getParam(*vocab2Path, *vocab2PathShort)

	// 验证必需参数
	if model == "" || lex == "" || vocab == "" {
		fmt.Println("错误: 必须提供模型、词典短列表和词汇表文件路径")
		fmt.Println("用法: ./mt --model <模型文件> --lex <词典短列表文件> --vocab <源语言词汇表> [--vocab2 <目标语言词汇表>] [待翻译文本]")
		fmt.Println("   或: ./mt --m <模型文件> --l <词典短列表文件> --v <源语言词汇表> [--v2 <目标语言词汇表>] [待翻译文本]")
		os.Exit(1)
	}

	// 获取待翻译文本
	var text string
	if args := flag.Args(); len(args) > 0 {
		text = strings.Join(args, " ")
	} else {
		// 如果命令行没有提供文本，则从标准输入读取
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取标准输入错误: %v\n", err)
			os.Exit(1)
		}
		text = string(bytes)
	}

	// 如果没有文本可翻译，显示错误信息
	if text == "" {
		fmt.Println("错误: 没有提供待翻译文本")
		fmt.Println("请在命令行参数中提供文本或通过标准输入传入")
		os.Exit(1)
	}

	// 打开文件
	modelFile, err := os.Open(model)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开模型文件错误: %v\n", err)
		os.Exit(1)
	}
	defer modelFile.Close()

	lexFile, err := os.Open(lex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开词典短列表文件错误: %v\n", err)
		os.Exit(1)
	}
	defer lexFile.Close()

	vocabFile, err := os.Open(vocab)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开词汇表文件错误: %v\n", err)
		os.Exit(1)
	}
	defer vocabFile.Close()

	// 准备翻译器配置
	bundle := gobergamot.FilesBundle{
		Model:            modelFile,
		LexicalShortlist: lexFile,
		Vocabularies:     []io.Reader{vocabFile},
	}

	// 如果提供了第二个词汇表，添加到配置中
	if vocab2 != "" {
		vocab2File, err := os.Open(vocab2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "打开目标语言词汇表文件错误: %v\n", err)
			os.Exit(1)
		}
		defer vocab2File.Close()
		bundle.Vocabularies = append(bundle.Vocabularies, vocab2File)
	}

	// 创建翻译器配置
	config := gobergamot.Config{
		FilesBundle:     bundle,
		CacheSize:       1000, // 设置适当的缓存大小
		BergamotOptions: gobergamot.DefaultBergamotOptions(),
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 创建翻译器
	translator, err := gobergamot.New(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建翻译器错误: %v\n", err)
		os.Exit(1)
	}
	defer translator.Close(ctx)

	// 执行翻译
	request := gobergamot.TranslationRequest{
		Text: text,
		Options: gobergamot.TranslationOptions{
			HTML: false, // 不处理HTML标签
		},
	}

	result, err := translator.Translate(ctx, request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "翻译错误: %v\n", err)
		os.Exit(1)
	}

	// 输出翻译结果
	fmt.Println(result)
}

// 获取参数值，优先使用长参数
func getParam(long, short string) string {
	if long != "" {
		return long
	}
	return short
}

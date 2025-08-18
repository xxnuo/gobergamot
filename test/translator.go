package gobergamot_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/tetratelabs/wazero"

	"github.com/xxnuo/gobergamot"
	"github.com/xxnuo/gobergamot/internal/wasm"
)

func TestTranslator_New(t *testing.T) {
	ctx := context.Background()
	cache := wazero.NewCompilationCache()
	tests := []struct {
		name    string
		cfg     gobergamot.Config
		wantErr bool
	}{
		{
			name: "no model",
			cfg: gobergamot.Config{
				WASMCache: cache,
			},
			wantErr: true,
		},
		{
			name: "no lexical shortlist",
			cfg: gobergamot.Config{
				FilesBundle: gobergamot.FilesBundle{
					Model:            bytes.NewReader(nil),
					LexicalShortlist: nil,
					Vocabularies:     []io.Reader{bytes.NewReader(nil)},
				},
				WASMCache: cache,
			},
			wantErr: true,
		},
		{
			name: "no vocabulary",
			cfg: gobergamot.Config{
				FilesBundle: gobergamot.FilesBundle{
					Model:            bytes.NewReader(nil),
					LexicalShortlist: bytes.NewReader(nil),
					Vocabularies:     nil,
				},
				WASMCache: cache,
			},
			wantErr: true,
		},
		{
			name: "invalid files",
			cfg: gobergamot.Config{
				FilesBundle: gobergamot.FilesBundle{
					Model:            bytes.NewReader(nil),
					LexicalShortlist: bytes.NewReader(nil),
					Vocabularies:     []io.Reader{bytes.NewReader(nil)},
				},
				WASMCache: cache,
			},
			wantErr: true,
		},
		{
			name: "valid",
			cfg: gobergamot.Config{
				FilesBundle: testBundle(t),
				WASMCache:   cache,
			},
			wantErr: false,
		},
		{
			name: "valid with slow paths",
			cfg: gobergamot.Config{
				FilesBundle: strictReaderWrapper(testBundle(t)),
				WASMCache:   cache,
			},
			wantErr: false,
		},
		{
			name: "valid with two vocabularies",
			cfg: gobergamot.Config{
				FilesBundle: testBundleWithTwoVocabularies(t),
				WASMCache:   cache,
			},
			wantErr: false,
		},
		{
			name: "valid with enzh model",
			cfg: gobergamot.Config{
				FilesBundle: testBundleEnZh(t),
				WASMCache:   cache,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
			tt.cfg.CompileConfig.Stdout = stdout
			tt.cfg.CompileConfig.Stderr = stderr
			translator, err := gobergamot.New(ctx, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"New() error = %v, wantErr %t\n\nstdout: %s\n\nstderr: %s",
					err,
					tt.wantErr,
					stdout.String(),
					stderr.String(),
				)
			}
			if tt.wantErr {
				return
			}
			if err := translator.Close(ctx); err != nil {
				t.Fatalf("Translator.Close() error = %v", err)
			}
		})
	}
}

func TestTranslator_Translate(t *testing.T) {
	ctx := context.Background()

	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	translator, err := gobergamot.New(ctx, gobergamot.Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: stderr,
			Stdout: stdout,
		},
		FilesBundle: testBundle(t),
	})
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}
	defer func() {
		if err := translator.Close(ctx); err != nil {
			t.Fatalf("failed to close translator: %v", err)
		}
	}()

	tests := []struct {
		name         string
		request      gobergamot.TranslationRequest
		wantErr      bool
		wantedOutput string
	}{
		{
			name: "Hello World!",
			request: gobergamot.TranslationRequest{
				Text: "Hello, World!",
			},
			wantedOutput: "Здравствуйте, Мир!",
		},
		{
			name: "error",
			request: gobergamot.TranslationRequest{
				Text: "Yaml: line 2: mapping values are not allowed in this context",
			},
			wantedOutput: "Ямл: строка 2: картирование значений не допускается в этом контексте",
		},
		{
			name: "error 2",
			request: gobergamot.TranslationRequest{
				Text: "Invalid format: invalid regex",
			},
			wantedOutput: "Неверный формат: недействительный regex",
		},
		{
			name: "text",
			request: gobergamot.TranslationRequest{
				Text: "Computers have become an integral part of our daily lives. They have a great impact on the way we live, work, and communicate. Computers have opened up new possibilities. Due to the Internet, students have access to information beyond traditional textbooks. They can conduct research, collaborate with peers on projects, expanding their knowledge horizons. In today’s world, being computer literate is essential for future success. By integrating computers into education, students can learn how to navigate digital tools, analyze and evaluate online information, and develop problem-solving and coding skills. ",
			},
			wantedOutput: "Компьютеры стали неотъемлемой частью нашей повседневной жизни. Они оказывают большое влияние на то, как мы живем, работаем и общаемся. Компьютеры открыли новые возможности. В связи с Интернетом, студенты имеют доступ к информации, помимо традиционных учебников. Они могут проводить исследования, сотрудничать с коллегами по проектам, расширяя горизонты своих знаний. В современном мире быть грамотным компьютером имеет важное значение для будущего успеха. Интегрируя компьютеры в образование, студенты могут узнать, как ориентироваться в цифровых инструментах, анализировать и оценивать онлайн-информацию, а также развивать навыки решения проблем и кодирования. ",
		},
		{
			name: "html hello world",
			request: gobergamot.TranslationRequest{
				Text:    "<a href=\"link.com/path/endpoint?query=parameter\">Hello, World!</a>",
				Options: gobergamot.TranslationOptions{HTML: true},
			},
			wantedOutput: "<a href=\"link.com/path/endpoint?query=parameter\">Здравствуйте, Мир!</a>",
		},
	}

	for _, tt := range tests {
		stdout.Reset()
		stderr.Reset()
		t.Run(tt.name, func(t *testing.T) {
			output, err := translator.Translate(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"got error %v, expected wantErr %t\n\nstdout: %s\n\nstderr: %s",
					err,
					tt.wantErr,
					stdout.String(),
					stderr.String(),
				)
			}
			if output != tt.wantedOutput {
				t.Errorf("\nexpected: %s\ngot: %s", tt.wantedOutput, output)
			}
		})
	}
}

func TestTranslator_TranslateEnZh(t *testing.T) {
	ctx := context.Background()

	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	translator, err := gobergamot.New(ctx, gobergamot.Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: stderr,
			Stdout: stdout,
		},
		FilesBundle: testBundleEnZh(t),
	})
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}
	defer func() {
		if err := translator.Close(ctx); err != nil {
			t.Fatalf("failed to close translator: %v", err)
		}
	}()

	tests := []struct {
		name         string
		request      gobergamot.TranslationRequest
		wantErr      bool
		wantedOutput string
	}{
		{
			name: "Hello World!",
			request: gobergamot.TranslationRequest{
				Text: "Hello, World!",
			},
			wantedOutput: "你好,世界!",
		},
		{
			name: "text",
			request: gobergamot.TranslationRequest{
				Text: "Computers have become an integral part of our daily lives.",
			},
			wantedOutput: "计算机已经成为我们日常生活中不可或缺的一部分。",
		},
		{
			name: "html hello world",
			request: gobergamot.TranslationRequest{
				Text:    "<a href=\"link.com/path/endpoint?query=parameter\">Hello, World!</a>",
				Options: gobergamot.TranslationOptions{HTML: true},
			},
			wantedOutput: "<a href=\"link.com/path/endpoint?query=parameter\">你好,世界!</a>",
		},
	}

	for _, tt := range tests {
		stdout.Reset()
		stderr.Reset()
		t.Run(tt.name, func(t *testing.T) {
			output, err := translator.Translate(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"got error %v, expected wantErr %t\n\nstdout: %s\n\nstderr: %s",
					err,
					tt.wantErr,
					stdout.String(),
					stderr.String(),
				)
			}
			if output != tt.wantedOutput {
				t.Errorf("\nexpected: %s\ngot: %s", tt.wantedOutput, output)
			}
		})
	}
}

// 从文件路径加载模型文件
func loadModelFile(path string) (io.Reader, error) {
	// 获取项目根目录
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(projectRoot, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	// 读取文件内容到内存中
	data, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	file.Close()
	return bytes.NewBuffer(data), nil
}

// 获取项目根目录
func getProjectRoot() (string, error) {
	// 当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 如果当前目录是test子目录，则返回父目录
	if filepath.Base(cwd) == "test" {
		return filepath.Dir(cwd), nil
	}

	// 检查是否已经在项目根目录
	if _, err := os.Stat(filepath.Join(cwd, "models")); err == nil {
		return cwd, nil
	}

	// 向上查找直到找到包含models目录的目录
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "models")); err == nil {
			return dir, nil
		}
	}

	return "", errors.New("cannot find project root directory containing models folder")
}

// 模型文件路径
var (
	testModelPath      = filepath.Join("models", "enru", "model.enru.intgemm.alphas.bin")
	testShortlistPath  = filepath.Join("models", "enru", "lex.50.50.enru.s2t.bin")
	testVocabularyPath = filepath.Join("models", "enru", "vocab.enru.spm")

	testModelEnZhPath         = filepath.Join("models", "enzh", "model.enzh.intgemm.alphas.bin")
	testShortlistEnZhPath     = filepath.Join("models", "enzh", "lex.50.50.enzh.s2t.bin")
	testSrcVocabularyEnZhPath = filepath.Join("models", "enzh", "srcvocab.enzh.spm")
	testTrgVocabularyEnZhPath = filepath.Join("models", "enzh", "trgvocab.enzh.spm")
)

func testBundle(t *testing.T) gobergamot.FilesBundle {
	if t != nil {
		t.Helper()
	}

	model, err := loadModelFile(testModelPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load model file: %v", err)
	}

	shortlist, err := loadModelFile(testShortlistPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load shortlist file: %v", err)
	}

	vocabulary, err := loadModelFile(testVocabularyPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load vocabulary file: %v", err)
	}

	return gobergamot.FilesBundle{
		Model:            model,
		LexicalShortlist: shortlist,
		Vocabularies:     []io.Reader{vocabulary},
	}
}

func testBundleWithTwoVocabularies(t *testing.T) gobergamot.FilesBundle {
	if t != nil {
		t.Helper()
	}

	model, err := loadModelFile(testModelPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load model file: %v", err)
	}

	shortlist, err := loadModelFile(testShortlistPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load shortlist file: %v", err)
	}

	vocabulary, err := loadModelFile(testVocabularyPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load vocabulary file: %v", err)
	}

	vocabulary2, err := loadModelFile(testVocabularyPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load second vocabulary file: %v", err)
	}

	return gobergamot.FilesBundle{
		Model:            model,
		LexicalShortlist: shortlist,
		Vocabularies:     []io.Reader{vocabulary, vocabulary2},
	}
}

func testBundleEnZh(t *testing.T) gobergamot.FilesBundle {
	if t != nil {
		t.Helper()
	}

	model, err := loadModelFile(testModelEnZhPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load enzh model file: %v", err)
	}

	shortlist, err := loadModelFile(testShortlistEnZhPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load enzh shortlist file: %v", err)
	}

	srcVocabulary, err := loadModelFile(testSrcVocabularyEnZhPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load enzh source vocabulary file: %v", err)
	}

	trgVocabulary, err := loadModelFile(testTrgVocabularyEnZhPath)
	if err != nil && t != nil {
		t.Fatalf("failed to load enzh target vocabulary file: %v", err)
	}

	return gobergamot.FilesBundle{
		Model:            model,
		LexicalShortlist: shortlist,
		Vocabularies:     []io.Reader{srcVocabulary, trgVocabulary},
	}
}

// wrapping files readers to avoid readers type assertion which is done for fast size/data access.
func strictReaderWrapper(bundle gobergamot.FilesBundle) gobergamot.FilesBundle {
	vocabularies := make([]io.Reader, len(bundle.Vocabularies))
	for i, vocab := range bundle.Vocabularies {
		vocabularies[i] = readerWrapper{r: vocab}
	}

	return gobergamot.FilesBundle{
		Model:            readerWrapper{r: bundle.Model},
		LexicalShortlist: readerWrapper{r: bundle.LexicalShortlist},
		Vocabularies:     vocabularies,
	}
}

type readerWrapper struct {
	r io.Reader
}

func (r readerWrapper) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

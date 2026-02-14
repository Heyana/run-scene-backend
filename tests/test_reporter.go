package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TestResult 测试结果
type TestResult struct {
	Timestamp   string              `json:"timestamp"`
	Status      string              `json:"status"` // PASS, FAIL
	TotalTests  int                 `json:"total_tests"`
	PassedTests int                 `json:"passed_tests"`
	FailedTests int                 `json:"failed_tests"`
	Duration    string              `json:"duration"`
	Tests       []TestCaseResult    `json:"tests"`
	Summary     string              `json:"summary"`
}

// TestCaseResult 单个测试用例结果
type TestCaseResult struct {
	Name     string   `json:"name"`
	Status   string   `json:"status"` // PASS, FAIL
	Duration string   `json:"duration"`
	Error    string   `json:"error,omitempty"`
	SubTests []string `json:"sub_tests,omitempty"`
}

// TestReporter 测试报告器
type TestReporter struct {
	startTime time.Time
	results   []TestCaseResult
}

// NewTestReporter 创建测试报告器
func NewTestReporter() *TestReporter {
	return &TestReporter{
		startTime: time.Now(),
		results:   make([]TestCaseResult, 0),
	}
}

// AddResult 添加测试结果
func (tr *TestReporter) AddResult(name, status, duration, errorMsg string, subTests []string) {
	tr.results = append(tr.results, TestCaseResult{
		Name:     name,
		Status:   status,
		Duration: duration,
		Error:    errorMsg,
		SubTests: subTests,
	})
}

// SaveResults 保存测试结果
func (tr *TestReporter) SaveResults() error {
	// 确保 results 目录存在
	resultsDir := "./results"
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return fmt.Errorf("创建 results 目录失败: %v", err)
	}

	// 计算统计信息
	totalTests := len(tr.results)
	passedTests := 0
	failedTests := 0
	for _, result := range tr.results {
		if result.Status == "PASS" {
			passedTests++
		} else {
			failedTests++
		}
	}

	// 生成测试结果
	duration := time.Since(tr.startTime)
	status := "PASS"
	if failedTests > 0 {
		status = "FAIL"
	}

	summary := fmt.Sprintf("总计 %d 个测试，通过 %d 个，失败 %d 个", totalTests, passedTests, failedTests)

	result := TestResult{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		Status:      status,
		TotalTests:  totalTests,
		PassedTests: passedTests,
		FailedTests: failedTests,
		Duration:    duration.String(),
		Tests:       tr.results,
		Summary:     summary,
	}

	// 生成文件名
	filename := fmt.Sprintf("test_result_%s.json", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(resultsDir, filename)

	// 保存为 JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化测试结果失败: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("保存测试结果失败: %v", err)
	}

	// 同时保存一份最新结果
	latestPath := filepath.Join(resultsDir, "latest.json")
	if err := os.WriteFile(latestPath, data, 0644); err != nil {
		return fmt.Errorf("保存最新测试结果失败: %v", err)
	}

	// 打印摘要
	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Printf("测试结果已保存到: %s\n", filePath)
	fmt.Printf("状态: %s\n", status)
	fmt.Printf("总计: %d 个测试\n", totalTests)
	fmt.Printf("通过: %d 个\n", passedTests)
	fmt.Printf("失败: %d 个\n", failedTests)
	fmt.Printf("耗时: %s\n", duration.String())
	fmt.Println(separator + "\n")

	return nil
}

// 全局测试报告器
var GlobalReporter *TestReporter

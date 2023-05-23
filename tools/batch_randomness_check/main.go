package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Trisia/randomness"
	"github.com/Trisia/randomness/detect"
)

type R struct {
	Name string
	P    []float64
	Q    []float64
}

func worker(jobs chan string, out chan *R) {
	for filename := range jobs {
		buf, _ := os.ReadFile(filename)
		bits := randomness.B2bitArr(buf)
		arrP := make([]float64, 0, 25)
		arrQ := make([]float64, 0, 25)

		p, q := randomness.MonoBitFrequencyTest(bits)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.FrequencyWithinBlockTest(bits)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.PokerProto(bits, 4)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.PokerProto(bits, 8)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)

		p1, p2, q1, q2 := randomness.OverlappingTemplateMatchingProto(bits, 3)
		arrP = append(arrP, p1, p2)
		arrQ = append(arrQ, q1, q2)
		p1, p2, q1, q2 = randomness.OverlappingTemplateMatchingProto(bits, 5)
		arrP = append(arrP, p1, p2)
		arrQ = append(arrQ, q1, q2)

		p, q = randomness.RunsTest(bits)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.RunsDistributionTest(bits)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.LongestRunOfOnesInABlockTest(bits, true)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)

		p, q = randomness.BinaryDerivativeProto(bits, 3)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.BinaryDerivativeProto(bits, 7)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)

		p, q = randomness.AutocorrelationProto(bits, 1)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.AutocorrelationProto(bits, 2)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.AutocorrelationProto(bits, 8)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.AutocorrelationProto(bits, 16)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)

		p, q = randomness.MatrixRankTest(bits)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.CumulativeTest(bits, true)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.ApproximateEntropyProto(bits, 2)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.ApproximateEntropyProto(bits, 5)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.LinearComplexityProto(bits, 500)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.LinearComplexityProto(bits, 1000)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.MaurerUniversalTest(bits)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)
		p, q = randomness.DiscreteFourierTransformTest(bits)
		arrP = append(arrP, p)
		arrQ = append(arrQ, q)

		fmt.Printf(">> 检测结束 文件 %s\n", filename)
		go func(file string) {
			out <- &R{path.Base(file), arrP, arrQ}
		}(filename)
	}
}

// 结果集写入文件工作器
func writeResult(w io.StringWriter, r *R) {
	w.WriteString(r.Name)
	for j := 0; j < len(r.P); j++ {
		_, _ = w.WriteString(fmt.Sprintf(", %0.6f", r.P[j]))
	}
	_, _ = w.WriteString("\n")
}

var (
	inputPath  string // 参数文件输入路径
	reportPath string // 生成的监测报告位置
)

func init() {
	flag.StringVar(&inputPath, "i", "data", "batched待检测随机数文件位置")
	flag.StringVar(&reportPath, "o", "report", "报告文件位置")
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `batch randomness 随机性检测

batch_randomness_check -i 待检测数据目录 [-o 生成报告位置]

	示例: rddetector -i rand_data -o report
	
	输入是一个根文件夹，下面包含多份数据的子文件夹
	输出是一个文件夹，下面包含以每一个子目录命名的报告csv文件

`)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if inputPath == "" {
		fmt.Fprintf(os.Stderr, "	-i 参数缺失\n\n")
		flag.Usage()
		return
	}
	_ = os.MkdirAll(reportPath, os.FileMode(0755))

	subDirs, err := os.ReadDir(inputPath)
	if err != nil {
		panic(err)
	}

	for i, subDir := range subDirs {
		if !subDir.IsDir() {
			continue
		}

		dataSetName := subDir.Name()
		dataSetDir := path.Join(inputPath, dataSetName)
		reportPath := path.Join(reportPath, dataSetName+"_report.csv")
		start := time.Now()
		runDataSet(dataSetDir, reportPath)
		dur := time.Since(start)
		fmt.Printf("finish run %d of %d in %s \n", i, len(subDirs), dur.String())
	}
}

func runDataSet(dataDir, reportPath string) {
	n := runtime.NumCPU()
	jobs := make(chan string)

	w, err := os.Create(reportPath)
	if err != nil {
		fmt.Fprint(os.Stderr, "无法打开写入文件 "+reportPath)
		return
	}
	defer w.Close()
	_, _ = w.WriteString(
		"源数据," +
			"单比特频数检测," +
			"块内频数检测 m=10000," +
			"扑克检测 m=4," +
			"扑克检测 m=8," +
			"重叠子序列检测 m=3 P1,重叠子序列检测 m=2 P2," +
			"重叠子序列检测 m=5 P1,重叠子序列检测 m=5 P2," +
			"游程总数检测," +
			"游程分布检测," +
			"块内最大游程检测 m=10000," +
			"二元推导检测 k=3," +
			"二元推导检测 k=7," +
			"自相关检测 d=1," +
			"自相关检测 d=2," +
			"自相关检测 d=8," +
			"自相关检测 d=16," +
			"矩阵秩检测," +
			"累加和检测," +
			"近似熵检测 m=2," +
			"近似熵检测 m=5," +
			"线性复杂度检测 m=500," +
			"线性复杂度检测 m=1000," +
			"Maurer通用统计检测 L=7 Q=1280," +
			"离散傅里叶检测\n")
	s := toBeTestFileNum(dataDir)
	out := make(chan *R)

	fmt.Printf(">> 开始执行随机性检测，待检测样本数 s = %d\n", s)

	// 检测工作器
	for i := 0; i < n; i++ {
		go worker(jobs, out)
	}
	// 结果工作器
	go func() {
		filepath.Walk(dataDir, func(p string, _ fs.FileInfo, _ error) error {
			if strings.HasSuffix(p, ".bin") || strings.HasSuffix(p, ".dat") {
				jobs <- p
			}
			return nil
		})
		close(jobs)
	}()

	check_result := make([]*R, 0, s)
	for i := 0; i < s; i++ {
		r := <-out
		check_result = append(check_result, r)
		writeResult(w, r)
	}

	dataSetResult := parseResult(check_result)
	pass := isPass(dataSetResult)

	_, _ = w.WriteString("总计")
	for i := 0; i < len(dataSetResult); i++ {
		_, _ = w.WriteString(fmt.Sprintf(", %d", dataSetResult[i].passes))
	}
	_, _ = w.WriteString("\n")

	_, _ = w.WriteString("Q分布")
	for i := 0; i < len(dataSetResult); i++ {
		_, _ = w.WriteString(fmt.Sprintf(", %0.6f", dataSetResult[i].qDis))
	}
	_, _ = w.WriteString("\n")

	_, _ = w.WriteString("结果")
	for i := 0; i < len(dataSetResult); i++ {
		_, _ = w.WriteString(fmt.Sprintf(", %t", pass))
	}
	_, _ = w.WriteString("\n")

	fmt.Println(">> 检测完成 检测报告: ", reportPath, " 结果: ", pass)
}

func toBeTestFileNum(p string) int {
	cnt := 0
	// 结果工作器
	filepath.Walk(p, func(p string, _ fs.FileInfo, _ error) error {
		if strings.HasSuffix(p, ".bin") || strings.HasSuffix(p, ".dat") {
			cnt++
		}

		return nil
	})
	return cnt
}

type DataSetResult struct {
	passes int
	qDis   float64
}

func parseResult(results []*R) []DataSetResult {
	o := make([]DataSetResult, 25)
	for i := 0; i < 25; i++ {
		qValues := make([]float64, 0, len(results))
		for _, r := range results {
			qValues = append(qValues, r.Q[i])
			if r.P[i] > 0.01 {
				o[i].passes++
			}
			o[i].qDis = detect.ThresholdQ(qValues)
		}
	}

	return o
}

func isPass(r []DataSetResult) bool {
	pass := true
	for _, rr := range r {
		if rr.passes < 981 {
			pass = false
		}
		if rr.qDis < 0.0001 {
			pass = false
		}
	}
	return pass
}

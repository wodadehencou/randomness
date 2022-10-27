package detect

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestFactoryDetect(t *testing.T) {
	hit := "通过"
	pass, err := FactoryDetect(rand.Reader)
	if err != nil {
		hit = err.Error()
	}
	fmt.Printf("15种算法 上电自检 50组 10^6 bit: %v, hit: %s\n", pass, hit)
}

func TestPowerOnDetect(t *testing.T) {
	hit := "通过"
	pass, err := PowerOnDetect(rand.Reader)
	if err != nil {
		hit = err.Error()
	}
	fmt.Printf("15种算法 上电自检 20组 10^6 bit: %v, hit: %s\n", pass, hit)
}

func TestPeriodDetect(t *testing.T) {
	hit := "通过"
	pass, err := PeriodDetect(rand.Reader)
	if err != nil {
		hit = err.Error()
	}
	fmt.Printf("12种算法 周期检测 20组 20000 bit: %v, hit: %s\n", pass, hit)
}

func TestSingleDetect(t *testing.T) {
	pass, err := SingleDetect(rand.Reader, 16)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("单次检测 320 bit:", pass)
	pass, err = SingleDetect(rand.Reader, 320/8)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("扑克检测 单次检测 10^6 bit:", pass)
}

/*
func TestPowerOnDetect2(t *testing.T) {
	var files []io.Reader
	dirname := "D:\\Project\\cliven\\randomness\\tools\\rdgen\\target\\data"
	dirs, err := ioutil.ReadDir(dirname)
	if err != nil {
		panic(err)
	}
	defer func() {
		for _, file := range files {
			_ = file.(io.Closer).Close()
		}
	}()
	for _, fi := range dirs {
		name := filepath.Join(dirname, fi.Name())
		file, err := os.OpenFile(name, os.O_RDONLY, os.FileMode(666))
		if err != nil {
			panic(err)
		}
		files = append(files, file)
	}
	mReader := io.MultiReader(files...)
	hit := "通过"

	pass, err := PowerOnDetect(mReader)
	if err != nil {
		hit = err.Error()
	}
	fmt.Printf("15种算法 上电自检 20组 10^6 bit: %v, hit: %s\n", pass, hit)
}
*/

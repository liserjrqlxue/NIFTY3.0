package main

import (
	"flag"
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	xlsx = flag.String(
		"xlsx",
		"",
		"input excel",
	)
	input = flag.String(
		"input",
		"",
		"input.list",
	)
	outDir = flag.String(
		"outDir",
		".",
		"output dir",
	)
	sheetName = flag.String(
		"sheet",
		"建库",
		"sheet name of input",
	)
	libID = flag.String(
		"lib",
		"",
		"文库号,raw data in -dataPath/-proj/-libID",
	)
	dataPath = flag.String(
		"dataPath",
		"/share/nastj6/solexa_A/fqdata19A/Zebra/MGISEQ-2000",
		"数据路径",
	)
	proj = flag.String(
		"proj",
		"P18Z15000N0443",
		"项目编号",
	)
)

type Sample struct {
	sampleID    string
	libID       string
	subLibID    string
	positiveMut string
	fq1, fq2    []string
}

// regexp
var (
	fq1 = regexp.MustCompile(`_1.f(ast)?q(.gz)?$`)
	fq2 = regexp.MustCompile(`_2.f(ast)?q(.gz)?$`)
)

func (sample *Sample) findPE(rawPath string) (success bool) {
	dirs, err := filepath.Glob(filepath.Join(rawPath, "*"+sample.subLibID))
	simple_util.CheckErr(err)
	var fq1Count, fq2Count uint
	for _, dir := range dirs {
		fqs, err := filepath.Glob(filepath.Join(dir, "*.fq.gz"))
		simple_util.CheckErr(err)
		for _, fq := range fqs {
			if fq1.MatchString(fq) {
				sample.fq1 = append(sample.fq1, fq)
				fq1Count++
			} else if fq2.MatchString(fq) {
				sample.fq2 = append(sample.fq2, fq)
				fq2Count++
			} else {
				log.Fatalf("can not parse fq[%s]", fq)
			}
		}
		if fq1Count != fq2Count {
			log.Printf("findPE of %s with error:[%v %v]", sample.sampleID, sample.fq1, sample.fq2)
		}
	}
	if fq1Count > 0 {
		success = true
	} else {
		log.Printf("findPE of %s with error:[%v %v]", sample.sampleID, sample.fq1, sample.fq2)
	}
	return
}

func (sample *Sample) fprint(w io.Writer) (err error) {
	_, err = fmt.Fprintln(w, strings.Join([]string{sample.sampleID, sample.libID, sample.subLibID, strings.Join(sample.fq1, ","), strings.Join(sample.fq2, ","), sample.positiveMut}, "\t"))
	simple_util.CheckErr(err)
	return
}

var inputListFH *os.File
var mapArray []map[string]string
var err error

func main() {
	flag.Parse()
	if *input == "" && *xlsx == "" {
		flag.Usage()
		log.Print("-input or -xlsx is required")
		os.Exit(1)
	}
	if *libID == "" {
		flag.Usage()
		log.Print("-libID is required")
		os.Exit(1)
	}

	simple_util.CheckErr(os.MkdirAll(*outDir, 0755))

	info, err := os.Create(filepath.Join(*outDir, "sample.info"))
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(info)
	_, err = fmt.Fprintln(info, strings.Join([]string{"SampleID", "LibID", "SubLibID", "Fq1", "Fq2", "PositiveMut"}, "\t"))
	simple_util.CheckErr(err)

	var inputList = filepath.Join(*outDir, "input.list")
	if simple_util.FileExists(inputList) {
		inputList, err = filepath.Abs(inputList)
		simple_util.CheckErr(err)
	}
	var writeInputList = true

	if *input != "" {
		*input, err = filepath.Abs(*input)
		simple_util.CheckErr(err)
		mapArray, _ = simple_util.File2MapArray(*input, "\t", nil)
		if inputList == *input {
			log.Printf("reuse %s/input.list (may be fixed)", *outDir)
			writeInputList = false
		}
	} else {
		mapArray = parseXlsx(*input)
	}

	if writeInputList {
		inputListFH, err = os.Create(inputList)
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(inputListFH)
		_, err = fmt.Fprintln(inputListFH, strings.Join([]string{"序号", "样本编号", "子文库号", "突变位点"}, "\t"))
		simple_util.CheckErr(err)
	}

	var db = make(map[string]*Sample)
	for _, item := range mapArray {
		if writeInputList {
			_, err = fmt.Fprintln(inputListFH, strings.Join([]string{item["序号"], item["样本编号"], item["子文库号"], item["突变位点"]}, "\t"))
			simple_util.CheckErr(err)
		}
		sampleID := item["样本编号"]
		db[sampleID] = &Sample{
			sampleID:    sampleID,
			libID:       *libID,
			subLibID:    item["子文库号"],
			positiveMut: item["突变位点"],
		}
		if db[sampleID].findPE(filepath.Join(*dataPath, *proj, *libID)) {
			simple_util.CheckErr(db[sampleID].fprint(info))
		}
	}
}

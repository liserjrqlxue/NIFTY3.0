package main

import (
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/liserjrqlxue/simple-util"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	input = flag.String(
		"input",
		"",
		"input excel or fixed input.list",
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
		"文库号",
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
	}
	if fq1Count == 1 && fq2Count == 1 {
		success = true
	} else {
		log.Printf("findPE of %s with error:[%s %s]", sample.sampleID, sample.fq1, sample.fq2)
	}
	return
}

func (sample *Sample) fprint(w io.Writer) (err error) {
	fmt.Fprintln(w, strings.Join([]string{sample.sampleID, sample.libID, sample.subLibID, sample.fq1[0], sample.fq2[0], sample.positiveMut}, "\t"))
	return
}

var inputListFH *os.File
var mapArray []map[string]string
var err error

func main() {
	flag.Parse()
	if *input == "" || *libID == "" {
		flag.Usage()
		log.Print("-input and -libID is required")
		os.Exit(1)
	}
	*input, err = filepath.Abs(*input)
	simple_util.CheckErr(err)

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
	var writeInputList = false
	if inputList == *input {
		log.Printf("use input.list (may be fixed)")
		mapArray, _ = simple_util.File2MapArray(inputList, "\t", nil)
	} else {
		writeInputList = true
		inputListFH, err = os.Create(inputList)
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(inputListFH)
		_, err = fmt.Fprintln(inputListFH, strings.Join([]string{"序号", "样本编号", "子文库号", "突变位点"}, "\t"))
		simple_util.CheckErr(err)

		xlsxFh, err := excelize.OpenFile(*input)
		simple_util.CheckErr(err)
		rows, err := xlsxFh.GetRows(*sheetName)
		var skip = true
		var title []string
		for _, row := range rows {
			if row[0] == "序号" {
				title = row
				skip = false
				continue
			}
			if skip {
				continue
			}
			item := make(map[string]string)
			for i, key := range title {
				item[key] = row[i]
			}
			mapArray = append(mapArray, item)
		}
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

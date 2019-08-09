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
		"input excel",
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

func (sample *Sample) findPE(rawPath string) {
	dirs, err := filepath.Glob(filepath.Join(rawPath, "*"+sample.subLibID))
	simple_util.CheckErr(err)
	for _, dir := range dirs {
		fqs, err := filepath.Glob(filepath.Join(dir, "*.fq.gz"))
		simple_util.CheckErr(err)
		for _, fq := range fqs {
			if fq1.MatchString(fq) {
				sample.fq1 = append(sample.fq1, fq)
			} else if fq2.MatchString(fq) {
				sample.fq2 = append(sample.fq2, fq)
			} else {
				log.Fatalf("can not parse fq[%s]", fq)
			}
		}
	}
}

func (sample *Sample) fprint(w io.Writer) (err error) {
	fmt.Fprintln(w, strings.Join([]string{sample.sampleID, sample.libID, sample.subLibID, sample.fq1[0], sample.fq2[0], sample.positiveMut}, "\t"))
	return
}

func main() {
	flag.Parse()
	if *input == "" || *libID == "" {
		flag.Usage()
		log.Print("-input and -libID is required")
		os.Exit(1)
	}

	var inputList = filepath.Join(*outDir, "input.list")
	iL, err := os.Create(inputList)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(iL)

	info, err := os.Create(filepath.Join(*outDir, "sample.info"))
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(info)
	_, err = fmt.Fprintln(info, strings.Join([]string{"SampleID", "LibID", "SubLibID", "Fq1", "Fq2", "PositiveMut"}, "\t"))
	simple_util.CheckErr(err)

	xlsxFh, err := excelize.OpenFile(*input)
	simple_util.CheckErr(err)
	rows, err := xlsxFh.GetRows(*sheetName)
	var skip = true
	var title []string
	var db = make(map[string]*Sample)
	for _, row := range rows {
		if row[0] == "序号" {
			title = row
			skip = false
			_, err = fmt.Fprintln(iL, strings.Join([]string{"序号", "样本编号", "子文库号", "突变位点"}, "\t"))
			simple_util.CheckErr(err)
			continue
		}
		if skip {
			continue
		}
		item := make(map[string]string)
		for i, key := range title {
			item[key] = row[i]
		}
		_, err = fmt.Fprintln(iL, strings.Join([]string{item["序号"], item["样本编号"], item["子文库号"], item["突变位点"]}, "\t"))
		simple_util.CheckErr(err)
		sampleID := item["样本编号"]
		db[sampleID] = &Sample{
			sampleID:    sampleID,
			libID:       *libID,
			subLibID:    item["子文库号"],
			positiveMut: item["突变位点"],
		}
		db[sampleID].findPE(filepath.Join(*dataPath, *proj, *libID))
		simple_util.CheckErr(db[sampleID].fprint(info))
	}
}

package main

import (
	"flag"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/liserjrqlxue/simple-util"
	"log"
	"os"
	"path/filepath"
)

var (
	input = flag.String(
		"input",
		"",
		"input excel",
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
	subLibID    string
	positiveMut string
	fq1, fq2    string
}

func (sample *Sample) findPE(rawPath string) {
	dirs, err := filepath.Glob(filepath.Join(rawPath, "*"+sample.subLibID))
	simple_util.CheckErr(err)
	log.Printf("%+v", dirs)
}

func main() {
	flag.Parse()
	if *input == "" || *libID == "" {
		flag.Usage()
		log.Print("-input and -libID is required")
		os.Exit(1)
	}

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
			continue
		}
		if skip {
			continue
		}
		item := make(map[string]string)
		for i, key := range title {
			item[key] = row[i]
		}
		sampleID := item["样本编号"]
		db[sampleID] = &Sample{
			sampleID:    sampleID,
			subLibID:    item["子文库号"],
			positiveMut: item["突变位点"],
		}
		db[sampleID].findPE(filepath.Join(*dataPath, *proj))
	}
}

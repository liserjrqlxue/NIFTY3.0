package main

import (
	"github.com/360EntSecGroup-Skylar/excelize"
	simple_util "github.com/liserjrqlxue/simple-util"
)

func parseXlsx(xlsx string) (mapArray []map[string]string) {
	xlsxFh, err := excelize.OpenFile(xlsx)
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
	return
}

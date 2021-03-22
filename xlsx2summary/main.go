package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/version"
)

type Info struct {
	sampleID   string
	基因检测结果     []string
	基因         []GeneInfo
	geneList   []string
	geneMap    map[string]GeneInfo
	样本快递日期     string
	新筛编号       string
	华大基因检测编号   string
	性别         string
	生化筛查结果     string
	生化筛查结果拟诊   string
	基因检测结果报告日期 string
	基因检测结果总结   string
}

type GeneInfo struct {
	基因名称 string
	变异   []MutInfo
	疾病   string
	患病风险 string
	遗传方式 string
}
type MutInfo struct {
	外显子   string
	碱基改变  string
	氨基酸改变 string
}

// flag
var (
	anno = flag.String(
		"anno",
		"",
		"anno excel, comma as sep",
	)
	input = flag.String(
		"input",
		"",
		"info dir",
	)
	prefix = flag.String(
		"prefix",
		"",
		"prefix of output, output -prefix.date.xlsx",
	)
)

func main() {
	version.LogVersion()
	flag.Parse()
	if *anno == "" || *input == "" {
		flag.Usage()
		fmt.Printf("-anno/-input/-prefix required!")
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = *input
	}

	var outExcel = simpleUtil.HandleError(excelize.OpenFile(*input)).(*excelize.File)

	var db = make(map[string]Info)
	// 读解读表
	for _, in := range strings.Split(*anno, ",") {
		var inExcel = simpleUtil.HandleError(excelize.OpenFile(in)).(*excelize.File)
		var strSlice = simpleUtil.HandleError(inExcel.GetRows("All variants data")).([][]string)

		var title []string
		for i, strArray := range strSlice {
			if i == 0 {
				title = strArray
				continue
			}
			var item = make(map[string]string)
			for j := range title {
				if j < len(strArray) {
					item[title[j]] = strArray[j]
				}
			}
			if item["报告类别"] != "正式报告" {
				continue
			}
			var sampleID = item["SampleID"]
			var info, ok1 = db[sampleID]
			if !ok1 {
				info = Info{
					sampleID: sampleID,
					geneMap:  make(map[string]GeneInfo),
				}
			}
			var geneSymbol = item["Gene Symbol"]
			var geneInfo, ok2 = info.geneMap[geneSymbol]
			if !ok2 {
				geneInfo = GeneInfo{
					基因名称: geneSymbol,
					变异:   []MutInfo{},
					疾病:   item["疾病中文名"],
					患病风险: item["遗传模式判读"],
					遗传方式: item["遗传模式"],
				}
				info.geneList = append(info.geneList, geneSymbol)
			}
			var mutInfo = MutInfo{
				外显子:   item["ExIn_ID"],
				碱基改变:  item["cHGVS"],
				氨基酸改变: item["pHGVS"],
			}
			geneInfo.变异 = append(geneInfo.变异, mutInfo)
			info.geneMap[geneSymbol] = geneInfo
			db[sampleID] = info
		}
	}

	// load sample info
	var sheetName = "159基因结果汇总"
	simpleUtil.CheckErr(
		outExcel.SetColWidth(
			sheetName,
			"BM",
			"BM",
			10,
		),
	)
	simpleUtil.CheckErr(
		outExcel.SetColStyle(
			sheetName,
			"BM",
			simpleUtil.HandleError(outExcel.NewStyle(`{"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1}, "number_format": 14}`)).(int),
		),
	)
	var strSlice = simpleUtil.HandleError(
		simpleUtil.HandleError(
			excelize.OpenFile(*input),
		).(*excelize.File).GetRows(sheetName),
	).([][]string)
	for rIdx, strArray := range strSlice {
		if rIdx < 4 {
			continue
		}
		rIdx++
		var sampleID = strArray[3]
		if len(strArray) > 64 && strArray[64] != "" {
			fmt.Printf("警告：样品[%s]已有时间戳，跳过\n", sampleID)
			continue
		}
		var item, ok = db[sampleID]
		if !ok {
			fmt.Printf("警告：样品[%s]无解读信息，跳过\n", sampleID)
			continue
		}
		for _, geneSymbol := range item.geneList {
			var geneInfo = item.geneMap[geneSymbol]
			item.基因 = append(item.基因, geneInfo)
			item.基因检测结果 = append(item.基因检测结果, geneInfo.疾病)
		}
		for i, v := range item.基因检测结果 {
			WriteCellStr(outExcel, sheetName, 10+i, rIdx, v)
		}
		for i, geneInfo := range item.基因 {
			if i > 4 {
				fmt.Printf("警告：样品[%s]检出基因超出5个，基因[%s]跳过\n", sampleID, geneInfo.基因名称)
				break
			}
			WriteCellStr(outExcel, sheetName, 15+i*10, rIdx, geneInfo.基因名称)
			WriteCellStr(outExcel, sheetName, 22+i*10, rIdx, geneInfo.疾病)
			WriteCellStr(outExcel, sheetName, 23+i*10, rIdx, geneInfo.患病风险)
			WriteCellStr(outExcel, sheetName, 24+i*10, rIdx, geneInfo.遗传方式)
			for j, mutInfo := range geneInfo.变异 {
				if j > 1 {
					fmt.Printf("警告：样品[%s]在基因[%s]检出变异超出2个，变异[%s]跳过", sampleID, geneInfo.基因名称, mutInfo.碱基改变)
					break
				}
				WriteCellStr(outExcel, sheetName, 16+i*10+j*3, rIdx, mutInfo.外显子)
				WriteCellStr(outExcel, sheetName, 17+i*10+j*3, rIdx, mutInfo.碱基改变)
				WriteCellStr(outExcel, sheetName, 18+i*10+j*3, rIdx, mutInfo.氨基酸改变)
			}
		}
		WriteCellValue(outExcel, sheetName, 65, rIdx, time.Now().UTC())
	}

	simpleUtil.CheckErr(outExcel.SaveAs(fmt.Sprintf("%s.%s.xlsx", *prefix, time.Now().Format("2006-01-02"))))
}

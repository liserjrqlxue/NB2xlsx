package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/version"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	dbPath       = filepath.Join(exPath, "..", "db")
	etcPath      = filepath.Join(exPath, "..", "etc")
	templatePath = filepath.Join(exPath, "..", "template")
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
	input = flag.String(
		"input",
		"",
		"input excel",
	)
	infoDir = flag.String(
		"infoDir",
		"",
		"info dir",
	)
	output = flag.String(
		"output",
		"",
		"output excel, default is -input.结果汇总.xlsx")
	template = flag.String(
		"template",
		filepath.Join(templatePath, "结果汇总.xlsx"),
		"template of 结果汇总",
	)
)

func main() {
	version.LogVersion()
	flag.Parse()
	if *input == "" || *infoDir == "" {
		flag.Usage()
		fmt.Printf("-input/-infoDir required!")
		os.Exit(1)
	}
	if *output == "" {
		*output = *input + ".结果汇总.xlsx"
	}

	var inExcel = simpleUtil.HandleError(excelize.OpenFile(*input)).(*excelize.File)
	var outExcel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)
	var strSlice = simpleUtil.HandleError(inExcel.GetRows("All variants data")).([][]string)
	var sampleList []string
	var db = make(map[string]Info)
	var title []string
	for i, strArray := range strSlice {
		if i == 0 {
			title = strArray
			continue
		}
		var item = make(map[string]string)
		for j := range title {
			item[title[j]] = strArray[j]
		}
		var sampleID = item["SampleID"]
		var info, ok1 = db[sampleID]
		if !ok1 {
			info = Info{
				sampleID: sampleID,
				geneMap:  make(map[string]GeneInfo),
			}
			sampleList = append(sampleList, sampleID)
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

	// load sample info
	var fileInfos = simpleUtil.HandleError(ioutil.ReadDir(*infoDir)).([]os.FileInfo)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime().After(fileInfos[j].ModTime())

	})
	var sampleCount = len(sampleList)
	var count = 0
	for _, fileInfo := range fileInfos {
		var strSlice = simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(
					filepath.Join(*infoDir, fileInfo.Name()),
				),
			).(*excelize.File).GetRows("159基因结果汇总"),
		).([][]string)
		for i, strArray := range strSlice {
			if i < 4 {
				continue
			}
			var sampleID = strArray[3]
			var info, ok = db[sampleID]
			if !ok {
				continue
			}
			if info.华大基因检测编号 == sampleID {
				continue
			}
			count++
			info.样本快递日期 = strArray[1]
			info.新筛编号 = strArray[2]
			info.华大基因检测编号 = strArray[3]
			info.性别 = strArray[4]
			info.生化筛查结果 = strArray[5]
			info.生化筛查结果拟诊 = strArray[6]
			info.基因检测结果报告日期 = strArray[7]
			info.基因检测结果总结 = strArray[8]
			if count >= sampleCount {
				break
			}
		}
		if count >= sampleCount {
			break
		}
	}

	// write
	var rIdx = 5
	var sheetName = "159基因结果汇总"
	for n, sampleID := range sampleList {
		var item = db[sampleID]
		for _, geneSymbol := range item.geneList {
			var geneInfo = item.geneMap[geneSymbol]
			item.基因 = append(item.基因, geneInfo)
			item.基因检测结果 = append(item.基因检测结果, geneInfo.疾病)
		}
		WriteCellInt(outExcel, sheetName, 1, rIdx, n+1)
		WriteCellStr(outExcel, sheetName, 2, rIdx, item.样本快递日期)
		WriteCellStr(outExcel, sheetName, 3, rIdx, item.新筛编号)
		WriteCellStr(outExcel, sheetName, 4, rIdx, item.华大基因检测编号)
		WriteCellStr(outExcel, sheetName, 5, rIdx, item.性别)
		WriteCellStr(outExcel, sheetName, 6, rIdx, item.生化筛查结果)
		WriteCellStr(outExcel, sheetName, 7, rIdx, item.生化筛查结果拟诊)
		WriteCellStr(outExcel, sheetName, 8, rIdx, item.基因检测结果报告日期)
		WriteCellStr(outExcel, sheetName, 9, rIdx, item.基因检测结果总结)
		for i, v := range item.基因检测结果 {
			WriteCellStr(outExcel, sheetName, 10+i, rIdx, v)
		}
		// 15
		for i, geneInfo := range item.基因 {
			if i > 4 {
				break
			}
			WriteCellStr(outExcel, sheetName, 15+i*10, rIdx, geneInfo.基因名称)
			WriteCellStr(outExcel, sheetName, 22+i*10, rIdx, geneInfo.疾病)
			WriteCellStr(outExcel, sheetName, 23+i*10, rIdx, geneInfo.患病风险)
			WriteCellStr(outExcel, sheetName, 24+i*10, rIdx, geneInfo.遗传方式)
			for j, mutInfo := range geneInfo.变异 {
				if j > 1 {
					break
				}
				WriteCellStr(outExcel, sheetName, 16+i*10+j*3, rIdx, mutInfo.外显子)
				WriteCellStr(outExcel, sheetName, 17+i*10+j*3, rIdx, mutInfo.碱基改变)
				WriteCellStr(outExcel, sheetName, 18+i*10+j*3, rIdx, mutInfo.氨基酸改变)
			}
		}
		rIdx++
	}
	simpleUtil.CheckErr(outExcel.SaveAs(*output))
}

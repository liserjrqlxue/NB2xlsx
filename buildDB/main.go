package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	AES "github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
	"github.com/xuri/excelize/v2"
)

// os
var (
	ex, _  = os.Executable()
	exPath = filepath.Dir(ex)
)

// flag
var (
	input = flag.String(
		"input",
		"",
		"input db excel",
	)
	output = flag.String(
		"output",
		"",
		"encode db output, default is -input.json.aes",
	)
	sheetName = flag.String(
		"sheet",
		"Sheet1",
		"input db excel sheet name",
	)
	keys = flag.String(
		"keys",
		filepath.Join(exPath, "key.list"),
		"key list to build",
	)
	codeKey = flag.String(
		"code",
		"c3d112d6a47a0a04aad2b9d2d2cad266",
		"code key for aes",
	)
	extract = flag.String(
		"extract",
		"",
		"extract tsv, column names split by comma",
	)
)

func init() {
	version.LogVersion()
	flag.Parse()
	if *input == "" {
		flag.Usage()
		log.Fatalln("-input is required!")
	}
	if *output == "" {
		*output = *input + ".json.aes"
	}
}

func main() {
	var keyList = textUtil.File2Array(*keys)

	var extractFile *os.File
	var extractCols []string
	if *extract != "" {
		extractFile = osUtil.Create(*output + ".mut.tsv")
		extractCols = strings.Split(*extract, ",")
		fmtUtil.FprintStringArray(extractFile, extractCols, "\t")
	}
	defer func() {
		if *extract != "" {
			simpleUtil.DeferClose(extractFile)
		}
	}()

	var db, _ = simpleUtil.Slice2MapMapArray(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(*input),
			).(*excelize.File).
				GetRows(*sheetName),
		).([][]string),
		"Transcript", "cHGVS",
	)

	var liteDb = make(map[string]map[string]string)
	for mainKey, item := range db {
		var liteItem = make(map[string]string)
		for _, key := range keyList {
			liteItem[key] = item[key]
		}
		liteDb[mainKey] = liteItem

		if *extract != "" {
			var (
				strArray []string
				line     string
			)
			for _, col := range extractCols {
				strArray = append(strArray, item[col])
			}
			line = strings.Join(strArray, "\t")
			line = strings.ReplaceAll(line, "\n", "<br/>")
			fmtUtil.Fprintln(extractFile, line)
		}
	}

	var d = simpleUtil.HandleError(json.MarshalIndent(liteDb, "", "  ")).([]byte)
	var codeKeyBytes = []byte(*codeKey)
	AES.Encode2File(*output, d, codeKeyBytes)
}

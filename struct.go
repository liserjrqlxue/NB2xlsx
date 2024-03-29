package main

import (
	"github.com/liserjrqlxue/goUtil/stringsUtil"
	"regexp"
	"strings"
)

type Mode int

const (
	NBSP Mode = iota
	NBSIM
	WGSNB
	WGSCS
)

var ModeSet = []string{"NBSP", "NBSIM", "WGSNB", "WGSCS"}

var ModeMap = func() map[string]Mode {
	var db = make(map[string]Mode)
	for i, str := range ModeSet {
		db[str] = Mode(i)
	}
	return db
}()

func (m Mode) String() string {
	return ModeSet[m]
}

// GeneInfo : struct info of gene
type GeneInfo struct {
	gene                    string
	inheritance             string
	gender                  string
	PLP, hetPLP, VUS, HpVUS int
	cnv, cnv0               bool
	tag3                    string
	tag4                    bool
}

func (info *GeneInfo) new(item map[string]string) *GeneInfo {
	info.gene = item["Gene Symbol"]
	info.inheritance = geneInheritance[info.gene]
	return info
}

func (info *GeneInfo) count(item map[string]string) {

	if item["自动化判断"] == "VUS" {
		if item["Zygosity"] == "Het" && compositeP(item) {
			info.HpVUS++
		}
	}
	if isPLP(item) {
		item["P/LP*"] = "1"
		info.PLP++
	} else if isVUS(item) {
		item["VUS*"] = "1"
		info.VUS++
	}
}

func (info *GeneInfo) isAD() bool {
	switch info.inheritance {
	case "AD", "AD,AR", "AD,SMu", "Mi", "XLD", "XL":
		return true
	case "XLR":
		if info.gender == "M" {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func (info *GeneInfo) isAR() bool {
	switch info.inheritance {
	case "AR":
		return true
	case "XLR":
		if info.gender == "F" {
			return true
		}
	}
	return false

}

func (info *GeneInfo) lowADAF(item map[string]string) bool {
	if info.inheritance == "AD" || info.inheritance == "AD,AR" || info.inheritance == "AD,SMu" {
		for af := range af0List {
			if gt(item[af], afThreshold) {
				return false
			}
		}
	}
	return true
}

func (info *GeneInfo) getTag(item map[string]string) (tag string) {
	var tags []string
	var tag1 = 标签1(item, info)
	if tag1 != "" {
		tags = append(tags, tag1)
	}
	var tag2 = 标签2(item, info)
	if tag2 != "" {
		tags = append(tags, tag2)
	}
	var tag3 = 标签3(item, info)
	if tag3 != "" {
		tags = append(tags, tag3)
	}
	var tag5 = 标签5(item)
	if tag5 != "" {
		tags = append(tags, tag5)
	}
	var tag6 = 标签6(item)
	if tag6 != "" {
		tags = append(tags, tag6)
	}
	var tag7 = 标签7(item, info)
	if tag7 != "" {
		tags = append(tags, tag7)
	}
	return strings.Join(tags, ";")
}

func (info *GeneInfo) Tag4() {
	if info.cnv && info.isAD() {
		info.tag4 = true
	} else if info.cnv0 && info.isAR() {
		info.tag4 = true
	}
}

type Region struct {
	chr   string
	start int
	end   int
	gene  string
}

var (
	reg = regexp.MustCompile(`chr(.):(\d+)-(\d+)`)
)

func newRegion(region string) *Region {
	var match = reg.FindStringSubmatch(region)
	if len(match) == 4 {
		return &Region{
			chr:   match[1],
			start: stringsUtil.Atoi(match[2]),
			end:   stringsUtil.Atoi(match[3]),
			gene:  "",
		}
	}
	return nil
}

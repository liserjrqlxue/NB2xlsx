package main

type SampleInfo struct {
	sampleID string
	gender   string
	p0       string
	p1       string
	p2       string
	p3       string
}

func newSampleInfo(item map[string]string) SampleInfo {
	return SampleInfo{
		sampleID: item["sampleID"],
		gender:   item["gender_analysed"],
		p0:       item["P0"],
		p1:       item["P1"],
		p2:       item["P2"],
		p3:       item["P3"],
	}
}

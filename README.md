# NB2xlsx

[![GoDoc](https://godoc.org/github.com/liserjrqlxue/NB2xlsx?status.svg)](https://pkg.go.dev/github.com/liserjrqlxue/NB2xlsx)
[![Go Report Card](https://goreportcard.com/badge/github.com/liserjrqlxue/NB2xlsx)](https://goreportcard.com/report/github.com/liserjrqlxue/NB2xlsx)

安馨可生信注释格式

- [x] [编译安装](#编译安装)
- [ ] [All variants data](#all-variants-data)
  - [x] [过滤](#过滤)
  - [X] [标签](#标签)
    - [X] [定义](#定义)
  - [x] [疾病数据库](#疾病数据库)
  - [x] [已解读数据库](#已解读数据库)
  - [ ] [Other Columns](#other-columns)
    - [x] [LOF](#lof)
    - [x] [HGMDorClinvar](#hgmdorclinvar)
    - [x] [遗传模式判读](#遗传模式判读)
- [x] [lims.info](#limsinfo)
- [ ] [QC](#qc)
  - [x] [common](#common)
  - [ ] [others](#others)
- [x] [SMA_result](#sma_result)
- [x] [excel 格式](#excel-格式)
  - [x] [DataValidation](#datavalidation)
  - [x] [Background Color](#background-color)
- [ ] [modules](#modules)
  - [x] [anno](#anno2xlsxv2annno)
  - [x] [ACMG](#acmg2015)
    - [x] [init](#init)
    - [x] [use](#use)

## 使用方式

### 输入参数

| 参数             | 格式      | 默认值     | 说明                                                                                                          |
|----------------|---------|---------|-------------------------------------------------------------------------------------------------------------|
| -batch         | string  |         | batch name, use for hyperlink [batch].result_batCNV-dipin/chr11_chr16_chrX_cnemap/[sampleID]_W30S25_cne.jpg |
| -prefix        | string  | [batch] | output to [prefix].xlsx and [prefix].*                                                                      |
| -gender        | string  | F       | gender for all or gender map file                                                                           |
| -acmg          | boolean | false   | 是否启动ACMG2015注释                                                                                              |
| -autoPVS1      | boolean | false   | 是否使用autoPVS1的证据项结果                                                                                          |
| -cs            | boolean | false   | if use for CS                                                                                               |
| -im            | boolean | false   | if use for im                                                                                               |
| -wgs           | boolean | false   | if use for wgs                                                                                              |
| -all           | boolean | false   | 是否输出单样品所有变异excel                                                                                            |
| -batchCNV      | string  |         | batchCNV result                                                                                             |
| -annoDir       | string  |         | CS模式单样品excel输出目录，输出结果为[annoDir]/[sampleID]_vcfanno.xlsx                                                     |
| -detail        | string  |         | sample info                                                                                                 |
| -info          | string  |         | im info.txt                                                                                                 |
| -lims          | string  |         | lims.info                                                                                                   |
| -avd           | string  |         | All variants data file list, comma as sep                                                                   |
| -avdList       | string  |         | All variants data file list, one path per line                                                              |
| -dmd           | string  |         | DMD result file list, comma as sep                                                                          |
| -dmdList       | string  |         | DMD result file list, one path per line                                                                     |
| -lumpy         | string  |         | DMD-lumpy data                                                                                              |
| -nator         | string  |         | DMD-nator data                                                                                              |
| -dipin         | string  |         | dipin result file                                                                                           |
| -sma           | string  |         | sma result file                                                                                             |
| -sma2          | string  |         | sma result file                                                                                             |
| -feature       | string  |         | 个特 list                                                                                                     |
| -geneID        | string  |         | 基因ID list                                                                                                   |
| -qc            | string  |         | qc excel                                                                                                    |
| -bamPath       | string  |         | bamList file                                                                                                |
| -drug          | string  |         | drug result file                                                                                            |
| -drugSheetName | string  |         | 药物检测结果                                                                                                      |
| -threshold     | int     | 12      | threshold limit                                                                                             |

## goroutine

- main
  - useBatchCnv
    - go goWriteBatchCnv
      - holdChan(saveBatchCnvChan)
  - go createExcel
    - fillExcel
      - writeAvd2Sheet
        - go writeAvd
          - range dbChan
          - holdChan(runWrite)
        - go loadAvd
          - holdChan(throttle)
            - dbChan <- filterData
          - waitChan(throttle)
        - waitChan(runWrite)
        - holdChan(throttle)
    - holdChan(saveMainChan)
  - waitChan(saveMainChan, saveBatchCnvChan)

## 编译安装

```shell
git clone https://github.com/liserjrqlxue/NB2xlsx.git
cd NB2xlsx
go build -ldflags "-X 'main.codeKey=c3d112d6a47a0a04aad2b9d2d2cad266'" # 需要替换对应AES密钥
```

### 另一种方式

1. 修改 `generate.go` 内 `c3d112d6a47a0a04aad2b9d2d2cad266` 为实际使用的AES密钥
2. 安装 `vb` 或者将 `generate.go` 内 `vb -ldflags "-w -s"` 改成 `go build`
   1. `go install github.com/liserjrqlxue/version/vb@latest`
3. 运行 `go generate`

### 注意

部分数据库文件不在`git repo`内，需要拷贝到对应位置

## etc

### gene.exclude.list.txt

```text
1. 新生儿升级流程中以下10个基因不给“补充报告“的标签
PPM1K、GCSH、PRODH、BCAT1、BCAT2、HAL、CD320、ACAA1、ACAA2、LDLR
2. 新生儿流程中的标签
2.1  PPM1K、GCSH、PRODH、BCAT1、BCAT2、HAL、CD320、ACAA1、ACAA2、LDLR这10个基因上的变异不给标签
```

列表内基因：

1. "报告类别-原始"=="补充报告"时"报告类别-原始"置空
2. "Database"置空

## All variants data

### 过滤

```text
三 输出到解读表位点调整
满足以下任一一个条件就输出到sheet1：
1. 161基因上Clinvar的标签是Pathogenic或者Likely_pathogenic或者Pathogenic/Likely_pathogenic或者HGMD的标签是DM或者DM？或者DM/DM? 并且ESP6500/1000G/ExAC和GnomeAD总人频以及ExAC的East Asian、GnomAD的East Asian频率≤0.05，去除自动化致病等级是B，LB的变异（保留clinvar是P，LP，P/LP的位点）。
2. 161基因上Clinvar/HGMD数据库外ESP6500/1000G/ExAC和GnomeAD总人频以及ExAC的East Asian、GnomAD的East Asian的频率≤0.01，并且变异包括missense/nonsense/frameshift/cds-ins/cds-del/coding-synon/init-loss/ncRNA/splice-3/splice-5/20位以内的intron/SpliceAI预测影响剪切的ESP6500/1000G/ExAC和GnomeAD总人频以及ExAC的East Asian、GnomAD的East Asian的频率≤0.01的intron
3. 已解读数据库内位点
```

```text
第三条的具体条件如下：

（1）保留已解读数据库内位点
（2）过滤 etc/gene.list.txt 之外的基因
（3）过滤 "ESP6500 AF/1000G AF/ExAC AF/GnomeAD AF/ExAC EAS AF、GnomAD EAS AF" > 0.05
（4）过滤自动化致病等级是B，LB的变异（不包括clinvar是P，LP，P/LP的位点）。
（5）"ClinVar Significance" 是 Pathogenic 或者 Likely_pathogenic 或者 Pathogenic/Likely_pathogenic 的保留
（6）"HGMD Pred" 是 DM 或者 DM? 或者 DM/DM? 的保留
（7）"Function"列为 missense/nonsense/frameshift/cds-ins/cds-del/cds-indel/stop-loss/span/altstart/coding-synon/init-loss/ncRNA/splice-3/splice-5/splice+10/splice-10/splice+20/splice-20，且 "ESP6500 AF/1000G AF/ExAC AF/GnomeAD AF/ExAC EAS AF、GnomAD EAS AF" <= 0.01 的保留
（8）"Function"列为intron且SpliceAI预测影响剪切且"ESP6500 AF/1000G AF/ExAC AF/GnomeAD AF/ExAC EAS AF、GnomAD EAS AF" <= 0.01 的保留
```

1. **保留** 已解读数据库内位点
2. **过滤** "Gene Symbol" no in `etc/gene.list.txt`
3. **定义** 频率列表 `avdAfList` ["ESP6500 AF","1000G AF","ExAC AF","GnomAD AF","ExAC EAS AF","GnomAD EAS AF",]
4. **过滤** `avdAfList` > 0.05
5. **保留** "ClinVar Significance" in ["Pathogenic","Likely_pathogenic","Pathogenic/Likely_pathogenic"]
6. **过滤** "自动化判断" in ["B","LB"]
7. **保留** "HGMD Pred" in ["DM","DM?","DM/DM?"]
8. **过滤** `avdAfList` > 0.01
9. **保留** "Function" = "intron" 且 "SpliceAI Pred" = "D"
10. **过滤** "Function" in `etc/function.exclude.txt`
11. **保留** 剩余

### 标签

#### 定义

- 遗传模式

  - 来源：新生儿疾病库（-disease）的"遗传模式"  
- cnv

    拷贝数异常  
    使用batchCNV和DMD的分析流程，任一有检出≥2个连续exon

  - BatCNV的copyNumber或DMD CNV有~~≥2个连续~~exon的CopyNum列不为2拷贝
- cnv0

    使用batchCNV和DMD的分析流程，任一有检出0

  - BatCNV的copyNumber或DMD CNV有~~≥2个连续~~exon的CopyNum列为0拷贝
- P/LP*

    可能有害

  - Definition为P/LP
  - 烈性
    - nonsense/frameshift/stop-gain/span/altstart/init-loss/splice-3/splice-5
  - ClinVar收录P/LP
  - HGMD收录P/LP
    - ClinVar致病等级不为B/LB
- P/LP2*
  - Definition为P/LP
  - 烈性
    - nonsense/frameshift/stop-gain/span/altstart/init-loss/splice-3/splice-5
  - ClinVar收录P/LP
  - HGMD收录P/LP
    - ClinVar致病等级不为B/LB/**Conflicting interpretations of pathogenicity/VUS**
- VUS*

    可能意义未明
  
  - 不是P/LP*
    - ClinVar致病等级不为B/LB
      - VUS/P/LP（自动化判断）或者VUS（Definition）变异

- AD类遗传模式
  - AD
  - AD,AR
  - AD,SMu
  - Mi
  - XLD
  - XL
  - (XLR且男性)
- AD低频

    AF列表:"ESP6500 AF","1000G AF","ExAC AF","ExAC EAS AF","GnomAD AF","GnomAD EAS AF"

  - AD或AD,AR或AD,SMu
    - AF <1e-4 或 .
  - 其它遗传模式
- AR类遗传模式
  - AR
  - AR;AR
  - (XLR且女性)
- CDS*
  - cds-del/cds-ins/cds-indel/stop-loss变异
    - RepeatTag无标签
- Splice*
  - splice+10/splice-10/splice+20/splice-20/intron变异  

       dbscSNV_RF_SCORE（≥0.6为有影响）、dbscSNV_ADA_SCOR（≥0.6为有影响）、spliceAI（≥0.2为有影响）

    - 有害性预测至少2个软件有预测结果，均预测有害，其他无结果，
  - 仅spliceAI有预测结果（且结果为有害）
- SpliceCS*
  - splice+10/splice-10/splice+20/splice-20/intron/coding-synon变异

       dbscSNV_RF_SCORE（≥0.6为有影响）、dbscSNV_ADA_SCOR（≥0.6为有影响）、spliceAI（≥0.2为有影响）

    - 有害性预测至少2个软件有预测结果，均预测有害，其他无结果，
    - 仅spliceAI有预测结果（且结果为有害）
- NoSplice*
  - 除splice+10/splice-10/splice+20/splice-20/intron以外的变异
  - SIFT、Condel、MutationTaster、Polyphen2HVAR有害性预测至少2个软件有预测结果，均预测有害，其他无结果
- NoSpliceCS*
  - 除splice+10/splice-10/splice+20/splice-20/intron/coding-synon以外的变异
    - SIFT、Condel、MutationTaster、Polyphen2HVAR有害性预测至少2个软件有预测结果，均预测有害，其他无结果
- compositeP
  - Splice*
  - NoSplice*
  - CDS*
- compositePCS
  - SpliceCS*
  - NoSpliceCS*
  - CDS*

#### 规则

|遗传模式|P/LP*|compositeP|Zygosity|Function|自动化判断|Definition|ClinVar| HGMD |lowAF|VarCount| CNV |标签|CNV标签|
|-------|-----|----------|--------|--------|---------|----------|-------|------|-----|---------|----|----|-------|
|     AD|true |          |        |        |         |          |       |      |true |         |    |   1|       |
|     AR|true |          |Hom     |        |         |          |       |      |     |         |    |   1|       |
|     AR|true |          |Het     |        |         |          |       |      |     |LPL>1    |    |   1|       |
|     AR|true |          |Het     |        |         |          |       |      |     |PLPVUS>1 |    |   1|       |
|     AR|true |          |Het     |        |!PLPVUS  |          |       |      |     |PLP==1   |    |   1|       |
|     AR|false|          |        |        |VUS      |          |       |      |     |hetPLP==1|    |   1|       |
|     AD|false|true      |        |        |         |PLPVUS    |       |      |true |         |    |   2|       |
|     AR|     |true      |Hom     |        |VUS      |          |       |      |     |         |    |   2|       |
|     AR|     |true      |        |        |VUS      |          |       |      |     |HpVUS>1  |    |   2|       |
|     AR|true |          |        |        |         |          |       |      |     |VUS==0   |cnv |   3|3      |
|     AR|     |true      |        |        |VUS      |          |       |      |     |         |cnv |   3|3      |
|     AD|     |          |        |        |         |          |       |      |     |         |cnv |    |4      |
|     AR|     |          |        |        |         |          |       |      |     |         |cnv0|    |4      |
|       |     |          |        |        |         |P/LP      |       |      |     |         |    |   5|       |
|       |     |          |        |LoF     |         |          |       |      |     |         |    |   5|       |
|       |     |          |        |        |         |          |P/LP   |      |     |         |    |   5|       |
|       |     |          |        |        |         |          |!B/LB  |DM/DM?|     |         |    |   5|       |

### 疾病数据库

```text
使用'$$'分隔疾病信息
```

| key1        | key2   | note     |
|-------------|--------|----------|
| Gene Symbol | 基因     | main key |
| 疾病中文名       | 疾病     |          |
| 遗传模式        | 遗传模式   |          |
| 疾病简介        | 疾病简介   |          |
| 包装疾病分类      | 包装疾病分类 |          |
| 报告逻辑        | 报告逻辑   |          |

### 已解读数据库

```text
第三附件1中的BS列Database，在已解读数据库内并且已解读数据库的DU列是否是包装位点记录为“是”：标记NBS-in
                         在已解读数据库内并且已解读数据库的DU列是否是包装位点无记录：标记NBS-out
                         不在已解读数据库：标记.
第五附件1中的BW列报告类别，同孕前，数据库内包装的变异（BS列Database标记NBS-in）标记正式报告；数据库外的烈性（LOF列为YES）且低频(GnomAD≤1%，且千人≤1%)：标记补充报告
```

```text
对于已解读数据库的“是否是包装位点”列为“是”以外的烈性低频（条件未改变）标记补充报告
```

```text
sheet1里面的CC列“参考文献”，提取的是已解读数据库中的DM列“Reference”的内容
```

```text
2. 模板excel的CE列“位点关联疾病”匹配位点数据库CU列“Disease“
3. 模板excel的CF列“位点关联遗传模式“匹配位点数据库CV列” 遗传模式“
4. 模板excel的CG列“Evidence New + Check“匹配位点数据库DU列” 证据项“
5. 模板excel的CH列“Definition“匹配位点数据库CX列” Definition“
6. 位点数据库中的DV列“是否是包装位点“为“是”的在注释表中是正式报告
```

```text
“正式报告”补充报告“是输出在”报告类别“列，现在需要输出到”报告类别-原始“这一列
```

| key1                 | key2       | note               |
|----------------------|------------|--------------------|
| Transcript           | Transcript | main key 1         |
| cHGVS                | cHGVS      | main key 2         |
| 参考文献                 | Reference  ||
| 位点关联疾病               | Disease    ||
| 位点关联遗传模式             | 遗传模式       ||
| Evidence New + Check | 证据项        ||
| Definition           | Definition ||
| Database             |            | [NBS-in,NBS-out,.] |
| 报告类别-原始              |            | [正式报告,补充报告]        |
|                      | 是否是包装位点    ||

### Other Columns

| key1           | key2                         | note |
|----------------|------------------------------|------|
| ClinVar星级      | ClinVar Number of gold stars |      |
| 1000Gp3 AF     | 1000G AF                     |      |
| 1000Gp3 EAS AF | 1000G EAS AF                 |      |
| 引物设计           | anno.PrimerDesign(item)      |      |

#### LOF

```text
第二附件1中的BL列LOF同孕前：nonsense、frameshift、splice-3、splice-5类型且低频(GnomAD≤1%，且千人≤1%)，标记YES，否则标记NO。
```

[`updateLOF`](../367051a760349aac7a4b236ca081340d086c10bd/main.go#L361)

| key | value        |
|-----|--------------|
| LOF | ['YES','NO'] |

#### HGMDorClinvar

```go
item["HGMDorClinvar"] = "否"
if isHGMD[item["HGMD Pred"]] || isClinVar[item["ClinVar Significance"]] {
    item["HGMDorClinvar"] = "是"
}
```

#### 遗传模式判读

| 遗传模式                            | 杂合性                  | 个数   | 样品性别              | 遗传模式判读 |
|---------------------------------|----------------------|------|-------------------|--------|
| AR                              | ['Hom']              |      |                   | 可能患病   |
| AR                              | ['Het']              | =1   |                   | 携带者    |
| AR                              | ['Het']              | \>1  |                   | 可能患病   |
| 包含 AD                           | ['Hom','Het']        |      |                   | 可能患病   |
| Mi                              | ['Hom','Het']        |      |                   | 可能患病   |
| 包含 XLD                          | ['Hom','Het','Hemi'] |      |                   | 可能患病   |
| XLR                             | ['Hom','Het','Hemi'] |      | Male              | 可能患病   |
| XLR                             | ['Hom']              |      | Female            | 可能患病   |
| XLR                             | ['Het']              | =1   | Female            | 携带者    |
| XLR                             | ['Het']              | \>1  | Female            | 可能患病   |
| XL 仅当基因属于['OTC','GLA','PCDH19'] | ['Hemi','Hom',Het']  | \>=1 | ['Female','Male'] | 可能患病   |

## 补充实验

### 地贫标准写法

| 目前流程      | 更新后标准写法       | HTML                                |
|-----------|---------------|-------------------------------------|
| 3.7       | αα/-α3.7      | αα/-α<sup>3.7</sup>                 |
| 4.2       | αα/-α4.2      | αα/-α<sup>4.2</sup>                 |
| SEA       | αα/--SEA      | αα/--<sup>SEA</sup>                 |
| FIL       | αα/--FIL      | αα/--<sup>FIL</sup>                 |
| THAI      | αα/--THAI     | αα/--<sup>THAI</sup>                |
| 3.7;3.7   | -α3.7/-α3.7   | -α<sup>3.7</sup>/-α<sup>3.7</sup>   |
| 4.2;4.2   | -α4.2/-α4.2   | -α<sup>4.2</sup>/-α<sup>4.2</sup>   |
| 3.7;4.2   | -α3.7/-α4.2   | -α<sup>3.7</sup>/-α<sup>4.2</sup>   |
| 3.7;SEA   | -α3.7/--SEA   | -α<sup>3.7</sup>/--<sup>SEA</sup>   |
| 3.7;FIL   | -α3.7/--FIL   | -α<sup>3.7</sup>/--<sup>FIL</sup>   |
| 3.7;THAI  | -α3.7/--THAI  | -α<sup>3.7</sup>/--<sup>THAI</sup>  |
| 4.2;SEA   | -α4.2/--SEA   | -α<sup>4.2</sup>/--<sup>SEA</sup>   |
| 4.2;FIL   | -α4.2/--FIL   | -α<sup>4.2</sup>/--<sup>FIL</sup>   |
| 4.2;THAI  | -α4.2/--THAI  | -α<sup>4.2</sup>/--<sup>THAI</sup>  |
| SEA;FIL   | --SEA/--FIL   | --<sup>SEA</sup>/--<sup>FIL</sup>   |
| FIL;THAI  | --FIL/--THAI  | --<sup>FIL</sup>/--<sup>THAI</sup>  |
| SEA;THAI  | --SEA/--THAI  | --<sup>SEA</sup>/--<sup>THAI</sup>  |
| SEA;SEA   | --SEA/--SEA   | --<sup>SEA</sup>/--<sup>SEA</sup>   |
| FIL;FIL   | --FIL/--FIL   | --<sup>FIL</sup>/--<sup>FIL</sup>   |
| THAI;THAI | --THAI/--THAI | --<sup>THAI</sup>/--<sup>THAI</sup> |

## lims.info

| key1      | key2               | note     |
|-----------|--------------------|----------|
| SampleID  | MAIN_SAMPLE_NUM    | main key |
| 期数        | HYBRID_LIBRARY_NUM |          |
| flow ID   | FLOW_ID            |          |
| 产品编码      | PRODUCT_CODE       |          |
| 产品名称      | PRODUCT_NAME       |          |
| 产品编码_产品名称 || 产品编码+'_'+产品名称      |     |

## QC

```text
1. Order为序号：1，2，3，。。。
2. Sample列对应下机QC的 “Sample”列
3. Q20列对应下机QC的“Q20”
4. Q30列对应下机QC的”Q30”
5. AverageDepth对应下机QC的“Target Mean Depth[RM DUP]:”列
6. Depth>=10(%)对应下机QC的“Target coverage >=10X percentage[RM DUP]:”列
7. Coverage(%)对应下机QC的“Target coverage[RM DUP]:“列
8. GC(%)对应下机QC的“Total mapped GC Rate:“
9. Target coverage >=20X percentage对应下机QC的“Target coverage >=20X percentage[RM DUP]:“列
10. mitochondria Target Mean Depth[RM DUP]对应下机QC的“Target Mean Depth[RM DUP]: (mitochondria)“列
11. Gender为性别
12. RESULT为C-K列的一个判断，比如提示性别不一致，GC含量高，如果都合格就输出“YES“
13. 解读人和审核人为空
14. 产品编号对应临床新生儿的产品编号DX1968，多中心是DX1964
```

```text
11. Gender为性别，取下机QC的“gender_analysed”列
12. RESULT为C-K列的一个判断，比如提示性别不一致，GC含量高，如果都合格就输出“YES“
详细规则见下面玉婧的邮件
14. 产品编号对应临床新生儿的产品编号DX1968，多中心是DX1964
如微信沟通的方式实现
```

### common

[`etc/QC.txt`](etc/QC.txt)

### others

| title  | key                  | note             |
|--------|----------------------|------------------|
| Order  | i+1                  | index+1          |
| 产品编号   | lims["PRODUCT_CODE"] | from `lims.info` |
| RESULT | RESULT               |                  |

## SMA_result

1. titles:[`etc/title.sma.txt`](etc/title.sma.txt)
2. '产品编码_产品名称':[`updateABC`](#limsinfo)

## excel 格式

### DataValidation

```text
（1）CNV sheet的Y列“报告类别”有下拉选项：正式报告和补充报告
（2）CNV sheet的AA列“杂合性”有下拉选项：Hom和Het和Hemi
（3） CNV sheet的AB列“disGroup”有下拉选项：A和B
（4）CNV sheet的AC列”突变类型”有下拉选项：Pathogenic和Likely pathogenic和VUS
（5）补充实验sheet的I列“β地贫_最终结果”有下拉选项：阴性和SEA-HPFH和Chinese和SEA-HPFH；SEA-HPFH和Chinese；Chinese和SEA-HPFH；Chinese
（6）补充实验sheet的J列“α地贫_最终结果”有下拉选项：阴性和3.7和SEA和4.2和THAI和FIL和3.7;3.7和4.2;4.2和SEA;SEA和3.7;4.2和3.7;SEA和3.7;THAI和3.7;FIL和4.2;SEA和 4.2;THAI和4.2;FIL和SEA;THAI和SEA;FIL和THAI;THAI和THAI;FIL和FIL;FIL
（7）补充实验sheet的M列“SMN1 EX7 del最终结果”有下拉选项：阴性和杂合阳性和纯合阳性和杂合灰区和纯合灰区
```

etc/drop.list.txt 包含对应列的下拉表

### Background Color

```text
4. 需要验证位点的一行字体标记红色
5. 正式报告的位点一行标记蓝色底纹，补充报告位点一行标记绿色底纹
```

### HyperLink

```text
2.1Sheet1 All variants data的内容如下
2. “reads_picture”需要链接reads图，需要链接的位点要同时满足以下2个条件：
（1）正式报告或者补充报告或者clinvar收录是P,LP,P/LP或者HGMD收录DM,DM?,DM/DM?或者库内解读过的
（2）SNV:depth≤40或者A.Ratio≤0.4；Indel:depth≤60或者A.Ratio≤0.45
```

```text
2. “reads_picture”需要链接reads图，需要链接的位点要满足以下1个条件：
（1）正式报告或者补充报告或者clinvar收录是P,LP,P/LP或者HGMD收录DM,DM?,DM/DM?或者库内解读过的
PS：之前是满足2个条件，现在改为1个，高质量位点也是需要查看reads图的
```

```text
2.3 Sheet3 补充实验的内容如下
1. “β地贫_最终结果”“α地贫_最终结果”CNE图链接的添加
```

```go
item["HyperLink"] = filepath.Join(*batch+".result_batCNV-dipin", "chr11_chr16_chrX_cnemap", item["SampleID"]+"_W30S25_cne.jpg")
item["β地贫_最终结果_HyperLink"] = item["HyperLink"]
item["α地贫_最终结果_HyperLink"] = item["HyperLink"]
```

## modules

### anno2xlsx/v2/annno

```go
anno.Score2Pred(item)
anno.UpdateFunction(item)
anno.UpdateAutoRule(item)
item["引物设计"] = anno.PrimerDesign(item)
```

### acmg2015

#### init

```go
acmg2015.AutoPVS1 = *autoPVS1
var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(*acmgDb, "\t", false)).(map[string]string)
for k, v := range acmgCfg {
    acmgCfg[k] = filepath.Join(dbPath, v)
}
acmg2015.Init(acmgCfg)
```

#### use

```go
acmg2015.AddEvidences(item)
item["自动化判断"] = acmg2015.PredACMG2015(item, *autoPVS1)
```

## 多模式

| parameter               | NBSP-IM `-im`       | NBSP              | NBSP-WGS `-wgs`   | CS-WGS `-cs`      |
|-------------------------|---------------------|-------------------|-------------------|-------------------|
| `-avd` `-avdList`       | SNV&INDEL           | All variants data | All variants data | All variants data |
| `-dmd` `-dmdList`       | DMD CNV             | CNV               | CNV-原始            |                   |
| `-lumpy` `-nator`       |                     |                   | CNV               | DMD CNV           |
| `-dipin` `-sma` `-sam2` | THAL CNV + SMN1 CNV | 补充实验              | 补充实验              | 补充实验              |
| `-feature`              |                     | 个特                | 个特                | 个特                |
| `-geneID`               |                     | 基因ID              | 基因ID              | 基因ID              |
| `-drug`                 |                     | 药物检测结果            | 药物检测结果            | 药物检测结果            |
| `-qc`                   | QC                  | QC                | QC                | QC                |
| `-info` `-lims`         | Sample              | 样本信息              | 样本信息              | 样本信息              |
| `bamPath`               |                     | bam文件路径           | bam文件路径           | bam文件路径           |
|                         |                     | 任务单               | 任务单               | 任务单               |

## 模板

- [x] All variants data
- [x] CNV
- [x] DMD-lumpy
- [x] DMD-nator
- [x] 补充实验
- [x] 个特
- [x] 基因ID
- [ ] 药物检测结果
- [x] QC
- [ ] 样本信息
- [x] bam文件路径
- [ ] 任务单

## 一体机

### Sample

- [x] sampleID
  - [x] 样本编号
  - [x] sampleID
- [x] gender
  - [x] 性别
  - [x] Gender
- [x] SampleType
  - [x] 样本类型
  - [x] SampleType
- [x] Date of Birth
  - [x] 出生日期
  - [x] Date of Birth
- [x] Received Date
  - [x] 收样日期
  - [x] Received Date
- [x] ProductID_ProductName
  - [x] 产品套餐编码_产品套餐名称
  - [x] ProductID_ProductName
- [x] Clinical information
  - [x] 临床表现
  - [x] Clinical information
- [x] TaskID
  - [x] 任务编号
  - [x] TaskID
- [x] flow ID
  - [x] 芯片ID
  - [x] flow ID
- [x] pipeline
  - [x] 分析流程
  - [x] pipeline

### QC

- [x] sampleID
  - [x] sampleID
  - [x] sampleID
- [x] Lane ID
  - [x] Lane ID
  - [x] Lane ID
- [x] Barcode ID
  - [x] Barcode ID
  - [x] Barcode ID
- [x] Q20
  - [x] Q20
  - [x] Q20
- [x] Q30
  - [x] Q30
  - [x] Q30
- [x] Target Mean Depth[RM DUP]: (remove_mitochondria)
  - [x] AverageDepth
  - [x] AverageDepth
- [x] Target coverage >=20X percentage[RM DUP]:
  - [x] Target coverage >=20X percentage
  - [x] Target coverage >=20X percentage
- [x] Target Mean Depth[RM DUP]: (mitochondria)
  - [x] mitochondria Target Mean Depth[RM DUP]
  - [x] Mitochondria Target Mean Depth[RM DUP]
- [x] contamination
  - [x] contamination
  - [x] Contamination
- [x] Gender
  - [x] 性别
  - [x] Gender
- [x] 质控
  - [x] 质控
  - [x] QC
- [x] Failed reason
  - [x] 不合格原因
  - [x] Failed reason
- [x] Target coverage >=10X percentage[RM DUP]:
  - [x] Depth>=10(%)
  - [x] Depth>=10(%)
- [x] Target coverage[RM DUP]:
  - [x] Coverage(%)
  - [x] Coverage(%)
- [x] Total mapped GC Rate:
  - [x] GC(%)
  - [x] GC(%)
- [x] Operation Advice
  - [x] 操作建议
  - [x] Operation Advice
  -

### SNV&INDEL

- [x] TaskID
  - [x] 任务编号
  - [x] TaskID
- [x] flow ID
  - [x] 芯片ID
  - [x] flow ID
- [x] ProductID_ProductName
  - [x] 产品套餐编码_产品套餐名称
  - [x] ProductID_ProductName
- [x] SampleID
  - [x] 样本编号
  - [x] SampleID
- [x] Sex
  - [x] 性别
  - [x] Gender
- [x] #Chr
  - [x] Chr
  - [x] Chr
- [x] Zygosity
  - [x] 杂合性
  - [x] Zygosity
- [x] Gene Symbol
  - [x] 基因名称
  - [x] Gene Symbol
- [x] Transcript
  - [x] 转录本
  - [x] Transcript
- [x] cHGVS
  - [x] 碱基改变
  - [x] cHGVS
- [x] pHGVS
  - [x] 氨基酸改变
  - [x] pHGVS
- [x] ExIn_ID
  - [x] 外显子
  - [x] ExIn_ID
- [x] Function
  - [x] 突变类型
  - [x] Variant type
- [x] ClinVar星级
  - [x] ClinVar星级
  - [x] Gold stars in ClinVar
- [x] 自动化判断
  - [x] Automated Pathogenicity
  - [x] Automated Pathogenicity
- [x] 疾病名称
  - [x] 疾病名称
  - [x] DiseaseName
- [x] 遗传模式
  - [x] 遗传模式
  - [x] Genetic Patterns
- [x] 疾病简介
  - [x] 疾病简介
  - [x] Generalization
- [ ] Definition
  - [ ] 致病性
  - [x] Pathogenicity
- [x] 遗传模式判读
  - [x] 遗传模式判读
  - [x] Risk
- [x] 报告类别
  - [x] 是否报告
  - [x] Official Report
- [x] pre_class
  - [x] pre_quality
  - [x] pre_quality
- [x] In BGI database
  - [x] 是否库内变异
  - [x] In BGI database

### DMD CNV

- [x] TaskID
  - [x] 任务编号
  - [x] TaskID
- [x] flow ID
  - [x] flow ID
  - [x] flow ID
- [x] ProductID_ProductName
  - [x] 产品套餐编码_产品套餐名称
  - [x] ProductID_ProductName
- [x] #sample
  - [x] SampleID
  - [x] SampleID
- [x] gender
  - [x] 性别
  - [x] Gender
- [x] chr
  - [x] 染色体
  - [x] Chr
- [x] gene
  - [x] Gene Symbol
  - [x] Gene Symbol
- [x] NM
  - [x] 转录本
  - [x] Transcript
- [x] exon
  - [x] 外显子
  - [x] ExIn_ID
- [x] CNVType
  - [x] CNVType
  - [x] CNVType
- [x] 杂合性
  - [x] 杂合性
  - [x] Zygosity
- [x] 报告类别
  - [x] 是否报告
  - [x] Official Report
- [x] 核苷酸变化
  - [x] 核苷酸变化
  - [x] Variant
- [x] 中文-突变判定
  - [x] 突变类型
  - [x] Pathogenicity
- [x] P0
  - [x] 外显子图
  - [x] DMD Exon_picture
- [x] 新生儿一体机包装变异
  - [x] 是否在库内
  - [x] In BGI database

### THAL CNV

- [x] TaskID
  - [x] 任务编号
  - [x] TaskID
- [x] flow ID
  - [x] flow ID
  - [x] flow ID
- [x] ProductID_ProductName
  - [x] 产品套餐编码_产品套餐名称
  - [x] ProductID_ProductName
- [x] SampleID
  - [x] SampleID
  - [x] SampleID
- [x] Sex
  - [x] 性别
  - [x] Gender
- [x] 地贫_QC
  - [x] 地贫_QC
  - [x] THAL_QC
- [x] β地贫自动化结果
  - [x] β地贫自动化结果
  - [x] Result_HBB CNV-auto
- [x] α地贫自动化结果
  - [x] α地贫自动化结果
  - [x] Result_HBA1/HBA2 CNV-auto
- [x] β地贫_最终结果
  - [x] β地贫_最终结果
  - [x] Result_HBB CNV
- [x] α地贫_最终结果
  - [x] α地贫_最终结果
  - [x] Result_HBA1/HBA2 CNV

### SMN1 CNV

- [x] TaskID
  - [x] 任务编号
  - [x] TaskID
- [x] flow ID
  - [x] flow ID
  - [x] flow ID
- [x] ProductID_ProductName
  - [x] 产品套餐编码_产品套餐名称
  - [x] ProductID_ProductName
- [x] SampleID
  - [x] SampleID
  - [x] SampleID
- [x] Sex
  - [x] 性别
  - [x] Gender
- [x] SMN1_质控结果
  - [x] SMN1_质控结果
  - [x] SMN1_QC
- [x] SMN1 EX7 del最终结果
  - [x] SMN1 EX7 del最终结果
  - [x] Result_SMN1 EX7 DEL
- [x] Official Report
  - [x] 是否报告
  - [x] Official Report
- [x] SMN2_ex7_cn
  - [x] SMN2 EX7 del最终结果
  - [x] Copy of SMN2 EX7

## 携带者WGS

### All variants data

- [x] updateInfo()
  - [x] 期数
  - [x] flow ID
  - [x] 产品编码_产品名称
- [x] updateINDEX()
  - [x] 解读人
  - [x] 审核人
- [x] updateABC()
  - [x] Sex
- [x] updateAvd
  - [x] 1000Gp3 AF
    - 1000G AF
  - [x] 1000Gp3 EAS AF
    - 1000G EAS AF
  - [x] gene+cHGVS
    - Gene Symbol +":"+ cHGVS
  - [x] gene+pHGVS3
  - Gene Symbol +":"+ pHGVS3
  - [x] gene+pHGVS1
    - Gene Symbol +":"+ pHGVS1
  - [x] 引物设计
    - anno.PrimerDesign
- [x] 输入 相同列
  - [x] SampleID
  - [x] Start
  - [x] Stop
  - [x] A.Depth
  - [x] A.Ratio
  - [x] RepeatTag
  - [x] Filter
  - [x] MIM Gene ID
  - [x] MIM Pheno IDs
  - [x] MIM Inheritance
  - [x] TestCode
  - [x] Gene Symbol
  - [x] Transcript
  - [x] cHGVS
  - [x] pHGVS
  - [x] ExIn_ID
  - [x] Zygosity
  - [x] Flank
  - [x] FuncRegion
  - [x] Function
  - [x] Protein
  - [x] Strand
  - [x] CodonChange
  - [x] PolarChange
  - [x] MutationName
  - [x] dbSNP Allele Freq
  - [x] GnomAD EAS HomoAlt Count
  - [x] GnomAD EAS AF
  - [x] GnomAD HomoAlt Count
  - [x] GnomAD AF
  - [x] ESP6500 AF
  - [x] ExAC EAS HomoAlt Count
  - [x] ExAC EAS AF
  - [x] ExAC HomoAlt Count
  - [x] ExAC AF
  - [x] rsID
  - [x] PVFD Homo Count
  - [x] PVFD AF
  - [x] Panel AlleleFreq
  - [x] HGMD pmID
  - [x] HGMD Pred
- [x] TOK1K无内容
  - [x] A.Index
  - [x] PVFD Homo Frequency
  - [x] PVFD AF等级
  - [x] Panel Pred
  - [x] HGMD MutName
  - [x] 备注
  - [x] 解读日期
  - [x] 报告期限
  - [x] 是否有修改
  - [x] 入库时间
  - [x] 同一坐标变异
- [ ] 待处理
  - [x] #Chr
    - [x] "Chr" + #Chr
  - [ ] reads_picture
    - readsPicture()
    - 画图过滤条件
    - 链接有效性
  - [ ] HGMD Disease
  - [x] LOF
    - updateLOF()
      - !LOF[item["Function"]] || gt(item["GnomAD AF"], 0.01) || gt(item["1000G AF"], 0.01)
      - 置空
  - [x] Database
    - Auto ACMG + Check PLP 库内 DX1605
  - [x] 是否是库内位点
    - TOK1K留空
    - Auto ACMG + Check PLP 库内 是
  - [x] PrePregnancyAnno
    - [x] 报告类别
      - 都写 "正式报告"
      - 无报告的写"正式报告"
    - [x] 遗传模式
      - Inheritance
    - [x] 疾病中文名
      - Chinese disease name
    - [x] 中文-疾病背景
      - Chinese disease introduction
      - 空的
    - [x] 中文-突变详情
      - Chinese mutation information
      - 空的
    - [x] Disease*
      - English disease name
    - [x] 英文-疾病背景
      - English disease introduction
      - 空的
    - [x] 英文-突变详情
      - English mutation information
        - 空的
    - [x] 参考文献
      - Reference-final-Info
    - [x] Evidence New + Check
    - [x] 地贫通用名
    - [x] 突变类型
      - Auto ACMG + Check
        - Auto ACMG + Check 判断
  - [x] disGroup
    - PP_disGroup
  - [ ] sanger_tookit
    - [ ] pre_class
    - [ ] pre_pro
  - [ ] 需验证的变异
    - TOP1K处理逻辑未实现
  - [x] 是否国内（际）包装变异
    - 基于TOP1K基因列表
  - [x] anno2xlsx
    - [x] 自动化判断
    - [x] autoRuleName
    - [x] 遗传相符
    - [x] 烈性突变
    - [x] 突变频谱

### CNV

- [x] updateInfo()
  - [x] 期数
  - [x] flow ID
  - [x] 产品编码_产品名称
- [x] updateINDEX()
  - [x] 解读人
  - [x] 审核人
- [ ] updateDMD
  - [x] #sample
    - Sample
- [x] updateABC
  - [x] gender
    - [x] Sex
- [x] Nator
  - [x] Chr
  - [x] Start
  - [x] End
  - [x] CNV_type
  - [x] Gene
  - [x] Gene_num
  - [x] Gene_num_score
  - [x] OMIM_EX
  - [x] Pathogenicity summary
- [ ] lumpy
  - [x] Chr
    - CHROM
  - [x] Start
    - POS
  - [x] End
    - END
  - [x] CNV_type
    - SVTYPE
  - [x] Gene
    - OMIM_Gene
  - [ ] Gene_num
  - [ ] Gene_num_score
  - [x] OMIM_EX
    - OMIM_exon
  - [x] Pathogenicity summary
- [ ] CopyNum
- [ ] 留空
  - [ ] 解读备注
  - [ ] 审核备注
- [ ] 报告类别
- [ ] 核苷酸变化
- [ ] 杂合性
- [ ] 突变类型
- [ ] 参考文献

### 地贫融合基因

-[x] 无，留空

### 补充实验

- [x] updateInfo()
  - [x] 期数
  - [x] flow ID
  - [x] 产品编码_产品名称
- [x] updateINDEX()
  - [x] 解读人
  - [x] 审核人
- [x] updateABC()
  - [x] Sex
    - [x] sex
- [x] sampleID
  - sample
  - Sample
- [x] updateDipin()
  - [x] 地贫_QC
  - [x] β地贫_chr11
  - [x] α地贫_chr16
  - [x] β地贫_最终结果
  - [x] α地贫_最终结果
- [x] updateSma2
  - [x] SMN1_检测结果
  - [x] SMN1_质控结果
  - [x] SMN1 EX7 del最终结果
- [x] 检测范围外
  - [x] F8int1h-1.5k&2k最终结果
  - [x] F8int22h-10.8k&12k最终结果

### QC

- [x] updateInfo()
  - [x] 产品编号
- [x] updateINDEX()
  - [x] 解读人
  - [x] 审核人
- [x] 输入 相同列
  - [x] Order
  - [x] Sample
  - [x] AverageDepth
  - [x] Depth>=30(%)
  - [x] GC(%)
  - [ ] Gender
    - [ ] 性别校验
  - [x] RESULT
- [x] Q20(%)
  - Q20
- [x] Q30(%)
  - Q30

### 样本信息

- [x] 空sheet

### bam文件路径

- [x] 直接填充bam路径

### 任务单

- [x] 只有空表头，继承自模板

### TO-DO

- 颜色
- 列表
- autoPVS1 注释
- sanger_tookit
- TOP1K 基因名 核对
- 性别校验

### WGS使用

```shell
# WGS 携带者 使用 NB2xlsx 生成 BB 表格的说明

# 软件目录$bin # /home/wangyaoshen/pipeline/cs/NB2xlsx
# 输出目录$workdir
# 输出前缀$prefix
# 输出结果$workdir/$prefix.xlsx
# 更新$prefix
prefix=$workdir/$prefix

# 样品信息
## 输入样品信息$info # input.list
## 每个样品一行，不能重复
## 必要字段：
##        sampleID  样品编号
##        gender  样品性别
##        TaskID  期数
##        ProductID_ProductName 产品编码_产品名称
##        ProductID 产品编号
##        flow ID flow_ID
## 字段命名确定后跟我对称下

# All variants data
## anno2xlsx目录$anno # /home/wangyaoshen/pipeline/anno2xlsx
## snv&indel注释结果$snv # carrier_2s.addHGMD.filter.func.vaf.rmdup.all.xlsx
## 输出$workdir
##            avd.Sheet1.txt
##            avd.addDisease.txt
##            avd.[sampleID].tsv
##            avd.list
### /home/wangyaoshen/local/bin/xlsx2txt
xlsx2txt -input $snv -prefix $workdir/avd 
$anno/addDiseaseInfo2SNV/addDiseaseInfo2SNV -input $workdir/avd.Sheet1.txt -output $workdir/avd.addDisease.tsv
### /home/wangyaoshen/local/bin/splitTsv 
### 第5列为样品编号
splitTsv -i $workdir/avd.addDisease.tsv -k 5 -p $workdir/avd -h > $workdir/avd.list 

# CNV
## lumpy结果$lumpy # DMD_batch6_57SAMPLE_lumpy.xls
## nator结果$nator # DMD_batch6_57SAMPLE_Nator.xls

# 补充实验
## 地贫结果$dipin # merge_dipin_annot.xls
## 新版SMA结果$sma # carrier_2s.tsv

# QC
## QC结果$qc # carrier_2s_qc.xlsx
## 输出$workdir/qc.qc_file.txt
xlsx2txt -input $qc -prefix $workdir/qc

# bam文件路径
## bam列表$bamList
### 只有一列bam路径
## 参考命令：
##        ls /home/cromwell/project/fuxiangke/wgs_carrier_2s_test_0228/*/bam_chr/*.final.merge.bam > $bamList

# 生信性别信息
## 输入性别信息$genderMap # gender.txt
### 第一列样品编号
### 第二列性别
## 参考命令：
##        cut -f 4,7 /home/cromwell/project/fuxiangke/wgs_carrier_2s_test_0228/*/QC/sex.txt > $genderMap

$bin/NB2xlsx \
  -template $bin/template/CS.BB.xlsx \
  -prefix $prefix \
  -info $info \
  -avdList $workdir/avd.list \
  -lumpy $lumpy \
  -nator $nator \
  -dipin $dipin \
  -sma2 $sma \
  -qc $workdir/qc.qc_file.txt \
  -bamPath $bamList \
  -gender $genderMap \
  -cs -acmg \
  # -autoPVS1 # avd注释有进行autoPVS1注释
```

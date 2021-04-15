package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	templatePath = filepath.Join(exPath, "template")
)

// flag
var (
	port = flag.String(
		"port",
		":9091",
		"port for server",
	)
)

func main() {
	// 设置访问路由
	http.HandleFunc("/summary", summary)
	http.HandleFunc("/summaryResult", summaryResult)
	http.HandleFunc("/NB2xlsx", NB2xlsx)
	http.HandleFunc("/NB2xlsxResult", NB2xlsxResult)
	http.HandleFunc("/", index)
	simpleUtil.CheckErr(http.ListenAndServe(*port, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		generalGet(w, r, "index.html")
	} else {
		http.ServeFile(w, r, strings.TrimPrefix(r.URL.Path, "/"))
	}
}

func generalGet(w http.ResponseWriter, r *http.Request, html string) {
	var t, e = template.ParseFiles(filepath.Join(templatePath, html))
	if e != nil {
		handleError(w, e)
		return
	}

	logRequest(r)

	e = t.Execute(w, nil)
	if e != nil {
		handleError(w, e)
		return
	}
}

func summary(w http.ResponseWriter, r *http.Request) {
	generalGet(w, r, "summary.html")
}

// NB2xlsx use NB2xlsx to get result excel
func NB2xlsx(w http.ResponseWriter, r *http.Request) {
	generalGet(w, r, "NB2xlsx.html")

}

func handleError(w http.ResponseWriter, e error, msg ...string) {
	log.Printf("%+v[%+v]\n", e, msg)
	_, e = fmt.Fprintf(w, "%+v[%+v]\n", e, msg)
	if e != nil {
		log.Printf("%+v[%+v]\n", e, msg)
	}
}

// Info contain Href and Message
type Info struct {
	Href    string
	Message string
}

// Result contain Title Tag Href and Message
type Result struct {
	Title   string
	Tag     string
	Href    string
	Message string
}

// SummaryResult contain Title Message and Info of Suumary Result []Anno
type SummaryResult struct {
	Title   string
	Message string
	Summary Info
	Anno    []Info
	Result  Info
}

func summaryResult(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var t, e = template.ParseFiles(filepath.Join(templatePath, "summary.result.html"))
		if e != nil {
			handleError(w, e)
			return
		}
		e = os.MkdirAll("input", 0755)
		if e != nil {
			handleError(w, e)
			return
		}
		e = os.MkdirAll("output", 0755)
		if e != nil {
			handleError(w, e)
			return
		}
		e = r.ParseMultipartForm(31 << 20)
		if e != nil {
			handleError(w, e)
			return
		}
		logRequest(r)

		summaryInfos, e := uploadFile(r, "summary")
		if e != nil {
			handleError(w, e)
			return
		}
		annoInfos, e := uploadFile(r, "anno")
		if e != nil {
			handleError(w, e)
			return
		}
		var result = SummaryResult{
			Summary: summaryInfos[0],
			Anno:    annoInfos,
			Result:  Info{},
		}
		result.Result.Message = fmt.Sprintf("%s.%s.xlsx", strings.TrimSuffix(result.Summary.Message, ".xlsx"), time.Now().Format("2006-01-02"))
		result.Title = result.Result.Message
		result.Result.Href = "output/" + result.Result.Message
		var annos []string
		for _, anno := range result.Anno {
			annos = append(annos, anno.Href)
		}
		var cmd = exec.Command(filepath.Join("..", "xlsx2summary"), "-input", result.Summary.Href, "-anno", strings.Join(annos, ","), "-prefix", filepath.Join("output", strings.TrimSuffix(result.Summary.Message, ".xlsx")))
		log.Println(cmd.String())
		output, e := cmd.CombinedOutput()
		log.Printf("%s", output)
		if e != nil {
			handleError(w, e, cmd.String(), "\n", string(output))
			return
		}
		result.Message = cmd.String() + "\n" + string(output)
		e = t.Execute(w, result)
		if e != nil {
			handleError(w, e)
			return
		}
	} else {
		var _, e = fmt.Fprintln(w, "only support POST method")
		log.Printf("%+v", e)
	}
}

func uploadFile(r *http.Request, key string) (infos []Info, e error) {
	var fhs, ok = r.MultipartForm.File[key]
	if !ok {
		e = errors.New(key + "not found!")
		return
	}
	for _, fh := range fhs {
		var f io.ReadCloser
		f, e = fh.Open()
		if e != nil {
			return
		}
		var sPath = "input/" + fh.Filename
		var info = Info{
			Href:    sPath,
			Message: fh.Filename,
		}
		e = upload(f, sPath)
		if e != nil {
			return
		}
		e = f.Close()
		if e != nil {
			return
		}
		infos = append(infos, info)
	}
	return
}

func upload(file io.Reader, dest string) error {
	var f, err = os.Create(dest)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, file)
	if err == nil {
		err = f.Close()
	} else {
		f.Close()
	}
	return err
}

func logRequest(r *http.Request) {
	log.Println(r.Form) //这些信息是输出到服务器端的打印信息
	log.Println("path", r.URL.Path)
	log.Println("scheme", r.URL.Scheme)
	log.Println(r.Form["url_long"])
	for k, v := range r.Form {
		log.Printf("key:%s\t", k)
		if len(v) < 1024 {
			log.Printf("key:[%s]\tval:[%v]\n", k, v)
		} else {
			log.Printf("key:[%s]\tval: large data!\n", k)
		}
	}
}

// NB2xlsxResult use NB2xlsx create result excel and return result page
func NB2xlsxResult(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var t, e = template.ParseFiles(filepath.Join(templatePath, "NB2xlsx.result.html"))
		if e != nil {
			handleError(w, e)
			return
		}
		e = os.MkdirAll("output/NBS", 0755)
		if e != nil {
			handleError(w, e)
			return
		}
		e = r.ParseMultipartForm(31 << 20)
		if e != nil {
			handleError(w, e)
			return
		}
		logRequest(r)
		var batchName = r.Form["batchName"][0]
		var cmd = exec.Command(
			"bash",
			"NB2xlsx.sh",
			batchName,
		)
		msg, e := cmd.CombinedOutput()
		var msgStr = string(msg)
		fmt.Printf("%s\n", msgStr)
		if e != nil {
			handleError(w, e, "\ncmd:\t", cmd.String(), "\nlog:\t", msgStr)
			return
		}
		var result = Result{
			Title:   "NB2xlsx",
			Tag:     batchName,
			Href:    "",
			Message: msgStr,
		}
		e = t.Execute(w, result)
		if e != nil {
			handleError(w, e, "\ncmd:\t", cmd.String(), "\nlog:\t", msgStr)
			return
		}
	} else {
		var _, e = fmt.Fprintln(w, "only support POST method")
		log.Printf("%+v", e)
	}
}

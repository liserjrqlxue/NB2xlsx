package main

import (
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

func main() {
	// 设置访问路由
	http.HandleFunc("/summary", summary)
	http.HandleFunc("/summaryResult", summaryResult)
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			// static file server
			http.ServeFile(w, r, strings.TrimPrefix(r.URL.Path, "/"))
			return
		},
	)
	simpleUtil.CheckErr(http.ListenAndServe(":9091", nil))
}

func summary(w http.ResponseWriter, r *http.Request) {
	var t, e = template.ParseFiles(filepath.Join(templatePath, "summary.html"))
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

func handleError(w http.ResponseWriter, e error, msg ...string) {
	log.Printf("%+v[%+v]\n", e, msg)
	_, e = fmt.Fprintf(w, "%+v[%+v]\n", e, msg)
	if e != nil {
		log.Printf("%+v[%+v]\n", e, msg)
	}
}

type Info struct {
	Href    string
	Message string
}

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
		log.Printf("Form:%+v\n", r.Form)
		log.Printf("PostForm:%+v\n", r.PostForm)
		log.Printf("MultipartForm:%+v\n", r.MultipartForm)
		log.Printf("File:%+v\n", r.MultipartForm.File)
		log.Printf("Value:%+v\n", r.MultipartForm.Value)
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
		result.Result.Message = fmt.Sprintf("%s.%s.xlsx", result.Summary.Message, time.Now().Format("2006-01-02"))
		result.Title = result.Result.Message
		result.Result.Href = "output/" + result.Result.Message
		var annos []string
		for _, anno := range result.Anno {
			annos = append(annos, anno.Href)
		}
		var cmd = exec.Command(filepath.Join("..", "xlsx2summary"), "-input", result.Summary.Href, "-anno", strings.Join(annos, ","), "-prefix", "output/"+result.Summary.Message)
		output, e := cmd.CombinedOutput()
		if e != nil {
			handleError(w, e, cmd.String(), string(output))
			return
		}
		result.Message = string(output)
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
	for _, fh := range r.MultipartForm.File[key] {
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
	var f, err = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer simpleUtil.DeferClose(f)
	_, err = io.Copy(f, file)
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

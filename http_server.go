package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

const (
	JsStr = `
	<script>
	;(function() {
		function setUrl(url){
			var pattern='t=([^&]*)';
			var replaceText='t='+new Date().getTime(); 
			if(url.match(pattern)){
				var tmp='/(t=)([^&]*)/gi';
				tmp=url.replace(eval(tmp),replaceText);
				return tmp;
			}else{ 
				if(url.match('[\?]')){ 
					return url+'&'+replaceText; 
				}else{ 
					return url+'?'+replaceText; 
				} 
			}
		}
	
		function flushRes() {
			links = document.querySelectorAll('link')
			links.forEach(element => {
				element.href = setUrl(element.href);
			});
	
			images = document.querySelectorAll('img')
			images.forEach(element =>{
				element.src = setUrl(element.src);
			});
	
			scripts = document.querySelectorAll('script')
			scripts.forEach(element =>{
				if (element.src){
					element.src = setUrl(element.src);
				}
			});
		}
	
		var conn = new WebSocket("ws://%s/ws");
		conn.onclose = function(evt) {
			document.querySelector('html').innerHTML = 'server closed';
		}
		conn.onmessage = function(evt) {
			if (evt.data == "reload") {
				window.location.reload()
			} else {
				flushRes()
			}
		}
		flushRes()
	})();
	</script>`
)

type fileHandler struct {
	root http.FileSystem
	h    http.Handler
}

func (f *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/" {
		path = "index.html"
	}

	filePath := filepath.Join(RootDir, path)

	ext := filepath.Ext(filePath)
	switch ext {
	case ".html", ".htm":
		if _, err := os.Stat(filePath); err == nil {
			ff, err := f.root.Open(filePath)
			if err == nil {
				defer ff.Close()
				buffer := bytes.Buffer{}
				buffer.ReadFrom(ff)
				ip := getLocalIP()
				buffer.WriteString(fmt.Sprintf(JsStr, ip+":"+Port))
				w.Write(buffer.Bytes())
				return
			}
		}
	}

	f.h.ServeHTTP(w, r)
}

func fileServer(root http.FileSystem, h http.Handler) http.Handler {
	return &fileHandler{root, h}
}

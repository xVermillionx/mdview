package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/browser"
	"gitlab.com/golang-commonmark/markdown"
)

const appVersion = "1.4.0"

func main() {
	var outfilePtr = flag.String("o", "", "Output filename. (Optional)")
	var versionPtr = flag.Bool("version", false, "Prints mdview version.")
	var helpPtr = flag.Bool("help", false, "Prints mdview help message.")
	var barePtr = flag.Bool("bare", false, "Bare HTML with no style applied.")
	var filepathPtr = flag.Bool("filepath", false, "Output filepath instead of html on pipe/redirect")
	var xhtmlPtr = flag.Bool("xhtml", false, "Choose XHTML instead of HTML")
	var darkPtr = flag.Bool("dark", false, "Darkmode")
	flag.BoolVar(versionPtr, "v", false, "Prints mdview version.")
	flag.BoolVar(helpPtr, "h", false, "Prints mdview help message.")
	flag.BoolVar(barePtr, "b", false, "Bare HTML with no style applied.")
	flag.BoolVar(filepathPtr, "f", false, "Output filelocation in pipe/redirect")
	flag.BoolVar(xhtmlPtr, "x", false, "Choose XHTML instead of HTML")
	flag.BoolVar(darkPtr, "d", false, "Darkmode")

	flag.Parse()
	inputFilename := flag.Arg(0)

	if *versionPtr {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if inputFilename == "" || *helpPtr {
		os.Stderr.WriteString("Usage:\nmdview [options] <filename>\nFormats markdown and launches it in a browser.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	dat, err := ioutil.ReadFile(inputFilename)
	check(err)

	var htmlorxhtml =  markdown.HTML(true)
	if *xhtmlPtr {
		htmlorxhtml = markdown.XHTMLOutput(true)
	}

	md := markdown.New(
		htmlorxhtml,
		markdown.Nofollow(true),
		markdown.Tables(true),
		markdown.Typographer(true))

	markdownTokens := md.Parse(dat)
	html := md.RenderTokensToString(markdownTokens)
	title := getTitle(markdownTokens)

	outfilePath := *outfilePtr
	if outfilePath == "" {
		outfilePath = tempFileName("mdview", ".html")
	}

	f, err := os.Create(outfilePath)
	check(err)
	defer f.Close()
	actualStyle := style
	if *barePtr {
		actualStyle = ""
	}
	if *darkPtr {
		actualStyle = darkstyle
	}
	_, err = fmt.Fprintf(f, template, actualStyle, title, html)
	check(err)
	f.Sync()

  o, _ := os.Stdout.Stat()
  if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice { //Terminal
    //Display info to the terminal
    err = browser.OpenFile(outfilePath)
    check(err)
  } else { //It is not the terminal
    // Display info to a pipe

    if *filepathPtr {
      _, err = fmt.Printf(outfilePath)
    } else {
      _, err = fmt.Printf(template, actualStyle, title, html)
    }
    check(err)
  }
}

func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(getTempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func getTempDir() string {
	if os.Getenv("SNAP_USER_COMMON") != "" {
		var tmpdir = os.Getenv("HOME") + "/mdview-temp"
		if _, err := os.Stat(tmpdir); os.IsNotExist(err) {
			err = os.Mkdir(tmpdir, 0700)
			check(err)
		}
		return tmpdir
	}
	return os.TempDir()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getTitle(tokens []markdown.Token) string {
	var result string
	if len(tokens) > 0 {
		for i := 0; i < len(tokens); i++ {
			if topLevelHeading, ok := tokens[i].(*markdown.HeadingOpen); ok {
				for j := i + 1; j < len(tokens); j++ {
					if token, ok := tokens[j].(*markdown.HeadingClose); ok && token.Lvl == topLevelHeading.Lvl {
						break
					}
					result += getText(tokens[j])
				}
				result = strings.TrimSpace(result)
				break
			}
		}
	}
	return result
}

func getText(token markdown.Token) string {
	switch token := token.(type) {
	case *markdown.Text:
		return token.Content
	case *markdown.Inline:
		result := ""
		for _, token := range token.Children {
			result += getText(token)
		}
		return result
	}
	return ""

}

const template = "<!DOCTYPE html><html><head><meta http-equiv=\"content-type\" content=\"text/html; charset=utf-8\"> <style>%s</style><title>%s</title></head><body class=\"markdown-body\">%s</body></html>"

const style = `.markdown-body {box-sizing: border-box;min-width: 200px;max-width:
	 	980px;margin: 0 auto;padding: 45px;}	@media (max-width: 767px) {.markdown-body
		{padding: 15px;}}.markdown-body hr::after,.markdown-body::after{clear:both}
		@font-face{font-family:octicons-link;src:url(data:font/woff;charset=utf-8;
		base64,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)
		format('woff')}.markdown-body{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%;
		color:#24292e;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",
		Helvetica,Arial,sans-serif,\"Apple Color Emoji\",\"Segoe UI Emoji\",\"Segoe UI Symbol\"
		;font-size:16px;line-height:1.5;word-wrap:break-word}.markdown-body .pl-c{color:#6a737d}
		.markdown-body .pl-c1,.markdown-body .pl-s .pl-v{color:#005cc5}.markdown-body .pl-e,
		.markdown-body .pl-en{color:#6f42c1}.markdown-body .pl-s .pl-s1,.markdown-body .pl-smi{color:#24292e}
		.markdown-body .pl-ent{color:#22863a}.markdown-body .pl-k{color:#d73a49}.markdown-body .pl-pds,
		.markdown-body .pl-s,.markdown-body .pl-s .pl-pse .pl-s1,.markdown-body .pl-sr,.markdown-body 
		.pl-sr .pl-cce,.markdown-body .pl-sr .pl-sra,.markdown-body .pl-sr .pl-sre{color:#032f62}.markdown-body
		.pl-smw,.markdown-body .pl-v{color:#e36209}.markdown-body .pl-bu{color:#b31d28}.markdown-body
		.pl-ii{color:#fafbfc;background-color:#b31d28}.markdown-body .pl-c2{color:#fafbfc;background-color:#d73a49}
		.markdown-body .pl-c2::before{content:\"^M\"}.markdown-body .pl-sr .pl-cce{font-weight:700;color:#22863a}
		.markdown-body .pl-ml{color:#735c0f}.markdown-body .pl-mh,.markdown-body .pl-mh .pl-en,.markdown-body
		.pl-ms{font-weight:700;color:#005cc5}.markdown-body .pl-mi{font-style:italic;color:#24292e}.markdown-body
		.pl-mb{font-weight:700;color:#24292e}.markdown-body .pl-md{color:#b31d28;background-color:#ffeef0}
		.markdown-body .pl-mi1{color:#22863a;background-color:#f0fff4}.markdown-body .pl-mc{color:#e36209;
		background-color:#ffebda}.markdown-body .pl-mi2{color:#f6f8fa;background-color:#005cc5}
		.markdown-body .pl-mdr{font-weight:700;color:#6f42c1}.markdown-body .pl-ba{color:#586069}
		.markdown-body .pl-sg{color:#959da5}.markdown-body .pl-corl{text-decoration:underline;
		color:#032f62}.markdown-body .octicon{display:inline-block;fill:currentColor;vertical-align:text-bottom}
		.markdown-body hr::after,.markdown-body hr::before,.markdown-body::after,
		.markdown-body::before{display:table;content:\"\"}.markdown-body a{background-color:transparent;
		color:#0366d6;text-decoration:none}.markdown-body a:active,.markdown-body a:hover{outline-width:0}
		.markdown-body h1{margin:.67em 0}.markdown-body img{border-style:none}.markdown-body hr{box-sizing:content-box}
		.markdown-body input{font:inherit;margin:0;overflow:visible;font-family:inherit;font-size:inherit;line-height:inherit}
		.markdown-body dl dt,.markdown-body strong,.markdown-body table th{font-weight:600}.markdown-body code,
		.markdown-body pre{font-family:SFMono-Regular,Consolas,\"Liberation Mono\",Menlo,Courier,monospace}
		.markdown-body [type=checkbox]{box-sizing:border-box;padding:0}.markdown-body *{box-sizing:border-box}
		.markdown-body a:hover{text-decoration:underline}.markdown-body td,.markdown-body th{padding:0}
		.markdown-body blockquote{margin:0}.markdown-body ol ol,.markdown-body ul ol{list-style-type:lower-roman}
		.markdown-body ol ol ol,.markdown-body ol ul ol,.markdown-body ul ol ol,.markdown-body
		ul ul ol{list-style-type:lower-alpha}.markdown-body dd{margin-left:0}.markdown-body 
		.pl-0{padding-left:0!important}.markdown-body .pl-1{padding-left:4px!important}.markdown-body 
		.pl-2{padding-left:8px!important}.markdown-body .pl-3{padding-left:16px!important}.markdown-body 
		.pl-4{padding-left:24px!important}.markdown-body .pl-5{padding-left:32px!important}.markdown-body 
		.pl-6{padding-left:40px!important}.markdown-body>:first-child{margin-top:0!important}
		.markdown-body>:last-child{margin-bottom:0!important}.markdown-body a:not([href]){color:inherit;
		text-decoration:none}.markdown-body .anchor{float:left;padding-right:4px;margin-left:-20px;
		line-height:1}.markdown-body .anchor:focus{outline:0}.markdown-body blockquote,
		.markdown-body dl,.markdown-body ol,.markdown-body p,.markdown-body pre,.markdown-body table,
		.markdown-body ul{margin-top:0;margin-bottom:16px}.markdown-body hr{overflow:hidden;background:#e1e4e8;
		height:.25em;padding:0;margin:24px 0;border:0}.markdown-body blockquote{padding:0 1em;color:#6a737d;
		border-left:.25em solid #dfe2e5}.markdown-body h1,.markdown-body h2{padding-bottom:.3em;
		border-bottom:1px solid #eaecef}.markdown-body blockquote>:first-child{margin-top:0}
		.markdown-body blockquote>:last-child{margin-bottom:0}.markdown-body h1,.markdown-body h2,
		.markdown-body h3,.markdown-body h4,.markdown-body h5,.markdown-body h6{margin-top:24px;
		margin-bottom:16px;font-weight:600;line-height:1.25}.markdown-body h1 .octicon-link,.markdown-body 
		h2 .octicon-link,.markdown-body h3 .octicon-link,.markdown-body h4 .octicon-link,.markdown-body 
		h5 .octicon-link,.markdown-body h6 .octicon-link{color:#1b1f23;vertical-align:middle;visibility:hidden}
		.markdown-body h1:hover .anchor,.markdown-body h2:hover .anchor,.markdown-body h3:hover .anchor,
		.markdown-body h4:hover .anchor,.markdown-body h5:hover .anchor,.markdown-body h6:hover .anchor{text-decoration:none}
		.markdown-body h1:hover .anchor .octicon-link,.markdown-body h2:hover .anchor .octicon-link,.markdown-body 
		h3:hover .anchor .octicon-link,.markdown-body h4:hover .anchor .octicon-link,.markdown-body h5:hover .anchor 
		.octicon-link,.markdown-body h6:hover .anchor .octicon-link{visibility:visible}.markdown-body h1{font-size:2em}
		.markdown-body h2{font-size:1.5em}.markdown-body h3{font-size:1.25em}.markdown-body h4{font-size:1em}.markdown-body
		h5{font-size:.875em}.markdown-body h6{font-size:.85em;color:#6a737d}.markdown-body ol,.markdown-body ul{padding-left:2em}
		.markdown-body ol ol,.markdown-body ol ul,.markdown-body ul ol,.markdown-body ul ul{margin-top:0;margin-bottom:0}
		.markdown-body li{word-wrap:break-all}.markdown-body li>p{margin-top:16px}.markdown-body li+li{margin-top:.25em}
		.markdown-body dl{padding:0}.markdown-body dl dt{padding:0;margin-top:16px;font-size:1em;font-style:italic}.markdown-body
		dl dd{padding:0 16px;margin-bottom:16px}.markdown-body table{border-spacing:0;border-collapse:collapse;display:block;
		width:100%;overflow:auto}.markdown-body table td,.markdown-body table th{padding:6px 13px;border:1px solid #dfe2e5}
		.markdown-body table tr{background-color:#fff;border-top:1px solid #c6cbd1}.markdown-body table 
		tr:nth-child(2n){background-color:#f6f8fa}.markdown-body img{max-width:100%;box-sizing:content-box;background-color:#fff}
		.markdown-body img[align=right]{padding-left:20px}.markdown-body img[align=left]{padding-right:20px}.markdown-body 
		code{padding:.2em .4em;margin:0;font-size:85%;background-color:rgba(27,31,35,.05);border-radius:3px}.markdown-body 
		pre{word-wrap:normal}.markdown-body pre>code{padding:0;margin:0;font-size:100%;word-break:normal;white-space:pre;background:0 0;
		border:0}.markdown-body .highlight{margin-bottom:16px}.markdown-body .highlight pre{margin-bottom:0;word-break:normal}
		.markdown-body .highlight pre,.markdown-body pre{padding:16px;overflow:auto;font-size:85%;line-height:1.45;
		background-color:#f6f8fa;border-radius:3px}.markdown-body pre code{display:inline;max-width:auto;padding:0;margin:0;
		overflow:visible;line-height:inherit;word-wrap:normal;background-color:transparent;border:0}.markdown-body 
		.full-commit .btn-outline:not(:disabled):hover{color:#005cc5;border-color:#005cc5}.markdown-body kbd{display:inline-block;
		padding:3px 5px;font:11px SFMono-Regular,Consolas,\"Liberation Mono\",Menlo,Courier,monospace;line-height:10px;color:#444d56;
		vertical-align:middle;background-color:#fafbfc;border:1px solid #d1d5da;border-bottom-color:#c6cbd1;border-radius:3px;
		box-shadow:inset 0 -1px 0 #c6cbd1}.markdown-body :checked+.radio-label{position:relative;z-index:1;border-color:#0366d6}
		.markdown-body .task-list-item{list-style-type:none}.markdown-body .task-list-item+.task-list-item{margin-top:3px}
		.markdown-body .task-list-item input{margin:0 .2em .25em -1.6em;vertical-align:middle}.markdown-body hr{border-bottom-color:#eee}`

const darkstyle = `.markdown-body {color: #e8e8e8;background-color:#282828;box-sizing: border-box;min-width: 200px;max-width:
	 	980px;margin: 0 auto;padding: 45px;}	@media (max-width: 767px) {.markdown-body
		{padding: 15px;}}.markdown-body hr::after,.markdown-body::after{clear:both}
		@font-face{font-family:octicons-link;src:url(data:font/woff;charset=utf-8;
		base64,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)
		format('woff')}.markdown-body{-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%;
		color:#e8e8e8;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",
		Helvetica,Arial,sans-serif,\"Apple Color Emoji\",\"Segoe UI Emoji\",\"Segoe UI Symbol\"
		;font-size:16px;line-height:1.5;word-wrap:break-word}.markdown-body .pl-c{color:#6a737d}
		.markdown-body .pl-c1,.markdown-body .pl-s .pl-v{color:#005cc5}.markdown-body .pl-e,
		.markdown-body .pl-en{color:#6f42c1}.markdown-body .pl-s .pl-s1,.markdown-body .pl-smi{color:#24292e}
		.markdown-body .pl-ent{color:#22863a}.markdown-body .pl-k{color:#d73a49}.markdown-body .pl-pds,
		.markdown-body .pl-s,.markdown-body .pl-s .pl-pse .pl-s1,.markdown-body .pl-sr,.markdown-body 
		.pl-sr .pl-cce,.markdown-body .pl-sr .pl-sra,.markdown-body .pl-sr .pl-sre{color:#032f62}.markdown-body
		.pl-smw,.markdown-body .pl-v{color:#e36209}.markdown-body .pl-bu{color:#b31d28}.markdown-body
		.pl-ii{color:#fafbfc;background-color:#b31d28}.markdown-body .pl-c2{color:#fafbfc;background-color:#d73a49}
		.markdown-body .pl-c2::before{content:\"^M\"}.markdown-body .pl-sr .pl-cce{font-weight:700;color:#22863a}
		.markdown-body .pl-ml{color:#735c0f}.markdown-body .pl-mh,.markdown-body .pl-mh .pl-en,.markdown-body
		.pl-ms{font-weight:700;color:#005cc5}.markdown-body .pl-mi{font-style:italic;color:#24292e}.markdown-body
		.pl-mb{font-weight:700;color:#24292e}.markdown-body .pl-md{color:#b31d28;background-color:#ffeef0}
		.markdown-body .pl-mi1{color:#22863a;background-color:#f0fff4}.markdown-body .pl-mc{color:#e36209;
		background-color:#ffebda}.markdown-body .pl-mi2{color:#f6f8fa;background-color:#005cc5}
		.markdown-body .pl-mdr{font-weight:700;color:#6f42c1}.markdown-body .pl-ba{color:#586069}
		.markdown-body .pl-sg{color:#959da5}.markdown-body .pl-corl{text-decoration:underline;
		color:#032f62}.markdown-body .octicon{display:inline-block;fill:currentColor;vertical-align:text-bottom}
		.markdown-body hr::after,.markdown-body hr::before,.markdown-body::after,
		.markdown-body::before{display:table;content:\"\"}.markdown-body a{background-color:transparent;
		color:#00bbff; text-decoration:none}.markdown-body a:active,.markdown-body a:hover{outline-width:0}
		.markdown-body h1{margin:.67em 0}.markdown-body img{border-style:none}.markdown-body hr{box-sizing:content-box}
		.markdown-body input{font:inherit;margin:0;overflow:visible;font-family:inherit;font-size:inherit;line-height:inherit}
		.markdown-body dl dt,.markdown-body strong,.markdown-body table th{font-weight:600}.markdown-body code,
		.markdown-body pre{font-family:SFMono-Regular,Consolas,\"Liberation Mono\",Menlo,Courier,monospace}
		.markdown-body [type=checkbox]{box-sizing:border-box;padding:0}.markdown-body *{box-sizing:border-box}
		.markdown-body a:hover{text-decoration:underline}.markdown-body td,.markdown-body th{padding:0}
		.markdown-body blockquote{margin:0}.markdown-body ol ol,.markdown-body ul ol{list-style-type:lower-roman}
		.markdown-body ol ol ol,.markdown-body ol ul ol,.markdown-body ul ol ol,.markdown-body
		ul ul ol{list-style-type:lower-alpha}.markdown-body dd{margin-left:0}.markdown-body 
		.pl-0{padding-left:0!important}.markdown-body .pl-1{padding-left:4px!important}.markdown-body 
		.pl-2{padding-left:8px!important}.markdown-body .pl-3{padding-left:16px!important}.markdown-body 
		.pl-4{padding-left:24px!important}.markdown-body .pl-5{padding-left:32px!important}.markdown-body 
		.pl-6{padding-left:40px!important}.markdown-body>:first-child{margin-top:0!important}
		.markdown-body>:last-child{margin-bottom:0!important}.markdown-body a:not([href]){color:inherit;
		text-decoration:none}.markdown-body .anchor{float:left;padding-right:4px;margin-left:-20px;
		line-height:1}.markdown-body .anchor:focus{outline:0}.markdown-body blockquote,
		.markdown-body dl,.markdown-body ol,.markdown-body p,.markdown-body pre,.markdown-body table,
		.markdown-body ul{margin-top:0;margin-bottom:16px}.markdown-body hr{overflow:hidden;background:#888888;
		height:.25em;padding:0;margin:24px 0;border:0}.markdown-body blockquote{padding:0 1em;color:#a8a8a8;
		border-left:.25em solid #888888}.markdown-body h1,.markdown-body h2{padding-bottom:.3em;
		border-bottom:1px solid #888888}.markdown-body blockquote>:first-child{margin-top:0}
		.markdown-body blockquote>:last-child{margin-bottom:0}.markdown-body h1,.markdown-body h2,
		.markdown-body h3,.markdown-body h4,.markdown-body h5,.markdown-body h6{margin-top:24px;
		margin-bottom:16px;font-weight:600;line-height:1.25}.markdown-body h1 .octicon-link,.markdown-body 
		h2 .octicon-link,.markdown-body h3 .octicon-link,.markdown-body h4 .octicon-link,.markdown-body 
		h5 .octicon-link,.markdown-body h6 .octicon-link{color:#1b1f23;vertical-align:middle;visibility:hidden}
		.markdown-body h1:hover .anchor,.markdown-body h2:hover .anchor,.markdown-body h3:hover .anchor,
		.markdown-body h4:hover .anchor,.markdown-body h5:hover .anchor,.markdown-body h6:hover .anchor{text-decoration:none}
		.markdown-body h1:hover .anchor .octicon-link,.markdown-body h2:hover .anchor .octicon-link,.markdown-body 
		h3:hover .anchor .octicon-link,.markdown-body h4:hover .anchor .octicon-link,.markdown-body h5:hover .anchor 
		.octicon-link,.markdown-body h6:hover .anchor .octicon-link{visibility:visible}.markdown-body h1{font-size:2em}
		.markdown-body h2{font-size:1.5em}.markdown-body h3{font-size:1.25em}.markdown-body h4{font-size:1em}.markdown-body
		h5{font-size:.875em}.markdown-body h6{font-size:.85em;color:#c8c8c8}.markdown-body ol,.markdown-body ul{padding-left:2em}
		.markdown-body ol ol,.markdown-body ol ul,.markdown-body ul ol,.markdown-body ul ul{margin-top:0;margin-bottom:0}
		.markdown-body li{word-wrap:break-all}.markdown-body li>p{margin-top:16px}.markdown-body li+li{margin-top:.25em}
		.markdown-body dl{padding:0}.markdown-body dl dt{padding:0;margin-top:16px;font-size:1em;font-style:italic}.markdown-body
		dl dd{padding:0 16px;margin-bottom:16px}.markdown-body table{border-spacing:0;border-collapse:collapse;display:block;
		width:100%;overflow:auto}.markdown-body table td,.markdown-body table th{padding:6px 13px;border:1px solid #888888}
		.markdown-body table tr{background-color:#323232;border-top:1px solid #c6cbd1}.markdown-body table 
		tr:nth-child(2n){background-color:#4e4e4e}.markdown-body img{max-width:100%;box-sizing:content-box;background-color:#282C34}
		.markdown-body img[align=right]{padding-left:20px}.markdown-body img[align=left]{padding-right:20px}.markdown-body 
		code{padding:.2em .4em;margin:0;font-size:85%;background-color:#282828;border-radius:3px}.markdown-body 
		pre{word-wrap:normal}.markdown-body pre>code{padding:0;margin:0;font-size:100%;word-break:normal;white-space:pre;background:0 0;
		border:0}.markdown-body .highlight{margin-bottom:16px}.markdown-body .highlight pre{margin-bottom:0;word-break:normal}
		.markdown-body .highlight pre,.markdown-body pre{padding:16px;overflow:auto;font-size:85%;line-height:1.45;
		background-color:transparent;border-radius:3px}.markdown-body pre code{display:inline;max-width:auto;padding:0;margin:0;
		overflow:visible;line-height:inherit;word-wrap:normal;background-color:transparent;border:0}.markdown-body 
		.full-commit .btn-outline:not(:disabled):hover{color:#005cc5;border-color:#005cc5}.markdown-body kbd{display:inline-block;
		padding:3px 5px;font:11px SFMono-Regular,Consolas,\"Liberation Mono\",Menlo,Courier,monospace;line-height:10px;color:#444d56;
		vertical-align:middle;background-color:#fafbfc;border:1px solid #d1d5da;border-bottom-color:#c6cbd1;border-radius:3px;
		box-shadow:inset 0 -1px 0 #c6cbd1}.markdown-body :checked+.radio-label{position:relative;z-index:1;border-color:#0366d6}
		.markdown-body .task-list-item{list-style-type:none}.markdown-body .task-list-item+.task-list-item{margin-top:3px}
		.markdown-body .task-list-item input{margin:0 .2em .25em -1.6em;vertical-align:middle}.markdown-body hr{border-bottom-color:#eee}`

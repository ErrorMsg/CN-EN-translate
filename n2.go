package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"bufio"
)

const (
	urlcn = "http://dict.youdao.com/w/eng/%s"
	urlen = "http://dict.youdao.com/w/%s"
	urlexp = "http://dict.youdao.com/example/blng/eng/%s"
	logo = `
	 ______________________________
	|                              |
	|   DDDD   II   CCCC  TTTTTT   |
	|   DD  D  II  CC       TT     |
	|   DDDD   II   CCCC    TT     |
	\______________________________/	
	`
)

func translate(words []string, withExample, isMulti bool){
	var url string
	word2Trans := strings.Join(words, " ")
	isChinese := isChinese(word2Trans)
	if isChinese{
		url = fmt.Sprintf(urlcn, word2Trans)
	}else{
		url = fmt.Sprintf(urlen, word2Trans)
	}
	resp,_:= http.Get(url)
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil{
		log.Fatal(err)
		return
	}
	if isChinese{
		doc.Find("div.trans-container > ul > p").Each(func(i int, s *goquery.Selection){
			/* wordType := s.Children().Not(".contentTitle").Text()
			if wordType == ""{
				fmt.Printf("%s", wordType)
			}
			var result []string
			s.Find(".contentTitle > .search-js").Each(func(j int, ss *goquery.Selection){
				result = append(result, ss.Text())
			})
			fmt.Printf("%s\n", strings.Join(result, ";")) */
			s.Children().Each(func(j int, ss *goquery.Selection){
				if str:=ss.Not(".contentTitle").Text();str!=""{
					fmt.Printf("%s", str)
				}else if str:=ss.Has("a");str!=nil{
					fmt.Printf("/ %s\n", ss.Find("a").Text())
				}
			})
		})
	}else{
		if ck := checkHint(doc);ck!=nil{
			fmt.Printf("\r\n!word '%s' not found, do you want to search: ", word2Trans)
			fmt.Println()
			for _, guess := range ck{
				fmt.Println(guess[0] + guess[1] + " /" + guess[2])
			}
			fmt.Println()
			return
		}
		/* if !isMulti{
			fmt.Printf("\r\n %s", getPronounce(doc))
		} */
		//result := doc.Find("div#phrsListTab > div.trans-container > ul").Text()
		//fmt.Println(result)
		fmt.Println()
		if sh:=doc.Has("div.baav");sh.Text()==""{
			fmt.Printf("mmm, %s not found\n", words)
		}
		doc.Find("div#phrsListTab > div.trans-container > ul >li").Each(func(i int, s *goquery.Selection){
			fmt.Printf("%d/ %s\n", i+1, s.Text())
		})
		fmt.Println()
	}
	
	if withExample{
		sentences := getSentences(words, isChinese)
	/* if len(sentences) > 0{
		fmt.Println()
		for i,sentence := range sentences{
			fmt.Printf("%2d.%s\n", i+1, sentence[0])
			fmt.Printf("  %s\n", sentence[1])
		}
		fmt.Println()
	} */
		PrintSentence:
		for {
			if len(sentences) > 5{
				for i,sentence := range sentences{
					if i > 4{
						break
					}
					fmt.Printf("-%s\n %s\n", sentence[0], sentence[1])
				}
				fmt.Println("\nPress n to see more examples.")
				r := bufio.NewReader(os.Stdin)
				w,_ := r.ReadString('\n')
				if strings.TrimSpace(w) == "n"{
					fmt.Println(w)
					sentences = sentences[5:]
					continue
				}else{
					break PrintSentence
				}
			}else{
				for _,sentence := range sentences{
					fmt.Printf("--%s\n %s\n", sentence[0], sentence[1])
					break PrintSentence
				}
			}
		}
	}
}

func getPronounce(doc *goquery.Document) string{
	var pronounce string
	doc.Find("div.baav > span.pronounce").Each(func(i int, s *goquery.Selection){
		if i == 0{
			p := fmt.Sprintf("%s: %s ", s.Text(), s.Find("span.phonetic").Text())
			pronounce += p
		}
		if i == 1{
			p := fmt.Sprintf("%s: %s ", s.Text(), s.Find("span.phonetic").Text())
			pronounce += p
		}
	})
	return pronounce
}

func checkHint(doc *goquery.Document) [][]string{
	hints := doc.Find("p.typo-rel")
	if hints.Length()== 0{
		return nil
	}
	var result [][]string
	hints.Each(func(i int, s *goquery.Selection){
		word := strings.TrimSpace(s.Find("a").Text())
		s.Children().Remove()
		meaning := strings.TrimSpace(s.Text())
		num := fmt.Sprintf("%d/ ", i+1)
		result = append(result, []string{num, word, meaning})
	})
	return result
}

func getSentences(words []string, isChinese bool) [][]string{
	var sentences [][]string
	var doc *goquery.Document
	var err error
	url := fmt.Sprintf(urlexp, strings.Join(words, "_"))
	resp,_ := http.Get(url)
	defer resp.Body.Close()
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil{
		log.Fatal(err)
		return sentences
	}

	doc.Find("div#bilingual > ul > li").Each(func(i int, s *goquery.Selection){
		var sentence []string
		s.Children().Each(func(j int, ss *goquery.Selection){
			if j == 2{
				return
			}
			word := ""
			ss.Children().Each(func(k int, sss *goquery.Selection){
				if text := strings.TrimSpace(sss.Text());text!=""{
					addSpace := (j == 1 && isChinese) || (j == 0 && !isChinese) && k != 0&& text != "."
					if addSpace{
						text = " " + text
					}
					word += text
				}
			})
			sentence = append(sentence, word)
		})
		if len(sentence) == 2{
			sentences = append(sentences, sentence)
		}
	})
	return sentences
}

func displayUsage(){
	fmt.Println(logo)
	fmt.Println("Usage: ")
	fmt.Println("dict: <word(s)> to translate")
	fmt.Println("dict: <word(s)> -e to see examples")
	fmt.Println("dict: -q to exit")
	fmt.Println("**********************************")
}

func isChinese(str string) bool{
	for _,v := range str{
		if unicode.Is(unicode.Scripts["Han"], v){
			fmt.Println()
			return true
		}
	}
	return false
}


func parseArgs(args []string) ([]string, bool){
	withExample := false
	for idx, word := range args{
		if strings.HasPrefix(word, "-e") && len(word) == 2{
			withExample = true
			return args[1:idx], withExample
		}
	}
	return args[1:], withExample
}

/* func main(){
	if len(os.Args) == 1{
		displayUsage()
		os.Exit(1)
	}
	words, withExample := parseArgs(os.Args)
	translate(words, withExample, len(words) > 1)
} */

func main(){
	for {
		displayUsage()
		f := true
		r := bufio.NewReader(os.Stdin)
		s,_ := r.ReadString('\n')
		args := strings.Split(s, " ")
		args[len(args)-1] = strings.TrimSpace(args[len(args)-1])
		/* for i,v := range args{
			fmt.Printf("%d %s.", i, v)
		}
		if strings.TrimSpace(s) == "q"{
			os.Exit(1)
		} */
		for idx, w := range args{
			if strings.HasPrefix(w, "-q"){
				os.Exit(1)
			}
			if strings.HasPrefix(w, "-e") && len(w) >= 2{
				translate(args[:idx], true, len(args[:idx]) > 1)
				f = false
			}
		}
		if f{
			translate(args, false, len(args) > 1)
		}
	}
}
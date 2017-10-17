package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type selpgArgs struct {
	startPage  int    // start page of the article
	endPage    int    // end page of the article
	inFilename string // name of the file to be read
	pageLen    int    /* number of lines in one page, default value is 72,
	   can be overriden by "-lNumber" on command line */
	pageType rune /* type of the article, 'l' for lines-delimited, 'f' for form-feed-delimited
	   default is 'l'. */
	printDest string // destination of result pages
}

var progname string // name of program, used to display error message


func usage() {
	fmt.Fprintf(os.Stderr,
		"\nUSAGE: %s -s=start_page(number) -e=end_page(number) [ -f | -l=lines_per_page(number) ] [ -ddest ] [ in_filename ]\n", progname)
}


func processArgs(argNums int, args []string, saAddr *selpgArgs) {
	// check if the number of arguments is valid
	if argNums < 3 {
		fmt.Fprintf(os.Stderr, "%s: not enough arguments\n", progname)
		usage()
		os.Exit(1)
	}

	// handle 1st arg - start page
	tmpStr := []rune(args[1])
	if tmpStr[0] != '-' || tmpStr[1] != 's'||tmpStr[2] != '=' {
		fmt.Fprintf(os.Stderr, "%s: 1st arg should be -sstart_page\n", progname)
		usage()
		os.Exit(2)
	}
	page, err := strconv.Atoi(string(tmpStr[3:]))
	if page < 1 || err != nil {
		fmt.Fprintf(os.Stderr, "%s: invalid start page %s\n", progname, string(tmpStr[3:]))
		usage()
		os.Exit(3)
	}
	saAddr.startPage = page

	// handle 2nd arg -end page
	tmpStr = []rune(args[2])
	if tmpStr[0] != '-' || tmpStr[1] != 'e' || tmpStr[2] != '=' {
		fmt.Fprintf(os.Stderr, "%s: 2nd arg should be -eend_page\n", progname)
		usage()
		os.Exit(4)
	}
	page, err = strconv.Atoi(string(tmpStr[3:]))
	if page < 1 || page < saAddr.startPage || err != nil {
		fmt.Fprintf(os.Stderr, "%s: invalid start page %s\n", progname, string(tmpStr[3:]))
		usage()
		os.Exit(5)
	}
	saAddr.endPage = page

	// handle optional args
	argIndex := 3
	for argIndex < argNums && []rune(args[argIndex])[0] == '-' {
		tmpStr = []rune(args[argIndex])

		switch tmpStr[1] {
		case 'l':
      if tmpStr[2]!='=' {
        usage()
        os.Exit(13)
      }
			lineNum, err := strconv.Atoi(string(tmpStr[3:]))
			if lineNum < 1 || err != nil {
				fmt.Fprintf(os.Stderr, "%s: invalid page length %s\n", progname, string(tmpStr[2:]))
				usage()
				os.Exit(6)
			}
			saAddr.pageLen = lineNum
			argIndex++

		case 'f':
			if strings.Compare(string(tmpStr), "-f") != 0 {
				fmt.Fprintf(os.Stderr, "%s: option should be \"-f\"\n", progname)
				usage()
				os.Exit(7)
			}
			saAddr.pageType = 'f'
			argIndex++

		case 'd':
			if len(tmpStr[2:]) < 1 {
				fmt.Fprintf(os.Stderr, "%s: -d option requires a printer destination\n", progname)
				usage()
				os.Exit(8)
			}
			saAddr.printDest = string(tmpStr[2:])
			argIndex++

		default:
			fmt.Fprintf(os.Stderr, "%s: unknown option %s\n", progname, string(tmpStr))
			usage()
			os.Exit(9)
		} // end switch
	} // end while

	// there is one more argument
	if argIndex <= argNums-1 {
		saAddr.inFilename = args[argIndex]
		// check if file exists
		f, err := os.Open(saAddr.inFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: input file \"%s\" does not exist\n", progname, saAddr.inFilename)
			os.Exit(10)
		}
		f.Close()
	}
}


func processInput(saAddr *selpgArgs) {
	// set the input source
	fin := os.Stdin
	var err error
	if saAddr.inFilename != "" {
		fin, err = os.Open(saAddr.inFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: could not open input file \"%s\"\n", progname, saAddr.inFilename)
			os.Exit(11)
		}
	}

	// set the ouput destination
	fout := os.Stdout
	var cmd *exec.Cmd
	if saAddr.printDest != "" {
		tmpStr := fmt.Sprintf("./%s", saAddr.printDest)
		cmd = exec.Command("sh", "-c", tmpStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: could not open pipe to \"%s\"\n", progname, tmpStr)
			os.Exit(12)
		}
	}

	// dealing with the page type
	var line string
	pageCnt := 1
	inputReader := bufio.NewReader(fin)
	rst := ""
	if saAddr.pageType == 'l' {
		lineCnt := 0

		for true {
			line, err = inputReader.ReadString('\n')
			if err != nil { // error or EOF
				break
			}
			lineCnt++
			if lineCnt > saAddr.pageLen {
				pageCnt++
				lineCnt = 1
			}
			if pageCnt >= saAddr.startPage && pageCnt <= saAddr.endPage {
				if saAddr.printDest == "" {
					fmt.Fprintf(fout, "%s", line)
				} else {
					rst += line
				}
			}
		}
	} else {
		for true {
			c, _, erro := inputReader.ReadRune()
			if erro != nil { // error or EOF
				break
			}
			if c == '\f' {
				pageCnt++
			}
			if pageCnt >= saAddr.startPage && pageCnt <= saAddr.endPage {
				if saAddr.printDest == "" {
					fmt.Fprintf(fout, "%c", c)
				} else {
					rst += string(c)
				}
			}
		}
	}

	if saAddr.printDest != "" {
		cmd.Stdin = strings.NewReader(rst)
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			fmt.Println("print error!")
		}
	}

	if pageCnt < saAddr.startPage {
		fmt.Fprintf(os.Stderr, "%s: start_page (%d) greater than total pages (%d), no output written\n", progname, saAddr.startPage, pageCnt)
	} else {
		if pageCnt < saAddr.endPage {
			fmt.Fprintf(os.Stderr, "%s: end_page (%d) greater than total pages (%d), less output than expected\n", progname, saAddr.endPage, pageCnt)
		}
	}

	fin.Close()
	fout.Close()
	fmt.Fprintf(os.Stderr, "%s: done\n", progname)
}


func main() {
	sa := new(selpgArgs)
	progname = os.Args[0] // get the name of the program

	// initial selpg's arguments to default values
	sa.startPage = -1
	sa.endPage = -1
	sa.inFilename = ""
	sa.pageLen = 20
	sa.pageType = 'l'
	sa.printDest = ""

	processArgs(len(os.Args), os.Args, sa)
	processInput(sa)
}
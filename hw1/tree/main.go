package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
)

const pathSeparator = string(os.PathSeparator)

type objName []os.DirEntry

func (p objName) Len() int { return len(p) }

func (p objName) Less(i, j int) bool { return p[i].Name() < p[j].Name() }

func (p objName) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func calcTab(curTab string, isLast bool) (tab string, nextTab string) {
	var head string
	if isLast {
		head = "└───"
	} else {
		head = "├───"
	}
	if isLast {
		nextTab = curTab + "\t"
	} else {
		nextTab = curTab + "│\t"
	}

	return curTab + head, nextTab
}

func dirTreeTab(w *bufio.Writer, path string, printFiles bool, curtab string) error {

	var dirSlice []os.DirEntry
	files, err := os.ReadDir(path)
	//остановка рекурсии
	if len(files) == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	if !printFiles {
		dirSlice = make([]os.DirEntry, 0)
		for _, v := range files {
			if v.IsDir() {
				dirSlice = append(dirSlice, v)
			}
		}
		files = dirSlice
	}

	sort.Sort(objName(files))

	if err != nil {
		return err
	}

	for i, file := range files {
		isLast := i == (len(files) - 1)
		if file.IsDir() {
			newPath := path + pathSeparator + file.Name()
			tab, nextTab := calcTab(curtab, isLast)
			w.WriteString(tab + file.Name() + "\n")
			err = dirTreeTab(w, newPath, printFiles, nextTab) // проваливаемся внутрб директории
			if err != nil {
				log.Println(err.Error())
			}
		}
		if printFiles && !file.IsDir() {
			tab, _ := calcTab(curtab, isLast)
			fileInfo, err := file.Info()
			if err != nil {
				log.Println(err.Error())
			}
			if size := strconv.Itoa(int(fileInfo.Size())); size == "0" {
				w.WriteString(tab + file.Name() + " (empty)\n")
			} else {
				w.WriteString(tab + file.Name() + " (" + size + "b)\n")
			}
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	w := bufio.NewWriter(out)
	defer w.Flush()
	err := dirTreeTab(w, path, printFiles, "")
	return err
}

func main() {
	out := new(bytes.Buffer)
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	io.MultiWriter(os.Stdout)
	err := dirTree(out, path, printFiles)

	if err != nil {
		panic(err.Error())
	}
}

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
)

func Preprocess(input []byte, imagedir string) ([]byte, error) {
	buf := bytes.NewBuffer(input)
	var output bytes.Buffer
	out := bufio.NewWriter(&output)

	r := regexp.MustCompile("^`{3}\\s*(.*)$")

	var imgbuf bytes.Buffer
	imageCount := 0
	inBlock := false
	var format string

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		if r.MatchString(scanner.Text()) {
			if !inBlock {
				sub := r.FindStringSubmatch(scanner.Text())
				if len(sub) > 1 {
					inBlock = true
					format = sub[1]
					imageCount++
				}
			} else {
				inBlock = false
				name, err := writeImage(format, imageCount, imagedir, imgbuf)
				if err != nil {
					fmt.Fprintf(out, "```%s\n", format)
					fmt.Fprintf(out, "%s\n", string(imgbuf.Bytes()[:]))
					fmt.Fprintf(out, "```\n")
				} else {
					fmt.Fprintf(out, "[![](%s)](%s)\n", name, name)
				}
				imgbuf.Reset()
			}
		} else {
			if inBlock {
				imgbuf.WriteString(scanner.Text())
				imgbuf.WriteString("\n")
			} else {
				fmt.Fprintf(out, "%s\n", scanner.Text())
			}
		}
	}
	out.Flush()

	return output.Bytes(), nil
}

func writeImage(format string, index int, imagedir string, buf bytes.Buffer) (string, error) {
	os.MkdirAll(imagedir, os.ModePerm)
	filename := fmt.Sprintf("%s/image-%d.svg", imagedir, index)
	tempfile := ""
	var cmd *exec.Cmd
	switch format {
	case "ditaa":
		filename = fmt.Sprintf("%s/image-%d.png", imagedir, index)
		_, err := exec.LookPath("ditaa")
		if err != nil {
			log.Print("ditaa not found.")
			return filename, err
		}
		file, err := ioutil.TempFile("", "ditaa")
		if err != nil {
			log.Print(err)
			return filename, err
		}
		tempfile = file.Name()
		file.Close()
		cmd = exec.Command("ditaa", "-o", tempfile, filename)
	case "dot":
		_, err := exec.LookPath("dot")
		if err != nil {
			log.Print("dot not found.")
			return filename, err
		}
		cmd = exec.Command("dot", "-Tsvg", fmt.Sprintf("-o%s", filename))
	case "plantuml":
		_, err := exec.LookPath("plantuml")
		if err != nil {
			log.Print("plantuml not found.")
			return filename, err
		}
		cmd = exec.Command("plantuml", "-tsvg", "-p")
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Print(err)
		return filename, err
	}
	if "" == tempfile {
		stdin.Write(buf.Bytes())
	} else {
		ioutil.WriteFile(tempfile, buf.Bytes(), os.ModePerm)
		defer os.Remove(tempfile)
	}
	stdin.Close()

	var outbuf bytes.Buffer
	writer := bufio.NewWriter(&outbuf)
	cmd.Stdout = writer

	err = cmd.Start()
	if err != nil {
		log.Print(err)
		log.Print(cmd)
		return filename, err
	}
	cmd.Wait()

	if format == "plantuml" {
		ioutil.WriteFile(filename, outbuf.Bytes(), os.ModePerm)
	}
	return filename, err
}

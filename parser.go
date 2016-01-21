package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

func Parse(input []byte) ([]byte, error) {
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

				// setup name
				sub := r.FindStringSubmatch(scanner.Text())
				if len(sub) > 1 {
					inBlock = true
					format = sub[1]
					imageCount++
				}
			} else {
				inBlock = false
				name, _ := writeImage(format, imageCount, imgbuf)
				fmt.Fprintf(out, "![](%s)\n", name)
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

func writeImage(format string, index int, buf bytes.Buffer) (string, error) {
	filename := fmt.Sprintf("image-%d.png", index)
	var cmd *exec.Cmd
	switch format {
	case "ditaa":
		cmd = exec.Command("ditaa", "-T", "-o", "-", filename)
	case "dot":
		cmd = exec.Command("dot", "-Tpng", fmt.Sprintf("-o%s", filename))
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	stdin.Write(buf.Bytes())
	stdin.Close()
	return filename, err
}

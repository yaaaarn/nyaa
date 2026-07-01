package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/mitchellh/go-wordwrap"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/spf13/cobra"
)

const maxCols = 72

var (
	rootUrl string
	sukebei bool
)

var colorConfig = renderer.ColorizedConfig{
	Column: renderer.Tint{
		FG: renderer.Colors{color.FgHiBlack},
		Columns: []renderer.Tint{
			{},
			{FG: renderer.Colors{color.Reset}},
		},
	},
	Border:    renderer.Tint{FG: renderer.Colors{color.FgHiBlack}},
	Separator: renderer.Tint{FG: renderer.Colors{color.FgHiBlack}},
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func cleanANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

type imgRef struct {
	alt, img, link string
}
type fetchResult struct {
	data []byte
	err  error
}

var mdImgRegex = regexp.MustCompile(`(?:\[!\[([^\]]*)\]\((.*?)\)\]\((.*?)\)|!\[([^\]]*)\]\((.*?)\))`)
var markerRegex = regexp.MustCompile(`IMGMARKER(\d+)REKRAMGMI`)

func plainLowdown(description string) {
	cmd := exec.Command("lowdown", "-tterm", "--term-hpadding=0")
	cmd.Stdin = strings.NewReader(description)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(description)
		return
	}
	fmt.Print(string(out))
}

func parseImgRefs(matches [][]string) []imgRef {
	refs := make([]imgRef, len(matches))
	for i, m := range matches {
		if m[2] != "" {
			refs[i] = imgRef{alt: m[1], img: m[2], link: m[3]}
		} else {
			refs[i] = imgRef{alt: m[4], img: m[5]}
		}
	}
	return refs
}

func fetchImages(refs []imgRef) []fetchResult {
	results := make([]fetchResult, len(refs))
	var wg sync.WaitGroup

	for i, ref := range refs {
		url := strings.TrimSpace(ref.img)

		if url == "" {
			continue
		}

		wg.Add(1)

		go func(i int, url string) {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				results[i] = fetchResult{err: err}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results[i] = fetchResult{err: fmt.Errorf("bad status code: %s", resp.Status)}
				return
			}

			data, err := io.ReadAll(resp.Body)
			results[i] = fetchResult{data: data, err: err}
		}(i, url)
	}
	wg.Wait()
	return results
}

func renderMarkdown(description string) {
	rawMatches := mdImgRegex.FindAllStringSubmatch(description, -1)

	if len(rawMatches) == 0 {
		plainLowdown(description)
		return
	}

	refs := parseImgRefs(rawMatches)
	results := fetchImages(refs)

	i := 0
	modifiedDesc := mdImgRegex.ReplaceAllStringFunc(description, func(string) string {
		token := fmt.Sprintf("IMGMARKER%dREKRAMGMI", i)
		i++
		return token
	})

	cmd := exec.Command("lowdown", "-tterm", "--term-hpadding=0")
	cmd.Stdin = strings.NewReader(modifiedDesc)
	stdout, err := cmd.StdoutPipe()
	if err != nil || cmd.Start() != nil {
		fmt.Println(description)
		return
	}

	var out bytes.Buffer
	renderImage := func(idx int) {
		ref, res := refs[idx], results[idx]
		if res.err != nil {
			fmt.Fprintf(&out, "\n[Error with image %s: %v]\n", ref.img, res.err)
			return
		}

		if alt := strings.TrimSpace(ref.alt); alt != "" {
			color.New(color.FgHiBlack, color.Bold).Fprintf(&out, wordwrap.WrapString("[%s]\n", maxCols), alt)
		} else {
			color.New(color.FgHiBlack).Fprint(&out, "[untitled image]\n")
		}

		chafa := exec.Command("chafa", "-f", "symbols", "--symbols", "all", "-O", "9", fmt.Sprintf("--size=%d", maxCols))
		chafa.Stdin = bytes.NewReader(res.data)
		chafa.Stdout = &out

		var stderr bytes.Buffer
		chafa.Stderr = &stderr
		if err := chafa.Run(); err != nil {
			fmt.Fprintf(&out, "[Chafa Error: %v | Details: %s]\n", err, strings.TrimSpace(stderr.String()))
		}

		if ref.link != "" {
			color.New(color.FgGreen).Fprintf(&out, "%s\n\n", strings.TrimSpace(ref.link))
		}
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		clean := cleanANSI(line)
		locs := markerRegex.FindAllStringSubmatchIndex(clean, -1)

		if locs == nil {
			fmt.Fprintln(&out, line)
			continue
		}

		last := 0
		for _, loc := range locs {
			if seg := strings.TrimSpace(clean[last:loc[0]]); seg != "" {
				fmt.Fprintln(&out, clean[last:loc[0]])
			}

			idx, _ := strconv.Atoi(clean[loc[2]:loc[3]])
			renderImage(idx)
			last = loc[1]
		}

		if seg := strings.TrimSpace(clean[last:]); seg != "" {
			fmt.Fprintln(&out, clean[last:])
		}
	}

	cmd.Wait()
	os.Stdout.Write(out.Bytes())
}

var rootCmd = &cobra.Command{
	Use:   "nyaa",
	Short: "a simple nyaa.si client",

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if sukebei {
			rootUrl = "https://sukebei.nyaa.si"
		} else {
			rootUrl = "https://nyaa.si"
		}
	},
}

func Execute() {
	color.NoColor = false
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&sukebei, "sukebei", "x", false, "use sukebei (nsfw)")
}

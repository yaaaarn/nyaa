package cmd

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/mitchellh/go-wordwrap"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/spf13/cobra"
)

var (
	download bool
)

var idCmd = &cobra.Command{
	Use:   "id <id>",
	Short: "get/download torrent info from id",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		u, _ := url.JoinPath(rootUrl, "view", args[0])

		res, err := http.Get(u)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			panic(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			panic(err)
		}

		title := strings.TrimSpace(doc.Find(".panel-title").First().Text())

		metadata := map[string]string{}
		var metadataKeys []string

		doc.Find(".panel-body .row").Each(func(i int, row *goquery.Selection) {
			row.Find(".col-md-1, .col-md-offset-6").Each(func(j int, label *goquery.Selection) {
				key := strings.TrimSpace(strings.Replace(label.Text(), ":", "", 1))
				val := strings.TrimSpace(label.NextFiltered(".col-md-5").Text())
				if val != "" && key != "" {
					metadata[key] = val
					metadataKeys = append(metadataKeys, key)
				}
			})
		})

		/*magnetLink, exists := doc.Find(`.panel-footer a[href^="magnet:"]`).Attr("href")
		if !exists {
			magnetLink = ""
		}*/
		torrentLink, exists := doc.Find(`.panel-footer a[href^="/download/"]`).Attr("href")
		if !exists {
			torrentLink = ""
		}
		description := strings.ReplaceAll(strings.TrimSpace(doc.Find(`#torrent-description`).Text()), "&#10;", "")
		if description == "" {
			description = "no description"
		}
		/*
			if magnetLink != "" {
				if torrentPrinted {
					fmt.Println()
				}

				color.Cyan("Magnet link")
				fmt.Println(magnetLink)
			}
		*/
		bold := color.New(color.Bold)

		fmt.Println()
		color.HiBlack(metadata["Category"])
		bold.Println(wordwrap.WrapString(title, maxCols))

		fmt.Println()
		table := [][]string{}
		for _, key := range metadataKeys {
			if key == "Completed" || key == "Leechers" || key == "Seeders" || key == "Category" {
				continue
			}
			table = append(table, []string{key, metadata[key]})
		}
		t := tablewriter.NewTable(os.Stdout, tablewriter.WithRenderer(renderer.NewColorized(colorConfig)))
		t.Bulk(table[0:])
		t.Render()
		fmt.Println()

		fmt.Println(
			color.GreenString(`[↑] %s`, metadata["Seeders"]), color.RedString("[↓] %s", metadata["Leechers"]), color.CyanString(`[#] %s`, metadata["Completed"]),
		)

		if !download {
			fmt.Println()
			renderMarkdown(description)
		}
		fmt.Println()

		if torrentLink != "" {
			u, err := url.JoinPath(rootUrl, torrentLink)
			if download {
				resp, err := http.Get(u)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					panic(fmt.Sprintf("failed to download torrent: status %d", resp.StatusCode))
				}

				var filename string
				contentDisposition := resp.Header.Get("Content-Disposition")
				if contentDisposition != "" {
					_, params, _ := mime.ParseMediaType(contentDisposition)
					filename = params["filename"]
				}

				if filename == "" {
					parsedURL, err := url.Parse(u)
					if err == nil {
						filename = filepath.Base(parsedURL.Path)
					}

					if filename == "" || filename == "." || filename == "/" {
						reg := regexp.MustCompile(`[^a-zA-Z0-9\-_\. ]+`)
						safeTitle := reg.ReplaceAllString(title, "")
						filename = fmt.Sprintf("%s.torrent", strings.TrimSpace(safeTitle))
					}
				}

				out, err := os.Create(filename)
				if err != nil {
					panic(err)
				}
				defer out.Close()

				_, err = io.Copy(out, resp.Body)
				if err != nil {
					panic(err)
				}

				color.Green("Successfully downloaded: %s", filename)
			} else {
				if err == nil {
					color.Cyan("Torrent link")
					fmt.Println(u)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(idCmd)

	idCmd.Flags().BoolVarP(&download, "download", "d", false, "download the .torrent file")
}

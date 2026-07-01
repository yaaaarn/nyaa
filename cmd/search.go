package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
)

var (
	category string
	filter   int
	page     int
	sort     string
	order    string
	limit    int
	query string
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "search torrents",
	Args: cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		categories := getCategories()
		isValid := false
		for _, cat := range categories {
			if cat.id == category || cat.name == category {	
				category = cat.id
				isValid = true
				break
			}
		}
		if isValid == false {
			return fmt.Errorf("invalid category")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0	{
			query = args[0]
		} else {
			query = "[none]"
		}

		u, _ := url.Parse(rootUrl)

		q := url.Values{}
		q.Add("c", category)
		q.Add("f", strconv.Itoa(filter))
		q.Add("p", strconv.Itoa(page))
		if sort != "" {
			q.Add("s", sort)
		}
		if order != "" {
			q.Add("o", order)
		}
		if query != "[none]" {
			q.Add("q", query)
		}	

		u.RawQuery = q.Encode()

		res, err := http.Get(u.String())
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			fmt.Println(u.String())
			panic(fmt.Sprintf("status code error: %s", res.Status))
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			panic(err)
		}

		fmt.Println()
		color.HiBlack("search results for: %s", query)	

		doc.Find("table tbody tr").Slice(0, limit).Each(func(i int, s *goquery.Selection) {
			cols := s.Find("td")

			category, exists := cols.Eq(0).Find("img").Attr("alt")
			if !exists {
				category = ""
			}

			title := wordwrap.WrapString(strings.TrimSpace(cols.Eq(1).Find("a:not(.comments)").Last().Text()), maxCols)

			pageLink, exists := cols.Eq(1).Find(`a[href^="/view/"]`).Attr("href")
			if !exists {
				pageLink = ""
			} else {
				pageLink = strings.Split(pageLink, "#comments")[0]
			}

			size := strings.TrimSpace(cols.Eq(3).Text())
			date := strings.TrimSpace(cols.Eq(4).Text())
			seeders, err := strconv.Atoi(strings.TrimSpace(cols.Eq(5).Text()))
			if err != nil {
				seeders = -1
			}
			leechers, err := strconv.Atoi(strings.TrimSpace(cols.Eq(6).Text()))
			if err != nil {
				seeders = -1
			}
			downloads, err := strconv.Atoi(strings.TrimSpace(cols.Eq(7).Text()))
			if err != nil {
				seeders = -1
			}

			filledBackground := color.New(color.BgHiBlack, color.FgHiWhite)
			bold := color.New(color.Bold)

			fmt.Println()
			fmt.Println(filledBackground.Sprintf(" %s ", strings.Replace(pageLink, "/view/", "", 1)), color.HiBlackString(category));
			bold.Println(title)
      fmt.Println(
        color.GreenString(`[↑] %s`, strconv.Itoa(seeders)), color.RedString("[↓] %s", strconv.Itoa(leechers)), color.CyanString(`[#] %s`, strconv.Itoa(downloads)),
        color.HiBlackString("|"), color.HiBlackString("(≡) %s", size), color.HiBlackString(`[@] %s`, date),
      );
		})
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&category, "category", "c", "0_0", "show only in catgeory (e.g., anime, anime:raw)")
	searchCmd.Flags().IntVarP(&filter, "filter", "f", 0, "filter torrents (1=no_remakles, 2=trusted_only)")
	searchCmd.Flags().IntVarP(&page, "page", "p", 1, "page index")
	searchCmd.Flags().StringVarP(&sort, "sort", "s", "", "sort by (comments, size, date, seeders, leechers, downloads)")
	searchCmd.Flags().StringVarP(&order, "order", "o", "", "sort order (asc or desc)")
	searchCmd.Flags().IntVarP(&limit, "limit", "l", 10, "limit results")
}

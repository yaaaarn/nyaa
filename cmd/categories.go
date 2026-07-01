package cmd

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Category struct {
	id          string
	pretty_name string
	name        string
}

func getCategories() []Category {
	res, err := http.Get(rootUrl)
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

	var categories []Category

	doc.Find("select[name=c]").Children().Each(func(i int, s *goquery.Selection) {
		name, titleExists := s.Attr("title")
		id, valueExists := s.Attr("value")
		if titleExists && valueExists {
			categories = append(categories, Category{
				id:          id,
				pretty_name: name,
				name:        strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(name), " - ", ":"), " ", "_"),
			})
		}
	})

	return categories
}

var categoriesCmd = &cobra.Command{
	Use:   "cats",
	Short: "list all categories",
	Run: func(cmd *cobra.Command, args []string) {
		cats := getCategories()

		maxCat := slices.MaxFunc(cats, func(a, b Category) int {
			return cmp.Compare(len(a.name), len(b.name))
		})
		getMaxLengthName := len(maxCat.name)

		for _, cat := range cats {
			fmt.Printf("%-*s %s\n", getMaxLengthName, cat.name, color.HiBlackString(cat.pretty_name))
		}
	},
}

func init() {
	rootCmd.AddCommand(categoriesCmd)
}

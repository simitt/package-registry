package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/elastic/integrations-registry/util"
)

var categoryTitles = map[string]string{
	"logs":    "Logs",
	"metrics": "Metrics",
}

type Category struct {
	Id    string `yaml:"id" json:"id"`
	Title string `yaml:"title" json:"title"`
	Count int    `yaml:"count" json:"count"`
}

// categoriesHandler is a dynamic handler as it will also allow filtering in the future.
func categoriesHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		packagePaths, err := util.GetPackagePaths(packagesBasePath)
		if err != nil {
			notFound(w, err)
			return
		}

		packageList := map[string]*util.Package{}
		// Get unique list of newest packages
		for _, i := range packagePaths {
			p, err := util.NewPackage(packagesBasePath, i)
			if err != nil {
				return
			}

			// Check if the version exists and if it should be added or not.
			if pp, ok := packageList[p.Name]; ok {
				// If the package in the list is newer, do nothing. Otherwise delete and later add the new one.
				if pp.IsNewer(p) {
					continue
				}
			}
			packageList[p.Name] = p
		}

		categories := map[string]*Category{}

		for _, p := range packageList {
			for _, c := range p.Categories {
				if _, ok := categories[c]; !ok {
					categories[c] = &Category{
						Id:    c,
						Title: c,
						Count: 0,
					}
				}

				categories[c].Count = categories[c].Count + 1
			}
		}

		data, err := getCategoriesOutput(categories)
		if err != nil {
			notFound(w, err)
			return
		}
		jsonHeader(w)
		fmt.Fprint(w, string(data))
	}
}

func getCategoriesOutput(categories map[string]*Category) ([]byte, error) {
	var keys []string
	for k := range categories {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var outputCategories []*Category
	for _, k := range keys {
		c := categories[k]
		if title, ok := categoryTitles[c.Title]; ok {
			c.Title = title
		}
		outputCategories = append(outputCategories, c)
	}

	return json.MarshalIndent(outputCategories, "", "  ")
}

/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"os"
	"path"
	"text/template"

	"github.com/spf13/cobra"
	strcase "github.com/stoewer/go-strcase"
	"gosalusa.com/spice/util"
)

//go:embed model.go.tpl
var modelSrc string

// makeModelCmd represents the makeModel command
var makeModelCmd = &cobra.Command{
	Use:   "make:model [name]",
	Short: "",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := template.New("model").Parse(modelSrc)
		if err != nil {
			return err
		}
		rawName := args[0]
		goName := strcase.UpperCamelCase(rawName)
		fileName := "./" + strcase.KebabCase(rawName) + ".go"
		c, err := util.LoadConfig(".")
		if err != nil {
			return err
		}

		err = os.MkdirAll(c.Model.Dir, 0755)
		if err != nil {
			return err
		}

		f, err := os.OpenFile(path.Join(c.Model.Dir, fileName), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		return t.Execute(f, map[string]string{
			"Package": c.Model.Pkg,
			"Name":    goName,
		})
	},
}

func init() {
	rootCmd.AddCommand(makeModelCmd)
}

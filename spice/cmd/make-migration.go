/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"gosalusa.com/database/migrate"
	"gosalusa.com/spice/pkg"
	"gosalusa.com/spice/util"
)

// makeMigrationCmd represents the makeMigration command
var makeMigrationCmd = &cobra.Command{
	Use:   "make:migration [name]",
	Short: "",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := util.MigrationName(args)

		c, err := util.LoadConfig(".")
		if err != nil {
			return err
		}

		err = os.MkdirAll(c.Migration.Dir, 0755)
		if err != nil {
			return err
		}
		run := "schema.Run(func(ctx context.Context, tx database.DB) error {\n" +
			"return nil\n" +
			"})"
		src, err := migrate.SrcFile(name, c.Migration.Pkg, pkg.Raw(run), pkg.Raw(run))
		if err != nil {
			return err
		}

		return os.WriteFile(path.Join(c.Migration.Dir, name+".go"), []byte(src), 0644)
	},
}

func init() {
	rootCmd.AddCommand(makeMigrationCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// makeMigrationCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// makeMigrationCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

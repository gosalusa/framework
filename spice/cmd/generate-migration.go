/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
	"gosalusa.com/spice/util"
)

var srcMain = `package main

import (
	"errors"
	"log"
	"os"

	"gosalusa.com/database/migrate"
	migrations %#v
	models %#v
)

func main() {
	m := migrations.Use()

	src, err := m.GenerateMigration(%#v, %#v, &models.%s{})
	if errors.Is(err, migrate.ErrNoChanges) {
		return
	} else if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(%#v, []byte(src), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
`

var srcMigrations = `package %s

import (
	"gosalusa.com/database/migrate"
)

var migrations = migrate.New()

func Use() *migrate.Migrations {
	return migrations
}
`

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:    "generate:migration",
	Short:  "Run from go generate",
	Long:   ``,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// info, err := util.PkgInfo(".")
		// if err != nil {
		// 	return err
		// }

		// packageName := "migrations"
		// migrationsDir := path.Join(info.PackageRoot, packageName)

		c, err := util.LoadConfig(".")
		if err != nil {
			return err
		}

		err = os.MkdirAll(c.Migration.Dir, 0755)
		if err != nil {
			return err
		}
		err = os.WriteFile(path.Join(c.Migration.Dir, "migrations.go"), fmt.Appendf(nil, srcMigrations, c.Migration.Pkg), 0644)
		if err != nil {
			return err
		}

		// modelPackage := os.Getenv("GOPACKAGE")
		modelFile := os.Getenv("GOFILE")
		modelLineStr := os.Getenv("GOLINE")
		modelLine, err := strconv.Atoi(modelLineStr)
		if err != nil {
			return err
		}

		b, err := os.ReadFile(modelFile)
		if err != nil {
			return err
		}

		line := bytes.Split(b, []byte("\n"))[modelLine]

		matches := regexp.MustCompile(`type ([A-Z]\w+) struct`).FindSubmatch(line)
		if len(matches) < 2 {
			return fmt.Errorf("could not find model struct")
		}
		name := util.MigrationName([]string{string(matches[1])})
		migrationFile := path.Join(c.Migration.Dir, name+".go")
		src := fmt.Appendf(nil, srcMain,
			// imports
			c.Migration.Import,
			c.Model.Import,
			// GenerateMigration
			name, c.Migration.Pkg, matches[1],
			// write file
			migrationFile,
		)
		// fmt.Printf("%s\n", src)

		tmp := os.TempDir()
		outFile := path.Join(tmp, fmt.Sprintf("spice-generate-main-%s.go", name))

		err = os.WriteFile(outFile, src, 0644)
		if err != nil {
			return err
		}
		defer func() {
			err = os.Remove(outFile)
			if err != nil {
				fmt.Printf("failed to remove temp file %s\n", outFile)
			}
		}()
		err = run("go", "run", outFile)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

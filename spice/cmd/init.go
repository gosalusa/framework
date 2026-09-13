/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"gosalusa.com/static"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init <package>",
	Short: "Initialize a new salusa repo",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		module := args[0]

		parts := strings.SplitN(module, "/", 2)
		gitURL := fmt.Sprintf("ssh://git@%s:%s", parts[0], parts[1])

		err := copyDir(static.Content, "template", ".", args[0])
		if err != nil {
			return err
		}

		err = exec.Command("git", "init").Run()
		if err != nil {
			return err
		}

		err = exec.Command("git", "remote", "add", "origin", gitURL).Run()
		if err != nil {
			return err
		}

		err = exec.Command("go", "mod", "init", module).Run()
		if err != nil {
			return err
		}

		err = exec.Command("go", "mod", "tidy").Run()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func copyDir(root fs.FS, src, dist, pkgPath string) error {
	files, err := fs.ReadDir(root, src)
	if err != nil {
		return err
	}
	for _, f := range files {
		srcPath := path.Join(src, f.Name())
		distPath := path.Join(dist, f.Name())
		if f.IsDir() {
			err = os.MkdirAll(distPath, 0755)
			if err != nil {
				return err
			}
			err = copyDir(root, srcPath, distPath, pkgPath)
			if err != nil {
				return err
			}
		} else {
			b, err := fs.ReadFile(root, srcPath)
			if err != nil {
				return err
			}
			b = bytes.ReplaceAll(b, []byte("gosalusa.com/static/template"), []byte(pkgPath))

			err = os.WriteFile(distPath, b, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

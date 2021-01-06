/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/jpillora/longestcommon"
)

var cfgFile string
var outputDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dircopy",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(outputDir) == 0 {
			log.Fatal("output-dir cannot be empty")
		}
		err := dirCopy(outputDir, args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "output-dir", "", "output directory")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output-dir", "", "output directory")
	rootCmd.MarkFlagRequired("output-dir")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".dircopy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".dircopy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func dirCopy(output string, files []string) error {
	basePath := longestcommon.Prefix(files)

	// if single file remove file
	fi, err := os.Stat(basePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to check if base path exists: %+v %w", err.(*os.PathError), err)
		}
		basePath = filepath.Dir(basePath)
		fi, err = os.Stat(basePath)
		if err != nil {
			return fmt.Errorf("unable to check if base path exists: %+v %w", err.(*os.PathError), err)
		}
	}

	if !fi.IsDir() {
		basePath = filepath.Dir(basePath)
	}

	return copyFiles(output, basePath, files)
}

func copyFiles(output, basePath string, files []string) error {
	output = filepath.Clean(output)
	err := os.RemoveAll(output)
	if err != nil {
		return err
	}
	toCopy := make([]string, 0)
	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		//do not return skip for files in basePath
		if info.IsDir() {
			// check to see if any of the paths has the prefix
			found := false
			for _, f := range files {
				if strings.HasPrefix(f, path) {
					found = true
					break
				}
			}

			if !found {
				return filepath.SkipDir
			}
			toCopy = append(toCopy, path)
		}

		return nil
	}

	err = filepath.Walk(basePath, fn)
	sort.Strings(toCopy)
	for _, v := range toCopy {
		fmt.Println(v)
	}

	prefixPath := filepath.Dir(strings.TrimSuffix(basePath, "/"))
	fmt.Println("prefix", prefixPath)
	for _, fp := range toCopy {
		nfp := strings.Replace(fp, prefixPath, output, 1)
		fmt.Printf("creating (%s) for (%s)\n", nfp, nfp)
		err := os.MkdirAll(nfp, os.ModePerm)
		if err != nil {
			return err
		}
	}

	for _, f := range files {
		fmt.Println("copying files...")
		destPath := strings.Replace(f, prefixPath, output, 1)

		src, err := os.Open(f)
		if err != nil {
			return err
		}

		dst, err := os.Create(destPath)
		if err != nil {
			return err
		}

		_ = src
		_ = dst

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}

	return err
}

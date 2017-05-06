// Copyright © 2017 Anthony Spring <anthonyspring@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// VERSION The version of the application
var VERSION string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "annotate-influxdb",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Inside rootCmd PreRun with args: %v\n", args)
		rootCmdPreRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Inside rootCmd Run with args: %v\n", args)
		rootCmdRun()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// General Flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.annotate-influxdb.yaml)")
	
	// Logging Flags
	//RootCmd.PersistentFlags().String("loglevel", "info", "The level of logging. Acceptable values: debug, info, warn, error, fatal, panic.")
	//viper.BindPFlag("loglevel", RootCmd.PersistentFlags().Lookup("loglevel"))

	// InfluxDB Related Flags
	RootCmd.PersistentFlags().String("url", "http://localhost:8086", "The URL to the InfluxDB server.")
	viper.BindPFlag("influxdb.url", RootCmd.PersistentFlags().Lookup("url"))

	RootCmd.PersistentFlags().String("database", "", "The name of the database to write to.")
	viper.BindPFlag("influxdb.database", RootCmd.PersistentFlags().Lookup("database"))

	RootCmd.PersistentFlags().String("measurement", "events", "The name of the measurement to write to.")
	viper.BindPFlag("influxdb.measurement", RootCmd.PersistentFlags().Lookup("measurement"))

	RootCmd.PersistentFlags().String("title", "", "A title for the annotation.")
	viper.BindPFlag("influxdb.title", RootCmd.PersistentFlags().Lookup("title"))

	RootCmd.PersistentFlags().StringSlice("tag", []string{}, "A tag for the annotation.")
	viper.BindPFlag("influxdb.tags", RootCmd.PersistentFlags().Lookup("tag"))

	RootCmd.PersistentFlags().String("description", "", "A description for the annotation.")
	viper.BindPFlag("influxdb.description", RootCmd.PersistentFlags().Lookup("description"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".annotate-influxdb") // name of config file (without extension)
	viper.AddConfigPath("$HOME")              // adding home directory as first search path
	viper.AutomaticEnv()                      // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// GetInfluxDBClient
func GetInfluxDBClient(host string) (influxdb.Client, error) {
	// Parse the url to get the auth pieces
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	user, pass := "", ""
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}

	return influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     host,
		Username: user,
		Password: pass,
	})
}

// rootCmdPreRun
func rootCmdPreRun() {
	// Make sure the required arguments are provided
	if viper.IsSet("influxdb.url") && viper.GetString("influxdb.url") == "" {
		log.Fatal("--url is a required parameter.")
		os.Exit(-1)
	}
}

// rootCmdRun
func rootCmdRun() {
	// Get a client
	client, err := GetInfluxDBClient(viper.GetString("influxdb.url"))
	if err != nil {
		log.Fatal(err)
	}
	//defer client.Close()

	// Create a new point batch
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  viper.GetString("influxdb.database"),
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a point and add to batch
	tags := map[string]string{}
	fields := map[string]interface{}{
		"description":  viper.GetString("influxdb.description"),
		"tags": 				strings.Join(viper.GetStringSlice("influxdb.tags"), ","),
		"title":   			viper.GetString("influxdb.title"),
	}

	pt, err := influxdb.NewPoint(viper.GetString("influxdb.measurement"), tags, fields, time.Now().UTC())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := client.Write(bp); err != nil {
		log.Fatal(err)
	}
}

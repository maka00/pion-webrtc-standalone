/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"pion-webrtc/internal/application"
	"pion-webrtc/internal/datachannel"
	"pion-webrtc/internal/dto"
	"pion-webrtc/internal/gstreamer"
	"pion-webrtc/internal/signalling"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command.
var serverCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "server",
	Short: "starts a webrtc server",
	Long: `
A webrtc server serving one video stream. 
Use the PIPELINE environment variable to define the 
GStreamer pipeline. One element must be named sink 
in order to get the encoded frames.`,
	Run: func(_ *cobra.Command, _ []string) {
		pipeline := viper.GetString("PIPELINE")
		log.Println(pipeline)

		rootDir := "static"
		if viper.GetString("ROOTDIR") != "" {
			rootDir = viper.GetString("ROOTDIR")
		}

		srv := signalling.NewHTTPServer(rootDir)
		sigCli := signalling.NewHTTPSignallerClient(srv)

		sigCli.Init()

		cdata := make(chan string)
		cframes := make(chan dto.VideoFrame)
		prevManager := application.NewPreviewManager(sigCli, cdata, cframes)

		prevManager.Init()
		srv.Start()
		prevManager.Run()

		datachan := datachannel.NewDataPump(cdata)
		datachan.Start()

		vid := gstreamer.NewVideo(pipeline, cframes)
		vid.Run()

		select {}
	},
}

func init() { //nolint:gochecknoinits
	cobra.OnInitialize(initServerConfig)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	serverCmd.Flags().StringP("rootdir", "r", "static", "The directory the server code is in.")
	serverCmd.Flags().StringP("pipeline", "p", "", "The gstreamer pipeline to use")

	rootCmd.AddCommand(serverCmd)
}

// initServerConfig reads in config file and ENV variables if set.
func initServerConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".webrtcpeer-webrtc" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("pion-webrtc")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

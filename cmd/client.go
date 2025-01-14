package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"pion-webrtc/internal/application"
	"pion-webrtc/internal/dto"
	"pion-webrtc/internal/signalling"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command.
var clientCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:   "client",
	Short: "client to connect to the server",
	Long:  ``,
	Run: func(_ *cobra.Command, args []string) {
		fmt.Println("client called")
		server := viper.GetString("WEBRTC_SERVER_LOCATION")
		if server == "" {
			if len(args) < 1 {
				fmt.Println("missing server address (eg.: localhost:8080")
				os.Exit(1)
			}

			server = args[0]
		}
		fmt.Printf("calling %s\n", server)
		conn, err := getWebSocket(server)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)

		}
		defer func() { conn.Close() }()
		wsClient := signalling.NewWebrtcClient(conn)
		cframes := make(chan dto.VideoFrame)
		previewClient := application.NewPreviewClient(wsClient, cframes)
		previewClient.Init()
		select {}

	},
}

func getWebSocket(urlstring string) (*websocket.Conn, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}

	con, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating websocket: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return con, nil
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(clientCmd)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

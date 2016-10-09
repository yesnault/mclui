package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	marathon "github.com/gambol99/go-marathon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yesnault/mclui/version"
	"golang.org/x/crypto/ssh/terminal"
)

var clients []client

type client struct {
	marathon marathon.Marathon
}

var mainCmd = &cobra.Command{
	Use:   "mclui",
	Short: "Marathon Commande Line UI",
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetEnvPrefix("mclui")
		viper.AutomaticEnv()

		log.SetLevel(log.DebugLevel)
		initClients()
		runUI()
	},
}

func init() {
	mainCmd.AddCommand(version.Cmd)

	flags := mainCmd.Flags()

	flags.StringSlice("marathon-url", nil, "URLs Marathon")
	viper.BindPFlag("marathon_url", flags.Lookup("marathon-url"))

	flags.Bool("with-auth-basic", true, "Ask HTTP Basic Auth at startup")
	viper.BindPFlag("with_auth_basic", flags.Lookup("with-auth-basic"))
}

func main() {
	mainCmd.Execute()
}

func initClients() {

	for _, marathonURL := range viper.GetStringSlice("marathon_url") {

		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Enter HTTP Basic Auth User for url %s: ", marathonURL)
		username, _ := reader.ReadString('\n')

		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		password := string(bytePassword)

		config := marathon.NewDefaultConfig()
		config.URL = marathonURL
		config.HTTPBasicAuthUser = username
		config.HTTPBasicPassword = password
		config.HTTPClient = &http.Client{
			Timeout: (time.Duration(10) * time.Second),
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
				}).Dial,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
		}

		mclient, err := marathon.NewClient(config)
		if err != nil {
			log.Fatalf("Failed to create a client for marathon, error: %s", err)
		}

		clients = append(clients, client{marathon: mclient})
	}
}

func listApps() {

	for _, c := range clients {
		applications, err := c.marathon.Applications(nil)
		if err != nil {
			log.Fatalf("Failed to list applications %s", err)
		}

		for _, application := range applications.Apps {
			fmt.Printf("%s %d %d %d %d \n", application.ID, application.TasksHealthy, application.TasksRunning, application.TasksStaged, application.TasksUnhealthy)
		}
	}

}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	host           string
	port           string
	password       string
	channels       []string
	credentialFile string
	spreadsheetID  string
)

var ctx = context.Background()

// EventData struct to store event data
type EventData struct {
	Asset     string `json:"asset"`
	Position  int    `json:"position"`
	Timestamp int    `json:"timestamp"`
	Group     string `json:"group"`
	Text      string `json:"text"`
	Mode      string `json:"mode"`
	Name      string `json:"name"`
}

func main() {
	var rootCmd = &cobra.Command{Use: "redis2sheet"}
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Using redis as a message broker to update google sheets",
		Run: func(cmd *cobra.Command, args []string) {

			log.WithFields(log.Fields{
				"host":           host,
				"port":           port,
				"channels":       channels,
				"totalChannels":  len(channels),
				"credentialFile": credentialFile,
				"spreadsheetID":  spreadsheetID,
			}).Info("using parameters")

			if spreadsheetID == "" {
				spreadsheetID = "Spreadsheet not set"
			}

			data, err := os.ReadFile(credentialFile)
			if err != nil {
				log.Fatal(err)
			}

			creds, err := google.CredentialsFromJSON(ctx, data, "https://www.googleapis.com/auth/spreadsheets")
			if err != nil {
				log.Fatal(err)
			}

			sheetsService, err := sheets.NewService(ctx, option.WithCredentials(creds))
			if err != nil {
				log.Fatal(err)
			}

			client := redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%s", host, port),
				Password: password,
				DB:       0,
			})

			pubsub := client.Subscribe(channels...)
			_, err = pubsub.Receive()
			if err != nil {
				log.Fatal(err)
			}

			log.WithFields(log.Fields{
				"client":   client,
				"pubsub":   pubsub,
				"channels": channels,
			}).Info("connected redis server")

			event := EventData{}

			for {
				msg, err := pubsub.ReceiveMessage()
				if err != nil {
					log.Fatal(err)
				}

				log.WithFields(log.Fields{
					"payload": msg.Payload,
					"channel": msg.Channel,
					"pattern": msg.Pattern,
				}).Info("received message")

				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					log.WithFields(log.Fields{
						"payload": msg.Payload,
						"channel": msg.Channel,
						"pattern": msg.Pattern,
						"error":   err,
					}).Error("invalid message")
				}

				sheetName := fmt.Sprintf("%s-%s", event.Mode, event.Name)
				if err := ensureSheetExists(sheetsService, spreadsheetID, sheetName); err != nil {
					log.WithFields(log.Fields{
						"sheetName":     sheetName,
						"spreadsheetID": spreadsheetID,
						"error":         err,
					}).Error("error get or creating sheet")
				}

				writeDataToSheet(sheetsService, spreadsheetID, sheetName, msg.Channel, event)
			}

		},
	}

	cmd.Flags().StringVarP(&host, "host", "", "localhost", "Redis Host")
	cmd.Flags().StringVarP(&port, "port", "p", "6379", "Redis Port")
	cmd.Flags().StringVarP(&password, "password", "w", "", "Redis Password")
	cmd.Flags().StringSliceVarP(&channels, "channels", "c", []string{"events"}, "Redis Topic Name")
	cmd.Flags().StringVarP(&credentialFile, "credentialFile", "f", "credentials.json", "Google Credential File")
	cmd.Flags().StringVarP(&spreadsheetID, "spreadsheetID", "s", "", "Spreadsheet ID")

	rootCmd.AddCommand(cmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ensureSheetExists(srv *sheets.Service, spreadsheetID, sheetName string) error {
	// Get the spreadsheet details
	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return err
	}

	// Check if the sheet exists
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			// Sheet already exists
			return nil
		}
	}

	// If we reach here, the sheet does not exist, so we create it
	addSheetRequest := sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: sheetName,
			},
		},
	}

	batchUpdateRequest := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&addSheetRequest},
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, &batchUpdateRequest).Do()
	return err
}

func writeDataToSheet(srv *sheets.Service, spreadsheetID string, sheetName string, channel string, event EventData) {
	valueRange := sheets.ValueRange{
		MajorDimension: "ROWS",
		Values: [][]interface{}{
			{event.Asset, event.Position, event.Timestamp, event.Group, event.Text, event.Mode, event.Name},
		},
	}

	insertDataOption := "INSERT_ROWS"
	_, err := srv.Spreadsheets.Values.Append(spreadsheetID, sheetName, &valueRange).ValueInputOption("RAW").InsertDataOption(insertDataOption).Do()
	if err != nil {

		log.WithFields(log.Fields{
			"channel":   channel,
			"asset":     event.Asset,
			"position":  event.Position,
			"timestamp": event.Timestamp,
			"group":     event.Group,
			"mode":      event.Mode,
			"name":      event.Name,
			"error":     err,
		}).Error("error update sheet")

	} else {
		log.WithFields(log.Fields{
			"channel":   channel,
			"asset":     event.Asset,
			"position":  event.Position,
			"timestamp": event.Timestamp,
			"group":     event.Group,
			"mode":      event.Mode,
			"name":      event.Name,
		}).Debug("sheet updated")
	}
}
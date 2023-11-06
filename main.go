package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	url            string
	password       string
	channels       []string
	credentialFile string
	spreadsheetID  string
)

const (
	dateFormule = "=EPOCHTODATE(INDIRECT(\"C\" & ROW()))-TIME(3;0;0)"
)

var ctx = context.Background()

// EventData struct to store event data
type EventData struct {
	Asset     string `json:"asset"`
	Position  int    `json:"position"`
	Quantity  int    `json:"quantity"`
	Timestamp int    `json:"timestamp"`
	Group     string `json:"group"`
	Text      string `json:"text"`
	Mode      string `json:"mode"`
	Name      string `json:"name"`
}

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	var rootCmd = &cobra.Command{Use: "redis2sheet"}
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Using redis as a message broker to update google sheets",
		Run: func(cmd *cobra.Command, args []string) {

			logger.Info(
				"using parameters",
				"url", url,
				"channels", channels,
				"totalChannels", len(channels),
				"credentialFile", credentialFile,
				"spreadsheetID", spreadsheetID,
			)

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

			logger.Info("connected google sheets", "service", sheetsService)
			_, err = sheetsService.Spreadsheets.Get(spreadsheetID).Do()
			if err != nil {
				log.Fatal(fmt.Sprintf("error getting spreadsheet %s", spreadsheetID), err)
			}

			opt, _ := redis.ParseURL(url)
			client := redis.NewClient(opt)

			pubsub := client.Subscribe(channels...)
			r, err := pubsub.Receive()
			if err != nil {
				log.Fatal(err)
			}

			logger.Info(
				"connected redis server",
				"client", client,
				"pubsub", pubsub,
				"channels", channels,
				"received", r,
			)

			event := EventData{}

			for {
				msg, err := pubsub.ReceiveMessage()
				if err != nil {
					log.Fatal(err)
				}

				logger.Info(
					"received message",
					"payload", msg.Payload,
					"channel", msg.Channel,
					"pattern", msg.Pattern,
				)

				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					logger.Error(
						"invalid message",
						"payload", msg.Payload,
						"channel", msg.Channel,
						"pattern", msg.Pattern,
						"error", err,
					)
				}

				sheetName := fmt.Sprintf("%s-%s", event.Mode, event.Name)
				logger.Info(
					"checking exists sheet",
					"sheetName", sheetName,
					"spreadsheetID", spreadsheetID,
				)

				if err := ensureSheetExists(sheetsService, spreadsheetID, sheetName); err != nil {
					logger.Error(
						"error get or creating sheet",
						"sheetName", sheetName,
						"spreadsheetID", spreadsheetID,
						"error", err,
					)
				}

				resp, err := writeDataToSheet(sheetsService, spreadsheetID, sheetName, msg.Channel, event)
				if err != nil {
					logger.Error(
						"error update sheet",
						"channel", msg.Channel,
						"asset", event.Asset,
						"position", event.Position,
						"quantity", event.Quantity,
						"timestamp", event.Timestamp,
						"group", event.Group,
						"mode", event.Mode,
						"name", event.Name,
						"error", err,
					)
				} else {
					logger.Info(
						"sheet updated",
						"channel", msg.Channel,
						"asset", event.Asset,
						"position", event.Position,
						"quantity", event.Quantity,
						"timestamp", event.Timestamp,
						"group", event.Group,
						"mode", event.Mode,
						"name", event.Name,
						"resp", resp,
					)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&url, "url", "", "redis://localhost:6379", "Redis URL")
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

func writeDataToSheet(srv *sheets.Service, spreadsheetID string, sheetName string, channel string, event EventData) (*sheets.AppendValuesResponse, error) {
	valueRange := sheets.ValueRange{
		MajorDimension: "ROWS",
		Values: [][]interface{}{
			{event.Asset, event.Position, event.Quantity, event.Timestamp, event.Group, event.Text, event.Mode, event.Name, dateFormule},
		},
	}

	insertDataOption := "INSERT_ROWS"
	res, err := srv.Spreadsheets.Values.Append(spreadsheetID, sheetName, &valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption(insertDataOption).
		Do()
	return res, err
}

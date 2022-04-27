package main

import (
	"encoding/json"
	"fmt"

	"github.com/inconshreveable/log15"
	"github.com/urfave/cli/v2"

	"github.com/nsec/askgod/api"
	"github.com/nsec/askgod/internal/utils"
)

func (c *client) cmdAdminMonitorLog(ctx *cli.Context) error {
	// Parse the arguments
	logLvl, err := log15.LvlFromString(ctx.String("loglevel"))
	if err != nil {
		return err
	}

	// Connection handler
	conn, err := c.websocket("/events?type=logging")
	if err != nil {
		return err
	}

	// Process the messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		event := api.Event{}
		err = json.Unmarshal(data, &event)
		if err != nil {
			continue
		}

		if event.Type != "logging" {
			continue
		}

		logEntry := api.EventLogging{}
		err = json.Unmarshal(event.Metadata, &logEntry)
		if err != nil {
			continue
		}

		lvl, err := log15.LvlFromString(logEntry.Level)
		if err != nil {
			continue
		}

		if lvl > logLvl {
			continue
		}

		ctx := []interface{}{}
		for k, v := range logEntry.Context {
			ctx = append(ctx, k)
			ctx = append(ctx, v)
		}

		record := log15.Record{
			Time: event.Timestamp,
			Lvl:  lvl,
			Msg:  logEntry.Message,
			Ctx:  ctx,
		}

		format := log15.TerminalFormat()
		fmt.Printf("[%s] %s", event.Server, format.Format(&record))
	}

	return nil
}

func (c *client) cmdAdminMonitorFlags(ctx *cli.Context) error {
	// Connection handler
	conn, err := c.websocket("/events?type=flags")
	if err != nil {
		return err
	}

	const layout = "2006/01/02 15:04"
	// Process the messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		event := api.Event{}
		err = json.Unmarshal(data, &event)
		if err != nil {
			continue
		}

		if event.Type != "flags" {
			continue
		}

		score := api.EventFlag{}
		err = json.Unmarshal(event.Metadata, &score)
		if err != nil {
			continue
		}

		team := fmt.Sprintf("id=%d", score.Team.ID)
		if score.Team.Tags["infra"] != "" {
			team = score.Team.Tags["infra"]
		}

		if score.Type == "valid" {
			fmt.Printf("[%s][%s] Team \"%s\" (%s) scored %d points with \"%s\" (id=%d) (%s)\n",
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, team, score.Value, score.Input, score.Flag.ID, utils.PackTags(score.Flag.Tags))
		} else if score.Type == "duplicate" {
			fmt.Printf("[%s][%s] Team \"%s\" (%s) re-submitted \"%s\" (id=%d) (%s)\n",
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, team, score.Input, score.Flag.ID, utils.PackTags(score.Flag.Tags))
		} else if score.Type == "invalid" {
			fmt.Printf("[%s][%s] Team \"%s\" (%s) submitted invalid flag \"%s\"\n",
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, team, score.Input)
		}
	}

	return nil
}

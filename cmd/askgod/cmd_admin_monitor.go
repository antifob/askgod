package main

import (
	"encoding/json"
	"fmt"

	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/urfave/cli.v1"

	"github.com/nsec/askgod/api"
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

		if score.Result == "valid" {
			fmt.Printf("[%s][%s] Team \"%s\" (%d) scored %d points with \"%s\" (%d)\n",
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, score.Team.ID, score.Value, score.Input, score.Flag.ID)
		} else if score.Result == "duplicate" {
			fmt.Printf("[%s][%s] Team \"%s\" (%d) re-submitted \"%s\" (%d)\n",
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, score.Team.ID, score.Input, score.Flag.ID)
		} else if score.Result == "invalid" {
			fmt.Printf("[%s][%s] Team \"%s\" (%d) submitted invalid flag \"%s\"\n",
				event.Server, event.Timestamp.Local().Format(layout), score.Team.Name, score.Team.ID, score.Input)
		}
	}

	return nil
}
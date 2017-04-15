package rest

import (
	"database/sql"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/nsec/askgod/api"
)

func (r *rest) getTimeline(writer http.ResponseWriter, request *http.Request, logger log15.Logger) {
	var team *api.AdminTeam

	if r.config.Scoring.HideOthers {
		if !r.hasAccess("team", request) {
			r.errorResponse(403, "Scoreboard is hidden", writer, request)
			return
		}

		if !r.hasAccess("admin", request) {
			// Extract the client IP
			ip, err := r.getIP(request)
			if err != nil {
				logger.Error("Failed to get the client's IP", log15.Ctx{"error": err})
				r.errorResponse(500, "Internal Server Error", writer, request)
				return
			}

			// Look for a matching team
			team, err = r.db.GetTeamForIP(*ip)
			if err == sql.ErrNoRows {
				logger.Warn("No team found for IP", log15.Ctx{"ip": ip.String()})
				r.errorResponse(404, "No team found for IP", writer, request)
				return
			} else if err != nil {
				logger.Error("Failed to get the team", log15.Ctx{"error": err})
				r.errorResponse(500, "Internal Server Error", writer, request)
				return
			}
		}
	}

	timeline, err := r.db.GetTimeline(team)
	if err != nil {
		logger.Error("Failed to get the timeline", log15.Ctx{"error": err})
		r.errorResponse(500, "Internal Server Error", writer, request)
		return
	}

	r.jsonResponse(timeline, writer, request)
}

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost/server/public/plugin"
)

// handleHTTP routes HTTP requests
func (p *Plugin) handleHTTP(w plugin.ResponseWriter, r *plugin.Request) {
	switch r.URL.Path {
	case "/api/groups":
		p.handleGetGroups(w, r)
	case "/api/groups/autocomplete":
		p.handleAutocomplete(w, r)
	default:
		http.NotFound(w, &http.Request{Method: r.Method, URL: r.URL})
	}
}

// handleGetGroups returns groups for a team
func (p *Plugin) handleGetGroups(w plugin.ResponseWriter, r *plugin.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id parameter required", http.StatusBadRequest)
		return
	}

	groups, err := p.listGroups(teamID)
	if err != nil {
		p.logError("Failed to list groups", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Filter groups based on visibility
	visibleGroups := make([]*Group, 0)
	for _, group := range groups {
		if p.canViewGroup(userID, group) {
			visibleGroups = append(visibleGroups, group)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visibleGroups)
}

// handleAutocomplete returns group suggestions for autocomplete
func (p *Plugin) handleAutocomplete(w plugin.ResponseWriter, r *plugin.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id parameter required", http.StatusBadRequest)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	limit := 20

	groups, err := p.listGroups(teamID)
	if err != nil {
		p.logError("Failed to list groups", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Filter and match groups
	type AutocompleteResult struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		MemberCount int    `json:"member_count"`
	}

	results := make([]AutocompleteResult, 0)
	for _, group := range groups {
		// Check visibility permissions
		if !p.canViewGroup(userID, group) {
			continue
		}

		// Match query
		if query == "" || strings.HasPrefix(strings.ToLower(group.Name), query) {
			description := ""
			if group.Visibility == "private" {
				description = "🔒 Private"
			} else {
				description = "🔓 Public"
			}

			results = append(results, AutocompleteResult{
				Name:        group.Name,
				DisplayName: "@" + group.Name,
				Description: description + " - " + string(rune(len(group.Members))) + " members",
				Visibility:  group.Visibility,
				MemberCount: len(group.Members),
			})

			if len(results) >= limit {
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// writeJSON writes a JSON response
func writeJSON(w plugin.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// readJSON reads a JSON request body
func readJSON(r *plugin.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

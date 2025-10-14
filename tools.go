package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// GetToolDefinitions returns the list of tools available to the voice assistant
func GetToolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"type":        "function",
			"name":        "find_branches",
			"description": "Find medical clinic branches by city or district",
			"strict":      true,
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"city": map[string]any{
						"type":        "string",
						"description": "City name (e.g., Tel Aviv, Haifa)",
					},
					"district": map[string]any{
						"type":        "string",
						"description": "District name",
					},
				},
				"additionalProperties": false,
			},
		},
		// TODO: Add more tool definitions here
		// - find_doctors
		// - list_free_slots
		// - create_or_get_patient
		// - book_appointment
		// - cancel_appointment
		// - list_patient_appointments
		// - reschedule_appointment
	}
}

func ExecuteVoiceFunction(name, argsJSON string, debug bool) any {
	if debug {
		log.Printf("=== FUNCTION CALL: %s ===", name)
		log.Printf("Arguments: %s", argsJSON)
	}

	switch name {
	case "find_branches":
		return executeFindBranches(argsJSON)
		// TODO: Add more function handlers
	default:
		return map[string]any{
			"error": "unknown function",
			"name":  name,
		}
	}
}

func executeFindBranches(argsJSON string) any {
	var args struct {
		City     string `json:"city"`
		District string `json:"district"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	// TODO: Implement actual database query
	//example  response
	return map[string]any{
		"branches": []map[string]any{
			{
				"id":      1,
				"name":    "סניף תל אביב מרכז",
				"city":    "Tel Aviv",
				"address": "Dizengoff 123",
			},
		},
	}
}

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

	query := `
        SELECT branch_id, name, district, city, address_line, phone
        FROM branches
        WHERE active = true
    `
	var queryArgs []any
	argIndex := 1

	if args.City != "" {
		query += fmt.Sprintf(" AND city ILIKE $%d", argIndex)
		queryArgs = append(queryArgs, "%"+args.City+"%")
		argIndex++
	}

	if args.District != "" {
		query += fmt.Sprintf(" AND district ILIKE $%d", argIndex)
		queryArgs = append(queryArgs, "%"+args.District+"%")
		argIndex++
	}

	query += " ORDER BY city, name"

	rows, err := DB.Query(query, queryArgs...)
	if err != nil {
		log.Printf("Query error: %v", err)
		return map[string]any{"error": fmt.Sprintf("database query failed: %v", err)}
	}
	defer rows.Close()

	var branches []map[string]any
	for rows.Next() {
		var branchID int
		var name, district, city, addressLine, phone string

		if err := rows.Scan(&branchID, &name, &district, &city, &addressLine, &phone); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		branches = append(branches, map[string]any{
			"id":       branchID,
			"name":     name,
			"district": district,
			"city":     city,
			"address":  addressLine,
			"phone":    phone,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return map[string]any{"error": fmt.Sprintf("error reading results: %v", err)}
	}

	if len(branches) == 0 {
		return map[string]any{
			"branches": []map[string]any{},
			"message":  "לא נמצאו סניפים תואמים",
		}
	}

	return map[string]any{
		"branches": branches,
		"count":    len(branches),
	}
}

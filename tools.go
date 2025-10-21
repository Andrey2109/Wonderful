package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func getIsraelTime() time.Time {
	loc, err := time.LoadLocation("Asia/Jerusalem")
	if err != nil {
		log.Printf("Failed to load Israel timezone: %v", err)
		return time.Now()
	}
	return time.Now().In(loc)
}

func GetToolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"type":        "function",
			"name":        "find_branches",
			"description": "Find medical clinic branches by city or district. City names must be in Hebrew",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"city": map[string]any{
						"type":        "string",
						"description": "City name in Hebrew (e.g., תל אביב, חיפה, ירושלים)",
					},
					"district": map[string]any{
						"type":        "string",
						"description": "District name in Hebrew",
					},
				},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "find_doctors",
			"description": "Find doctors by specialty, branch, district, city, or name",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"specialty": map[string]any{
						"type":        "string",
						"description": "Medical specialty (e.g., רופא שיניים, קרדיולוג, דרמטולוג)",
					},
					"branch_id": map[string]any{
						"type":        "integer",
						"description": "Branch ID to filter doctors",
					},
					"doctor_name": map[string]any{
						"type":        "string",
						"description": "Doctor's first or last name",
					},
					"district": map[string]any{
						"type":        "string",
						"description": "District name in Hebrew (e.g. מחוז תל אביב, מחוז חיפה, מחוז הדרום)",
					},
					"city": map[string]any{
						"type":        "string",
						"description": "City name in Hebrew (e.g., תל אביב, חיפה, ירושלים)",
					},
				},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "list_free_slots",
			"description": "List available appointment slots for a specific doctor at a branch on a given date",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"doctor_branch_id": map[string]any{
						"type":        "integer",
						"description": "Doctor-branch association ID",
					},
					"date": map[string]any{
						"type":        "string",
						"description": "Date in YYYY-MM-DD format",
					},
				},
				"required":             []string{"doctor_branch_id", "date"},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "create_or_get_patient",
			"description": "Create a new patient or get existing patient by national ID",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"first_name": map[string]any{
						"type":        "string",
						"description": "Patient's first name",
					},
					"last_name": map[string]any{
						"type":        "string",
						"description": "Patient's last name",
					},
					"national_id": map[string]any{
						"type":        "string",
						"description": "Patient's national ID number (תעודת זהות)",
					},
					"phone": map[string]any{
						"type":        "string",
						"description": "Patient's phone number (optional for contact)",
					},
				},
				"required":             []string{"first_name", "last_name", "national_id"},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "book_appointment",
			"description": "Book a new appointment",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"doctor_branch_id": map[string]any{
						"type":        "integer",
						"description": "Doctor-branch association ID",
					},
					"doctor_id": map[string]any{
						"type":        "integer",
						"description": "Doctor ID",
					},
					"branch_id": map[string]any{
						"type":        "integer",
						"description": "Branch ID",
					},
					"patient_id": map[string]any{
						"type":        "integer",
						"description": "Patient ID",
					},
					"start_at_utc": map[string]any{
						"type":        "string",
						"description": "Appointment start time in ISO 8601 UTC format",
					},
					"duration_minutes": map[string]any{
						"type":        "integer",
						"description": "Appointment duration in minutes",
					},
				},
				"required":             []string{"doctor_branch_id", "doctor_id", "branch_id", "patient_id", "start_at_utc", "duration_minutes"},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "cancel_appointment",
			"description": "Cancel an existing appointment",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"appointment_id": map[string]any{
						"type":        "string",
						"description": "Appointment UUID",
					},
					"reason": map[string]any{
						"type":        "string",
						"description": "Cancellation reason",
					},
				},
				"required":             []string{"appointment_id"},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "list_patient_appointments",
			"description": "List appointments for a patient",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_national_id": map[string]any{
						"type":        "string",
						"description": "Patient's national ID number (תעודת זהות)",
					},
					"from_date": map[string]any{
						"type":        "string",
						"description": "Start date (YYYY-MM-DD)",
					},
					"to_date": map[string]any{
						"type":        "string",
						"description": "End date (YYYY-MM-DD)",
					},
				},
				"required":             []string{"patient_national_id"},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "find_patient_appointment_for_cancel",
			"description": "Find patient appointments that can be cancelled",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_national_id": map[string]any{
						"type":        "string",
						"description": "Patient's national ID number (תעודת זהות)",
					},
					"date": map[string]any{
						"type":        "string",
						"description": "Appointment date to search for (YYYY-MM-DD)",
					},
				},
				"required":             []string{"patient_national_id"},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "reschedule_appointment",
			"description": "Reschedule an existing appointment",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"appointment_id": map[string]any{
						"type":        "string",
						"description": "Appointment UUID",
					},
					"new_start_at_utc": map[string]any{
						"type":        "string",
						"description": "New appointment start time in ISO 8601 UTC format",
					},
				},
				"required":             []string{"appointment_id", "new_start_at_utc"},
				"additionalProperties": false,
			},
		},
		{
			"type":        "function",
			"name":        "end_call",
			"description": "End the phone call when conversation is complete",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"reason": map[string]any{
						"type":        "string",
						"description": "Reason for ending the call",
					},
				},
			},
		},
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
	case "find_doctors":
		return executeFindDoctors(argsJSON)
	case "list_free_slots":
		return executeListFreeSlots(argsJSON)
	case "create_or_get_patient":
		return executeCreateOrGetPatient(argsJSON)
	case "book_appointment":
		return executeBookAppointment(argsJSON)
	case "cancel_appointment":
		return executeCancelAppointment(argsJSON)
	case "list_patient_appointments":
		return executeListPatientAppointments(argsJSON)
	case "find_patient_appointment_for_cancel":
		return executeFindPatientAppointmentForCancel(argsJSON)
	case "reschedule_appointment":
		return executeRescheduleAppointment(argsJSON)
	case "end_call":
		return map[string]any{
			"status":  "call_ended",
			"message": "להתראות!",
		}
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
		return map[string]any{"error": fmt.Sprintf("database query failed: %v", err)}
	}
	defer rows.Close()

	var branches []map[string]any
	for rows.Next() {
		var branchID int
		var name, district, city, addressLine, phone string

		if err := rows.Scan(&branchID, &name, &district, &city, &addressLine, &phone); err != nil {
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

func executeFindDoctors(argsJSON string) any {
	var args struct {
		Specialty  string `json:"specialty"`
		BranchID   int    `json:"branch_id"`
		DoctorName string `json:"doctor_name"`
		District   string `json:"district"`
		City       string `json:"city"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}
	if args.Specialty != "" {
		args.Specialty = normalizeSpecialty(args.Specialty)
	}

	log.Printf("DEBUG: Searching doctors with params: specialty=%s, branch_id=%d, city=%s, district=%s, doctor_name=%s",
		args.Specialty, args.BranchID, args.City, args.District, args.DoctorName)

	query := `
        SELECT DISTINCT 
            d.doctor_id,
            d.first_name,
            d.last_name,
            d.phone,
            s.name as specialty_name,
            db.doctor_branch_id,
            db.branch_id,
            b.name as branch_name,
            b.city as branch_city,
            b.district as branch_district
        FROM doctors d
        JOIN specialties s ON d.specialty_id = s.specialty_id
        JOIN doctor_branches db ON d.doctor_id = db.doctor_id
        JOIN branches b ON db.branch_id = b.branch_id
        WHERE d.active = true AND db.active = true AND b.active = true
    `

	var queryArgs []any
	argIndex := 1

	if args.Specialty != "" {
		query += fmt.Sprintf(" AND s.name ILIKE $%d", argIndex)
		queryArgs = append(queryArgs, "%"+args.Specialty+"%")
		argIndex++
	}

	if args.BranchID > 0 {
		query += fmt.Sprintf(" AND db.branch_id = $%d", argIndex)
		queryArgs = append(queryArgs, args.BranchID)
		argIndex++
	}

	if args.DoctorName != "" {
		// Search in first name, last name, or full name (both ways)
		query += fmt.Sprintf(" AND (d.first_name ILIKE $%d OR d.last_name ILIKE $%d OR (d.first_name || ' ' || d.last_name) ILIKE $%d OR (d.last_name || ' ' || d.first_name) ILIKE $%d)",
			argIndex, argIndex+1, argIndex+2, argIndex+3)
		searchPattern := "%" + args.DoctorName + "%"
		queryArgs = append(queryArgs, searchPattern, searchPattern, searchPattern, searchPattern)
		argIndex += 4
	}

	if args.District != "" {
		query += fmt.Sprintf(" AND b.district ILIKE $%d", argIndex)
		queryArgs = append(queryArgs, "%"+args.District+"%")
		argIndex++
	}

	if args.City != "" {
		query += fmt.Sprintf(" AND (b.city ILIKE $%d OR similarity(b.city, $%d) > 0.3)", argIndex, argIndex+1)
		queryArgs = append(queryArgs, "%"+args.City+"%", args.City)
		argIndex += 2
	}

	query += " ORDER BY d.last_name, d.first_name, b.name"

	log.Printf("DEBUG: Final query: %s", query)
	log.Printf("DEBUG: Query args: %v", queryArgs)

	rows, err := DB.Query(query, queryArgs...)
	if err != nil {
		log.Printf("DEBUG: Query error: %v", err)
		return map[string]any{"error": fmt.Sprintf("database query failed: %v", err)}
	}
	defer rows.Close()

	var doctors []map[string]any
	for rows.Next() {
		var doctorID, doctorBranchID, branchID int
		var firstName, lastName, specialty, branchName, branchCity, branchDistrict string
		var phone sql.NullString

		if err := rows.Scan(&doctorID, &firstName, &lastName, &phone, &specialty,
			&doctorBranchID, &branchID, &branchName, &branchCity, &branchDistrict); err != nil {
			log.Printf("DEBUG: Scan error: %v", err)
			continue
		}

		doctors = append(doctors, map[string]any{
			"doctor_id":        doctorID,
			"doctor_branch_id": doctorBranchID,
			"name":             fmt.Sprintf("ד\"ר %s %s", firstName, lastName),
			"first_name":       firstName,
			"last_name":        lastName,
			"specialty":        specialty,
			"branch_id":        branchID,
			"branch_name":      branchName,
			"branch_city":      branchCity,
			"branch_district":  branchDistrict,
		})
	}

	log.Printf("DEBUG: Found %d doctors", len(doctors))

	if len(doctors) == 0 {
		return map[string]any{
			"doctors": []map[string]any{},
			"message": "לא נמצאו רופאים תואמים",
		}
	}

	return map[string]any{
		"doctors": doctors,
		"count":   len(doctors),
	}
}
func normalizeSpecialty(specialty string) string {
	replacements := map[string]string{
		"אורטופד":    "אורתופד",
		"פיזיותרפיה": "פיזיותרפיסט",
		"גניקולוג":   "גינקולוג",
	}

	if normalized, ok := replacements[specialty]; ok {
		return normalized
	}
	return specialty
}

func executeListFreeSlots(argsJSON string) any {
	var args struct {
		DoctorBranchID int    `json:"doctor_branch_id"`
		Date           string `json:"date"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	targetDate, err := time.Parse("2006-01-02", args.Date)
	if err != nil {
		return map[string]any{"error": "invalid date format"}
	}

	var timezone string
	err = DB.QueryRow(`
        SELECT b.timezone 
        FROM doctor_branches db
        JOIN branches b ON db.branch_id = b.branch_id
        WHERE db.doctor_branch_id = $1
    `, args.DoctorBranchID).Scan(&timezone)

	if err != nil {
		return map[string]any{"error": "branch not found"}
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc, _ = time.LoadLocation("Asia/Jerusalem")
	}

	// Fix: Don't convert Sunday, keep PostgreSQL's 0-6 convention
	dow := int(targetDate.Weekday())

	var startTimeLocal, endTimeLocal time.Time
	var slotMinutes int

	err = DB.QueryRow(`
        SELECT start_time_local, end_time_local, slot_minutes
        FROM doctor_hours
        WHERE doctor_branch_id = $1 AND dow = $2
    `, args.DoctorBranchID, dow).Scan(&startTimeLocal, &endTimeLocal, &slotMinutes)

	if err != nil {
		return map[string]any{
			"slots":   []map[string]any{},
			"message": "הרופא לא עובד ביום זה",
		}
	}

	startDateTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		startTimeLocal.Hour(), startTimeLocal.Minute(), 0, 0, loc)
	endDateTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		endTimeLocal.Hour(), endTimeLocal.Minute(), 0, 0, loc)

	rows, err := DB.Query(`
        SELECT start_at, end_at
        FROM appointments
        WHERE doctor_branch_id = $1
        AND DATE(start_at AT TIME ZONE $2) = $3
        AND status IN ('scheduled', 'pending', 'rescheduled')
    `, args.DoctorBranchID, timezone, args.Date)

	if err != nil {
		return map[string]any{"error": "failed to check appointments"}
	}
	defer rows.Close()

	occupied := make(map[time.Time]bool)
	for rows.Next() {
		var start, end time.Time
		if err := rows.Scan(&start, &end); err != nil {
			continue
		}

		for t := start; t.Before(end); t = t.Add(time.Duration(slotMinutes) * time.Minute) {
			occupied[t.In(loc)] = true
		}
	}

	var slots []map[string]any
	now := getIsraelTime()

	for t := startDateTime; t.Before(endDateTime); t = t.Add(time.Duration(slotMinutes) * time.Minute) {

		if t.Before(now) {
			continue
		}

		if occupied[t] {
			continue
		}

		slots = append(slots, map[string]any{
			"time_local": t.Format("15:04"),
			"time_utc":   t.UTC().Format(time.RFC3339),
			"duration":   slotMinutes,
		})
	}

	return map[string]any{
		"slots":             slots,
		"count":             len(slots),
		"date":              args.Date,
		"slot_duration_min": slotMinutes,
	}
}

func executeCreateOrGetPatient(argsJSON string) any {
	var args struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		NationalID string `json:"national_id"`
		Phone      string `json:"phone"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	// Convert Hebrew number words to digits
	nationalID := convertHebrewNumbersToDigits(args.NationalID)

	// Remove all non-digit characters
	nationalID = strings.ReplaceAll(nationalID, "-", "")
	nationalID = strings.ReplaceAll(nationalID, " ", "")
	nationalID = strings.ReplaceAll(nationalID, ".", "")

	// Pad with leading zeros if needed (Israeli IDs should be 9 digits)
	if len(nationalID) < 9 && len(nationalID) > 0 {
		nationalID = fmt.Sprintf("%09s", nationalID)
	}

	if len(nationalID) != 9 {
		return map[string]any{
			"error":   "invalid_id",
			"message": fmt.Sprintf("תעודת זהות חייבת להכיל 9 ספרות. קיבלתי: %s (%d ספרות)", args.NationalID, len(nationalID)),
		}
	}

	var patientID int64
	var firstName, lastName, existingPhone sql.NullString

	err := DB.QueryRow(`
        SELECT patient_id, first_name, last_name, phone
        FROM patients
        WHERE national_id = $1
    `, nationalID).Scan(&patientID, &firstName, &lastName, &existingPhone)

	if err == sql.ErrNoRows {
		// Create new patient
		phone := ""
		if args.Phone != "" {
			phone = strings.ReplaceAll(args.Phone, "-", "")
			phone = strings.ReplaceAll(phone, " ", "")
		}

		err = DB.QueryRow(`
            INSERT INTO patients (first_name, last_name, national_id, phone)
            VALUES ($1, $2, $3, NULLIF($4, ''))
            ON CONFLICT (national_id) DO NOTHING
            RETURNING patient_id
        `, args.FirstName, args.LastName, nationalID, phone).Scan(&patientID)

		if err != nil {
			// Check if it was a conflict (patient was created by another request)
			err = DB.QueryRow(`
                SELECT patient_id, first_name, last_name
                FROM patients
                WHERE national_id = $1
            `, nationalID).Scan(&patientID, &firstName, &lastName)

			if err != nil {
				return map[string]any{"error": fmt.Sprintf("failed to create patient: %v", err)}
			}

			return map[string]any{
				"patient_id": patientID,
				"created":    false,
				"first_name": firstName.String,
				"last_name":  lastName.String,
				"message":    fmt.Sprintf("נמצא מטופל קיים: %s %s", firstName.String, lastName.String),
			}
		}

		return map[string]any{
			"patient_id": patientID,
			"created":    true,
			"message":    fmt.Sprintf("נוצר מטופל חדש: %s %s", args.FirstName, args.LastName),
		}
	}

	if err != nil {
		return map[string]any{"error": fmt.Sprintf("database error: %v", err)}
	}

	if args.Phone != "" && (!existingPhone.Valid || existingPhone.String != args.Phone) {
		phone := strings.ReplaceAll(args.Phone, "-", "")
		phone = strings.ReplaceAll(phone, " ", "")

		_, _ = DB.Exec(`
            UPDATE patients 
            SET phone = $1 
            WHERE patient_id = $2
        `, phone, patientID)
	}

	return map[string]any{
		"patient_id": patientID,
		"created":    false,
		"first_name": firstName.String,
		"last_name":  lastName.String,
		"message":    fmt.Sprintf("נמצא מטופל קיים: %s %s", firstName.String, lastName.String),
	}
}

func convertHebrewNumbersToDigits(input string) string {
	hebrewToDigit := map[string]string{
		"אפס": "0",
		"אחת": "1", "אחד": "1",
		"שתיים": "2", "שניים": "2", "שני": "2",
		"שלוש": "3", "שלושה": "3",
		"ארבע": "4", "ארבעה": "4",
		"חמש": "5", "חמישה": "5",
		"שש": "6", "שישה": "6",
		"שבע": "7", "שבעה": "7",
		"שמונה": "8",
		"תשע":   "9", "תשעה": "9",
	}

	result := input
	for hebrew, digit := range hebrewToDigit {
		result = strings.ReplaceAll(result, hebrew, digit)
	}

	return result
}

func executeBookAppointment(argsJSON string) any {
	var args struct {
		DoctorBranchID  int    `json:"doctor_branch_id"`
		DoctorID        int    `json:"doctor_id"`
		BranchID        int    `json:"branch_id"`
		PatientID       int    `json:"patient_id"`
		StartAtUTC      string `json:"start_at_utc"`
		DurationMinutes int    `json:"duration_minutes"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	startTime, err := time.Parse(time.RFC3339, args.StartAtUTC)
	if err != nil {
		return map[string]any{"error": "invalid time format"}
	}

	endTime := startTime.Add(time.Duration(args.DurationMinutes) * time.Minute)

	log.Printf("DEBUG: Booking appointment - doctor_branch_id=%d, doctor_id=%d, branch_id=%d, patient_id=%d, start=%s, end=%s",
		args.DoctorBranchID, args.DoctorID, args.BranchID, args.PatientID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	var appointmentID string
	err = DB.QueryRow(`
        INSERT INTO appointments (
            doctor_branch_id, doctor_id, branch_id, patient_id,
            start_at, end_at, created_by
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7
        )
        RETURNING appointment_id
    `, args.DoctorBranchID, args.DoctorID, args.BranchID, args.PatientID,
		startTime, endTime, "voice_assistant").Scan(&appointmentID)

	if err != nil {
		log.Printf("DEBUG: Failed to insert appointment: %v", err)
		if strings.Contains(err.Error(), "overlaps") || strings.Contains(err.Error(), "conflict") {
			return map[string]any{
				"error":   "time_conflict",
				"message": "התור תפוס. אנא בחר זמן אחר",
			}
		}
		return map[string]any{"error": fmt.Sprintf("failed to book: %v", err)}
	}

	log.Printf("DEBUG: Successfully created appointment_id=%s", appointmentID)

	var doctorFirst, doctorLast, branchName, branchCity string
	err = DB.QueryRow(`
        SELECT d.first_name, d.last_name, b.name, b.city
        FROM appointments a
        JOIN doctors d ON a.doctor_id = d.doctor_id
        JOIN branches b ON a.branch_id = b.branch_id
        WHERE a.appointment_id = $1
    `, appointmentID).Scan(&doctorFirst, &doctorLast, &branchName, &branchCity)

	if err != nil {
		return map[string]any{
			"appointment_id": appointmentID,
			"status":         "booked",
		}
	}

	// Get local time for display
	var timezone string
	DB.QueryRow("SELECT timezone FROM branches WHERE branch_id = $1", args.BranchID).Scan(&timezone)
	loc, _ := time.LoadLocation(timezone)
	localTime := startTime.In(loc)

	return map[string]any{
		"appointment_id": appointmentID,
		"status":         "booked",
		"doctor":         fmt.Sprintf("ד\"ר %s %s", doctorFirst, doctorLast),
		"branch":         fmt.Sprintf("%s, %s", branchName, branchCity),
		"date":           localTime.Format("02/01/2006"),
		"time":           localTime.Format("15:04"),
		"day_name":       getHebrewDayName(localTime.Weekday()),
		"message":        "התור נקבע בהצלחה",
	}
}

func executeCancelAppointment(argsJSON string) any {
	var args struct {
		AppointmentID string `json:"appointment_id"`
		Reason        string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	if args.Reason == "" {
		args.Reason = "ביטול לבקשת המטופל"
	}

	var startAt time.Time
	var doctorFirst, doctorLast, branchName string
	err := DB.QueryRow(`
        SELECT a.start_at, d.first_name, d.last_name, b.name
        FROM appointments a
        JOIN doctors d ON a.doctor_id = d.doctor_id
        JOIN branches b ON a.branch_id = b.branch_id
        WHERE a.appointment_id = $1 AND a.status = 'scheduled'
    `, args.AppointmentID).Scan(&startAt, &doctorFirst, &doctorLast, &branchName)

	if err == sql.ErrNoRows {
		return map[string]any{
			"error":   "not_found",
			"message": "התור לא נמצא או כבר בוטל",
		}
	}

	if err != nil {
		return map[string]any{"error": fmt.Sprintf("database error: %v", err)}
	}

	_, err = DB.Exec(`
        UPDATE appointments
        SET status = 'cancelled',
            cancellation_reason = $2,
            cancelled_at = $3
        WHERE appointment_id = $1
    `, args.AppointmentID, args.Reason, getIsraelTime())

	if err != nil {
		return map[string]any{"error": fmt.Sprintf("failed to cancel: %v", err)}
	}

	return map[string]any{
		"status":  "cancelled",
		"doctor":  fmt.Sprintf("ד\"ר %s %s", doctorFirst, doctorLast),
		"branch":  branchName,
		"date":    startAt.Format("02/01/2006"),
		"time":    startAt.Format("15:04"),
		"message": "התור בוטל בהצלחה",
	}
}

func executeListPatientAppointments(argsJSON string) any {
	var args struct {
		PatientNationalID string `json:"patient_national_id"`
		FromDate          string `json:"from_date"`
		ToDate            string `json:"to_date"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	nationalID := strings.ReplaceAll(args.PatientNationalID, "-", "")
	nationalID = strings.ReplaceAll(nationalID, " ", "")

	var patientID int64
	err := DB.QueryRow("SELECT patient_id FROM patients WHERE national_id = $1", nationalID).Scan(&patientID)
	if err != nil {
		return map[string]any{
			"error":   "patient_not_found",
			"message": "לא נמצא מטופל עם תעודת זהות זו",
		}
	}

	query := `
        SELECT 
            a.appointment_id,
            a.start_at,
            a.end_at,
            a.status,
            d.first_name,
            d.last_name,
            s.name as specialty,
            b.name as branch_name,
            b.city as branch_city,
            b.timezone
        FROM appointments a
        JOIN doctors d ON a.doctor_id = d.doctor_id
        JOIN specialties s ON d.specialty_id = s.specialty_id
        JOIN branches b ON a.branch_id = b.branch_id
        WHERE a.patient_id = $1
        AND a.status IN ('scheduled', 'pending', 'rescheduled')
    `

	var queryArgs []any = []any{patientID}
	argIndex := 2

	if args.FromDate != "" {
		query += fmt.Sprintf(" AND DATE(a.start_at) >= $%d", argIndex)
		queryArgs = append(queryArgs, args.FromDate)
		argIndex++
	} else {

		query += fmt.Sprintf(" AND a.start_at >= $%d", argIndex)
		queryArgs = append(queryArgs, getIsraelTime())
		argIndex++
	}

	if args.ToDate != "" {
		query += fmt.Sprintf(" AND DATE(a.start_at) <= $%d", argIndex)
		queryArgs = append(queryArgs, args.ToDate)
	}

	query += " ORDER BY a.start_at"

	rows, err := DB.Query(query, queryArgs...)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("query failed: %v", err)}
	}
	defer rows.Close()

	var appointments []map[string]any
	for rows.Next() {
		var appointmentID, status, doctorFirst, doctorLast, specialty, branchName, branchCity, timezone string
		var startAt, endAt time.Time

		if err := rows.Scan(&appointmentID, &startAt, &endAt, &status,
			&doctorFirst, &doctorLast, &specialty, &branchName, &branchCity, &timezone); err != nil {
			continue
		}

		loc, _ := time.LoadLocation(timezone)
		localStart := startAt.In(loc)

		appointments = append(appointments, map[string]any{
			"appointment_id": appointmentID,
			"doctor":         fmt.Sprintf("ד\"ר %s %s", doctorFirst, doctorLast),
			"specialty":      specialty,
			"branch":         fmt.Sprintf("%s, %s", branchName, branchCity),
			"date":           localStart.Format("02/01/2006"),
			"time":           localStart.Format("15:04"),
			"day_name":       getHebrewDayName(localStart.Weekday()),
			"status":         status,
		})
	}

	if len(appointments) == 0 {
		return map[string]any{
			"appointments": []map[string]any{},
			"message":      "לא נמצאו תורים",
		}
	}

	return map[string]any{
		"appointments": appointments,
		"count":        len(appointments),
	}
}

func executeFindPatientAppointmentForCancel(argsJSON string) any {
	var args struct {
		PatientNationalID string `json:"patient_national_id"`
		Date              string `json:"date"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	nationalID := strings.ReplaceAll(args.PatientNationalID, "-", "")
	nationalID = strings.ReplaceAll(nationalID, " ", "")

	var patientID int64
	err := DB.QueryRow("SELECT patient_id FROM patients WHERE national_id = $1", nationalID).Scan(&patientID)
	if err != nil {
		return map[string]any{
			"error":   "patient_not_found",
			"message": "לא נמצא מטופל עם תעודת זהות זו",
		}
	}

	query := `
        SELECT 
            a.appointment_id,
            a.start_at,
            d.first_name,
            d.last_name,
            s.name as specialty,
            b.name as branch_name,
            b.timezone
        FROM appointments a
        JOIN doctors d ON a.doctor_id = d.doctor_id
        JOIN specialties s ON d.specialty_id = s.specialty_id
        JOIN branches b ON a.branch_id = b.branch_id
        WHERE a.patient_id = $1
        AND a.status IN ('scheduled', 'pending', 'rescheduled')
    `

	var queryArgs []any = []any{patientID}

	if args.Date != "" {
		query += " AND DATE(a.start_at) = $2"
		queryArgs = append(queryArgs, args.Date)
	} else {
		query += " AND a.start_at >= $2"
		queryArgs = append(queryArgs, getIsraelTime())
	}

	query += " ORDER BY a.start_at LIMIT 5"

	rows, err := DB.Query(query, queryArgs...)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("query failed: %v", err)}
	}
	defer rows.Close()

	var appointments []map[string]any
	for rows.Next() {
		var appointmentID, doctorFirst, doctorLast, specialty, branchName, timezone string
		var startAt time.Time

		if err := rows.Scan(&appointmentID, &startAt, &doctorFirst, &doctorLast,
			&specialty, &branchName, &timezone); err != nil {
			continue
		}

		loc, _ := time.LoadLocation(timezone)
		localStart := startAt.In(loc)

		appointments = append(appointments, map[string]any{
			"appointment_id": appointmentID,
			"doctor":         fmt.Sprintf("ד\"ר %s %s", doctorFirst, doctorLast),
			"specialty":      specialty,
			"branch":         branchName,
			"date":           localStart.Format("02/01/2006"),
			"time":           localStart.Format("15:04"),
			"day_name":       getHebrewDayName(localStart.Weekday()),
		})
	}

	if len(appointments) == 0 {
		return map[string]any{
			"appointments": []map[string]any{},
			"message":      "לא נמצאו תורים לביטול",
		}
	}

	return map[string]any{
		"appointments": appointments,
		"count":        len(appointments),
		"message":      "נמצאו התורים הבאים. איזה תור לבטל?",
	}
}

func executeRescheduleAppointment(argsJSON string) any {
	var args struct {
		AppointmentID string `json:"appointment_id"`
		NewStartAtUTC string `json:"new_start_at_utc"`
	}

	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{"error": fmt.Sprintf("invalid arguments: %v", err)}
	}

	newStartTime, err := time.Parse(time.RFC3339, args.NewStartAtUTC)
	if err != nil {
		return map[string]any{"error": "invalid time format"}
	}

	var doctorBranchID, doctorID, branchID, patientID int
	var currentStart, currentEnd time.Time
	var slotMinutes int

	err = DB.QueryRow(`
        SELECT a.doctor_branch_id, a.doctor_id, a.branch_id, a.patient_id,
               a.start_at, a.end_at,
               EXTRACT(EPOCH FROM (a.end_at - a.start_at))/60 as duration_min
        FROM appointments a
        WHERE a.appointment_id = $1 AND a.status = 'scheduled'
    `, args.AppointmentID).Scan(&doctorBranchID, &doctorID, &branchID, &patientID,
		&currentStart, &currentEnd, &slotMinutes)

	if err != nil {
		return map[string]any{
			"error":   "not_found",
			"message": "התור לא נמצא או כבר בוטל",
		}
	}

	newEndTime := newStartTime.Add(time.Duration(slotMinutes) * time.Minute)

	_, err = DB.Exec(`
        UPDATE appointments
        SET start_at = $2,
            end_at = $3,
            slot = tstzrange($2, $3, '[)'),
            status = 'rescheduled'
        WHERE appointment_id = $1
    `, args.AppointmentID, newStartTime, newEndTime)

	if err != nil {
		if strings.Contains(err.Error(), "overlaps") || strings.Contains(err.Error(), "conflict") {
			alternatives := getAlternativeSlots(doctorBranchID, newStartTime, slotMinutes)
			return map[string]any{
				"error":        "time_conflict",
				"message":      "הזמן המבוקש תפוס",
				"alternatives": alternatives,
			}
		}
		return map[string]any{"error": fmt.Sprintf("failed to reschedule: %v", err)}
	}

	var doctorFirst, doctorLast, branchName string
	var timezone string
	DB.QueryRow(`
        SELECT d.first_name, d.last_name, b.name, b.timezone
        FROM appointments a
        JOIN doctors d ON a.doctor_id = d.doctor_id
        JOIN branches b ON a.branch_id = b.branch_id
        WHERE a.appointment_id = $1
    `, args.AppointmentID).Scan(&doctorFirst, &doctorLast, &branchName, &timezone)

	loc, _ := time.LoadLocation(timezone)
	localTime := newStartTime.In(loc)

	return map[string]any{
		"status":   "rescheduled",
		"doctor":   fmt.Sprintf("ד\"ר %s %s", doctorFirst, doctorLast),
		"branch":   branchName,
		"new_date": localTime.Format("02/01/2006"),
		"new_time": localTime.Format("15:04"),
		"new_day":  getHebrewDayName(localTime.Weekday()),
		"message":  "התור עודכן בהצלחה",
	}
}

func getAlternativeSlots(doctorBranchID int, requestedTime time.Time, duration int) []map[string]any {
	var alternatives []map[string]any

	for offset := 20; offset <= 60 && len(alternatives) < 3; offset += 20 {
		checkTime := requestedTime.Add(time.Duration(offset) * time.Minute)

		var exists bool
		err := DB.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM appointments
                WHERE doctor_branch_id = $1
                AND status IN ('scheduled', 'pending', 'rescheduled')
                AND tstzrange($2, $3, '[)') && slot
            )
        `, doctorBranchID, checkTime, checkTime.Add(time.Duration(duration)*time.Minute)).Scan(&exists)

		if err == nil && !exists {
			alternatives = append(alternatives, map[string]any{
				"time_utc":   checkTime.Format(time.RFC3339),
				"time_local": checkTime.Format("15:04"),
			})
		}
	}

	return alternatives
}

func getHebrewDayName(day time.Weekday) string {
	days := map[time.Weekday]string{
		time.Sunday:    "יום ראשון",
		time.Monday:    "יום שני",
		time.Tuesday:   "יום שלישי",
		time.Wednesday: "יום רביעי",
		time.Thursday:  "יום חמישי",
		time.Friday:    "יום שישי",
		time.Saturday:  "שבת",
	}
	return days[day]
}

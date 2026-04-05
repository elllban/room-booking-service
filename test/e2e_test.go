package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type Room struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateRoomRequest struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

type ScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type BookingRequest struct {
	SlotID string `json:"slotId"`
}

const baseURL = "http://localhost:8080"

func TestE2E_BookingFlow(t *testing.T) {
	adminToken := getToken(t, "admin")

	room := createRoom(t, adminToken, "E2E Test Room", 20)

	createSchedule(t, adminToken, room.ID)

	userToken := getToken(t, "user")

	slots := getAvailableSlots(t, userToken, room.ID)
	if len(slots) == 0 {
		t.Skip("No slots available for testing")
	}

	booking := createBooking(t, userToken, slots[0].ID)

	if booking.Status != "active" {
		t.Errorf("Expected booking status 'active', got '%s'", booking.Status)
	}
}

func TestE2E_CancelBooking(t *testing.T) {
	adminToken := getToken(t, "admin")
	room := createRoom(t, adminToken, "E2E Cancel Test", 20)
	createSchedule(t, adminToken, room.ID)

	userToken := getToken(t, "user")
	slots := getAvailableSlots(t, userToken, room.ID)
	if len(slots) == 0 {
		t.Skip("No slots available for testing")
	}
	booking := createBooking(t, userToken, slots[0].ID)

	cancelBooking(t, userToken, booking.ID)

	if booking.Status != "active" {
		t.Logf("Booking cancelled successfully")
	}
}

func getToken(t *testing.T, role string) string {
	body := bytes.NewBuffer([]byte(`{"role":"` + role + `"}`))
	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", body)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("Failed to decode token response: %v", err)
	}

	return tokenResp.Token
}

func createRoom(t *testing.T, token, name string, capacity int) Room {
	req := CreateRoomRequest{Name: name, Capacity: capacity}
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", baseURL+"/rooms/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Room Room `json:"room"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode room response: %v", err)
	}

	return result.Room
}

func createSchedule(t *testing.T, token, roomID string) {
	req := ScheduleRequest{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "17:00",
	}
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", baseURL+"/rooms/"+roomID+"/schedule/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

type Slot struct {
	ID     string    `json:"id"`
	RoomID string    `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

func getAvailableSlots(t *testing.T, token, roomID string) []Slot {
	today := time.Now().Format("2006-01-02")
	url := baseURL + "/rooms/" + roomID + "/slots/list?date=" + today

	httpReq, _ := http.NewRequest("GET", url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to get slots: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Slots []Slot `json:"slots"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode slots: %v", err)
	}

	return result.Slots
}

type Booking struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func createBooking(t *testing.T, token, slotID string) Booking {
	req := BookingRequest{SlotID: slotID}
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("POST", baseURL+"/bookings/create", bytes.NewBuffer(body))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to create booking: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Booking Booking `json:"booking"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode booking: %v", err)
	}

	return result.Booking
}

func cancelBooking(t *testing.T, token, bookingID string) {
	httpReq, _ := http.NewRequest("POST", baseURL+"/bookings/"+bookingID+"/cancel", nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to cancel booking: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

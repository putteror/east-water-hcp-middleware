package service

import (
	"context"
	"errors"
	"testing"

	"eastwater-backend/internal/adapter/hikcentral"
	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/schema"
)

func TestPersonService(t *testing.T) {
	mockPort := hikcentral.NewMockAdapter()
	svc := NewPersonService(mockPort)
	ctx := context.Background()

	t.Run("GetPeople - returns seeded people", func(t *testing.T) {
		query := schema.PersonSearchQuery{}
		list, page, err := svc.GetPeople(ctx, query)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(list) < 2 {
			t.Errorf("expected at least 2 people, got %d", len(list))
		}
		if page.TotalRows < 2 {
			t.Errorf("expected total rows >= 2, got %d", page.TotalRows)
		}
	})

	t.Run("GetPersonByCode - existing record", func(t *testing.T) {
		resp, err := svc.GetPersonByCode(ctx, "P001")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.PersonCode != "P001" {
			t.Errorf("expected P001, got %s", resp.PersonCode)
		}
	})

	t.Run("CreatePerson - valid new person", func(t *testing.T) {
		req := &schema.CreatePersonRequest{
			PersonName:        "Anan Sukkasem",
			PersonCode:        "P099",
			LicensePlateTexts: []string{"3กก9999", "4ขข8888"},
			StartDateTime:     "2026-05-01T08:00:00Z",
			EndDateTime:       "2026-12-31T18:00:00Z",
		}

		resp, err := svc.CreatePerson(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.PersonCode != "P099" {
			t.Errorf("expected PersonCode P099, got %s", resp.PersonCode)
		}
		if resp.Action != "created" {
			t.Errorf("expected action created, got %s", resp.Action)
		}
	})

	t.Run("CreatePerson - conflict already exists", func(t *testing.T) {
		req := &schema.CreatePersonRequest{
			PersonName: "Duplicate Somchai",
			PersonCode: "P001",
		}
		_, err := svc.CreatePerson(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, common.ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})

	t.Run("UpdatePerson - success on existing person", func(t *testing.T) {
		req := &schema.UpdatePersonRequest{
			PersonName:        "Somchai Updated Name",
			LicensePlateTexts: []string{"9กก9999"},
		}
		resp, err := svc.UpdatePerson(ctx, "P001", req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.PersonName != "Somchai Updated Name" {
			t.Errorf("expected updated name, got %s", resp.PersonName)
		}
		if resp.Action != "updated" {
			t.Errorf("expected action updated, got %s", resp.Action)
		}
	})

	t.Run("UpdatePerson - not found on unknown person", func(t *testing.T) {
		req := &schema.UpdatePersonRequest{
			PersonName: "Unknown Person",
		}
		_, err := svc.UpdatePerson(ctx, "NON_EXISTENT", req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("UpsertPerson - updates existing", func(t *testing.T) {
		req := &schema.CreatePersonRequest{
			PersonCode: "P002",
			PersonName: "Wichai Upserted",
		}
		resp, err := svc.UpsertPerson(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Action != "updated" {
			t.Errorf("expected action updated, got %s", resp.Action)
		}
		if resp.PersonName != "Wichai Upserted" {
			t.Errorf("expected Wichai Upserted, got %s", resp.PersonName)
		}
	})

	t.Run("UpsertPerson - creates new", func(t *testing.T) {
		req := &schema.CreatePersonRequest{
			PersonCode: "P999_NEW",
			PersonName: "Brand New Person",
		}
		resp, err := svc.UpsertPerson(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Action != "created" {
			t.Errorf("expected action created, got %s", resp.Action)
		}
	})

	t.Run("CreatePerson - with partial deviceNames matching", func(t *testing.T) {
		req := &schema.CreatePersonRequest{
			PersonName:        "Anan Sukkasem",
			PersonCode:        "P100",
			LicensePlateTexts: []string{"3กก9999"},
			DeviceNames:       []string{"Main Entrance", "building a"},
			StartDateTime:     "2026-05-01T08:00:00Z",
			EndDateTime:       "2026-12-31T18:00:00Z",
		}

		resp, err := svc.CreatePerson(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.PersonCode != "P100" {
			t.Errorf("expected PersonCode P100, got %s", resp.PersonCode)
		}
		if len(resp.DeviceNames) == 0 {
			t.Fatal("expected deviceNames to be resolved and returned, got empty")
		}
		// "Main Entrance" should match "Main Entrance Door 1" and "Main Entrance Turnstile"
		// "building a" should match "Building A Entrance" and "Building A Turnstile"
		var foundDoor1, foundTurnstile, foundBuildingA bool
		for _, d := range resp.DeviceNames {
			if d == "Main Entrance Door 1" {
				foundDoor1 = true
			}
			if d == "Main Entrance Turnstile" {
				foundTurnstile = true
			}
			if d == "Building A Entrance" {
				foundBuildingA = true
			}
		}
		if !foundDoor1 || !foundTurnstile || !foundBuildingA {
			t.Errorf("expected resolved device names, got: %v", resp.DeviceNames)
		}
	})

	t.Run("UpdatePerson - update deviceNames", func(t *testing.T) {
		req := &schema.UpdatePersonRequest{
			DeviceNames: []string{"Pump"}, // should match "Pump Station Door 1"
		}
		resp, err := svc.UpdatePerson(ctx, "P001", req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(resp.DeviceNames) != 1 || resp.DeviceNames[0] != "Pump Station Door 1" {
			t.Errorf("expected ['Pump Station Door 1'], got: %v", resp.DeviceNames)
		}
	})

	t.Run("DeletePeople - existing person", func(t *testing.T) {
		delReq := &schema.DeletePeopleRequest{
			PersonCodes: []string{"P001"},
		}
		delResp, err := svc.DeletePeople(ctx, delReq)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if delResp.DeletedCount != 1 {
			t.Errorf("expected 1 deleted, got %d", delResp.DeletedCount)
		}
	})
}

func TestMatchDeviceNames(t *testing.T) {
	known := []string{
		"Main Entrance Door 1",
		"Main Entrance Turnstile",
		"Building A Entrance",
		"Building B Entrance",
		"East Water Gate Barrier In",
	}

	t.Run("Partial substring match case-insensitive", func(t *testing.T) {
		matched := schema.MatchDeviceNames([]string{"main entrance", "building b"}, known)
		if len(matched) != 3 {
			t.Fatalf("expected 3 matches, got %d: %v", len(matched), matched)
		}
		if matched[0] != "Main Entrance Door 1" || matched[1] != "Main Entrance Turnstile" || matched[2] != "Building B Entrance" {
			t.Errorf("unexpected matched items: %v", matched)
		}
	})

	t.Run("Unknown device preserved as-is", func(t *testing.T) {
		matched := schema.MatchDeviceNames([]string{"Special Secret Door"}, known)
		if len(matched) != 1 || matched[0] != "Special Secret Door" {
			t.Errorf("expected unknown device preserved, got: %v", matched)
		}
	})

	t.Run("Empty input", func(t *testing.T) {
		matched := schema.MatchDeviceNames(nil, known)
		if matched != nil {
			t.Errorf("expected nil for empty input, got: %v", matched)
		}
	})
}

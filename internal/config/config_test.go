package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("MENU_API_URL", "http://from-env/menu")
	t.Setenv("TEXTBELT_PHONE", "5550000000")
	t.Setenv("TEXTBELT", "env-key")
}

func writeConfigFile(t *testing.T, dir, name, contents string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	return path
}

func TestInitExplicitPath(t *testing.T) {
	setRequiredEnv(t)

	path := filepath.Join("..", "..", "testdata", "menufy.yml")

	Init(&path)

	cfg := GetConfig()

	if cfg.MenuAPIURL != "http://menu-api.test/api/CalendarView/GetDailyMenuitemsByGrade" {
		t.Errorf("MenuAPIURL = %q, want %q", cfg.MenuAPIURL, "http://menu-api.test/api/CalendarView/GetDailyMenuitemsByGrade")
	}
	if len(cfg.Schools) != 2 {
		t.Fatalf("Schools = %+v, want 2 schools", cfg.Schools)
	}
	if cfg.Schools[0].Name != "Test Elementary" || cfg.Schools[0].SchoolID != "11111111-1111-1111-1111-111111111111" || cfg.Schools[0].Grade != "04" {
		t.Errorf("Schools[0] = %+v, want name %q, id %q, and grade %q", cfg.Schools[0], "Test Elementary", "11111111-1111-1111-1111-111111111111", "04")
	}
	if cfg.Schools[1].Name != "Another Test Elementary" || cfg.Schools[1].SchoolID != "22222222-2222-2222-2222-222222222222" || cfg.Schools[1].Grade != "06" {
		t.Errorf("Schools[1] = %+v, want name %q, id %q, and grade %q", cfg.Schools[1], "Another Test Elementary", "22222222-2222-2222-2222-222222222222", "06")
	}
	if cfg.ServingLine != "Test Serving Line" {
		t.Errorf("ServingLine = %q, want %q", cfg.ServingLine, "Test Serving Line")
	}
	if cfg.MealType != "Test Meal Type" {
		t.Errorf("MealType = %q, want %q", cfg.MealType, "Test Meal Type")
	}
	if cfg.PersonID != "null" {
		t.Errorf("PersonID = %q, want %q", cfg.PersonID, "null")
	}
	if cfg.TextbeltPhone != "5555555555" {
		t.Errorf("TextbeltPhone = %q, want %q", cfg.TextbeltPhone, "5555555555")
	}
	if cfg.TextbeltKey != "test-api-key" {
		t.Errorf("TextbeltKey = %q, want %q", cfg.TextbeltKey, "test-api-key")
	}
	if cfg.Date != "09/29/2026" {
		t.Errorf("Date = %q, want %q", cfg.Date, "09/29/2026")
	}
}

func TestInitDefaultPath(t *testing.T) {
	setRequiredEnv(t)

	dir := t.TempDir()
	writeConfigFile(t, dir, "menufy.yml", `menu_api_url: "http://default-path/menu"
schools:
  - name: "Test Elementary"
    school_id: "test-school-id"
    grade: "04"
`)

	t.Chdir(dir)

	Init(nil)

	cfg := GetConfig()

	if cfg.MenuAPIURL != "http://default-path/menu" {
		t.Errorf("MenuAPIURL = %q, want %q", cfg.MenuAPIURL, "http://default-path/menu")
	}
	if len(cfg.Schools) != 1 {
		t.Fatalf("Schools = %+v, want 1 school", cfg.Schools)
	}
	if cfg.Schools[0].Name != "Test Elementary" || cfg.Schools[0].SchoolID != "test-school-id" || cfg.Schools[0].Grade != "04" {
		t.Errorf("Schools[0] = %+v, want name %q, id %q, and grade %q", cfg.Schools[0], "Test Elementary", "test-school-id", "04")
	}
	if cfg.ServingLine != "Specials of the Day" {
		t.Errorf("ServingLine = %q, want %q", cfg.ServingLine, "Specials of the Day")
	}
	if cfg.MealType != "Lunch" {
		t.Errorf("MealType = %q, want %q", cfg.MealType, "Lunch")
	}
	if cfg.PersonID != "null" {
		t.Errorf("PersonID = %q, want %q", cfg.PersonID, "null")
	}
	if cfg.TextbeltPhone != "5550000000" {
		t.Errorf("TextbeltPhone = %q, want %q", cfg.TextbeltPhone, "5550000000")
	}
	if cfg.TextbeltKey != "env-key" {
		t.Errorf("TextbeltKey = %q, want %q", cfg.TextbeltKey, "env-key")
	}
}

func TestInitConfigDirPath(t *testing.T) {
	setRequiredEnv(t)

	dir := t.TempDir()
	configDir := filepath.Join(dir, ".config", "menufy")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}
	writeConfigFile(t, configDir, "menufy.yml", `menu_api_url: "http://config-dir/menu"
schools: []
`)

	t.Chdir(dir)

	Init(nil)

	cfg := GetConfig()

	if cfg.MenuAPIURL != "http://config-dir/menu" {
		t.Errorf("MenuAPIURL = %q, want %q", cfg.MenuAPIURL, "http://config-dir/menu")
	}
	if len(cfg.Schools) != 0 {
		t.Errorf("Schools = %+v, want none", cfg.Schools)
	}
}

func TestInitEnvDefaults(t *testing.T) {
	setRequiredEnv(t)

	path := writeConfigFile(t, t.TempDir(), "config.yml", `schools: []
`)

	Init(&path)

	cfg := GetConfig()

	if cfg.MenuAPIURL != "http://from-env/menu" {
		t.Errorf("MenuAPIURL = %q, want %q", cfg.MenuAPIURL, "http://from-env/menu")
	}
	if len(cfg.Schools) != 0 {
		t.Errorf("Schools = %+v, want none", cfg.Schools)
	}
	if cfg.TextbeltPhone != "5550000000" {
		t.Errorf("TextbeltPhone = %q, want %q", cfg.TextbeltPhone, "5550000000")
	}
	if cfg.TextbeltKey != "env-key" {
		t.Errorf("TextbeltKey = %q, want %q", cfg.TextbeltKey, "env-key")
	}
	if cfg.ServingLine != "Specials of the Day" {
		t.Errorf("ServingLine = %q, want %q", cfg.ServingLine, "Specials of the Day")
	}
	if cfg.MealType != "Lunch" {
		t.Errorf("MealType = %q, want %q", cfg.MealType, "Lunch")
	}
	if cfg.PersonID != "null" {
		t.Errorf("PersonID = %q, want %q", cfg.PersonID, "null")
	}
	if _, err := time.Parse("01/02/2006", cfg.Date); err != nil {
		t.Errorf("Date = %q, want a date formatted as MM/DD/YYYY", cfg.Date)
	}
}

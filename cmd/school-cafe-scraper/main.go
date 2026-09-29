package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := loadConfig()

	body, err := getMenu(cfg)
	if err != nil {
		return fmt.Errorf("getting menu: %w", err)
	}

	menu, err := parseBody(body)
	if err != nil {
		return fmt.Errorf("parsing menu: %w", err)
	}

	entrees := getEntrees(menu)

	if err := sendMessage(cfg, entrees); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	return nil
}

type config struct {
	menuAPIURL    string
	schoolID      string
	servingLine   string
	mealType      string
	grade         string
	personID      string
	textbeltPhone string
	textbeltKey   string
	today         string
}

func loadConfig() config {
	return config{
		menuAPIURL:    mustEnv("MENU_API_URL"),
		schoolID:      mustEnv("SCHOOL_ID"),
		servingLine:   mustEnv("SERVING_LINE"),
		mealType:      mustEnv("MEAL_TYPE"),
		grade:         mustEnv("GRADE"),
		personID:      mustEnv("PERSON_ID"),
		textbeltPhone: mustEnv("TEXTBELT_PHONE"),
		textbeltKey:   os.Getenv("TEXTBELT"),
		today:         time.Now().Format("01/02/2006"),
	}
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return value
}

func getMenu(cfg config) (string, error) {
	baseURL, err := url.Parse(cfg.menuAPIURL)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}

	params := url.Values{}
	params.Add("SchoolId", cfg.schoolID)
	params.Add("ServingDate", cfg.today)
	params.Add("ServingLine", cfg.servingLine)
	params.Add("MealType", cfg.mealType)
	params.Add("Grade", cfg.grade)
	params.Add("PersonId", cfg.personID)

	baseURL.RawQuery = params.Encode()

	resp, err := httpClient.Get(baseURL.String())
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}

	return string(body), nil
}

func parseBody(body string) ([]map[string]any, error) {
	var result map[string][]map[string]any

	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling body: %w", err)
	}

	return result["ENTREES"], nil
}

func getEntrees(menu []map[string]any) []string {
	var entrees []string

	for _, v := range menu {
		for key, value := range v {
			if key == "MenuItemDescription" {
				entrees = append(entrees, value.(string))
			}
			continue
		}
	}

	return entrees
}

func sendMessage(cfg config, entrees []string) error {
	entreeList := strings.Join(entrees, "\n")
	msgBody := fmt.Sprintf("Entrees for %s at Standing Bear Elementary:\n\n%s", cfg.today, entreeList)

	params := url.Values{}
	params.Add("phone", cfg.textbeltPhone)
	params.Add("message", msgBody)
	params.Add("key", cfg.textbeltKey)

	_, err := httpClient.PostForm("https://textbelt.com/text", params)
	if err != nil {
		return fmt.Errorf("submitting form: %w", err)
	}

	return nil
}

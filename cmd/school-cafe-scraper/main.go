package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dhollinger/school-menu-scraper-go/internal/config"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func main() {
	configPath := flag.String("config", "", "path to the config file")
	flag.Parse()

	config.Init(configPath)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.GetConfig()

	for _, school := range cfg.Schools {

		body, err := getMenu(cfg, school.Name, school.SchoolID)
		if err != nil {
			return fmt.Errorf("getting menu: %w", err)
		}

		menu, err := parseBody(body)
		if err != nil {
			return fmt.Errorf("parsing menu: %w", err)
		}

		entrees := getEntrees(menu)

		if err := sendMessage(cfg, school.Name, entrees); err != nil {
			return fmt.Errorf("sending message: %w", err)
		}

	}
	return nil
}

func getMenu(cfg config.Config, school, id string) (string, error) {
	baseURL, err := url.Parse(cfg.MenuAPIURL)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}

	params := url.Values{}
	params.Add("SchoolId", id)
	params.Add("ServingDate", cfg.Date)
	params.Add("ServingLine", cfg.ServingLine)
	params.Add("MealType", cfg.MealType)
	params.Add("Grade", cfg.Grade)
	params.Add("PersonId", cfg.PersonID)

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

func sendMessage(cfg config.Config, school string, entrees []string) error {
	entreeList := strings.Join(entrees, "\n")
	msgBody := fmt.Sprintf("Entrees for %s at %s:\n\n%s", cfg.Date, school, entreeList)

	params := url.Values{}
	params.Add("phone", cfg.TextbeltPhone)
	params.Add("message", msgBody)
	params.Add("key", cfg.TextbeltKey)

	_, err := httpClient.PostForm("https://textbelt.com/text", params)
	if err != nil {
		return fmt.Errorf("submitting form: %w", err)
	}

	return nil
}

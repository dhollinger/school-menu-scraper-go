package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dhollinger/school-menu-scraper-go/internal/config"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func stubHTTPClient(t *testing.T, rt http.RoundTripper) {
	t.Helper()

	orig := httpClient
	httpClient = &http.Client{Timeout: time.Second, Transport: rt}
	t.Cleanup(func() { httpClient = orig })
}

func interceptTextbelt(t *testing.T, next http.RoundTripper) *[]url.Values {
	t.Helper()

	var posts []url.Values

	stubHTTPClient(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "textbelt.com" {
			return next.RoundTrip(req)
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		form, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}

		posts = append(posts, form)

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"success":true}`)),
			Request:    req,
		}, nil
	}))

	return &posts
}

func TestGetEntrees(t *testing.T) {
	tests := []struct {
		name string
		menu []map[string]any
		want []string
	}{
		{"empty menu", nil, nil},
		{"single item", []map[string]any{{"MenuItemDescription": "Cheese Pizza"}}, []string{"Cheese Pizza"}},
		{"multiple items with extra keys", []map[string]any{
			{"MenuItemDescription": "Cheese Pizza", "ItemID": 1.0},
			{"MenuItemDescription": "Orange Chicken"},
			{"ItemID": 3.0},
		}, []string{"Cheese Pizza", "Orange Chicken"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEntrees(tt.menu)
			if !slices.Equal(got, tt.want) {
				t.Errorf("getEntrees() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBody(t *testing.T) {
	t.Run("entrees present", func(t *testing.T) {
		want := []map[string]any{
			{"MenuItemDescription": "Cheese Pizza"},
			{"MenuItemDescription": "Orange Chicken"},
		}

		got, err := parseBody(`{"ENTREES":[{"MenuItemDescription":"Cheese Pizza"},{"MenuItemDescription":"Orange Chicken"}]}`)
		if err != nil {
			t.Fatalf("parseBody() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("parseBody() = %v, want %v", got, want)
		}
	})

	t.Run("no entrees", func(t *testing.T) {
		got, err := parseBody(`{"DRINKS":[]}`)
		if err != nil {
			t.Fatalf("parseBody() error = %v, want nil", err)
		}
		if got != nil {
			t.Errorf("parseBody() = %v, want nil", got)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseBody(`{`)
		if err == nil || !strings.Contains(err.Error(), "unmarshaling body") {
			t.Errorf("parseBody() error = %v, want it to contain %q", err, "unmarshaling body")
		}
	})
}

func TestGetMenu(t *testing.T) {
	newConfig := func(menuAPIURL string) config.Config {
		return config.Config{
			MenuAPIURL:  menuAPIURL,
			ServingLine: "Specials of the Day",
			MealType:    "Lunch",
			PersonID:    "null",
			Date:        "09/29/2026",
		}
	}

	wantBody := `{"ENTREES":[{"MenuItemDescription":"Cheese Pizza"}]}`

	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			for _, tc := range []struct {
				key  string
				want string
			}{
				{"SchoolId", "test-school-id"},
				{"ServingDate", "09/29/2026"},
				{"ServingLine", "Specials of the Day"},
				{"MealType", "Lunch"},
				{"PersonId", "null"},
			} {
				if got := query.Get(tc.key); got != tc.want {
					t.Errorf("query param %s = %q, want %q", tc.key, got, tc.want)
				}
			}
			_, _ = fmt.Fprint(w, wantBody)
		}))
		defer srv.Close()

		got, err := getMenu(newConfig(srv.URL), "Test Elementary", "test-school-id", "04")
		if err != nil {
			t.Fatalf("getMenu() error = %v, want nil", err)
		}
		if got != wantBody {
			t.Errorf("getMenu() = %q, want %q", got, wantBody)
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := getMenu(newConfig("http://[::1"), "Test Elementary", "test-school-id", "04")
		if err == nil || !strings.Contains(err.Error(), "parsing base URL") {
			t.Errorf("getMenu() error = %v, want it to contain %q", err, "parsing base URL")
		}
	})

	t.Run("request failure", func(t *testing.T) {
		_, err := getMenu(newConfig("http://127.0.0.1:1"), "Test Elementary", "test-school-id", "04")
		if err == nil || !strings.Contains(err.Error(), "making request") {
			t.Errorf("getMenu() error = %v, want it to contain %q", err, "making request")
		}
	})

	t.Run("reading body failure", func(t *testing.T) {
		stubHTTPClient(t, roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     make(http.Header),
				Body:       io.NopCloser(errorReader{}),
				Request:    req,
			}, nil
		}))

		_, err := getMenu(newConfig("http://stub.invalid"), "Test Elementary", "test-school-id", "04")
		if err == nil || !strings.Contains(err.Error(), "reading body") {
			t.Errorf("getMenu() error = %v, want it to contain %q", err, "reading body")
		}
	})
}

func TestSendMessage(t *testing.T) {
	cfg := config.Config{
		TextbeltPhone: "5550000000",
		TextbeltKey:   "test-key",
		Date:          "09/29/2026",
	}
	entrees := []string{"Cheese Pizza", "Orange Chicken"}

	t.Run("success", func(t *testing.T) {
		posts := interceptTextbelt(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("unexpected request")
		}))

		if err := sendMessage(cfg, "Test Elementary", entrees); err != nil {
			t.Fatalf("sendMessage() error = %v, want nil", err)
		}

		if len(*posts) != 1 {
			t.Fatalf("textbelt posts = %d, want 1", len(*posts))
		}

		form := (*posts)[0]
		wantMsg := "Entrees for 09/29/2026 at Test Elementary:\n\nCheese Pizza\nOrange Chicken"
		if form.Get("phone") != "5550000000" {
			t.Errorf("phone = %q, want %q", form.Get("phone"), "5550000000")
		}
		if form.Get("key") != "test-key" {
			t.Errorf("key = %q, want %q", form.Get("key"), "test-key")
		}
		if form.Get("message") != wantMsg {
			t.Errorf("message = %q, want %q", form.Get("message"), wantMsg)
		}
	})

	t.Run("request failure", func(t *testing.T) {
		stubHTTPClient(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		}))

		err := sendMessage(cfg, "Test Elementary", entrees)
		if err == nil || !strings.Contains(err.Error(), "submitting form") {
			t.Errorf("sendMessage() error = %v, want it to contain %q", err, "submitting form")
		}
	})
}

const testMenuAPIURL = "http://menu-api.test/api/CalendarView/GetDailyMenuitemsByGrade"

func TestRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"ENTREES":[{"MenuItemDescription":"Cheese Pizza"},{"MenuItemDescription":"Orange Chicken"}]}`)
	}))
	defer srv.Close()

	t.Setenv("MENU_API_URL", srv.URL)
	t.Setenv("GRADE", "04")
	t.Setenv("TEXTBELT_PHONE", "5550000000")
	t.Setenv("TEXTBELT", "test-key")

	fixture, err := os.ReadFile(filepath.Join("..", "..", "testdata", "menufy.yml"))
	if err != nil {
		t.Fatalf("reading test config fixture: %v", err)
	}

	configYAML := strings.ReplaceAll(string(fixture), testMenuAPIURL, srv.URL)

	path := filepath.Join(t.TempDir(), "menufy.yml")
	if err := os.WriteFile(path, []byte(configYAML), 0o600); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	config.Init(&path)

	posts := interceptTextbelt(t, http.DefaultTransport)

	if err := run(); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}

	if len(*posts) != 2 {
		t.Fatalf("textbelt posts = %d, want 2", len(*posts))
	}

	schools := []string{"Test Elementary", "Another Test Elementary"}
	for i, school := range schools {
		form := (*posts)[i]
		if form.Get("phone") != "5555555555" {
			t.Errorf("phone = %q, want %q", form.Get("phone"), "5555555555")
		}
		if form.Get("key") != "test-api-key" {
			t.Errorf("key = %q, want %q", form.Get("key"), "test-api-key")
		}
		wantMsg := fmt.Sprintf("Entrees for 09/29/2026 at %s:\n\nCheese Pizza\nOrange Chicken", school)
		if form.Get("message") != wantMsg {
			t.Errorf("message = %q, want %q", form.Get("message"), wantMsg)
		}
	}
}

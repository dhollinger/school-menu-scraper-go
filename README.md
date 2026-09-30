# school-menu-scraper

A small CLI that fetches the daily lunch entrees for configured schools from the
web API of your choice and texts them via
[Textbelt](https://textbelt.com).

This is a proof of concept that is still actively worked on — expect rapid
changes.

## Usage

Required environment variables:

| Variable | Purpose |
| --- | --- |
| `MENU_API_URL` | School daily menu endpoint URL |
| `GRADE` | Grade level to fetch the menu for |
| `TEXTBELT_PHONE` | Phone number that receives the menu |
| `TEXTBELT` | Textbelt API key |

Configuration is read from a `menufy.yml` file, looked up in `.` and
`.config/menufy/`, or passed explicitly:

```sh
go run ./cmd/school-cafe-scraper -config path/to/menufy.yml
```

A minimal config selects the schools to fetch — the name appears in the text
message and the id is the School Cafe school identifier:

```yaml
schools:
  - name: Some Elementary
    school_id: 00000000-0000-0000-0000-000000000000
```

All other settings are optional and fall back to the environment variables
above or built-in defaults (`serving_line`, `meal_type`, `person_id`, `date`).

## Tests

```sh
go test ./...
```

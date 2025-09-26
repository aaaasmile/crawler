# Crawler
The Crawler downloads stock charts and sends them via secure email. It's used to monitor stock charts weekly without manually checking each individual chart.

## How it works
First, the Scraper program (scraper/main.go) downloads SVG files and converts them into PNG format. The Crawler then uses the Scraper's output to generate an email with embedded chart images. The email is sent using an SMTP mail relay over TLS transport.

## Database
The stock list is stored in a SQLite database created with the SQL script stockinfo-ref-db.sql.

## Scraper
The Scraper is needed to convert complex SVG files with styles into simple PNG files that are easy to embed in emails. The SVG to PNG conversion is done using chromedp screenshot functionality.
# Web Page Analyzer (Golang)

A simple web application to analyze a given webpage by its URL. It extracts and displays:

- HTML version
- Page title
- Headings count by level (H1–H6)
- Internal and external link counts
- Inaccessible link count
- Whether a login form is present

## 🔧 Technologies Used

- Go (Golang)
- `html/template` for rendering HTML
- `golang.org/x/net/html` for parsing HTML DOM
- Concurrency (Goroutines and channels) for link checks

## 🚀 Getting Started

### 1. Clone the repository

git clone https://github.com/nasminspy/lucytech.git
cd lucytech

### 2. Install dependencies

go mod tidy

### 3. Run the application

go run main.go

### 4. Open in browser

Go to: [http://localhost:8080](http://localhost:8080)

---

## 🛠 Directory Structure

├── handler/
│   └── handler.go
├── parser/
│   └── analyzer.go
├── templates/
│   └── index.html
├── main.go
├── go.mod
└── README.md

## 📌 Assumptions & Design Decisions

* If the user-provided URL doesn’t include `http://` or `https://`, it defaults to `https://`.
* HTML version is detected based on the `<!DOCTYPE>` declaration.
* Login form detection is based on the presence of `<input type="password">`.
* Link accessibility is checked using `HEAD` requests with a timeout.
* Internal/external links are categorized based on domain matching.


## 🔄 Suggestions for Improvement

* Show a detailed list of inaccessible links.
* Visual indicators for headings and link types.
* Asynchronous frontend updates using AJAX/WebSocket.
* Add support for analyzing local HTML files.
* Offer a JSON API for programmatic access.

## 📝 License

This project is provided for educational and demonstration purposes. Feel free to adapt or extend it as needed
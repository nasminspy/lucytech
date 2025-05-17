# 📘 LucyTech – Web Page Analyzer

## 🤩 Project Overview

**LucyTech** is a lightweight web application developed in Go (Golang) that analyzes a webpage provided by its URL. It extracts and displays:

* HTML version
* Page title
* Headings count by level (H1–H6)
* Internal and external link counts
* Inaccessible link count
* Presence of a login form

This tool is particularly useful for SEO specialists, developers, and QA engineers who need quick insights into webpage structures.

---

## ⚙️ Prerequisites

* **Go (Golang)**: Version 1.20 or higher
* **Git**: For cloning the repository

---

## 🛠️ Technologies Used

### Backend

* **Go (Golang)**: Main programming language
* **html/template**: For rendering HTML templates
* **golang.org/x/net/html**: For parsing HTML DOM
* **Goroutines and Channels**: For concurrent link checking

### DevOps

* **Docker**: Containerization of the application
* **Prometheus**: Metrics collection
* **GitHub Actions**: CI/CD pipeline (if configured)

---

## 📦 External Dependencies

The project uses Go modules for dependency management.

To install dependencies:

```bash
go mod tidy
```

---

## 🚀 Setup Instructions

1. **Clone the repository**

   ```bash
   git clone https://github.com/nasminspy/lucytech.git
   cd lucytech
   ```

2. **Install dependencies**

   ```bash
   go mod tidy
   ```

3. **Run the application**

   ```bash
   go run main.go
   ```

4. **Access the application**

   Open your browser and navigate to: [http://localhost:8080](http://localhost:8080)

---

## 🔪 Testing

The project includes unit tests for handlers and parsers.

To run tests:

```bash
go test ./...
```

---

## 🐳 Dockerization

A Dockerfile is provided for containerizing the application.

To build and run the Docker container:

```bash
docker build -t lucytech .
docker run -p 8080:8080 lucytech
```

---

## 📈 Metrics

The application exposes Prometheus metrics at: [http://localhost:6060/metrics](http://localhost:6060/metrics)

Ensure Prometheus is configured to scrape metrics from this endpoint.

---

## 🔍 Application Usage

* **Home Page (`/`)**: Provides a form to input the URL of the webpage to analyze.
* **Analyze Endpoint (`/analyze`)**: Processes the submitted URL and displays the analysis results, including HTML version, title, headings count, link counts, inaccessible links, and login form presence.

---

## 🧰 Main Functionalities

* **HTML Parsing**: Utilizes `golang.org/x/net/html` to parse and traverse the HTML DOM.
* **Link Classification**: Differentiates between internal and external links based on the base URL.
* **Accessibility Check**: Performs HTTP HEAD requests to determine if links are accessible.
* **Login Form Detection**: Checks for the presence of `<input type="password">` to identify login forms.
* **Concurrency**: Employs goroutines and channels to perform link accessibility checks concurrently, improving performance.
* **Metrics Collection**: Exposes application metrics for monitoring via Prometheus.

---

## 🧠 Challenges Faced

* **Efficient Link Checking**: Ensuring that link accessibility checks do not block the main thread required implementing concurrency using goroutines and channels.
* **HTML Parsing Complexity**: Handling various HTML structures and edge cases necessitated robust parsing logic.
* **Error Handling**: Implementing comprehensive error handling to manage malformed URLs and network issues gracefully.

---

## 🌟 Possible Improvements

* **Frontend Enhancement**: Integrate a modern frontend framework (e.g., React or Vue.js) for a more interactive UI.
* **Authentication**: Implement user authentication to allow users to save and manage their analysis history.
* **API Documentation**: Provide Swagger or OpenAPI documentation for the application's endpoints.
* **CI/CD Pipeline**: Set up GitHub Actions for automated testing and deployment.
* **Database Integration**: Store analysis results in a database for historical data and reporting.

---

## 📄 License

This project is licensed under the MIT License.

---

## 📬 Contact

For any inquiries or feedback, please contact [nasminspy](https://github.com/nasminspy).

---

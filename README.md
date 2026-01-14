# DataHub Domain Scanner (High-Performance Data Ingestion Engine)

## Overview
This repository contains a high-performance network data ingestion pipeline built in Go. It is designed to perform large-scale discovery and extraction of structured information from unstructured and semi-structured internet sources, including DNS records, SSL/TLS certificates, HTTP headers, HTML metadata, and decentralized protocol descriptors (NodeInfo/Matrix).

While written in Go for concurrency and performance, the core logic focuses on **ETL (Extract, Transform, Load)** principles: ingesting raw network responses and normalizing them into a canonical, validated JSON schema for downstream analysis.

## Key Data Engineering Features

### 1. Robust Data Extraction & Parsing
The pipeline extracts key fields from various "messy" real-world sources:
*   **HTML Metadata:** Utilizes a scraping layer to extract OpenGraph, Twitter Cards, and standard meta tags, filtering them against a blacklist to ensure data relevance.
*   **X.509 Certificate Sanitization:** Implements a custom mapping of OIDs (Object Identifiers) to human-readable names, transforming complex cryptographic structures into a flat, searchable schema.
*   **Protocol Discovery:** Automatically detects and parses `.well-known` configurations for Matrix and NodeInfo protocols, handling multi-step JSON lookups and redirection logic.

### 2. Data Normalization & Schema Validation
To ensure the output is ready for a structured database, the project:
*   Defines a **Canonical Schema** (`structs.go`) that enforces consistency across different scan types.
*   **Normalizes Redirects:** Includes logic to differentiate between local, scheme-based, and external redirections to prevent "dumb redirection" loops and ensure data lineage is preserved.
*   **Unit & Format Standardization:** Converts timestamps to a standard ISO-like format and maps varied DNS record types into a unified key-value structure.

### 3. Data Quality & Error Handling
Working with the public internet requires handling high-entropy data. This scanner implements:
*   **Validation Checks:** Logic to detect and tag "invalid-redirects" or "external-redirects."
*   **Timeout & Resource Management:** Configurable request limits (`MaxRespLen`, `MaxRedir`) to prevent the pipeline from being hung by oversized or malicious responses.
*   **Sanitization:** Cleaning of domain names using Punycode/IDNA standards before processing.

### 4. High-Concurrency Architecture
Designed for speed, the scanner utilizes Go’s concurrency primitives (`sync.WaitGroup`) to process thousands of domains in parallel, making it an ideal "ingestion front-end" for a data pipeline.

---

## Technical Stack
*   **Language:** Go (Golang)
*   **Parsing:** `net/http`, `github.com/anaskhan96/soup` (HTML), `encoding/json`
*   **Network/DNS:** `github.com/miekg/dns`
*   **Testing:** Go Test framework with concurrent execution patterns.

## Project Structure
*   `domain_scanner.go`: The main orchestration layer for the domain ETL process.
*   `x509_names.go` & `certificate_sanitizer.go`: Data normalization logic for SSL/TLS data.
*   `http_scanner.go`: Logic for extracting information from the HTTP/HTML layer.
*   `structs.go`: The "Canonical Schema" definition.
*   `cfg.go`: Pipeline constants and data quality thresholds.

## Getting Started

### Prerequisites
*   Go 1.23+

### Installation
```bash
git clone https://github.com/your-username/datahub-gomain-scanner.git
cd datahub-gomain-scanner
go mod tidy
```

### Usage
To scan a single domain and output the structured JSON report:
```bash
go run . example.com
```

### Running Tests
The project includes a test suite that demonstrates the ability to handle a list of domains via concurrent workers:
```bash
go test ./...
```

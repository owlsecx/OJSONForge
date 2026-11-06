# 🦉 OJSONForge

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Linux-0a0c10?style=flat-square"/>
  <img src="https://img.shields.io/badge/Category-Core%20/%20Utilities-blue?style=flat-square"/>
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square"/>
  <img src="https://img.shields.io/badge/Part%20of-OwlSec%20Toolkit-7b5ea7?style=flat-square"/>
  <img src="https://img.shields.io/badge/Version-1.0-cyan?style=flat-square"/>
</p>

> **OJSONForge** — JSON Query, Transform, Validate & Pretty ║

---

> ⚠️ **AUTHORISED USE ONLY** — Use only on systems you own or have explicit written permission to test. Unauthorised use is illegal.

---

## 📌 Overview

OJSONForge — JSON Query, Transform, Validate & Pretty. JSON Query / Transform / Validate / Pretty Pretty-print, minify, validate (with line/column error position), query by dotted/bracket path, and profile the structure of any JSON document — from a file, inline text, or standard input.

---

## ✨ Features

- Query, transform, validate and pretty-print JSON
- CLI-friendly pipelines for config and API payloads
- Validation with clear error localization
- Transform helpers for common reshaping tasks
- Interactive cyan UI for exploration
- Export cleaned JSON and validation reports

---

## ⚙️ Requirements

- **Linux**
- **Go 1.21+** (stdlib only; `CGO_ENABLED=0`)

---

## 🚀 Usage

```bash
go build -ldflags="-s -w" -o OJSONForge OJSONForge.go
./OJSONForge --help
./OJSONForge --input data.json --pretty
./OJSONForge --input data.json --validate
./OJSONForge --input data.json --query 'items[*].id'
```

---

## 📦 Part of OwlSec Toolkit

This tool is part of the **OwlSec** suite — a collection of 300+ security and privacy tools.

🔗 [owlsec.org](https://owlsec.org)

---

## ©️ License

MIT License — © Khaled S. Haddad

<div align="center">

# 🃏 RoyalFlush

**Online Poker Platform**

[![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![Angular](https://img.shields.io/badge/Angular-DD0031?style=flat&logo=angular&logoColor=white)](https://angular.io/)
[![SQLite](https://img.shields.io/badge/SQLite-07405E?style=flat&logo=sqlite&logoColor=white)](https://www.sqlite.org/)

</div>

---

## 📖 Project Description

**RoyalFlush** is a multiplayer Texas Hold'em poker application designed to demonstrate modern full-stack development practices. The platform allows users to create private game rooms, invite friends via codes, and play poker in real-time.

The application emphasizes low-latency communication and a responsive user experience. It utilizes a **Go (Golang)** backend to handle game logic and WebSocket connections, paired with an **Angular** frontend for the interactive game interface. The architecture focuses on clean code, secure authentication, and reliable state synchronization using **GORM** for data management.

### 🌟 Key Features
* **Real-Time Gameplay:** Instant state updates using WebSockets.
* **Private Rooms:** Create rooms and invite friends via unique codes.
* **Secure Auth:** User registration and login protected by JWT and bcrypt.
* **Persistent Data:** User profiles and game history stored in SQLite.

---

## 🧱 Frontend Baseline Decisions

This project has a locked frontend baseline that must stay stable before feature work.

### Node Policy
* Required runtime: **Node 20 LTS**
* Source of truth: `.nvmrc` contains `20`
* Enforcement:
  * `frontend/package.json` -> `"engines": { "node": "20.x" }`
  * `frontend/.npmrc` -> `engine-strict=true`
* Developer workflow: run `nvm use 20` before any frontend `npm` command

### Angular Version Policy
* All `@angular/*` packages are pinned to `21.1.4` in `frontend/package.json`
* No caret ranges are used for Angular packages to prevent team drift

### UI Library Decision
* Chosen UI library: **Angular Material**
* Installed and pinned packages:
  * `@angular/material@21.1.4`
  * `@angular/cdk@21.1.4`
  * `@angular/animations@21.1.4`
* Configuration:
  * Theme import in `frontend/src/styles.scss`
  * Animations provider in `frontend/src/app/app.config.ts`
* Baseline usage example: `frontend/src/app/pages/login/login.html` renders Material components (`mat-card`, `mat-raised-button`)

### Routing Architecture
* Public routes:
  * `/login`
  * `/register`
* Protected-shell routes (auth guard intentionally deferred):
  * `/dashboard`
  * `/lobby`
  * `/table/:id`
* Redirects:
  * `/` -> `/login`
  * `**` -> `/login`
* Route config location: `frontend/src/app/app.routes.ts`

### Scope Guard
* Baseline setup includes **only** structural placeholders and configuration
* Authentication logic and feature logic are intentionally out of scope

---

## 🛠 Tech Stack

| Domain | Technology | Usage |
| :---: | :---: | :---: |
| **Backend** | ![Go](https://img.shields.io/badge/-Go-00ADD8?logo=go&logoColor=white&style=flat) **Gorilla Mux**, **GORM** | Game Engine, API, Routing, ORM |
| **Frontend** | ![Angular](https://img.shields.io/badge/-Angular-DD0031?logo=angular&logoColor=white&style=flat) **RxJS**, **Lodash** | Interactive UI, State Management, Utilities |
| **Database** | ![SQLite](https://img.shields.io/badge/-SQLite-07405E?logo=sqlite&logoColor=white&style=flat) | Persistent Data Storage |

---

## 👥 Team Members

| Role | Name |
| :---: | :---: |
| **Backend** | Sai Puneeth Bonagiri<br>Himanshu Potham Shetty |
| **Frontend** | Sai Shravanth Reddy Madem<br>Devi Sanikommu |

---

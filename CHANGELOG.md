# Changelog

All notable changes to the **RoyalFlush** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- [Backend: User Model & Registration API](https://github.com/SE-RoyalFlush/Poker/issues/6)
    - Defined User data model using GORM with secure password hashing (bcrypt)
    - Implemented registration endpoint `POST /api/register` with validation
    - Added database auto-migration for User model
    - Integrated unit tests for authentication and models
    - Updated OpenAPI specification and Bruno collection for registration
- [Basic Go Setup: Health Check and APIs](https://github.com/SE-RoyalFlush/Poker/issues/3)
    - Initial Go backend setup with Gorilla Mux router
    - Health check endpoint at `/health`
    - Structured JSON 404 error responses
    - Bruno API test collection for backend endpoints

### Changed
- Moved all backend API routes under the `/api/` prefix for better separation (e.g., `/api/health`).

### Fixed

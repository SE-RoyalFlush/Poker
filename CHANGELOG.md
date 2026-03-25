# Changelog

All notable changes to the **RoyalFlush** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- [Frontend: Authentication Service Implementation](https://github.com/SE-RoyalFlush/Poker/pull/42)
    - Implemented `AuthService` in `src/app/core/services/auth.service.ts` as single source of truth for user authentication state
    - Created User model interfaces (`User`, `LoginCredentials`, `RegisterData`) in `src/app/core/models/user.model.ts`
    - Implemented RxJS BehaviorSubject pattern for state management with `currentUser$` observable for components
    - Implemented `login(credentials)` method - POSTs to `/api/login`, chains to `checkSession()` for user profile fetch
    - Implemented `register(data)` method - POSTs to `/api/register` and returns the created user payload
    - Implemented `logout()` method - POSTs to `/api/logout`, clears user state locally
    - Implemented `checkSession()` method - Calls `GET /api/me` to validate session and restore user state on app initialization
    - Configured `APP_INITIALIZER` to restore user session automatically on app startup
    - Added synchronous helper methods `isAuthenticated()` and `getCurrentUser()` for route guards
    - Comprehensive unit tests (29 tests, 100% passing) using `HttpClientTestingModule`
    - Updated frontend README with authentication architecture, usage examples, and API reference
    - Added detailed authentication and backend auth endpoint documentation to existing project docs
    - **Note**: Backend endpoints `/api/login`, `/api/logout`, and `/api/me` are required but not yet implemented
- [Frontend: Secure HTTP Client with AuthInterceptor](https://github.com/SE-RoyalFlush/Poker/pull/41)
    - Implemented AuthInterceptor in `src/app/core/interceptors/auth.interceptor.ts` for secure HTTP communication
    - Configured interceptor to automatically attach `withCredentials: true` to all HTTP requests for cookie-based authentication
    - Implemented X-CSRF-Token extraction from cookies and automatic injection into request headers
    - Added global error handling in interceptor with automatic redirect to login on 401 (Unauthorized) responses
    - Registered HTTP interceptor in application configuration for all outgoing requests
    - Added comprehensive unit tests for interceptor functionality and error handling
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

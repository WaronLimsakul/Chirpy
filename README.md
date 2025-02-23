# Chirpy API

## Overview

(This project come from the course [Learn HTTP Servers](https://www.boot.dev/courses/learn-http-servers-golang)
Chirpy is a lightweight microblogging API that allows users to create, retrieve, and delete short messages called "chirps." This API ensures proper user authentication, message validation, and content moderation.

## Features

- User authentication via JWT
- Chirp validation (max 140 characters, filtered words)
- CRUD operations for chirps
- Sorting and filtering chirps by author and timestamp

---

## _Endpoints_

## Users

### 1.Register a User

**Endpoint:** `POST /api/users`

**Description:** Creates a new user account.

**Request Body:**

```json
{
  "username": "example_user",
  "password": "securepassword"
}
```

**Response:**

```json
{
  "id": "uuid",
  "username": "example_user",
  "created_at": "timestamp"
}
```

**Errors:**

- `400 Bad Request` if the request body is invalid.
- `409 Conflict` if the username is already taken.

---

### 2. Authenticate a User

**Endpoint:** `POST /api/users/authenticate`

**Description:** Authenticates a user and returns a JWT token.

**Request Body:**

```json
{
  "username": "example_user",
  "password": "securepassword"
}
```

**Response:**

```json
{
  "token": "jwt_token"
}
```

**Errors:**

- `401 Unauthorized` if credentials are invalid.

---

### 3. Get User Profile

**Endpoint:** `GET /api/users/{user_id}`

**Description:** Retrieves public details of a user.

**Response:**

```json
{
  "id": "uuid",
  "username": "example_user",
  "created_at": "timestamp"
}
```

**Errors:**

- `404 Not Found` if the user does not exist.

---

### 4. Update User Profile

**Endpoint:** `PUT /api/users/{user_id}`

**Description:** Updates the authenticated user’s profile.

**Headers:**

```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "username": "new_username",
  "password": "newpassword"
}
```

**Response:**

```json
{
  "id": "uuid",
  "username": "new_username",
  "updated_at": "timestamp"
}
```

**Errors:**

- `401 Unauthorized` if the token is invalid or missing.
- `403 Forbidden` if the user is not authorized to update this profile.
- `404 Not Found` if the user does not exist.

---

### 5. Delete User Account

**Endpoint:** `DELETE /api/users/{user_id}`

**Description:** Deletes the authenticated user’s account.

**Headers:**

```
Authorization: Bearer <token>
```

**Response:**

```
204 No Content
```

**Errors:**

- `401 Unauthorized` if the token is invalid or missing.
- `403 Forbidden` if the user is not authorized to delete this account.
- `404 Not Found` if the user does not exist.

## Chirps

### **1. Validate Chirp**

**Endpoint:** `POST /validate-chirp`

**Description:**
Validates a chirp's content to ensure it follows length restrictions and filters out inappropriate words.

**Request Body:**

```json
{
  "body": "Some chirp message"
}
```

**Response:**

```json
{
  "valid": true,
  "error": "",
  "cleaned_body": "Some cleaned message"
}
```

**Errors:**

- `400 Bad Request` if chirp is too long
- `500 Internal Server Error` if request parsing fails

---

### **2. Create Chirp**

**Endpoint:** `POST /chirps`

**Authentication Required:** ✅

**Description:**
Creates a new chirp after validation and stores it in the database.

**Request Header:**

```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "body": "Hello, world!"
}
```

**Response:**

```json
{
  "id": "<chirp_uuid>",
  "created_at": "<timestamp>",
  "updated_at": "<timestamp>",
  "body": "Hello, world!",
  "user_id": "<user_uuid>"
}
```

**Errors:**

- `400 Bad Request` if chirp is too long
- `401 Unauthorized` if authentication fails
- `500 Internal Server Error` if chirp creation fails

---

### **3. Get All Chirps**

**Endpoint:** `GET /chirps`

**Description:**
Fetches all chirps, with optional filtering and sorting.

**Query Parameters:**

- `author_id=<uuid>` - Filters chirps by author
- `sort=asc|desc` - Sorts chirps by creation timestamp

**Response:**

```json
[
  {
    "id": "<chirp_uuid>",
    "created_at": "<timestamp>",
    "updated_at": "<timestamp>",
    "body": "Hello, world!",
    "user_id": "<user_uuid>"
  }
]
```

**Errors:**

- `500 Internal Server Error` if retrieval fails

---

### **4. Get Chirp by ID**

**Endpoint:** `GET /chirps/{chirp_id}`

**Description:**
Fetches a single chirp by its ID.

**Response:**

```json
{
  "id": "<chirp_uuid>",
  "created_at": "<timestamp>",
  "updated_at": "<timestamp>",
  "body": "Hello, world!",
  "user_id": "<user_uuid>"
}
```

**Errors:**

- `404 Not Found` if chirp does not exist
- `500 Internal Server Error` if request fails

---

### **5. Delete Chirp**

**Endpoint:** `DELETE /chirps/{chirp_id}`

**Authentication Required:** ✅

**Description:**
Deletes a chirp if the authenticated user is the owner.

**Request Header:**

```
Authorization: Bearer <token>
```

**Errors:**

- `400 Bad Request` if chirp ID is missing
- `401 Unauthorized` if authentication fails
- `403 Forbidden` if user does not own the chirp
- `404 Not Found` if chirp does not exist
- `500 Internal Server Error` if deletion fails

---

## Tech Stack

- **Go** (Golang) - API implementation
- **PostgreSQL** - Database for storing chirps
- **JWT** - Authentication mechanism
- **UUID** - Unique identifier for chirps and users

## Setup Instructions

1. Clone the repository:

   ```sh
   git clone https://github.com/WaronLimsakul/Chirpy.git
   cd Chirpy
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Set up environment variables:

   ```sh
   export TOKEN_SECRET="your-secret-key"
   export DATABASE_URL="your-database-url"
   ```

4. Run the API:
   ```sh
   go run main.go
   ```

## Contribution Guidelines

- Fork the repository
- Create a feature branch
- Submit a pull request with detailed descriptions

## License

This project is licensed under the MIT License.

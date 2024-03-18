# quick-match

# Project Setup

## Overview

This document provides instructions on how to set up the QuickMatch project, including building the application, running services with Docker Compose, and configuring AWS resources with Terraform.

## Prerequisites

- Docker and Docker Compose installed on your machine.
- Terraform installed for AWS resource configuration.
- An AWS account and AWS CLI configured with access credentials.

## Building the Application

The application is containerized using Docker. To build the Docker image, navigate to the project's root directory and run:

```bash
docker build -t quickmatch .
```

This command builds the Docker image based on the instructions in the Dockerfile, preparing it to be run.

## Running Services with Docker Compose

To start the necessary services, including LocalStack for AWS service emulation and the QuickMatch application, use Docker Compose:

```bash
docker-compose up -d
```

This command starts the services defined in `docker-compose.yml`, setting up LocalStack for DynamoDB and Elasticsearch services and running the QuickMatch application.

## Configuring AWS Resources with Terraform

Terraform is used to configure AWS resources like DynamoDB tables and Elasticsearch domain for development purposes. To apply the Terraform configuration, navigate to the directory containing your Terraform files and run:

### Initialize Terraform:

```bash
terraform init
```

### Apply Config:
```bash
terraform apply
```

## Accessing the Application

Once the Docker Compose services are up, and Terraform has successfully applied the configurations, the QuickMatch application will be accessible at `http://localhost:8080`.

## Notes

- The LocalStack service emulates AWS services locally for development purposes. It's configured to run DynamoDB and Elasticsearch, mapped to port 4566.
- The app service depends on localstack, ensuring AWS services are available before the application starts.
- AWS resources (DynamoDB tables and Elasticsearch domain) are created with minimal configuration suitable for development and testing. 


## CreateUser Endpoint

### Overview

The `CreateUser` endpoint automatically generates new users with complete profile details. Upon a successful request, the newly created user is inserted into both DynamoDB and ElasticSearch, with sensitive data appropriately stripped for security.

### URL

`POST /user/create`

### Method

`POST`

### URL Params

None required.

### Data Params

No data needs to be provided in the request body. This endpoint does not require input from the caller, as it auto-generates all necessary user information internally.

### Success Response

- **Code**: `201 Created`
- **Content**: The response body contains the newly created user's details in JSON format, excluding sensitive information.

Example:
```bash
curl -X POST http://localhost:8080/user/create -H "Content-Type: application/json"
```

```json
{
  "userID": "uniqueUserID",
  "username": "generatedUsername",
  "email": "user@example.com",
  // Other user details...
}
```

## Login Endpoint

### Overview

The `Login` endpoint authenticates users by validating their email and password. Upon successful authentication, it generates an authentication token that can be used for accessing secured endpoints. The login process involves validating the provided login credentials against the stored values, including comparing the hashed password.

### URL

`POST /login`

### Method

`POST`

### URL Params

None required.

### Data Params

The request body must include login credentials with the following structure:

```json
{
  "email": "user@example.com",
  "password": "userPassword"
}
```
- `email` (required): The user's email address. Must be a valid email format.
- `password` (required): The user's password.

### Success Response

- **Code**: `200 OK`
- **Content**: A JSON object containing the authentication token.

Example:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Error Response

Possible error responses include:

- **Code**: `400 Bad Request`
  - **Content**: `"Invalid request"`
    - Occurs when the request body cannot be decoded or fails validation checks.

- **Code**: `401 Unauthorized`
  - **Content**: `"Invalid credentials"`
    - Returned if the email or password does not match stored values.

- **Code**: `500 Internal Server Error`
  - **Content**: `"Server error"` or `"Failed to generate token"`
    - Indicates a problem with the server, such as failure to access the user repository or token service.

### Sample Call

Example using `curl`:

```bash
curl -X POST http://localhost:8080/login \
-H "Content-Type: application/json" \
-d '{"email": "user@example.com", "password": "userPassword"}'
```

## Swipe Endpoint

### Overview

The `Swipe` endpoint processes swipe actions (like or dislike) between users. It updates the swipe action with the user's ID and checks for potential matches based on mutual likes. A unique `MatchID` is generated for matched users, facilitating further interactions or messaging between them.

### URL

`POST /swipe`

### Method

`POST`

### URL Params

None required.

### Data Params

The request body must include swipe information with the following structure:

```json
{
  "SwipedUserID": "targetUserID",
  "preference": true // or false for dislike
}
```

- `SwipedUserID` (required): The ID of the user being swiped on.
- `preference` (required): A boolean where `true` indicates a like and `false` indicates a dislike.

### Success Response

- **Code**: `200 OK`
- **Content**: A JSON object indicating whether a match was made.

Example for a match:
```json
{
  "matched": true,
  "matchId": "uniqueMatchID"
}
```

Example for no match:

```json
{
  "matched": false
}
```

### Error Response

Possible error responses include:

- **Code**: `400 Bad Request`
    - **Content**: `"Invalid request body"`
        - Occurs when the request body cannot be decoded.

- **Code**: `500 Internal Server Error`
    - **Content**: `"Failed to authenticate"`, `"Failed to insert swipe record"`, or `"Failed to check for swipe match"`
        - Indicates a problem with server processing, such as failing to authenticate the user, insert the swipe record, or check for a match.

### Sample Call

Example using `curl`:

```bash
curl -X POST http://localhost:8080/swipe \
-H "Authorization: Bearer {your_jwt_token}" \
-H "Content-Type: application/json" \
-d '{"SwipedUserID": "targetUserID", "preference": true}'
```

### Notes

- The `UserID` is extracted from the request context, assuming it's set by a preceding JWT middleware that authenticates the user.
- A swipe action is considered a potential match only if both users have swiped right (liked) on each other.
- The endpoint requires a valid JWT token to authenticate the user making the swipe action.

## Discover Endpoint

### Overview

The `DiscoverUserInsert` endpoint processes user discovery requests, leveraging user IDs to exclude previously swiped users from the discovery results. It utilizes authenticated users' details, fetched from Elasticsearch—including their location—to filter compatible users based on provided discovery filters. This functionality allows for a dynamic and personalized user discovery experience.

### URL

`POST /discover`

### Method

`POST`

### URL Params

None required.

### Data Params

Discovery filters can be provided in the request body. None of the filters are mandatory, allowing for flexible user discovery based on optional criteria:

```json
{
  "gender": "female",
  "maxAge": 30,
  "minAge": 20,
  "maxLocation": 100 // The maximum distance in kilometers
}
```

- `gender` (optional): Filter users by gender.
- `maxAge` (optional): The maximum age of users to discover.
- `minAge` (optional): The minimum age of users to discover.
- `maxLocation` (optional): The maximum distance (in kilometers) from the user's location to consider for discovering other users.

### Success Response

- **Code**: `200 OK`
- **Content**: A list of users that match the discovery filters.

Example:
```json
{
  "users": [
    {
      "UserID": "user123",
      "name": "Jane Doe",
      "gender": "female",
      "age": 25,
      "location": {
        "lat": 52.5200,
        "lon": 13.4050
      }
    }
    // Additional matching users...
  ]
}
```

### Error Response

Possible error responses include:

- **Code**: `400 Bad Request`
    - **Content**: `"Invalid request body"`
        - Occurs when the request body cannot be decoded.

- **Code**: `500 Internal Server Error`
    - **Content**: `"Failed to find Swiped IDs"`, `"Failed to fetch user from Elasticsearch"`, or `"Failed to search users with df"`
        - Indicates a problem with server processing, such as failing to retrieve swiped IDs, fetch user details from Elasticsearch, or perform the user search based on the discovery filters.

### Sample Call

Example using `curl`:

```bash
curl -X POST http://localhost:8080/discover \
-H "Authorization: Bearer {your_jwt_token}" \
-H "Content-Type: application/json" \
-d '{"gender": "female", "maxAge": 30, "minAge": 20, "maxLocation": 100}'
```

### Notes

- User IDs are extracted from the request context, set by a preceding JWT middleware that authenticates the user.
- The discovery process excludes users that the authenticated user has already swiped on, ensuring fresh and relevant discovery results.
- The endpoint requires a valid JWT token to authenticate the user making the discovery request.


# Address Validation Webhook with Google Maps API

A secure, containerized webhook service for validating addresses using the Google Maps API, with geofencing capabilities.

![Address Validation](https://developers.google.com/static/maps/images/landing/hero_geocoding_api.png)

## Features

- **Address Validation**: Validate addresses using Google Maps API
- **Geofencing**: Restrict validation to addresses within a specific geographic area
- **Security Measures**:
  - Input sanitization to prevent injection attacks
  - Rate limiting to prevent API abuse
  - Suspicious pattern detection
  - HTTPS requirement option
- **Docker Support**:
  - Secure storage of API key in .env file
  - Volume mounting for configuration
  - Multi-stage build for minimal image size
- **Logging**: Structured logging with slog
- **Health Check**: Basic health check endpoint

## Setup

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for containerized deployment)
- Google Maps API key

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```
ENVIRONMENT=DEVELOPMENT
REQUIRE_HTTPS=false
PORT=8080

# Rate limiting settings
RATE_LIMIT_MAX_REQUESTS=10
RATE_LIMIT_TIME_WINDOW_SECONDS=60

# Logger Settings
LEVEL=DEBUG
ENCODING=console
OUTPUT_PATH=stdout
ERROR_PATH=stdout


# Map settings
GOOGLE_MAPS_API_KEY=your_api_key_here
MAP_MAX_DISTANCE=2
MAP_DISTANCE_UNIT=mi
MAP_CENTER_LAT=40.8313747
MAP_CENTER_LNG=-73.8272283
```

### Running Locally

1. Clone the repository
2. Set up your `.env` file with your Google Maps API key
3. Run the application:

```bash
go mod tidy
go run main.go
```

### Running with Docker

1. Clone the repository
2. Set up your `.env` file with your Google Maps API key
3. Build and run with Docker Compose:

```bash
docker-compose up -d
```

## Architecture

The application follows a hexagonal architecture pattern:

- **Ports**: Define interfaces between layers
- **Adapters**: Implement external services (Google Maps API)
- **Services**: Contain business logic
- **Handlers**: Handle HTTP requests and responses

![Hexagonal Architecture](https://miro.medium.com/v2/resize:fit:1400/1*yR4C1B-YfMh5zqpbHzTyag.png)

## How It Works

### Address Validation

1. The client sends a POST request to `/validate` with an address
2. The handler validates the request and applies rate limiting
3. The service sanitizes the input and checks for suspicious patterns
4. The adapter calls the Google Maps API to validate the address
5. The service checks if the address is within the geofence
6. The handler returns the validation result

### Geofencing

1. When an address is validated, the service calculates the distance between the address and the center of the geofence using the Haversine formula
2. If the distance is less than or equal to the maximum allowed distance, the address is considered within the geofence (`inRange=true`)
3. If the distance is greater than the maximum allowed distance, the address is considered outside the geofence (`inRange=false`)

![Geofencing Illustration](https://miro.medium.com/v2/resize:fit:1400/1*qcAZgT4Sk37ZPVQZ-M_aAQ.png)

### Security Measures

- **Input Sanitization**: Removes dangerous characters to prevent injection attacks
- **Rate Limiting**: Limits the number of requests per time window to prevent API abuse
- **Suspicious Pattern Detection**: Rejects addresses with suspicious patterns
- **HTTPS Requirement**: Option to require HTTPS for all requests

## API Documentation

### Validate Address

Validates an address and checks if it's within the geofence.

**Endpoint**: `POST /validate`

**Request Body**:
```json
{
  "address": "123 Main St, New York, NY"
}
```

**Response**:
```json
{
  "isValid": true,
  "formattedAddress": "123 Main St, New York, NY 10001, USA",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "inRange": true
}
```

| Field | Description |
|-------|-------------|
| `isValid` | Whether the address is valid |
| `formattedAddress` | The formatted address from Google Maps |
| `latitude` | The latitude of the address |
| `longitude` | The longitude of the address |
| `inRange` | Whether the address is within the geofence |
| `error` | Error message (if any) |

### Health Check

Checks if the service is running.

**Endpoint**: `GET /health`

**Response**:
```
OK
```

## Examples

### Address Within Geofence

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"address": "123 Main St, Bronx, NY"}'
```

Response:
```json
{
  "isValid": true,
  "formattedAddress": "123 Main St, Bronx, NY 10456, USA",
  "latitude": 40.8448,
  "longitude": -73.8648,
  "inRange": true
}
```

### Address Outside Geofence

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"address": "123 Main St, Manhattan, NY"}'
```

Response:
```json
{
  "isValid": true,
  "formattedAddress": "123 Main St, Manhattan, NY 10001, USA",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "inRange": false
}
```

### Invalid Address

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"address": "This is not a valid address"}'
```

Response:
```json
{
  "isValid": false,
  "error": "Address not found"
}
```

## Testing

The application includes comprehensive unit tests for all components:

- **Adapter Tests**: Test the Google Maps API adapter and geofencing
- **Service Tests**: Test the business logic
- **Handler Tests**: Test the HTTP handlers

Run the tests with:

```bash
go test -v ./...
```

## Docker Configuration

The application uses Docker for containerization:

- **Multi-stage Build**: Minimizes the final image size
- **Volume Mounting**: Securely mounts the `.env` file
- **Non-root User**: Runs as a non-root user for security
- **Resource Limits**: Sets CPU and memory limits

## License

This project is licensed under the MIT License - see the LICENSE file for details.
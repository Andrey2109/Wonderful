# OpenAI Realtime API Voice Assistant with SIP Integration

This Go-based application connects phone calls to OpenAI's Realtime API using SIP (Session Initiation Protocol). It enables natural voice conversations with AI by handling incoming calls through webhooks and establishing real-time WebSocket connections.

## Features

- **SIP Integration**: Connect phone calls directly to OpenAI's Realtime API via SIP trunking providers (e.g., Twilio)
- **Webhook Handler**: Automatically accepts and processes incoming call events from OpenAI
- **Real-time Voice Conversations**: Bi-directional audio streaming for natural AI-powered phone conversations
- **WebSocket Management**: Manages persistent WebSocket connections for call monitoring and control
- **Database Integration**: PostgreSQL database for storing call logs and session data
- **Configurable Assistant Instructions**: Customize the AI assistant's behavior, voice, and personality
- **Docker Support**: Containerized deployment with Docker Compose

## Prerequisites

Before setting up this project, you'll need:

- **Go 1.25.1 or later**
- **OpenAI API key** with Realtime API access
- **OpenAI Project ID** (found in Settings > General on platform.openai.com)
- **OpenAI Webhook Secret** (created during webhook setup)
- **SIP Trunking Provider** (e.g., Twilio with a phone number)
- **Ngrok** (for local development tunneling)
- **PostgreSQL** (for database storage)
- **Docker & Docker Compose** (optional, for containerized deployment)

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/Andrey2109/Wonderful.git
cd Wonderful
```

### 2. Configure OpenAI Webhook

1. Go to [OpenAI Platform Settings](https://platform.openai.com/settings/)
2. Navigate to **Webhooks** in the sidebar
3. Click **Create a webhook**
4. Set the **URL** to your ngrok endpoint (e.g., `https://your-domain.ngrok.app/webhook`)
5. Set **Event Type** to `realtime.call.incoming`
6. Copy the **Webhook Secret** for later use

### 3. Configure SIP Trunking (Twilio Example)

1. **Purchase a Phone Number** with voice capabilities from [Twilio Console](https://console.twilio.com/)
2. **Create a SIP Trunk** in Twilio Console > SIP Trunks
3. **Add Origination URI**: `sip:YOUR_PROJECT_ID@sip.api.openai.com;transport=tls`
   - Replace `YOUR_PROJECT_ID` with your OpenAI Project ID
4. **Connect your phone number** to the SIP Trunk

### 4. Environment Configuration

Create a `.env` file in the project root with the following variables:

```env
OPENAI_API_KEY=sk-proj-your_api_key_here
OPENAI_WEBHOOK_SECRET=whsec_your_webhook_secret_here
OPENAI_PROJECT_ID=proj_your_project_id_here
PORT=8000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_db_password
DB_NAME=wonderful_voice_assistant
```

### 5. Database Setup

If using PostgreSQL locally:

```bash
# Create the database
psql -U postgres -c "CREATE DATABASE wonderful_voice_assistant;"

# Run the initialization script
psql -U postgres -d wonderful_voice_assistant -f init-db/wonderful_voice_assistant.sql
```

### 6. Install Go Dependencies

```bash
go mod download
```

## Usage

### Running Locally with Ngrok

1. **Start Ngrok tunnel** (in a separate terminal):
   ```bash
   ngrok http 8000
   ```
   
   Or with a static domain (paid plan):
   ```bash
   ngrok http 8000 --domain=your-domain.ngrok.app
   ```

2. **Update your OpenAI webhook URL** with the ngrok endpoint if it changed

3. **Run the application**:
   ```bash
   go run .
   ```
   
   For debugging:
   ```bash
   go run . -debug=true
   ```

4. **Call your phone number** and start talking to the AI assistant!

### Running with Docker Compose

For a complete deployment with PostgreSQL:

```bash
docker-compose up -d
```

This will:
- Start a PostgreSQL database
- Initialize the database schema
- Run the Go application
- Expose the webhook endpoint on port 8000

### Running with Pre-built Docker Image

```bash
docker pull andrey2109/wonderful-cli:latest

docker run -d \
  -p 8000:8000 \
  -e OPENAI_API_KEY=your_api_key \
  -e OPENAI_WEBHOOK_SECRET=your_webhook_secret \
  -e OPENAI_PROJECT_ID=your_project_id \
  andrey2109/wonderful-cli:latest
```

## Customizing the Assistant

### Voice Instructions

Modify `voice_instructions.txt` to customize the assistant's:
- Personality and tone
- Language and communication style
- Domain expertise
- Response patterns
- Call handling behavior

Example:
```
You are a friendly customer support agent for Acme Corp.
Speak warmly and professionally. Ask clarifying questions when needed.
If the customer needs to speak to a human, transfer them using the transfer tool.
```

### Configuration Options

The application supports various configuration options through environment variables or code modifications:

- **Voice Selection**: Change the AI voice (alloy, echo, fable, onyx, nova, shimmer)
- **Model**: Configure the Realtime model version
- **Instructions**: Set system-level behavior and personality
- **Response Templates**: Customize initial greetings and responses


## Architecture & Design

### How It Works

1. **Incoming Call**: When someone dials your phone number, the SIP trunking provider routes the call to OpenAI's SIP endpoint
2. **Webhook Event**: OpenAI fires a `realtime.call.incoming` webhook to your server with call details
3. **Call Acceptance**: The application accepts the call and configures the Realtime session (voice, instructions, model)
4. **WebSocket Connection**: A persistent WebSocket connection is established to stream audio and events
5. **Real-time Conversation**: Bi-directional audio flows between the caller and the AI model
6. **Call Management**: The application can monitor, log, and control the call flow

### Components

#### SIP Server (`sipserver.go`)
Handles incoming webhook events from OpenAI:
- Validates webhook signatures for security
- Accepts/rejects incoming calls
- Configures Realtime session parameters
- Returns appropriate HTTP responses

#### WebSocket Client (`client.go`)
Manages real-time communication:
- Establishes WebSocket connections to OpenAI's Realtime API
- Sends initial response instructions
- Handles audio streaming and events
- Manages connection lifecycle (open, message, error, close)

#### Database Layer (`db.go`)
Stores call data and session information:
- Call logs with timestamps and metadata
- Session details and conversation history
- Error logging and debugging information

#### Main Application (`main.go`)
Orchestrates the components:
- Initializes database connections
- Starts the HTTP webhook server
- Manages graceful shutdown
- Handles environment configuration

### WebSocket Event Flow

```
Caller → SIP Provider → OpenAI SIP Endpoint → Webhook → Your Server
                                                              ↓
                                                    Accept Call + Configure
                                                              ↓
                                              WebSocket Connection Established
                                                              ↓
                                              ← Audio/Events Bi-directional →
```

## Project Structure

```
Wonderful/
├── main.go                          # Application entry point
├── sipserver.go                     # Webhook handler for SIP events
├── client.go                        # WebSocket client for Realtime API
├── db.go                            # Database operations
├── tools.go                         # Function calling tools
├── utils.go                         # Utility functions
├── utils_test.go                    # Unit tests
├── voice_instructions.txt           # AI assistant instructions
├── instructions.txt                 # Legacy CLI instructions
├── go.mod                           # Go dependencies
├── Dockerfile                       # Container configuration
├── docker-compose.yml               # Multi-container orchestration
├── .env                             # Environment variables (create this)
├── init-db/
│   └── wonderful_voice_assistant.sql # Database schema
└── README.md                        # This file
```

## API Endpoints

### POST `/webhook`
Receives incoming call webhooks from OpenAI

**Request Headers:**
- `webhook-signature`: OpenAI signature for verification
- `webhook-id`: Unique webhook event ID
- `webhook-timestamp`: Unix timestamp

**Request Body:**
```json
{
  "object": "event",
  "type": "realtime.call.incoming",
  "data": {
    "call_id": "rtc_...",
    "sip_headers": [...]
  }
}
```

**Response:** `200 OK` on successful acceptance

## Troubleshooting

### Common Issues

**1. Webhook not receiving calls**
- Verify ngrok is running and the URL matches your OpenAI webhook configuration
- Check that the webhook event type is set to `realtime.call.incoming`
- Ensure your SIP trunk origination URI is correctly formatted

**2. Call connects but no audio**
- Confirm the call acceptance request returns 200 OK
- Check WebSocket connection is established successfully
- Verify `voice_instructions.txt` exists and contains valid instructions

**3. Invalid webhook signature**
- Ensure `OPENAI_WEBHOOK_SECRET` in `.env` matches the webhook secret from OpenAI
- Check that the request body is being read as raw bytes for signature verification

**4. Database connection errors**
- Verify PostgreSQL is running and accessible
- Check database credentials in `.env`
- Ensure the database schema has been initialized

**5. Ngrok URL changes**
- Use ngrok static domains (paid feature) for consistent URLs
- Update OpenAI webhook URL whenever ngrok restarts with a new random domain

### Debug Mode

Enable debug logging to see WebSocket events:

```bash
go run . -debug=true
```

This will output:
- WebSocket connection status
- Incoming/outgoing messages
- Call acceptance responses
- Error details

## Testing

Run unit tests:

```bash
go test ./...
```

Run specific tests:

```bash
go test -v -run TestFunctionName
```

## Additional Resources

### OpenAI Documentation
- [Realtime API with SIP Guide](https://platform.openai.com/docs/guides/realtime-sip)
- [Realtime API Reference](https://platform.openai.com/docs/api-reference/realtime)
- [Webhook Events](https://platform.openai.com/docs/api-reference/webhook-events)

### Twilio Documentation
- [Elastic SIP Trunking](https://www.twilio.com/docs/sip-trunking)
- [Connect OpenAI Realtime API with Twilio](https://www.twilio.com/en-us/blog/developers/tutorials/product/openai-realtime-api-elastic-sip-trunking)

### Tools & Services
- [Ngrok Documentation](https://ngrok.com/docs)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is provided as-is for educational and development purposes.

## Acknowledgments

- OpenAI for the Realtime API and SIP integration
- Twilio for SIP trunking infrastructure
- Gorilla WebSocket library for Go WebSocket support
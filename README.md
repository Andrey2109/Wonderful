# Real-time GPT-4o-mini CLI with Function Calling

This Go-based CLI application provides a real-time interface to OpenAI's GPT-4o-mini model using WebSockets. It features streaming responses and function calling capabilities.

## Features

- **Real-time Streaming Responses**: Messages appear incrementally as they're generated, providing a natural conversation experience
- **Function Calling**: Demonstrates OpenAI's function calling capabilities with a multiplication function
- **Configurable Assistant Instructions**: Easily customize the assistant's behavior through environment variables
- **Robust Error Handling**: Gracefully handles connection issues and API errors

## Installation

### Prerequisites

- Go 1.25.1 or later
- OpenAI API key with access to GPT-4o-mini

### Setup

1. Clone this repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Configure your OpenAI API key in the [`.env`](.env ) file (see Configuration section)

## Configuration

Application uses a [`.env`](.env ) file for configuration. Create your own file in the project root:

```
OPENAI_API_KEY=your_openai_api_key_here
```

### Customizing Assistant Instructions

The assistant's instructions are defined in the `instructions.txt` file in the project root. 
You can modify this file to change the assistant's behavior, personality, or capabilities.

## Usage

### Running the CLI

Start the application with:

```bash
go run .
```

For debugging WebSocket communication:

```bash
go run . -debug=true
```


## Architecture & Design Choices

### WebSocket Implementation

The application uses the Gorilla WebSocket library to establish a real-time connection with OpenAI's API. This enables **true streaming functionality** where:

- Responses appear incrementally character-by-character
- The experience mimics a natural conversation flow

The streaming implementation leverages Go's concurrency features:
- A dedicated goroutine (`readLoop`) continuously reads incoming WebSocket messages
- Each message is processed and displayed in real-time
- The main thread remains responsive for user input

### Function Calling

The application demonstrates OpenAI's function calling capabilities with a multiplication function:

1. The function is registered during session initialization
2. When the model detects a multiplication request, it calls the function
3. The CLI processes the arguments, performs the calculation, and returns the result
4. The model incorporates the result into its response

### Error Handling

The application includes robust error handling for:
- WebSocket connection issues
- API authentication problems
- Invalid function arguments
- Context cancellation (for clean shutdown)

## Testing

To run tests:

```bash
go test
```

## Acknowledgments

- OpenAI for providing the API
- Gorilla WebSocket library
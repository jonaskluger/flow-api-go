# Flow Production Tracking REST API Go Client

A simple Go package for interacting with the Flow Production Tracking (formerly ShotGrid/Shotgun) REST API.

## Installation

```bash
go get github.com/jonaskluger/flow-api-go
```

## Configuration in Flow Production Tracking

To use this package, you need to create an API Script in Flow Production Tracking:

### Creating an API Script

1. **Log into Flow Production Tracking** as an admin user
2. **Navigate to Admin menu** (click your profile icon in the top right)
3. **Select "Scripts"** from the dropdown
4. **Click "+ Script"** button to create a new script
5. **Fill in the details:**
   - **Script Name:** Give it a descriptive name (e.g., "My Go Application")
   - **Description:** Optional description of what the script does
   - **Status:** Set to "Active"
6. **Click "Create Script"**
7. **Copy the credentials:**
   - **Application Key** (this is your `script_key`/`client_secret`)
   - Keep these credentials secure - they provide full API access!

### Finding Your Site URL

Your site URL is: `https://myproject.shotgrid.autodesk.com`

You can find it by looking at your browser's address bar when logged into Flow Production Tracking.

## Usage

### Basic Authentication Example

```go
package main

import (
	"fmt"
	"log"

	flowapi "github.com/jonaskluger/flow-api-go"
)

func main() {
	// Create a new client with your credentials
	client, err := flowapi.NewClient(flowapi.Config{
		SiteURL:    "https://myproject.shotgrid.autodesk.com",
		ScriptName: "your_script_name",
		ScriptKey:  "your_script_key",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Check if authenticated
	if client.IsAuthenticated() {
		fmt.Println("Successfully authenticated!")
	}

	// Get the access token
	token, err := client.GetAccessToken()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Access Token: %s\n", token)
}
```

### Using Environment Variables (Recommended)

For security, it's recommended to use environment variables:

```go
package main

import (
	"fmt"
	"log"
	"os"

	flowapi "github.com/jonaskluger/flow-api-go"
)

func main() {
	client, err := flowapi.NewClient(flowapi.Config{
		SiteURL:    os.Getenv("FLOW_SITE_URL"),
		ScriptName: os.Getenv("FLOW_SCRIPT_NAME"),
		ScriptKey:  os.Getenv("FLOW_SCRIPT_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Client created successfully!")
}
```

Then set the environment variables:

```bash
export FLOW_SITE_URL="https://myproject.shotgrid.autodesk.com"
export FLOW_SCRIPT_NAME="your_script_name"
export FLOW_SCRIPT_KEY="your_script_key"
```

## Token Management

The client automatically handles token management:

- Tokens are automatically obtained when you create a new client
- Tokens are automatically refreshed when they expire (with a 60-second buffer)
- The default token lifetime is typically 1 hour

```go
// The token is automatically managed
token, err := client.GetAccessToken()
if err != nil {
    log.Fatal(err)
}

// Use the token in your API requests
// Authorization: Bearer <token>
```

## Creating a Client from Environment Variables

For convenience, you can create a client using environment variables:

```go
package main

import (
	"log"
	
	flowapi "github.com/jonaskluger/flow-api-go"
)

func main() {
	// Automatically loads .env file and creates client
	client, err := flowapi.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	
	// Client is ready to use
}
```

Create a `.env` file in your project root:

```env
FLOW_SITE_URL=https://myproject.shotgrid.autodesk.com
FLOW_SCRIPT_NAME=your_script_name
FLOW_SCRIPT_KEY=your_script_key
```

## Working with Entities

### Finding Entities

Search for entities with filters:

```go
// Find all shots in a project
filters := []interface{}{
	[]interface{}{"project", "is", map[string]interface{}{
		"type": "Project",
		"id":   123,
	}},
}
shots, err := client.FindEntities("shots", filters, []string{"code", "description", "sg_status_list"})
if err != nil {
	log.Fatal(err)
}

for _, shot := range shots {
	fmt.Printf("Shot: %s\n", shot["code"])
}
```

### Getting a Single Entity

Retrieve an entity by ID:

```go
// Get a specific shot
shot, err := client.GetEntity("shots", 456, []string{"code", "description", "sg_status_list"})
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Shot: %s - %s\n", shot["code"], shot["description"])
```

### Creating Entities

Create a new entity:

```go
// Create a new note
noteData := map[string]interface{}{
	"subject": "Review feedback",
	"content": "This shot looks great!",
	"project": map[string]interface{}{
		"type": "Project",
		"id":   123,
	},
	"note_links": []interface{}{
		map[string]interface{}{
			"type": "Shot",
			"id":   456,
		},
	},
}

note, err := client.CreateEntity("notes", noteData)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Created note with ID: %d\n", note["id"])
```

## Helper Functions

The package includes convenient helper functions for common queries:

### User Queries

```go
// Get user by login name
user, err := client.GetUserByLogin("john.doe")
if err != nil {
	log.Fatal(err)
}

// Get user by display name
user, err := client.GetUserByName("John Doe")
if err != nil {
	log.Fatal(err)
}
```

### Shot Queries

```go
// Get all shots (no filters)
allShots, err := client.GetShots(0, []string{"code", "description"})
if err != nil {
	log.Fatal(err)
}

// Get shots for a specific project
projectShots, err := client.GetShots(123, []string{"code", "sg_status_list"})
if err != nil {
	log.Fatal(err)
}

// Get shots assigned to a user
userShots, err := client.GetShotsForUser(789, []string{"code", "project"})
if err != nil {
	log.Fatal(err)
}
```

### Task Queries

```go
// Get tasks for a specific shot
tasks, err := client.GetTasksForShot(456, []string{"content", "sg_status_list"})
if err != nil {
	log.Fatal(err)
}

// Get tasks assigned to a user
userTasks, err := client.GetTasksForUser(789, []string{"content", "entity", "project"})
if err != nil {
	log.Fatal(err)
}
```

## Features

- ✅ Client credentials authentication (API Script)
- ✅ Automatic token refresh with expiration handling
- ✅ Entity search with filters
- ✅ Get single entity by ID
- ✅ Create entities
- ✅ Helper functions for users, shots, and tasks
- ✅ Environment variable configuration with .env file support
- ⏳ Update entities (coming soon)
- ⏳ Delete entities (coming soon)
- ⏳ File upload/download (coming soon)

## Documentation

- [Flow Production Tracking REST API Documentation](https://developers.shotgridsoftware.com/rest-api/)
- [Authentication Guide](https://developers.shotgridsoftware.com/rest-api/#authentication)

## License

MIT License

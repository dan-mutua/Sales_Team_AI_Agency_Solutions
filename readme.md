# Sales Team AI Agency Solutions

A modern GraphQL-based CRM system designed specifically for sales agencies, featuring AI-powered lead management and automated client interactions.

## Features

- 🤖 AI-powered lead management and engagement
- 📊 Advanced campaign management and metrics
- 👥 Client relationship management
- 📈 Performance tracking and analytics
- 📧 Multi-channel communication support
- 🎯 Target audience management
- 📚 Training program management

## Tech Stack

- **Backend**: Go
- **API**: GraphQL
- **Database**: PostgreSQL
- **Framework**: gqlgen
- **Router**: chi

## Prerequisites

- Go 1.16 or higher
- PostgreSQL 12 or higher
- Docker (optional)

## Getting Started

1. Clone the repository:
```bash
git clone https://github.com/yourusername/Sales_Team_AI_Agency_Solutions.git
cd Sales_Team_AI_Agency_Solutions
```

2. Create a `.env` file in the root directory:
```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/salesagency?sslmode=disable
PORT=8080
```

3. Install dependencies:
```bash
go mod download
```

4. Start the server:
```bash
go run main.go
```

The GraphQL playground will be available at `http://localhost:8080`

## Database Schema

The application uses PostgreSQL with the following main entities:
- Leads
- Clients
- AI Agents
- Campaigns
- Interactions
- Message Templates
- Training Programs

## API Structure

### Main Types
- `Lead`: Represents potential customers
- `Client`: Represents active clients
- `AIAgent`: AI-powered automation agents
- `Campaign`: Marketing and sales campaigns
- `Interaction`: Communication records
- `MessageTemplate`: Reusable message templates

### Key Operations

#### Leads
```graphql
# Query leads
query { leads(filter: LeadFilterInput) { id name email status } }

# Create lead
mutation { createLead(input: LeadInput!) { id name } }
```

#### AI Agents
```graphql
# Trigger AI agent
mutation { triggerAIAgentRun(id: "agent-id") }

# Get AI agent stats
query { aiAgent(id: "agent-id") { stats { leadsEngaged responseRate } } }
```

## Development


### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o server main.go
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
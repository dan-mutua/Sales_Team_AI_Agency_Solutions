package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"salesagency/graph/model"

	_ "github.com/lib/pq"
)

// DB wraps the SQL database connection
type DB struct {
	conn *sql.DB
}

func Initialize() (*DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgresql://postgres:postgres@localhost:5432/salesagency?sslmode=disable"
	}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) beginTx(ctx context.Context) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, nil)
}

func (db *DB) GetLeadByID(ctx context.Context, id string) (*model.Lead, error) {
	query := `SELECT id, name, email, phone, company, position, status, intent_score, 
              tags, source, last_contact, next_follow_up, notes, created_at, updated_at 
              FROM leads WHERE id = $1`

	var lead model.Lead
	var tagsArray []sql.NullString
	var updatedAt sql.NullTime
	var lastContact, nextFollowUp sql.NullTime
	var phone, company, position, source, notes sql.NullString

	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&lead.ID, &lead.Name, &lead.Email, &phone, &company, &position, &lead.Status, &lead.IntentScore,
		&tagsArray, &source, &lastContact, &nextFollowUp, &notes, &lead.CreatedAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No lead found
		}
		return nil, fmt.Errorf("error fetching lead: %w", err)
	}

	if phone.Valid {
		lead.Phone = &phone.String
	}
	if company.Valid {
		lead.Company = &company.String
	}
	if position.Valid {
		lead.Position = &position.String
	}
	if source.Valid {
		lead.Source = &source.String
	}
	if notes.Valid {
		lead.Notes = &notes.String
	}
	if lastContact.Valid {
		lead.LastContact = &lastContact.Time
	}
	if nextFollowUp.Valid {
		lead.NextFollowUp = &nextFollowUp.Time
	}
	if updatedAt.Valid {
		lead.UpdatedAt = &updatedAt.Time
	}

	lead.Tags = make([]string, 0, len(tagsArray))
	for _, tag := range tagsArray {
		if tag.Valid {
			lead.Tags = append(lead.Tags, tag.String)
		}
	}

	return &lead, nil
}

func (db *DB) GetLeadsByFilter(ctx context.Context, filter *model.LeadFilterInput, limit *int, offset *int) ([]*model.Lead, error) {
	query := `SELECT id, name, email, phone, company, position, status, intent_score, 
              tags, source, last_contact, next_follow_up, notes, created_at, updated_at 
              FROM leads WHERE 1=1`

	var args []interface{}
	argCount := 1

	if filter != nil {
		if filter.Status != nil && len(filter.Status) > 0 {
			query += fmt.Sprintf(" AND status = ANY($%d)", argCount)
			args = append(args, filter.Status)
			argCount++
		}

		if filter.MinIntentScore != nil {
			query += fmt.Sprintf(" AND intent_score >= $%d", argCount)
			args = append(args, *filter.MinIntentScore)
			argCount++
		}

		if filter.Tags != nil && len(filter.Tags) > 0 {
			query += fmt.Sprintf(" AND tags && $%d", argCount)
			args = append(args, filter.Tags)
			argCount++
		}

		if filter.Source != nil {
			query += fmt.Sprintf(" AND source = $%d", argCount)
			args = append(args, *filter.Source)
			argCount++
		}

		if filter.LastContactAfter != nil {
			query += fmt.Sprintf(" AND last_contact >= $%d", argCount)
			args = append(args, *filter.LastContactAfter)
			argCount++
		}

		if filter.LastContactBefore != nil {
			query += fmt.Sprintf(" AND last_contact <= $%d", argCount)
			args = append(args, *filter.LastContactBefore)
			argCount++
		}
	}

	query += " ORDER BY created_at DESC"
	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, *limit)
		argCount++
	}

	if offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, *offset)
	}

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying leads: %w", err)
	}
	defer rows.Close()

	var leads []*model.Lead
	for rows.Next() {
		var lead model.Lead
		var tagsArray []sql.NullString
		var updatedAt sql.NullTime
		var lastContact, nextFollowUp sql.NullTime
		var phone, company, position, source, notes sql.NullString

		err := rows.Scan(
			&lead.ID, &lead.Name, &lead.Email, &phone, &company, &position, &lead.Status, &lead.IntentScore,
			&tagsArray, &source, &lastContact, &nextFollowUp, &notes, &lead.CreatedAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning lead row: %w", err)
		}

		if phone.Valid {
			lead.Phone = &phone.String
		}
		if company.Valid {
			lead.Company = &company.String
		}
		if position.Valid {
			lead.Position = &position.String
		}
		if source.Valid {
			lead.Source = &source.String
		}
		if notes.Valid {
			lead.Notes = &notes.String
		}
		if lastContact.Valid {
			lead.LastContact = &lastContact.Time
		}
		if nextFollowUp.Valid {
			lead.NextFollowUp = &nextFollowUp.Time
		}
		if updatedAt.Valid {
			lead.UpdatedAt = &updatedAt.Time
		}

		lead.Tags = make([]string, 0, len(tagsArray))
		for _, tag := range tagsArray {
			if tag.Valid {
				lead.Tags = append(lead.Tags, tag.String)
			}
		}

		leads = append(leads, &lead)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lead rows: %w", err)
	}

	return leads, nil
}

func (db *DB) CreateLead(ctx context.Context, lead *model.Lead) (*model.Lead, error) {
	query := `INSERT INTO leads (name, email, phone, company, position, status, intent_score, 
              tags, source, notes, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
              RETURNING id`

	err := db.conn.QueryRowContext(
		ctx, query, lead.Name, lead.Email, lead.Phone, lead.Company, lead.Position,
		lead.Status, lead.IntentScore, lead.Tags, lead.Source, lead.Notes, lead.CreatedAt,
	).Scan(&lead.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating lead: %w", err)
	}

	return lead, nil
}

func (db *DB) UpdateLead(ctx context.Context, lead *model.Lead) (*model.Lead, error) {
	query := `UPDATE leads SET 
              name = $1, email = $2, phone = $3, company = $4, position = $5, 
              status = $6, intent_score = $7, tags = $8, source = $9, 
              notes = $10, updated_at = $11 
              WHERE id = $12`

	_, err := db.conn.ExecContext(
		ctx, query, lead.Name, lead.Email, lead.Phone, lead.Company, lead.Position,
		lead.Status, lead.IntentScore, lead.Tags, lead.Source, lead.Notes, lead.UpdatedAt, lead.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("error updating lead: %w", err)
	}

	return lead, nil
}

func (db *DB) DeleteLead(ctx context.Context, id string) (bool, error) {
	query := "DELETE FROM leads WHERE id = $1"

	result, err := db.conn.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("error deleting lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error getting rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func (db *DB) AssignLeadToAIAgent(ctx context.Context, leadID string, aiAgentID string) (*model.Lead, error) {
	tx, err := db.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback() 

	query := "INSERT INTO lead_ai_agent (lead_id, ai_agent_id, assigned_at) VALUES ($1, $2, $3)"
	_, err = tx.ExecContext(ctx, query, leadID, aiAgentID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error assigning lead to AI agent: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return db.GetLeadByID(ctx, leadID)
}

func (db *DB) GetInteractionsByLeadID(ctx context.Context, leadID string) ([]*model.Interaction, error) {
	query := `SELECT id, lead_id, type, channel, message, ai_agent_id, template_id, 
              timestamp, response, status, notes, created_at 
              FROM interactions WHERE lead_id = $1 ORDER BY timestamp DESC`

	rows, err := db.conn.QueryContext(ctx, query, leadID)
	if err != nil {
		return nil, fmt.Errorf("error querying interactions: %w", err)
	}
	defer rows.Close()

	var interactions []*model.Interaction
	for rows.Next() {
		var interaction model.Interaction
		var aiAgentID, templateID, message, response, notes sql.NullString

		err := rows.Scan(
			&interaction.ID, &leadID, &interaction.Type, &interaction.Channel,
			&message, &aiAgentID, &templateID, &interaction.Timestamp,
			&response, &interaction.Status, &notes, &interaction.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning interaction row: %w", err)
		}

		lead := &model.Lead{ID: leadID}
		interaction.Lead = lead

		if message.Valid {
			interaction.Message = &message.String
		}
		if response.Valid {
			interaction.Response = &response.String
		}
		if notes.Valid {
			interaction.Notes = &notes.String
		}

		interactions = append(interactions, &interaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating interaction rows: %w", err)
	}

	return interactions, nil
}

func (db *DB) GetClientByID(ctx context.Context, id string) (*model.Client, error) {
	query := `SELECT id, name, industry, website, contact_person, email, phone, 
              address, start_date, status, notes, created_at, updated_at 
              FROM clients WHERE id = $1`

	var client model.Client
	var updatedAt, website, phone, address, notes sql.NullString
	var updatedAtTime sql.NullTime

	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&client.ID, &client.Name, &client.Industry, &website, &client.ContactPerson, &client.Email,
		&phone, &address, &client.StartDate, &client.Status, &notes, &client.CreatedAt, &updatedAtTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching client: %w", err)
	}

	if website.Valid {
		client.Website = &website.String
	}
	if phone.Valid {
		client.Phone = &phone.String
	}
	if address.Valid {
		client.Address = &address.String
	}
	if notes.Valid {
		client.Notes = &notes.String
	}
	if updatedAtTime.Valid {
		client.UpdatedAt = &updatedAtTime.Time
	}

	return &client, nil
}

func (db *DB) GetClientsByStatus(ctx context.Context, status *model.ClientStatus, limit *int, offset *int) ([]*model.Client, error) {
	query := `SELECT id, name, industry, website, contact_person, email, phone, 
              address, start_date, status, notes, created_at, updated_at 
              FROM clients`

	var args []interface{}
	argCount := 1

	if status != nil {
		query += fmt.Sprintf(" WHERE status = $%d", argCount)
		args = append(args, *status)
		argCount++
	}

	query += " ORDER BY name ASC"
	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, *limit)
		argCount++
	}

	if offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, *offset)
	}

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying clients: %w", err)
	}
	defer rows.Close()

	var clients []*model.Client
	for rows.Next() {
		var client model.Client
		var website, phone, address, notes sql.NullString
		var updatedAt sql.NullTime

		err := rows.Scan(
			&client.ID, &client.Name, &client.Industry, &website, &client.ContactPerson,
			&client.Email, &phone, &address, &client.StartDate, &client.Status,
			&notes, &client.CreatedAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning client row: %w", err)
		}

		if website.Valid {
			client.Website = &website.String
		}
		if phone.Valid {
			client.Phone = &phone.String
		}
		if address.Valid {
			client.Address = &address.String
		}
		if notes.Valid {
			client.Notes = &notes.String
		}
		if updatedAt.Valid {
			client.UpdatedAt = &updatedAt.Time
		}

		clients = append(clients, &client)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client rows: %w", err)
	}

	return clients, nil
}

func (db *DB) CreateClient(ctx context.Context, client *model.Client) (*model.Client, error) {
	query := `INSERT INTO clients (name, industry, website, contact_person, email, phone, 
              address, start_date, status, notes, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
              RETURNING id`

	err := db.conn.QueryRowContext(
		ctx, query, client.Name, client.Industry, client.Website, client.ContactPerson,
		client.Email, client.Phone, client.Address, client.StartDate, client.Status,
		client.Notes, client.CreatedAt,
	).Scan(&client.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return client, nil
}

func (db *DB) AssignServicesToClient(ctx context.Context, clientID string, serviceIDs []string) error {
	tx, err := db.beginTx(ctx)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	query := "INSERT INTO client_service (client_id, service_id) VALUES ($1, $2)"
	for _, serviceID := range serviceIDs {
		_, err = tx.ExecContext(ctx, query, clientID, serviceID)
		if err != nil {
			return fmt.Errorf("error assigning service to client: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (db *DB) GetServicesByClientID(ctx context.Context, clientID string) ([]*model.Service, error) {
	query := `SELECT s.id, s.name, s.description, s.price, s.features, s.created_at, s.updated_at 
              FROM services s 
              JOIN client_service cs ON s.id = cs.service_id 
              WHERE cs.client_id = $1`

	rows, err := db.conn.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("error querying services: %w", err)
	}
	defer rows.Close()

	var services []*model.Service
	for rows.Next() {
		var service model.Service
		var featuresArray []string
		var updatedAt sql.NullTime

		err := rows.Scan(
			&service.ID, &service.Name, &service.Description, &service.Price,
			&featuresArray, &service.CreatedAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning service row: %w", err)
		}

		service.Features = featuresArray

		if updatedAt.Valid {
			service.UpdatedAt = &updatedAt.Time
		}

		services = append(services, &service)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service rows: %w", err)
	}

	return services, nil
}

func (db *DB) GetAIAgentByID(ctx context.Context, id string) (*model.AIAgent, error) {
	query := `SELECT id, name, purpose, description, status, last_run, created_at, updated_at 
              FROM ai_agents WHERE id = $1`

	var agent model.AIAgent
	var description sql.NullString
	var lastRun, updatedAt sql.NullTime

	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&agent.ID, &agent.Name, &agent.Purpose, &description, &agent.Status,
		&lastRun, &agent.CreatedAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching AI agent: %w", err)
	}

	if description.Valid {
		agent.Description = &description.String
	}
	if lastRun.Valid {
		agent.LastRun = &lastRun.Time
	}
	if updatedAt.Valid {
		agent.UpdatedAt = &updatedAt.Time
	}

	return &agent, nil
}

func (db *DB) GetLeadsByAIAgentID(ctx context.Context, aiAgentID string) ([]*model.Lead, error) {
	query := `SELECT l.id, l.name, l.email, l.phone, l.company, l.position, l.status, 
              l.intent_score, l.tags, l.source, l.last_contact, l.next_follow_up, 
              l.notes, l.created_at, l.updated_at 
              FROM leads l 
              JOIN lead_ai_agent laa ON l.id = laa.lead_id 
              WHERE laa.ai_agent_id = $1`

	rows, err := db.conn.QueryContext(ctx, query, aiAgentID)
	if err != nil {
		return nil, fmt.Errorf("error querying leads for AI agent: %w", err)
	}
	defer rows.Close()

	var leads []*model.Lead
	for rows.Next() {
		var lead model.Lead
		var tagsArray []sql.NullString
		var updatedAt sql.NullTime
		var lastContact, nextFollowUp sql.NullTime
		var phone, company, position, source, notes sql.NullString

		err := rows.Scan(
			&lead.ID, &lead.Name, &lead.Email, &phone, &company, &position,
			&lead.Status, &lead.IntentScore, &tagsArray, &source, &lastContact,
			&nextFollowUp, &notes, &lead.CreatedAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning lead row: %w", err)
		}

		if phone.Valid {
			lead.Phone = &phone.String
		}
		if company.Valid {
			lead.Company = &company.String
		}
		if position.Valid {
			lead.Position = &position.String
		}
		if source.Valid {
			lead.Source = &source.String
		}
		if notes.Valid {
			lead.Notes = &notes.String
		}
		if lastContact.Valid {
			lead.LastContact = &lastContact.Time
		}
		if nextFollowUp.Valid {
			lead.NextFollowUp = &nextFollowUp.Time
		}
		if updatedAt.Valid {
			lead.UpdatedAt = &updatedAt.Time
		}

		lead.Tags = make([]string, 0, len(tagsArray))
		for _, tag := range tagsArray {
			if tag.Valid {
				lead.Tags = append(lead.Tags, tag.String)
			}
		}

		leads = append(leads, &lead)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lead rows: %w", err)
	}

	return leads, nil
}

func (db *DB) GetAgentStats(ctx context.Context, aiAgentID string) (*model.AgentStats, error) {
	query := `SELECT id, agent_id, leads_engaged, messages_delivered, response_rate, 
              conversion_rate, avg_response_time, period, created_at 
              FROM agent_stats 
              WHERE agent_id = $1 
              ORDER BY created_at DESC LIMIT 1`

	var stats model.AgentStats

	err := db.conn.QueryRowContext(ctx, query, aiAgentID).Scan(
		&stats.ID, &aiAgentID, &stats.LeadsEngaged, &stats.MessagesDelivered,
		&stats.ResponseRate, &stats.ConversionRate, &stats.AvgResponseTime,
		&stats.Period, &stats.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			stats = model.AgentStats{
				AgentID:           aiAgentID,
				LeadsEngaged:      0,
				MessagesDelivered: 0,
				ResponseRate:      0,
				ConversionRate:    0,
				AvgResponseTime:   0,
				Period:            "all",
				CreatedAt:         time.Now(),
			}

			insertQuery := `INSERT INTO agent_stats 
                           (agent_id, leads_engaged, messages_delivered, response_rate, 
                           conversion_rate, avg_response_time, period, created_at) 
                           VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
                           RETURNING id`

			err = db.conn.QueryRowContext(
				ctx, insertQuery, stats.AgentID, stats.LeadsEngaged, stats.MessagesDelivered,
				stats.ResponseRate, stats.ConversionRate, stats.AvgResponseTime,
				stats.Period, stats.CreatedAt,
			).Scan(&stats.ID)

			if err != nil {
				return nil, fmt.Errorf("error creating default agent stats: %w", err)
			}

			agent := &model.AIAgent{ID: aiAgentID}
			stats.Agent = agent

			return &stats, nil
		}

		return nil, fmt.Errorf("error fetching agent stats: %w", err)
	}

	agent := &model.AIAgent{ID: aiAgentID}
	stats.Agent = agent

	return &stats, nil
}

func (db *DB) GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error) {
	query := `SELECT id, name, description, client_id, start_date, end_date, 
              status, budget, created_at, updated_at 
              FROM campaigns WHERE id = $1`

	var campaign model.Campaign
	var description, clientID sql.NullString
	var endDate, updatedAt sql.NullTime
	var budget sql.NullFloat64

	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&campaign.ID, &campaign.Name, &description, &clientID, &campaign.StartDate,
		&endDate, &campaign.Status, &budget, &campaign.CreatedAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil 
		}
		return nil, fmt.Errorf("error fetching campaign: %w", err)
	}

	if description.Valid {
		campaign.Description = &description.String
	}
	if clientID.Valid {
		campaign.ClientID = &clientID.String
	}
	if endDate.Valid {
		campaign.EndDate = &endDate.Time
	}
	if budget.Valid {
		campaign.Budget = &budget.Float64
	}
	if updatedAt.Valid {
		campaign.UpdatedAt = &updatedAt.Time
	}

	return &campaign, nil
}

func (db *DB) GetCampaignsByClientID(ctx context.Context, clientID string) ([]*model.Campaign, error) {
	query := `SELECT id, name, description, client_id, start_date, end_date, 
              status, budget, created_at, updated_at 
              FROM campaigns WHERE client_id = $1`

	rows, err := db.conn.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("error querying campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*model.Campaign
	for rows.Next() {
		var campaign model.Campaign
		var description sql.NullString
		var endDate, updatedAt sql.NullTime
		var budget sql.NullFloat64

		err := rows.Scan(
			&campaign.ID, &campaign.Name, &description, &clientID, &campaign.StartDate,
			&endDate, &campaign.Status, &budget, &campaign.CreatedAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning campaign row: %w", err)
		}

		campaign.ClientID = &clientID

		if description.Valid {
			campaign.Description = &description.String
		}
		if endDate.Valid {
			campaign.EndDate = &endDate.Time
		}
		if budget.Valid {
			campaign.Budget = &budget.Float64
		}
		if updatedAt.Valid {
			campaign.UpdatedAt = &updatedAt.Time
		}

		campaigns = append(campaigns, &campaign)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaign rows: %w", err)
	}

	return campaigns, nil
}

func (db *DB) GetTargetsByCampaignID(ctx context.Context, campaignID string) ([]*model.TargetAudience, error) {
	query := `SELECT id, name, industry, company_size, location, decision_maker_role, 
              pain_points, campaign_id, created_at, updated_at 
              FROM target_audiences WHERE campaign_id = $1`

	rows, err := db.conn.QueryContext(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("error querying target audiences: %w", err)
	}
	defer rows.Close()

	var targets []*model.TargetAudience
	for rows.Next() {
		var target model.TargetAudience
		var location, decisionMakerRole sql.NullString
		var painPoints []sql.NullString
		var updatedAt sql.NullTime

		err := rows.Scan(
			&target.ID, &target.Name, &target.Industry, &target.CompanySize,
			&location, &decisionMakerRole, &painPoints, &campaignID,
			&target.CreatedAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning target audience row: %w", err)
		}

		target.CampaignID = &campaignID

		if location.Valid {
			target.Location = &location.String
		}
		if decisionMakerRole.Valid {
			target.DecisionMakerRole = &decisionMakerRole.String
		}
		if updatedAt.Valid {
			target.UpdatedAt = &updatedAt.Time
		}

		target.PainPoints = make([]string, 0, len(painPoints))
		for _, point := range painPoints {
			if point.Valid {
				target.PainPoints = append(target.PainPoints, point.String)
			}
		}

		targets = append(targets, &target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating target audience rows: %w", err)
	}

	return targets, nil
}

func (db *DB) CreateTargetAudience(ctx context.Context, target *model.TargetAudience) (*model.TargetAudience, error) {
	query := `INSERT INTO target_audiences (name, industry, company_size, location, 
              decision_maker_role, pain_points, campaign_id, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
              RETURNING id`

	err := db.conn.QueryRowContext(
		ctx, query, target.Name, target.Industry, target.CompanySize,
		target.Location, target.DecisionMakerRole, target.PainPoints,
		target.CampaignID, target.CreatedAt,
	).Scan(&target.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating target audience: %w", err)
	}

	return target, nil
}

func (db *DB) UpdateTargetAudience(ctx context.Context, target *model.TargetAudience) (*model.TargetAudience, error) {
	query := `UPDATE target_audiences SET 
              name = $1, industry = $2, company_size = $3, location = $4,
              decision_maker_role = $5, pain_points = $6, updated_at = $7 
              WHERE id = $8`

	_, err := db.conn.ExecContext(
		ctx, query, target.Name, target.Industry, target.CompanySize,
		target.Location, target.DecisionMakerRole, target.PainPoints,
		target.UpdatedAt, target.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("error updating target audience: %w", err)
	}

	return target, nil
}

func (db *DB) DeleteTargetAudience(ctx context.Context, id string) (bool, error) {
	query := "DELETE FROM target_audiences WHERE id = $1"

	result, err := db.conn.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("error deleting target audience: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error getting rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

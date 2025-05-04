package graph

import (
	"context"
	"salesagency/graph/model"
	"salesagency/internal/database"
	"time"
)

type Resolver struct {
	DB *database.DB
}

func (r *Resolver) Lead() LeadResolver {
	return &leadResolver{r}
}

type leadResolver struct{ *Resolver }

func (r *leadResolver) Interactions(ctx context.Context, obj *model.Lead) ([]*model.Interaction, error) {
	return r.DB.GetInteractionsByLeadID(ctx, obj.ID)
}

func (r *Resolver) Client() ClientResolver {
	return &clientResolver{r}
}

type clientResolver struct{ *Resolver }

func (r *clientResolver) ActiveServices(ctx context.Context, obj *model.Client) ([]*model.Service, error) {
	return r.DB.GetServicesByClientID(ctx, obj.ID)
}

func (r *clientResolver) Campaigns(ctx context.Context, obj *model.Client) ([]*model.Campaign, error) {
	return r.DB.GetCampaignsByClientID(ctx, obj.ID)
}

func (r *Resolver) AIAgent() AIAgentResolver {
	return &aiAgentResolver{r}
}

type aiAgentResolver struct{ *Resolver }

func (r *aiAgentResolver) Leads(ctx context.Context, obj *model.AIAgent) ([]*model.Lead, error) {
	return r.DB.GetLeadsByAIAgentID(ctx, obj.ID)
}

func (r *aiAgentResolver) Campaigns(ctx context.Context, obj *model.AIAgent) ([]*model.Campaign, error) {
	return r.DB.GetCampaignsByAIAgentID(ctx, obj.ID)
}

func (r *aiAgentResolver) Templates(ctx context.Context, obj *model.AIAgent) ([]*model.MessageTemplate, error) {
	return r.DB.GetTemplatesByAIAgentID(ctx, obj.ID)
}

func (r *aiAgentResolver) Stats(ctx context.Context, obj *model.AIAgent) (*model.AgentStats, error) {
	return r.DB.GetAgentStats(ctx, obj.ID)
}

func (r *Resolver) Campaign() CampaignResolver {
	return &campaignResolver{r}
}

type campaignResolver struct{ *Resolver }

func (r *campaignResolver) Client(ctx context.Context, obj *model.Campaign) (*model.Client, error) {
	if obj.ClientID == nil {
		return nil, nil
	}
	return r.DB.GetClientByID(ctx, *obj.ClientID)
}

func (r *campaignResolver) Targets(ctx context.Context, obj *model.Campaign) ([]*model.TargetAudience, error) {
	return r.DB.GetTargetsByCampaignID(ctx, obj.ID)
}

func (r *campaignResolver) Messages(ctx context.Context, obj *model.Campaign) ([]*model.MessageTemplate, error) {
	return r.DB.GetTemplatesByCampaignID(ctx, obj.ID)
}

func (r *campaignResolver) AIAgents(ctx context.Context, obj *model.Campaign) ([]*model.AIAgent, error) {
	return r.DB.GetAIAgentsByCampaignID(ctx, obj.ID)
}

func (r *campaignResolver) Metrics(ctx context.Context, obj *model.Campaign) (*model.CampaignMetrics, error) {
	return r.DB.GetCampaignMetrics(ctx, obj.ID)
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateLead(ctx context.Context, input model.LeadInput) (*model.Lead, error) {
	lead := &model.Lead{
		Name:       input.Name,
		Email:      input.Email,
		Phone:      input.Phone,
		Company:    input.Company,
		Position:   input.Position,
		Tags:       input.Tags,
		Source:     input.Source,
		Notes:      input.Notes,
		CreatedAt:  time.Now(),
	}
	
	if input.Status != nil {
		lead.Status = *input.Status
	} else {
		defaultStatus := model.LeadStatusNew
		lead.Status = defaultStatus
	}
	
	if input.IntentScore != nil {
		lead.IntentScore = *input.IntentScore
	} else {
		lead.IntentScore = 0.5
	}
	
	return r.DB.CreateLead(ctx, lead)
}

func (r *mutationResolver) UpdateLead(ctx context.Context, id string, input model.LeadInput) (*model.Lead, error) {
	lead, err := r.DB.GetLeadByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	lead.Name = input.Name
	lead.Email = input.Email
	
	if input.Phone != nil {
		lead.Phone = input.Phone
	}
	if input.Company != nil {
		lead.Company = input.Company
	}
	if input.Position != nil {
		lead.Position = input.Position
	}
	if input.Status != nil {
		lead.Status = *input.Status
	}
	if input.IntentScore != nil {
		lead.IntentScore = *input.IntentScore
	}
	if input.Tags != nil {
		lead.Tags = input.Tags
	}
	if input.Source != nil {
		lead.Source = input.Source
	}
	if input.Notes != nil {
		lead.Notes = input.Notes
	}
	
	lead.UpdatedAt = &time.Time{}
	*lead.UpdatedAt = time.Now()
	
	return r.DB.UpdateLead(ctx, lead)
}

func (r *mutationResolver) DeleteLead(ctx context.Context, id string) (bool, error) {
	return r.DB.DeleteLead(ctx, id)
}

func (r *mutationResolver) AssignLeadToAIAgent(ctx context.Context, leadID string, aiAgentID string) (*model.Lead, error) {
	return r.DB.AssignLeadToAIAgent(ctx, leadID, aiAgentID)
}

func (r *mutationResolver) CreateClient(ctx context.Context, input model.ClientInput) (*model.Client, error) {
	client := &model.Client{
		Name:          input.Name,
		Industry:      input.Industry,
		Website:       input.Website,
		ContactPerson: input.ContactPerson,
		Email:         input.Email,
		Phone:         input.Phone,
		Address:       input.Address,
		StartDate:     input.StartDate,
		Notes:         input.Notes,
		CreatedAt:     time.Now(),
	}
	
	if input.Status != nil {
		client.Status = *input.Status
	} else {
		defaultStatus := model.ClientStatusActive
		client.Status = defaultStatus
	}
	
	newClient, err := r.DB.CreateClient(ctx, client)
	if err != nil {
		return nil, err
	}
	
	if input.ServiceIds != nil {
		err = r.DB.AssignServicesToClient(ctx, newClient.ID, input.ServiceIds)
		if err != nil {
			return nil, err
		}
	}
	
	return newClient, nil
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Lead(ctx context.Context, id string) (*model.Lead, error) {
	return r.DB.GetLeadByID(ctx, id)
}

func (r *queryResolver) Leads(ctx context.Context, filter *model.LeadFilterInput, limit *int, offset *int) ([]*model.Lead, error) {
	return r.DB.GetLeadsByFilter(ctx, filter, limit, offset)
}

func (r *queryResolver) Client(ctx context.Context, id string) (*model.Client, error) {
	return r.DB.GetClientByID(ctx, id)
}

func (r *queryResolver) Clients(ctx context.Context, status *model.ClientStatus, limit *int, offset *int) ([]*model.Client, error) {
	return r.DB.GetClientsByStatus(ctx, status, limit, offset)
}

func (r *queryResolver) AIAgent(ctx context.Context, id string) (*model.AIAgent, error) {
	return r.DB.GetAIAgentByID(ctx, id)
}

func (r *queryResolver) AIAgents(ctx context.Context, status *model.AgentStatus, purpose *string, limit *int, offset *int) ([]*model.AIAgent, error) {
	return r.DB.GetAIAgentsByFilter(ctx, status, purpose, limit, offset)
}

func (r *queryResolver) Campaign(ctx context.Context, id string) (*model.Campaign, error) {
	return r.DB.GetCampaignByID(ctx, id)
}

func (r *queryResolver) Campaigns(ctx context.Context, filter *model.CampaignFilterInput, limit *int, offset *int) ([]*model.Campaign, error) {
	return r.DB.GetCampaignsByFilter(ctx, filter, limit, offset)
}

func (r *mutationResolver) TriggerAIAgentRun(ctx context.Context, id string) (bool, error) {
	return r.DB.TriggerAIAgentRun(ctx, id)
}

func (r *mutationResolver) PauseAIAgent(ctx context.Context, id string) (bool, error) {
	return r.DB.UpdateAIAgentStatus(ctx, id, model.AgentStatusPaused)
}

func (r *mutationResolver) ResumeAIAgent(ctx context.Context, id string) (bool, error) {
	return r.DB.UpdateAIAgentStatus(ctx, id, model.AgentStatusActive)
}
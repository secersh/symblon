package store

import (
	"context"
	"fmt"
	"sync"

	"undercover/internal/overseer/agent"
	"undercover/internal/overseer/evaluation"
)

// InMemoryAgentStore is a thread-safe in-memory implementation of AgentStore
// intended for development and testing.
type InMemoryAgentStore struct {
	mu     sync.RWMutex
	agents map[string]*agent.Agent
}

// NewInMemoryAgentStore returns an initialised in-memory agent store.
func NewInMemoryAgentStore() *InMemoryAgentStore {
	return &InMemoryAgentStore{agents: make(map[string]*agent.Agent)}
}

func (s *InMemoryAgentStore) Save(_ context.Context, a *agent.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[a.ID] = a
	return nil
}

func (s *InMemoryAgentStore) Get(_ context.Context, id string) (*agent.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[id]
	if !ok {
		return nil, fmt.Errorf("agent %q not found", id)
	}
	return a, nil
}

func (s *InMemoryAgentStore) List(_ context.Context) ([]*agent.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agent.Agent, 0, len(s.agents))
	for _, a := range s.agents {
		result = append(result, a)
	}
	return result, nil
}

func (s *InMemoryAgentStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agents, id)
	return nil
}

func (s *InMemoryAgentStore) ListByRoutingKey(_ context.Context, routingKey string) ([]*agent.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*agent.Agent
	for _, a := range s.agents {
		if a.Trigger.RoutingKey == routingKey {
			result = append(result, a)
		}
	}
	return result, nil
}

// InMemoryEvaluationStore is a thread-safe in-memory implementation of
// EvaluationStore intended for development and testing.
type InMemoryEvaluationStore struct {
	mu          sync.RWMutex
	evaluations []*evaluation.Evaluation
}

// NewInMemoryEvaluationStore returns an initialised in-memory evaluation store.
func NewInMemoryEvaluationStore() *InMemoryEvaluationStore {
	return &InMemoryEvaluationStore{}
}

func (s *InMemoryEvaluationStore) Append(_ context.Context, e *evaluation.Evaluation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evaluations = append(s.evaluations, e)
	return nil
}

func (s *InMemoryEvaluationStore) ListByAgent(_ context.Context, agentID string) ([]*evaluation.Evaluation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*evaluation.Evaluation
	for i := len(s.evaluations) - 1; i >= 0; i-- {
		if s.evaluations[i].AgentID == agentID {
			result = append(result, s.evaluations[i])
		}
	}
	return result, nil
}

func (s *InMemoryEvaluationStore) ListByActor(_ context.Context, actorLogin string) ([]*evaluation.Evaluation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*evaluation.Evaluation
	for i := len(s.evaluations) - 1; i >= 0; i-- {
		if s.evaluations[i].ActorLogin == actorLogin {
			result = append(result, s.evaluations[i])
		}
	}
	return result, nil
}

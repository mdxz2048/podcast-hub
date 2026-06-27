package content

import (
	"context"
	"fmt"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) ListStagingPrograms(ctx context.Context) ([]Program, error) {
	return s.store.ListStagingPrograms(ctx)
}

func (s *Service) GetProgram(ctx context.Context, id string) (Program, error) {
	item, found, err := s.store.GetProgram(ctx, id)
	if err != nil {
		return Program{}, err
	}
	if !found {
		return Program{}, fmt.Errorf("program not found")
	}
	return item, nil
}

func (s *Service) ListStagingEpisodes(ctx context.Context) ([]Episode, error) {
	return s.store.ListStagingEpisodes(ctx)
}

func (s *Service) GetEpisode(ctx context.Context, id string) (Episode, error) {
	item, found, err := s.store.GetEpisode(ctx, id)
	if err != nil {
		return Episode{}, err
	}
	if !found {
		return Episode{}, fmt.Errorf("episode not found")
	}
	return item, nil
}

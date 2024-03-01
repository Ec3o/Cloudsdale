package service

import (
	"crypto"
	"encoding/hex"
	"github.com/elabosak233/cloudsdale/internal/model"
	"github.com/elabosak233/cloudsdale/internal/model/dto/request"
	"github.com/elabosak233/cloudsdale/internal/model/dto/response"
	"github.com/elabosak233/cloudsdale/internal/repository"
	"github.com/mitchellh/mapstructure"
	"math"
)

type IGameService interface {
	Find(req request.GameFindRequest) (games []response.GameResponse, pageCount int64, total int64, err error)
	Create(req request.GameCreateRequest) (err error)
	Update(req request.GameUpdateRequest) (err error)
	Delete(req request.GameDeleteRequest) (err error)
}

type GameService struct {
	gameRepository repository.IGameRepository
}

func NewGameService(appRepository *repository.Repository) IGameService {
	return &GameService{
		gameRepository: appRepository.GameRepository,
	}
}

func (g *GameService) Find(req request.GameFindRequest) (games []response.GameResponse, pageCount int64, total int64, err error) {
	games, count, err := g.gameRepository.Find(req)
	if req.Size >= 1 && req.Page >= 1 {
		pageCount = int64(math.Ceil(float64(count) / float64(req.Size)))
	} else {
		pageCount = 1
	}
	return games, pageCount, count, err
}

func (g *GameService) Create(req request.GameCreateRequest) (err error) {
	game := model.Game{}
	err = mapstructure.Decode(req, &game)
	if req.Password != "" {
		hasher := crypto.SHA256.New()
		hasher.Write([]byte(req.Password))
		hashBytes := hasher.Sum(nil)
		game.Password = hex.EncodeToString(hashBytes)
	}
	_, err = g.gameRepository.Insert(game)
	return err
}

func (g *GameService) Update(req request.GameUpdateRequest) (err error) {
	game := model.Game{}
	err = mapstructure.Decode(req, &game)
	if req.Password != "" {
		hasher := crypto.SHA256.New()
		hasher.Write([]byte(req.Password))
		hashBytes := hasher.Sum(nil)
		game.Password = hex.EncodeToString(hashBytes)
	}
	err = g.gameRepository.Update(game)
	return err
}

func (g *GameService) Delete(req request.GameDeleteRequest) (err error) {
	return g.gameRepository.Delete(req)
}

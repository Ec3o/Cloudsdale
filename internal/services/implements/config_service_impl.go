package implements

import (
	"github.com/elabosak233/pgshub/internal/repositorys"
	"github.com/elabosak233/pgshub/internal/services"
	"github.com/spf13/viper"
)

type ConfigServiceImpl struct {
}

func NewConfigServiceImpl(appRepository *repositorys.AppRepository) services.ConfigService {
	return &ConfigServiceImpl{}
}

func (c ConfigServiceImpl) FindAll() map[string]any {
	return viper.GetStringMap("Global")
}
package service

import (
	"errors"
	"fmt"
	"github.com/elabosak233/pgshub/internal/config"
	"github.com/elabosak233/pgshub/internal/container/manager"
	"github.com/elabosak233/pgshub/internal/model"
	"github.com/elabosak233/pgshub/internal/model/dto/request"
	"github.com/elabosak233/pgshub/internal/model/dto/response"
	"github.com/elabosak233/pgshub/internal/repository"
	"github.com/elabosak233/pgshub/pkg/convertor"
	"github.com/elabosak233/pgshub/pkg/generator"
	"strings"
	"sync"
	"time"
)

var (
	// UserPodRequestMap 用于存储用户上次请求的时间
	UserPodRequestMap = struct {
		sync.RWMutex
		m map[uint]int64
	}{m: make(map[uint]int64)}

	// ContainerManagerPtrMap is a mapping of ID and manager pointer.
	ContainerManagerPtrMap = make(map[any]interface{})

	// PodMap is a mapping of IDs and ID
	PodMap = make(map[uint][]uint)
)

// GetUserInstanceRequestMap 返回用户上次请求的时间
func GetUserInstanceRequestMap(userID uint) int64 {
	UserPodRequestMap.RLock()
	defer UserPodRequestMap.RUnlock()
	return UserPodRequestMap.m[userID]
}

// SetUserInstanceRequestMap 设置用户上次请求的时间
func SetUserInstanceRequestMap(userID uint, t int64) {
	UserPodRequestMap.Lock()
	defer UserPodRequestMap.Unlock()
	UserPodRequestMap.m[userID] = t
}

type IPodService interface {
	Create(req request.PodCreateRequest) (res response.PodStatusResponse, err error)
	Status(id uint) (rep response.PodStatusResponse, err error)
	Renew(req request.PodRenewRequest) (removedAt int64, err error)
	Remove(req request.PodRemoveRequest) (err error)
	FindById(id uint) (rep response.PodResponse, err error)
	Find(req request.PodFindRequest) (rep []response.PodResponse, err error)
}

type PodService struct {
	MixinService        IMixinService
	ChallengeRepository repository.IChallengeRepository
	PodRepository       repository.IPodRepository
	NatRepository       repository.INatRepository
	FlagGenRepository   repository.IFlagGenRepository
	InstanceRepository  repository.IInstanceRepository
}

func NewPodService(appRepository *repository.Repository) IPodService {
	return &PodService{
		MixinService:        NewMixinService(appRepository),
		ChallengeRepository: appRepository.ChallengeRepository,
		InstanceRepository:  appRepository.ContainerRepository,
		FlagGenRepository:   appRepository.FlagGenRepository,
		PodRepository:       appRepository.PodRepository,
		NatRepository:       appRepository.NatRepository,
	}
}

func (t *PodService) IsLimited(userID uint, limit int64) (remainder int64) {
	if userID == 0 {
		return 0
	}
	ti := GetUserInstanceRequestMap(userID)
	if ti != 0 {
		if time.Now().Unix()-ti < limit {
			return limit - (time.Now().Unix() - ti)
		}
	}
	return 0
}

func (t *PodService) Create(req request.PodCreateRequest) (res response.PodStatusResponse, err error) {
	remainder := t.IsLimited(req.UserID, int64(config.AppCfg().Global.Container.RequestLimit))
	if remainder != 0 {
		return res, errors.New(fmt.Sprintf("请等待 %d 秒后再次请求", remainder))
	}
	switch config.AppCfg().Container.Provider {
	case "docker":
		SetUserInstanceRequestMap(req.UserID, time.Now().Unix()) // 保存用户请求时间
		challenges, _, _ := t.ChallengeRepository.Find(request.ChallengeFindRequest{
			IDs:       []uint{req.ChallengeID},
			IsDynamic: convertor.TrueP(),
		})
		challenges, _ = t.MixinService.MixinChallenge(challenges)
		challenge := challenges[0]
		isGame := req.GameID != nil && req.TeamID != nil

		// Parallel container limit
		if config.AppCfg().Global.Container.ParallelLimit > 0 {
			var availablePods []model.Pod
			var count int64
			if !isGame {
				availablePods, count, _ = t.PodRepository.Find(request.PodFindRequest{
					UserID:      req.UserID,
					IsAvailable: convertor.TrueP(),
				})
			} else {
				availablePods, count, _ = t.PodRepository.Find(request.PodFindRequest{
					TeamID:      req.TeamID,
					GameID:      req.GameID,
					IsAvailable: convertor.TrueP(),
				})
			}
			needToBeDeactivated := count - int64(config.AppCfg().Global.Container.ParallelLimit) + 1
			if needToBeDeactivated > 0 {
				for _, pod := range availablePods {
					if needToBeDeactivated == 0 {
						break
					}
					go func() {
						_ = t.Remove(request.PodRemoveRequest{
							ID: pod.ID,
						})
					}()
					needToBeDeactivated -= 1
				}
			}
		}

		var ctnMap = make(map[uint]model.Instance)

		removedAt := time.Now().Add(time.Duration(challenge.Duration) * time.Second).Unix()

		// Insert Pod model, get Pod's ID
		pod, _ := t.PodRepository.Insert(model.Pod{
			ChallengeID: req.ChallengeID,
			UserID:      req.UserID,
			RemovedAt:   removedAt,
		})

		if _, ok := PodMap[pod.ID]; !ok {
			PodMap[pod.ID] = make([]uint, 0)
		}

		// Select the first one as the target flag which will be injected
		var flag model.Flag
		var flagStr string
		for _, f := range challenge.Flags {
			if f.Type == "dynamic" {
				flag = *f
				flagStr = generator.GenerateFlag(flag.Value)
			} else if f.Type == "static" {
				flag = *f
				flagStr = f.Value
			}
		}

		_, _ = t.FlagGenRepository.Insert(model.FlagGen{
			Flag:  flagStr,
			PodID: pod.ID,
		})

		for _, image := range challenge.Images {
			var envs = make([]*model.Env, 0)
			for _, e := range image.Envs {
				ee := *e
				envs = append(envs, &ee)
			}

			// This Flag Env is only a temporary model. It will not be persisted.
			envs = append(envs, &model.Env{
				Key:   flag.Env,
				Value: flagStr,
			})

			ctnManager := manager.NewDockerManager(
				image.Name,
				image.Ports,
				envs,
				image.MemoryLimit,
				image.CPULimit,
				time.Duration(challenge.Duration)*time.Second,
			)
			assignedPorts, _ := ctnManager.Setup()

			// Instance -> Pod
			container, _ := t.InstanceRepository.Insert(model.Instance{
				PodID:   pod.ID,
				ImageID: image.ID,
			})
			PodMap[pod.ID] = append(PodMap[pod.ID], container.ID)

			ctn := ctnMap[container.ID]
			for src, dsts := range assignedPorts {
				srcPort := convertor.ToIntD(strings.Split(string(src), "/")[0], 0)
				for _, dst := range dsts {
					dstPort := convertor.ToIntD(dst.HostPort, 0)
					entry := fmt.Sprintf(
						"%s:%d",
						config.AppCfg().Container.Docker.PublicEntry,
						dstPort,
					)
					// Nat -> Instance
					nat, _ := t.NatRepository.Insert(model.Nat{
						InstanceID: container.ID,
						SrcPort:    srcPort,
						DstPort:    dstPort,
						Entry:      entry,
					})
					ctn.Nats = append(ctn.Nats, &nat)
				}
			}
			ctnMap[container.ID] = ctn

			ctnManager.SetInstanceID(container.ID)
			ContainerManagerPtrMap[container.ID] = ctnManager

			// Start removal plan
			go func() {
				if ctnManager.RemoveAfterDuration(ctnManager.CancelCtx) {
					delete(ContainerManagerPtrMap, container.ID)
				}
			}()
		}
		var ctns []*model.Instance
		for _, ctn := range ctnMap {
			cctn := ctn
			ctns = append(ctns, &cctn)
		}
		return response.PodStatusResponse{
			ID:        pod.ID,
			Instances: ctns,
			RemovedAt: removedAt,
		}, err
	case "k8s":
		SetUserInstanceRequestMap(req.UserID, time.Now().Unix()) // 保存用户请求时间
		challenges, _, _ := t.ChallengeRepository.Find(request.ChallengeFindRequest{
			IDs:       []uint{req.ChallengeID},
			IsDynamic: convertor.TrueP(),
		})
		challenges, _ = t.MixinService.MixinChallenge(challenges)
		challenge := challenges[0]
		isGame := req.GameID != nil && req.TeamID != nil

		// Parallel container limit
		if config.AppCfg().Global.Container.ParallelLimit > 0 {
			var availablePods []model.Pod
			var count int64
			if !isGame {
				availablePods, count, _ = t.PodRepository.Find(request.PodFindRequest{
					UserID:      req.UserID,
					IsAvailable: convertor.TrueP(),
				})
			} else {
				availablePods, count, _ = t.PodRepository.Find(request.PodFindRequest{
					TeamID:      req.TeamID,
					GameID:      req.GameID,
					IsAvailable: convertor.TrueP(),
				})
			}
			needToBeDeactivated := count - int64(config.AppCfg().Global.Container.ParallelLimit) + 1
			if needToBeDeactivated > 0 {
				for _, pod := range availablePods {
					if needToBeDeactivated == 0 {
						break
					}
					go func() {
						_ = t.Remove(request.PodRemoveRequest{
							ID: pod.ID,
						})
					}()
					needToBeDeactivated -= 1
				}
			}
		}

		var ctnMap = make(map[int64]model.Instance)

		removedAt := time.Now().Add(time.Duration(challenge.Duration) * time.Second).Unix()

		// Insert Pod model, get Pod's ID
		pod, _ := t.PodRepository.Insert(model.Pod{
			ChallengeID: req.ChallengeID,
			UserID:      req.UserID,
			RemovedAt:   removedAt,
		})

		if _, ok := PodMap[pod.ID]; !ok {
			PodMap[pod.ID] = make([]uint, 0)
		}

		// Select the first one as the target flag which will be injected
		var flag model.Flag
		var flagStr string
		for _, f := range challenge.Flags {
			if f.Type == "dynamic" {
				flag = *f
				flagStr = generator.GenerateFlag(flag.Value)
			} else if f.Type == "static" {
				flag = *f
				flagStr = f.Value
			}
		}

		_, _ = t.FlagGenRepository.Insert(model.FlagGen{
			Flag:  flagStr,
			PodID: pod.ID,
		})

		ctnManager := manager.NewK8sManager(
			pod.ID,
			challenge.Images,
			flag,
			time.Duration(challenge.Duration),
		)

		instances, _ := ctnManager.Setup()

		for _, instance := range instances {
			PodMap[pod.ID] = append(PodMap[pod.ID], instance.ID)
			for _, nat := range instance.Nats {
				nat.InstanceID = instance.ID
				_, _ = t.NatRepository.Insert(*nat)
			}
			_, _ = t.InstanceRepository.Insert(*instance)
		}
		go func() {
			if ctnManager.RemoveAfterDuration(ctnManager.CancelCtx) {
				delete(ContainerManagerPtrMap, pod.ID)
			}
		}()
		var ctns []model.Instance
		for _, ctn := range ctnMap {
			ctns = append(ctns, ctn)
		}
		return response.PodStatusResponse{
			ID:        pod.ID,
			Instances: instances,
			RemovedAt: removedAt,
		}, err
	}
	return res, errors.New("创建失败")
}

func (t *PodService) Status(podID uint) (rep response.PodStatusResponse, err error) {
	rep = response.PodStatusResponse{}
	if config.AppCfg().Container.Provider == "docker" {
		instance, err := t.PodRepository.FindById(podID)
		if ContainerManagerPtrMap[podID] != nil {
			ctn := ContainerManagerPtrMap[podID].(*manager.DockerManager)
			status, _ := ctn.GetContainerStatus()
			if status != "removed" {
				rep.ID = podID
				rep.Status = status
				//rep.Entry = instance.Entry
				rep.RemovedAt = instance.RemovedAt
				return rep, nil
			}
		}
		rep.Status = "removed"
		return rep, err
	}
	return rep, errors.New("获取失败")
}

func (t *PodService) Renew(req request.PodRenewRequest) (removedAt int64, err error) {
	remainder := t.IsLimited(req.UserID, int64(config.AppCfg().Global.Container.RequestLimit))
	if remainder != 0 {
		return 0, errors.New(fmt.Sprintf("请等待 %d 秒后再次请求", remainder))
	}
	if config.AppCfg().Container.Provider == "docker" {
		SetUserInstanceRequestMap(req.UserID, time.Now().Unix()) // 保存用户请求时间
		pod, err := t.PodRepository.FindById(req.ID)
		if err != nil || PodMap[req.ID] == nil {
			return 0, errors.New("实例不存在")
		}
		var duration time.Duration
		for _, ctnID := range PodMap[req.ID] {
			if ContainerManagerPtrMap[ctnID] != nil {
				ctn := ContainerManagerPtrMap[ctnID].(*manager.DockerManager)
				duration = ctn.Duration
				ctn.Renew(duration)
			}
		}
		pod.RemovedAt = time.Now().Add(duration).Unix()
		err = t.PodRepository.Update(pod)
		return pod.RemovedAt, err
	}
	return 0, errors.New("续期失败")
}

func (t *PodService) Remove(req request.PodRemoveRequest) (err error) {
	remainder := t.IsLimited(req.UserID, int64(config.AppCfg().Global.Container.RequestLimit))
	if remainder != 0 {
		return errors.New(fmt.Sprintf("请等待 %d 秒后再次请求", remainder))
	}
	if config.AppCfg().Container.Provider == "docker" {
		_ = t.PodRepository.Update(model.Pod{
			ID:        req.ID,
			RemovedAt: time.Now().Unix(),
		})
		for _, ctnID := range PodMap[req.ID] {
			if ContainerManagerPtrMap[ctnID] != nil {
				ctn := ContainerManagerPtrMap[ctnID].(*manager.DockerManager)
				ctn.Remove()
			}
		}
		return err
	}
	go func() {
		delete(ContainerManagerPtrMap, req.ID)
	}()
	return errors.New("移除失败")
}

func (t *PodService) FindById(id uint) (rep response.PodResponse, err error) {
	if config.AppCfg().Container.Provider == "docker" {
		instance, err := t.PodRepository.FindById(id)
		if err != nil || ContainerManagerPtrMap[id] == nil {
			return rep, errors.New("实例不存在")
		}
		ctn := ContainerManagerPtrMap[id].(*manager.DockerManager)
		status, _ := ctn.GetContainerStatus()
		rep = response.PodResponse{
			ID:          id,
			RemovedAt:   instance.RemovedAt,
			ChallengeID: instance.ChallengeID,
			Status:      status,
		}
		return rep, nil
	}
	return rep, errors.New("获取失败")
}

func (t *PodService) Find(req request.PodFindRequest) (pods []response.PodResponse, err error) {
	if config.AppCfg().Container.Provider == "docker" {
		if req.TeamID != nil && req.GameID != nil {
			req.UserID = 0
		}
		podResponse, _, err := t.PodRepository.Find(req)
		podResponse, err = t.MixinService.MixinPod(podResponse)
		for _, pod := range podResponse {
			var ctn *manager.DockerManager
			status := "removed"
			for _, ctnID := range PodMap[pod.ID] {
				if ContainerManagerPtrMap[ctnID] != nil {
					ctn = ContainerManagerPtrMap[ctnID].(*manager.DockerManager)
					s, _ := ctn.GetContainerStatus()
					if s == "running" {
						status = s
					}
				}
			}
			pods = append(pods, response.PodResponse{
				ID:          pod.ID,
				RemovedAt:   pod.RemovedAt,
				Instances:   pod.Instances,
				ChallengeID: pod.ChallengeID,
				Status:      status,
			})
		}
		return pods, err
	}
	return nil, errors.New("获取失败")
}

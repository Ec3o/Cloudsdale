package entity

type Container struct {
	ContainerID int64 `xorm:"'id' pk autoincr" json:"id,omitempty"`
	ImageID     int64 `xorm:"'image_id' notnull" json:"image_id,omitempty"`
	ChallengeID int64 `xorm:"'challenge_id' notnull" json:"challenge_id,omitempty"`
	PodID       int64 `xorm:"'pod_id' notnull" json:"pod_id,omitempty"`

	Nats  []Nat `xorm:"-" json:"nats,omitempty"`
	Image Image `xorm:"-" json:"image,omitempty"`
}

func (c *Container) TableName() string {
	return "container"
}
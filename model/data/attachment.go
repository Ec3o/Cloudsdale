/*
题目附件
*/

package data

import "time"

type Attachment struct {
	Id          string    `xorm:"pk unique 'id' notnull" json:"id"`
	RemoteUrl   string    `xorm:"text 'remote_url'" json:"remote_url"`
	LocalFileId string    `xorm:"text 'local_file_id'" json:"local_file_id"`
	CreatedAt   time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}
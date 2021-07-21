package db

import (
	"example/internal/support"
	"github.com/GabrielCarpr/cqrs/bus"
	"example/users/entities"
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ UserRepository = (*DBUserRepository)(nil)

func fromUser(u entities.User) dbUser {
	roles := make([]dbRoleID, len(u.RoleIDs))
	for i, id := range u.RoleIDs {
		roles[i] = dbRoleID{u.ID, id.String()}
	}
	var last sql.NullTime
	if u.LastSignedIn != nil {
		last = sql.NullTime{Time: *u.LastSignedIn, Valid: true}
	} else {
		last = sql.NullTime{Time: time.Time{}, Valid: false}
	}

	return dbUser{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email.String(),
		RoleIds:      roles,
		Hash:         u.Hash,
		Active:       u.Active,
		LastSignedIn: last,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Version: u.CurrentVersion(),
	}
}

type dbRoleID struct {
	UserId support.ID `gorm:"primaryKey;foreignKey:ID"`
	RoleId string     `gorm:"primaryKey;foreignKey:ID"`
}

func (d dbRoleID) TableName() string {
	return "user_roles"
}

type dbUser struct {
	ID           support.ID `gorm:"primaryKey;column:id"`
	Name         string
	Email        string     `gorm:"unique"`
	RoleIds      []dbRoleID `gorm:"foreignKey:user_id;"`
	Hash         string
	Active       bool
	LastSignedIn sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version		 int64
}

func (d dbUser) TableName() string {
	return "users"
}

func (d dbUser) User() entities.User {
	roles := make([]entities.RoleID, len(d.RoleIds))
	for i, r := range d.RoleIds {
		roles[i] = entities.NewRoleID(r.RoleId)
	}
	email, _ := support.NewEmail(d.Email)
	var time *time.Time
	if d.LastSignedIn.Valid {
		time = &d.LastSignedIn.Time
	} else {
		time = nil
	}

	u := entities.User{
		ID:           d.ID,
		Name:         d.Name,
		Email:        email,
		RoleIDs:      roles,
		Hash:         d.Hash,
		Active:       d.Active,
		LastSignedIn: time,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		EventBuffer:   bus.NewEventBuffer(d.ID),
	}
	u.ForceVersion(d.Version)
	return u
}

func NewDBUserRepository(db *gorm.DB) *DBUserRepository {
	return &DBUserRepository{db}
}

type DBUserRepository struct {
	db *gorm.DB
}

func (r DBUserRepository) Persist(users ...entities.User) error {
	records := make([]dbUser, len(users))
	for i, u := range users {
		records[i] = fromUser(u)
	}

	res := r.db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&records, 300)
	return support.TransduceError(res.Error)
}

func (r DBUserRepository) Find(ids ...support.ID) ([]entities.User, error) {
	var records []dbUser
	res := r.db.Preload("RoleIds").Find(&records, "id IN(?)", ids)
	if res.Error != nil {
		return []entities.User{}, support.TransduceError(res.Error)
	}

	users := make([]entities.User, len(records))
	for i, record := range records {
		users[i] = record.User()
	}
	return users, nil
}

func (r DBUserRepository) Delete(ids ...support.ID) error {
	panic("Not implemented")
}

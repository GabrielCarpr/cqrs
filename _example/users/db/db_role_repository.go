package db

import (
	"example/users/entities"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func fromRole(role entities.Role) dbRole {
	scopes := scopes(role)
	return dbRole{
		ID:     role.ID.String(),
		Label:  role.Label,
		Scopes: scopes,
		Version: role.CurrentVersion(),
	}
}

func scopes(roles ...entities.Role) []dbScope {
	var scopes []dbScope
	for _, role := range roles {
		for _, s := range role.Scopes() {
			scopes = append(scopes, dbScope{Name: s.Name})
		}
	}

	return scopes
}

type dbRole struct {
	ID    string `gorm:"primaryKey"`
	Label string
	Version int64

	Scopes []dbScope `gorm:"many2many:role_scopes;joinForeignKey:role_id;joinReferences:scope_id"`
}

func (dbRole) TableName() string {
	return "roles"
}

func (r dbRole) Role() entities.Role {
	role := entities.BuildRole(
		r.ID,
		r.Label,
	)
	for _, s := range r.Scopes {
		role = role.ApplyScopes(s.Name)
	}
	role.ForceVersion(r.Version)
	return role
}

type dbScope struct {
	Name string `gorm:"primaryKey"`
}

func (s dbScope) Scope() entities.Scope {
	return entities.Scope{Name: s.Name}
}

func (dbScope) TableName() string {
	return "scopes"
}

func NewDBRoleRepository(db *gorm.DB) *DBRoleRepository {
	return &DBRoleRepository{db}
}

type DBRoleRepository struct {
	db *gorm.DB
}

func (r DBRoleRepository) Persist(roles ...entities.Role) error {
	err := r.persistScopes(roles...)
	if err != nil {
		return fmt.Errorf("persistScopes: %w", err)
	}

	records := make([]dbRole, len(roles))
	for i, role := range roles {
		records[i] = fromRole(role)
	}

	res := r.db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(records, 50)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r DBRoleRepository) persistScopes(roles ...entities.Role) error {
	records := scopes(roles...)

	res := r.db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(records, 50)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r DBRoleRepository) Find(ids ...entities.RoleID) ([]entities.Role, error) {
	strIds := make([]string, len(ids))
	for i, id := range ids {
		strIds[i] = id.String()
	}

	var records []dbRole
	res := r.db.Preload("Scopes").Find(&records, strIds)
	if res.Error != nil {
		return []entities.Role{}, res.Error
	}

	roles := make([]entities.Role, len(records))
	for i, record := range records {
		roles[i] = record.Role()
	}
	return roles, nil
}

func (r DBRoleRepository) All() ([]entities.Role, error) {
	var records []dbRole
	res := r.db.Preload("Scopes").Find(&records)
	if res.Error != nil {
		return []entities.Role{}, res.Error
	}

	roles := make([]entities.Role, len(records))
	for i, record := range records {
		roles[i] = record.Role()
	}
	return roles, nil
}

package db

import "example/users/entities"

// RoleRepository handles storage of roles
type RoleRepository interface {
	Persist(roles ...entities.Role) error
	Find(ids ...entities.RoleID) ([]entities.Role, error)
	All() ([]entities.Role, error)
}

func NewMemoryRoleRepository() *MemoryRoleRepository {
	r := &MemoryRoleRepository{}
	r.roles = make(map[entities.RoleID]entities.Role)
	r.Persist(entities.CreateRole("User"))
	r.Persist(entities.CreateRole("Admin"))
	return r
}

type MemoryRoleRepository struct {
	roles map[entities.RoleID]entities.Role
}

func (r *MemoryRoleRepository) Persist(roles ...entities.Role) error {
	for _, role := range roles {
		r.roles[role.ID] = role
	}
	return nil
}

func (r *MemoryRoleRepository) Find(ids ...entities.RoleID) ([]entities.Role, error) {
	result := []entities.Role{}
	for _, id := range ids {
		role, ok := r.roles[id]
		if !ok {
			continue
		}
		result = append(result, role)
	}
	return result, nil
}

func (r *MemoryRoleRepository) All() ([]entities.Role, error) {
	results := []entities.Role{}
	for _, role := range r.roles {
		results = append(results, role)
	}
	return results, nil
}

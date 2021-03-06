package db

import (
	"example/internal/support"
	"example/users/entities"
)

type UserRepository interface {
	Persist(users ...entities.User) error
	Find(ids ...support.ID) ([]entities.User, error)
	Delete(ids ...support.ID) error
}

var _ UserRepository = (*MemoryUserRepository)(nil)

func NewMemoryUserRepository() *MemoryUserRepository {
	r := &MemoryUserRepository{}
	r.users = make(map[string]entities.User)
	return r
}

type MemoryUserRepository struct {
	users map[string]entities.User
}

func (r *MemoryUserRepository) Persist(users ...entities.User) error {
	for _, user := range users {
		r.users[user.ID.String()] = user
	}
	return nil
}

func (r *MemoryUserRepository) Find(ids ...support.ID) (users []entities.User, err error) {
	for _, id := range ids {
		u, ok := r.users[id.String()]
		if ok {
			users = append(users, u)
		}
	}
	return
}

func (r MemoryUserRepository) Delete(ids ...support.ID) error {
	return nil
}

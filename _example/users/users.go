package users

import (
    "github.com/GabrielCarpr/cqrs/bus"

    "example/internal/config"
	"example/users/commands"
	"example/users/db"
	"example/users/queries"

	"gorm.io/gorm"

	"github.com/jmoiron/sqlx"
	"github.com/sarulabs/di/v2"
)

type Users struct {}

func (u Users) Commands(b bus.CmdBuilder) {
    b.Command(commands.Register{}).Handled(commands.RegisterHandler{})
    b.Command(commands.UpdateUser{}).Handled(commands.UpdateUserHandler{})
    b.Command(commands.UpdateRole{}).Handled(commands.UpdateRoleHandler{})
}

func (u Users) Queries(b bus.QueryBuilder) {
    b.Query(queries.Login{}).Handled(queries.LoginHandler{})
    b.Query(queries.User{}).Handled(queries.UserHandler{})
    b.Query(queries.Roles{}).Handled(queries.RolesHandler{})
    b.Query(queries.Users{}).Handled(queries.UsersHandler{})
    b.Query(queries.Role{}).Handled(queries.RoleHandler{})
}

func (u Users) EventRules() bus.EventRules {
    return bus.EventRules{}
}

func (u Users) Services() []bus.Def {
    return []bus.Def{
        {
			Name:  "user-repository",
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				dbConn := ctn.Get("gorm").(*gorm.DB)
				return db.NewDBUserRepository(dbConn), nil
			},
		},

		{
			Name:  "role-repository",
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				dbConn := ctn.Get("gorm").(*gorm.DB)
				return db.NewDBRoleRepository(dbConn), nil
			},
		},

		{
			Name:  commands.RegisterHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				users := ctn.Get("user-repository").(db.UserRepository)
				roles := ctn.Get("role-repository").(db.RoleRepository)
				return commands.NewRegisterHandler(users, roles), nil
			},
		},

		{
			Name:  queries.LoginHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				users := ctn.Get("user-repository").(db.UserRepository)
				roles := ctn.Get("role-repository").(db.RoleRepository)
				db := ctn.Get("db").(*sqlx.DB)
				return queries.NewLoginHandler(&config.Values, users, roles, db), nil
			},
		},

		{
			Name:  commands.UpdateUserHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				users := ctn.Get("user-repository").(db.UserRepository)
				roles := ctn.Get("role-repository").(db.RoleRepository)
				return commands.NewUpdateUserHandler(users, roles), nil
			},
		},

		{
			Name:  queries.UserHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				users := ctn.Get("user-repository").(db.UserRepository)
				db := ctn.Get("db").(*sqlx.DB)
				return queries.NewUserHandler(users, db), nil
			},
		},

		{
			Name:  queries.RolesHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				roles := ctn.Get("role-repository").(db.RoleRepository)
				return queries.NewRolesHandler(roles), nil
			},
		},

		{
			Name:  queries.UsersHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				users := ctn.Get("user-repository").(db.UserRepository)
				db := ctn.Get("db").(*sqlx.DB)
				return queries.NewUsersHandler(users, db), nil
			},
		},

		{
			Name:  commands.UpdateRoleHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				roles := ctn.Get("role-repository").(db.RoleRepository)
				return commands.NewUpdateRoleHandler(roles), nil
			},
		},

		{
			Name:  commands.CreateRoleHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				roles := ctn.Get("role-repository").(db.RoleRepository)
				return commands.NewCreateRoleHandler(roles), nil
			},
		},

		{
			Name:  queries.RoleHandler{},
			Scope: di.Request,
			Build: func(ctn di.Container) (interface{}, error) {
				roles := ctn.Get("role-repository").(db.RoleRepository)
				return queries.NewRoleHandler(roles), nil
			},
		},
    }
}

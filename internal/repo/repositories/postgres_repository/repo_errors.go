package postgresrepository

import "errors"

var ErrUserExists = errors.New("user already exists")
var ErrTeamExists = errors.New("team already exists")
var ErrPRExists = errors.New("pr already exists")

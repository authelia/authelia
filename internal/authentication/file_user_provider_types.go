package authentication

import (
	"github.com/go-crypt/crypt"
)

// YAMLDatabaseModel is the model of users file database.
type YAMLDatabaseModel struct {
	Users map[string]YAMLUserDetailsModel `yaml:"users" valid:"required"`
}

func (m YAMLDatabaseModel) ToDatabaseModel() (model *DatabaseModel, err error) {
	model = &DatabaseModel{
		Users: map[string]UserDetailsModel{},
	}

	var udm *UserDetailsModel

	for user, details := range m.Users {
		if udm, err = details.ToDatabaseUserDetailsModel(); err != nil {
			return nil, err
		}

		model.Users[user] = *udm
	}

	return model, nil
}

// YAMLUserDetailsModel is the model of user details in the file database.
type YAMLUserDetailsModel struct {
	HashedPassword string   `yaml:"password" valid:"required"`
	DisplayName    string   `yaml:"displayname" valid:"required"`
	Email          string   `yaml:"email"`
	Groups         []string `yaml:"groups"`
}

func (m YAMLUserDetailsModel) ToDatabaseUserDetailsModel() (model *UserDetailsModel, err error) {
	var d crypt.Digest

	if d, err = crypt.Decode(m.HashedPassword); err != nil {
		return nil, err
	}

	return &UserDetailsModel{
		Digest:      d,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		Groups:      m.Groups,
	}, nil
}

type DatabaseModel struct {
	Users map[string]UserDetailsModel
}

func (m DatabaseModel) ToYAMLDatabaseModel() YAMLDatabaseModel {
	model := YAMLDatabaseModel{
		Users: map[string]YAMLUserDetailsModel{},
	}

	for user, details := range m.Users {
		model.Users[user] = details.ToYAMLUserDetailsModel()
	}

	return model
}

// UserDetailsModel is the model of user details in the file database.
type UserDetailsModel struct {
	Digest      crypt.Digest
	DisplayName string
	Email       string
	Groups      []string
}

func (m UserDetailsModel) ToYAMLUserDetailsModel() YAMLUserDetailsModel {
	return YAMLUserDetailsModel{
		HashedPassword: m.Digest.Encode(),
		DisplayName:    m.DisplayName,
		Email:          m.Email,
		Groups:         m.Groups,
	}
}

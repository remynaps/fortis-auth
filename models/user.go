package models

// Json structure required for user signin
type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Id          string
	DisplayName string
}

type UserIdentity struct {
	Id          string
	DisplayName string
}

func GetUser() (*User, error) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usr := new(User)
	usrErr := rows.Scan(&usr.Id, &usr.DisplayName)
	if usrErr != nil {
		return nil, err
	}
	if usrErr = rows.Err(); err != nil {
		return nil, err
	}
	return usr, nil
}

func RegisterUser() (*User, error) {
	return nil, nil
}

func UserExists() (*User, error) {

	return nil, nil
}

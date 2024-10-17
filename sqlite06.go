// INFO: Since we are using an external package in this code, this program
// should be placed under ~/go/src folder.
// In my case, it is located under `/home/tuna/go/src/github.com/tunacinsoy/mGo4th/ch06/sqlite06`
// After placing it there, the commands `go mod init` and `go mod tidy` should be run.

// Source: https://github.com/tunacinsoy/sqlite06
// The aim of this program is to define most-used operations on sqlite db and provide these
// functions as go package named sqlite06
package sqlite06

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Filename = ""
)

// Most of the time, you need as many structures as there are database tables
type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

// This function is private and only accessed within the scope of this package (starts with lowercase letter)
func openConnection() (*sql.DB, error) {
	// Before calling this func, programmer has to set `Filename` variable using:
	// sqlite06.Filename = "ch06.db" for instance.
	// SQLite3 does not require a username or a password and does not operate over a TCP/IP network.
	db, err := sql.Open("sqlite3", Filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// This function is also private
// Returns the ID of a user whose username is provided in as input parameter
// Returns -1 if there's an error, or user is not found
func exists(username string) int {
	username = strings.ToLower(username)
	// As said above, we can use openConnection() function within the scope of this package
	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// This one is prone to sql injection attacks
	// statement := fmt.Sprintf(`SELECT ID FROM Users where Username = '%s'`, username)

	statement := "SELECT ID FROM Users WHERE Username = ?"
	rows, err := db.Query(statement, username)
	if err != nil {
		fmt.Println("Error retrieving username:", err)
		return -1
	}
	defer rows.Close()

	userID := -1
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println("exists() Scan", err)
			return -1
		}
		userID = id
	}
	return userID
}

// AddUser adds a new user to the database
// Returns new User ID
// -1 if there was an error
func AddUser(d Userdata) int {
	d.Username = strings.ToLower(d.Username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()
	userID := exists(d.Username)
	if userID != -1 {
		fmt.Println("User already exists:", d.Username)
		return -1
	}

	insertStatement := `INSERT INTO Users values (NULL,?)`

	_, err = db.Exec(insertStatement, d.Username)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	// User has been created in Users table, let's check it
	userID = exists(d.Username)
	if userID == -1 {
		return -1
	}

	// `userID` field of Userdata table is the same value from Users table `ID` field
	insertStatement = `INSERT INTO Userdata values (?,?,?,?)`
	_, err = db.Exec(insertStatement, userID, d.Name, d.Surname, d.Description)

	if err != nil {
		fmt.Println(err)
		return -1
	}
	return userID
}

func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Check ID existance
	statement := `SELECT Username FROM Users WHERE ID = ?`
	rows, err := db.Query(statement, id)

	if err != nil {
		return err
	}

	defer rows.Close()

	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}

	if exists(username) == -1 {
		return fmt.Errorf("user with ID %d does not exist", id)
	}

	// At this point, we are sure that userID exists in both tables
	deleteStatement := `DELETE FROM Userdata WHERE UserID = ?`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	deleteStatement = `DELETE FROM Users WHERE ID = ?`

	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}
	return nil
}

func ListUsers() ([]Userdata, error) {
	Data := []Userdata{}
	db, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statement := `SELECT ID, Username, Name, Surname, Description
		FROM USERS, Userdata WHERE Users.ID = Userdata.UserID`

	rows, err := db.Query(statement)

	if err != nil {
		return Data, err
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var description string

		err = rows.Scan(&id, &username, &name, &surname, &description)
		if err != nil {
			return nil, err
		}

		temp := Userdata{ID: id, Username: username, Name: name, Surname: surname, Description: description}
		Data = append(Data, temp)
	}
	return Data, nil
}

// UpdateUser is for updating an existing user
func UpdateUser(d Userdata) error {
	db, err := openConnection()

	if err != nil {
		return err
	}
	defer db.Close()

	// Let's check if the user exists first
	userID := exists(d.Username)

	if userID == -1 {
		return fmt.Errorf("the user %s does not exist", d.Username)
	}

	statement := `UPDATE Userdata set Name = ?, Surname = ?, Description = ? WHERE UserID = ?`

	_, err = db.Exec(statement, d.Name, d.Surname, d.Description, d.ID)

	if err != nil {
		return err
	}

	return nil
}

// Source: https://github.com/tunacinsoy/sqlite06
// The aim of this program is to define most-used operations on sqlite db and provide these
// functions as go package named sqlite06

/*
	It is highly recommended that each Go package you create has a block comment
	preceding the package declaration that introduces developers to the package, and
	also explains what the package does.

The package works on 2 tables on an SQLite database.

The names of the tables are:

  - Users
  - Userdata

The definitions of the tables are:

	CREATE TABLE Users (
	    ID INTEGER PRIMARY KEY,
	    Username TEXT
	);

	CREATE TABLE Userdata (
	    UserID INTEGER NOT NULL,
	    Name TEXT,
	    Surname TEXT,
	    Description TEXT
	);


	// BUG(1): Function ListUsers() not working as expected
	// BUG(2): Function AddUser() is too slow
*/
package sqlite06

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

/*
This global variable holds the SQLite3 database filepath

	Filename: Is the filepath to the database file
*/

var (
	Filename = ""
)

// Most of the time, you need as many structures as there are database tables, however,
// since this a light package, we can combine fields of Users and Userdata table in one struct
type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

// This function is private and only accessed within the scope of this package (starts with lowercase letter)
// openConnection() is for opening the SQLite3 connection
// in order to be used by the other functions of the package
func openConnection() (*sql.DB, error) {
	// Before calling this func, programmer has to set `Filename` variable using:
	// sqlite06.Filename = "ch06.db" for instance.
	// SQLite3 does not require a username or a password and does not operate over a TCP/IP network.

	db, err := sql.Open("sqlite3", Filename)
	if err != nil {
		fmt.Println("Database connection could not be established in func openConnection().")
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
		fmt.Println("Database connection could not be established in func exists().")
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	// This one is prone to sql injection attacks
	// statement := fmt.Sprintf(`SELECT ID FROM Users where Username = '%s'`, username)
	/*
		 	For instance, if a hacker sets username as ' OR '1'='1'
			This changes the query to SELECT ID FROM Users where Username = '' OR '1'='1'
			This query will return all records in the Users table, potentially giving the attacker unauthorized access.
	*/

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
		fmt.Println("Database connection could not be established in func AddUser().")
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

	userID = exists(d.Username)

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
		fmt.Println("Database connection could not be established in func DeleteUser().")
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
		fmt.Println("Database connection could not be established in func ListUsers().")
		return nil, err
	}
	defer db.Close()

	// statement := `SELECT ID, Username, Name, Surname, Description
	// 	FROM USERS, Userdata WHERE Users.ID = Userdata.UserID`

	statement := `SELECT ID, Username, Name, Surname, Description
              FROM Users INNER JOIN Userdata ON Users.ID = Userdata.UserID`

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
	d.Username = strings.ToLower(d.Username)
	userID := exists(d.Username)

	if userID == -1 {
		return fmt.Errorf("the user %s does not exist", d.Username)
	}

	d.ID = userID

	statement := `UPDATE Userdata SET Name = ?, Surname = ?, Description = ? WHERE UserID = ?`

	_, err = db.Exec(statement, d.Name, d.Surname, d.Description, d.ID)

	if err != nil {
		return err
	}

	return nil
}

package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"proyecto/models"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	auth "github.com/hunterlong/authorizecim"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Response format
type response struct {
	ID      int64  `json:"id, omitempty"`
	Message string `json:"message, omitempty"`
}

// Create connection with postgres db
func createConnection() *sql.DB {

	// Open the connection
	postgresInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s database=%s sslmode=disable",
		"localhost", 5432, "root", "root", "project")
	db, err := sql.Open("postgres", postgresInfo)
	if err != nil {
		panic(err)
	}

	// Check the connection
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	// Return the connection
	return db

}

// CreateUser create a user in the postgres db
func CreateUser(w http.ResponseWriter, r *http.Request) {
	// Set the header to content type x-www-form-urlencoded
	// Allow all origin to handle cors issue
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Printf("Unable to decode the request body. %v", err)
	}

	insertID := insertUser(user)

	res := response{
		ID:      insertID,
		Message: "User created successfully",
	}

	json.NewEncoder(w).Encode(res)
}

func Payment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	var bill models.Billto
	err := json.NewDecoder(r.Body).Decode(&bill)

	if err != nil {
		log.Printf("Unable to decode the request body. %v", err)
	}

	Status := createPayment(bill)

	if Status {
		savePaymentInfo(bill)
		saveAddres(bill)
		id := getMaxPaymentID()
		paymentID, _ := strconv.ParseInt(id, 10, 64)
		saveTransaction(paymentID, bill.TypeRNC)

	}

	res := response{
		ID:      1,
		Message: "User created successfully",
	}

	json.NewEncoder(w).Encode(res)

}

// GetUser will return a single user by its id
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Printf("Unable to convert the string into int. %v", err)
	}

	user, err := getUserById(int64(id))

	if err != nil {
		log.Printf("Unable to get user. %v", err)
	}

	json.NewEncoder(w).Encode(user)
}

// GetAllUser will return all the users
func GetAllUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	users, err := getAllUser()

	if err != nil {
		log.Printf("Unable to get all user. %v", err)
	}

	json.NewEncoder(w).Encode(users)
}

// UpdateUser update user's detail in the postgres db
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Printf("Unable to conver the string into int. %v", err)
	}

	var user models.User

	err = json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Printf("Unable to decode the request Body. %v", err)
	}

	updatedRows := updateUser(int64(id), user)

	msg := fmt.Sprintf("User updated successfully. Total rows/record affected", updatedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

// DeleteUser delete user's detail in the postgres db
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Printf("Unable to convert the string into int. %v", err)
	}

	deletedRows := deleteUser(int64(id))

	msg := fmt.Sprintf("User deleted successfully. Total rows/record affected %v", deletedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

//login
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var result response
	var user models.Login

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Printf("Unable to decode the request body. %v", err)
	}

	email := user.Email

	pass, err := userPassword(user)

	if err != nil {
		log.Printf("Unable to get user. %v", err)
	}

	idReruned := getUserIdByEmail(email)

	if CheckPasswordHash(user.Password, pass) {
		result.ID = idReruned
		result.Message = "Success"

	} else {
		result.ID = 0
		result.Message = "Error, incorrect password"
	}

	json.NewEncoder(w).Encode(result)

}

//Send data to payment, like cities and countries
func GetCities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	users, err := getCities()

	if err != nil {
		log.Printf("Unable to get all cities. %v", err)
	}

	json.NewEncoder(w).Encode(users)
}

func GetCountries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	users, err := getCountries()

	if err != nil {
		log.Printf("Unable to get all cities. %v", err)
	}

	json.NewEncoder(w).Encode(users)
}

func GetPaymentSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Printf("Unable to convert the string into int. %v", err)
	}

	ccID := getCCID(id)

	CCData, err := getPaymentSettings(ccID)

	if err != nil {
		log.Printf("Unable to get user settings. %v", err)
	}

	json.NewEncoder(w).Encode(CCData)
}

func UpdatePaymentSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	log.Printf("Estos son los datos: %v", id)

	if err != nil {
		log.Printf("Unable to conver the string into int. %v", err)
	}

	var payment models.PaymentSettings

	err = json.NewDecoder(r.Body).Decode(&payment)

	if err != nil {
		log.Printf("Unable to decode the request Body. %v", err)
	}

	updatedRows := updateCreditCard(payment)

	msg := fmt.Sprintf("User updated successfully. Total rows/record affected ", updatedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func GetProfileSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Printf("Unable to convert the string into int. %v", err)
	}

	userSettings, err := getUserSettings(id)

	if err != nil {
		log.Printf("Unable to get user settings. %v", err)
	}

	json.NewEncoder(w).Encode(userSettings)
}

func UpdateContactSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	log.Printf("Estos son los datos: %v", id)

	if err != nil {
		log.Printf("Unable to conver the string into int. %v", err)
	}

	var payment models.ContactSettings

	err = json.NewDecoder(r.Body).Decode(&payment)

	if err != nil {
		log.Printf("Unable to decode the request Body. %v", err)
	}

	updatedRows := updateContactSettings(id, payment)

	msg := fmt.Sprintf("User updated successfully. Total rows/record affected ", updatedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	log.Printf("Estos son los datos: %v", id)

	if err != nil {
		log.Printf("Unable to conver the string into int. %v", err)
	}

	var payment models.UserSettings

	err = json.NewDecoder(r.Body).Decode(&payment)

	if err != nil {
		log.Printf("Unable to decode the request Body. %v", err)
	}

	updatedRows := updateUserSettings(id, payment)

	msg := fmt.Sprintf("User updated successfully. Total rows/record affected ", updatedRows)

	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func InsertRNC(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	var Rnc models.RNC
	err := json.NewDecoder(r.Body).Decode(&Rnc)

	if err != nil {
		log.Printf("Unable to decode the request body. %v", err)
	}

	Status := insertRNC(Rnc)

	res := response{
		ID:      1,
		Message: "User created successfully",
	}

	if !Status {
		res = response{
			ID:      0,
			Message: "Error created successfully",
		}
	}

	json.NewEncoder(w).Encode(res)

}

func SelectRnc(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "POST")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	var Rnc models.SelectRnc
	err := json.NewDecoder(r.Body).Decode(&Rnc)

	if err != nil {
		log.Printf("Unable to decode the request body. %v", err)
	}

	Rnc, _ = selectRnc()
	json.NewEncoder(w).Encode(Rnc)

}

/*
func GetVouchers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "GET")
	w.Header().Set("Access-Control-Allow-Origin", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Printf("Unable to conver the string into int. %v", err)
	}

	vouchers, err := getAllVouchers(id)

	if err != nil {
		log.Printf("Unable to get all vouchers. %v", err)
	}

	json.NewEncoder(w).Encode(vouchers)
}
*/
// ------------------------------------------ Handler Functions ------------------------//
// Insert one user in the DB
func insertUser(user models.User) int64 {

	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO users (name, lastname, email, password) VALUES ($1, $2, $3, $4) RETURNING id`

	var id int64

	hash, _ := HashPassword(user.Password)

	err := db.QueryRow(sqlStatement, user.Name, user.LastName, user.Email, hash).Scan(&id)

	if err != nil {
		log.Printf("Unable to execute the query 1. %v", err)
	}

	fmt.Printf("Inserted a single record %v\n", id)

	return id
}

var newTransactionId string

func createPayment(bill models.Billto) bool {

	apiName := "6zKpL87drYM"
	apiKey := "5ksj9N7789JU7p6g"

	auth.SetAPIInfo(apiName, apiKey, "test")
	status, _ := auth.IsConnected()
	if status {
		fmt.Println("Connected to Authorize.net!")
	}

	done := ChargeCustomer(bill)

	VoidTransaction()
	return done
}

func ChargeCustomer(bill models.Billto) bool {

	newTransaction := auth.NewTransaction{
		Amount: "99.99",
		CreditCard: auth.CreditCard{
			CardNumber:     bill.CardNumber,
			ExpirationDate: bill.ExpirationDate,
			CardCode:       bill.CardCode,
		},
		BillTo: &auth.BillTo{
			FirstName:   bill.FirstName,
			LastName:    bill.LastName,
			Address:     bill.Address,
			City:        bill.City,
			State:       bill.Province,
			Zip:         "51000",
			Country:     bill.Country,
			PhoneNumber: bill.PhoneNumber,
		},
	}
	response, _ := newTransaction.Charge()

	if response.Approved() {
		newTransactionId = response.TransactionID()
		fmt.Println("Transaction was Approved! #", newTransactionId)
		return true
	}

	return false
}

func VoidTransaction() {

	newTransaction := auth.PreviousTransaction{
		RefId: newTransactionId,
	}
	response, _ := newTransaction.Void()
	if response.Approved() {
		fmt.Println("Transaction was Voided!")
	}

}

// Get one user from the DB by userid
func getUserById(id int64) (models.User, error) {

	db := createConnection()

	defer db.Close()

	var user models.User

	sqlStatement := `SELECT id, name, lastname, email FROM users WHERE id=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&id, &user.Name, &user.LastName, &user.Email)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Printf("Unable to scan the row. %v", err)
	}

	return user, err
}

// Get onse user from the DB by it's userid
func getAllUser() ([]models.User, error) {

	db := createConnection()

	defer db.Close()

	var users []models.User

	sqlStatement := `SELECT name, lastname, email, password FROM users`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Printf("Unable to execute the query 2. %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var user models.User

		err := rows.Scan(&user.Name, &user.LastName, &user.Email, &user.Password)

		if err != nil {
			log.Printf("Unable to scan the row. %v", err)
		}

		users = append(users, user)
	}

	return users, err
}

/*
func getAllVouchers(id int) ([]models.Vouchers, error) {

	db := createConnection()

	defer db.Close()

	var Vouche []models.Vouchers

	sqlStatement := `SELECT name, lastname, email, password FROM users`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Printf("Unable to execute the query 2. %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var user models.User

		err := rows.Scan(&user.Name, &user.LastName, &user.Email, &user.Password)

		if err != nil {
			log.Printf("Unable to scan the row. %v", err)
		}

		users = append(users, user)
	}

	return users, err
}
*/
// Update user in the DB
func updateUser(id int64, user models.User) int64 {

	db := createConnection()

	defer db.Close()

	hash, _ := HashPassword(user.Password)

	sqlStatement := `UPDATE users SET name=$2, lastname=$3, email=$4, password=$5 WHERE userid=$1`

	res, err := db.Exec(sqlStatement, id, user.Name, user.LastName, user.Email, hash)
	if err != nil {
		log.Printf("Unable to execute the query 3. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error while checking the affected rows. %v", err)
	}
	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

// Delete user in the DB
func deleteUser(id int64) int64 {

	db := createConnection()

	defer db.Close()

	sqlStatement := `DELETE FROM users WHERE id=$1`

	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Printf("Unable to execute the query 4. %v", err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Printf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)
	return rowsAffected
}

func userPassword(user models.Login) (string, error) {

	db := createConnection()

	defer db.Close()

	sqlStatemnt := "SELECT password FROM users WHERE email=$1"

	row := db.QueryRow(sqlStatemnt, user.Email)

	err := row.Scan(&user.Password)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return user.Password, nil
	case nil:
		return user.Password, nil
	default:
		log.Printf("Unable to scan the row. %v", err)
	}

	return user.Password, err

	//return 1
}

func getMaxPaymentID() string {

	db := createConnection()

	defer db.Close()

	var id string

	sqlStatemnt := `SELECT max(id+1) FROM payments`

	row := db.QueryRow(sqlStatemnt)

	err := row.Scan(&id)

	if err != nil {
		fmt.Printf("Este es el error: %v", err)
	}

	return id
}

func saveCreditCardVSUser(userID int64, ccID int64) bool {
	db := createConnection()

	defer db.Close()
	idCC := strconv.FormatInt(ccID, 10)
	idUSER := strconv.FormatInt(userID, 10)

	var id int

	sqlStatement := `INSERT INTO uservcard (userid, creditcardid) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(sqlStatement, idUSER, idCC).Scan(&id)

	if err != nil {
		log.Printf("Unable to execute the query 5. %v", err)
		return false
	}

	return true
}

func saveCreditCard(userID int64, bill models.Billto) (int64, bool) {
	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO creditCard (number, cvv, date) VALUES ($1, $2, $3) RETURNING id`

	var id int64

	err := db.QueryRow(sqlStatement, bill.CardNumber, bill.CardCode, bill.ExpirationDate).Scan(&id)

	if err != nil {
		log.Printf("Unable to execute the query 6. %v", err)
		return 0, false
	}

	fmt.Printf("Inserted a single record %v", id)
	if saveCreditCardVSUser(userID, id) {
		return id, true
	}
	return 0, false
}

func getUserIdByEmail(Email string) int64 {

	db := createConnection()

	defer db.Close()

	var id int64

	sqlStatement := `SELECT id FROM users WHERE email=$1`

	row := db.QueryRow(sqlStatement, Email)

	err := row.Scan(&id)

	switch err {
	case sql.ErrNoRows:
		fmt.Printf("No rows were returned: %v", err)
		return 0
	}

	return id
}

func updateUserPhoneAndAddress(bill models.Billto) {

	var id int

	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE users SET rncid=$2, phone=$3 WHERE email=$1 RETURNING id`

	err := db.QueryRow(sqlStatement, bill.Email, bill.RNC, bill.PhoneNumber).Scan(&id)

	if err != nil {
		log.Printf("Unable to execute the query 12. %v", err)
	}

	log.Printf("User successfully update phone and RNC %v", id)

}

func savePaymentInfo(bill models.Billto) {

	//Save Credit Card and get the ID
	userID := getUserIdByEmail(bill.Email)
	ccID, statusCC := saveCreditCard(userID, bill)
	updateUserPhoneAndAddress(bill)
	RNCData, _ := selectRnc()
	voucher := RNCData.Rnc
	log.Printf("Este es el dato que debe ser el boucher%v", voucher)

	if statusCC {

		db := createConnection()

		defer db.Close()

		sqlStatement := `INSERT INTO payments (date, plansid, rnc, voucher, creditcardid) VALUES ($1, $2, $3, $4, $5) RETURNING id`

		currentTime := time.Now()

		var id int

		err := db.QueryRow(sqlStatement, currentTime, "1", bill.RNC, voucher, ccID).Scan(&id)
		if err != nil {
			log.Printf("Unable to execute the query 16. %v", err)
		}

		fmt.Printf("Inserted a single record on 16: %v", id)

	}
}

func updateUserID(id int, mail string) {

	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE users SET addressid=$1 WHERE email=$2 RETURNING id`

	err := db.QueryRow(sqlStatement, id, mail).Scan(&id)
	if err != nil {
		log.Printf("Unable to execute the query 17. %v", err)
	}

}

func saveAddres(bill models.Billto) {

	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO address (name, cityID) VALUES ($1, $2) RETURNING id`

	var id int

	err := db.QueryRow(sqlStatement, bill.Address, bill.City).Scan(&id)
	if err != nil {
		log.Printf("Unable to execute the query 7. %v", err)
	}

	updateUserID(id, bill.Email)

}

func saveTransaction(paymentID int64, transType string) bool {

	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO transactions (paymentid, typeid) VALUES ($1, $2) RETURNING id`

	var id int

	err := db.QueryRow(sqlStatement, paymentID, transType).Scan(&id)

	log.Printf("1:%v   , 2:%v", paymentID, transType)

	if err != nil {
		log.Printf("Unable to execute the query 8. %v", err)
		return false
	}

	fmt.Printf("Inserted a single record: %v", id)

	return true
}

func getCities() ([]models.City, error) {
	db := createConnection()

	defer db.Close()

	var cities []models.City

	sqlStatement := `SELECT * FROM city`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var data models.City

		err := rows.Scan(&data.Id, &data.Name, &data.Province)

		if err != nil {
			log.Printf("Unable to scan the row. %v", err)
		}

		cities = append(cities, data)
	}

	return cities, err

}

func getCountries() ([]models.Province, error) {
	db := createConnection()

	defer db.Close()

	var cities []models.Province

	sqlStatement := `SELECT id, name FROM province`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var data models.Province

		err := rows.Scan(&data.Id, &data.Name)

		if err != nil {
			log.Printf("Unable to scan the row. %v", err)
		}

		cities = append(cities, data)
	}

	return cities, err

}

func getCCID(id int) int {

	db := createConnection()

	defer db.Close()

	sqlStatement := `SELECT creditcardid FROM uservcard WHERE userid=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&id)
	log.Printf("Este es el ID de la tarjeta: %v", id)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return 0
	}

	return id

}

func getPaymentSettings(id int) (models.PaymentSettings, error) {

	db := createConnection()

	defer db.Close()

	var cc models.PaymentSettings

	sqlStatement := `SELECT number, cvv, date FROM creditcard WHERE id=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&cc.Number, &cc.Cvv, &cc.Date)

	cc.Id = strconv.Itoa(id)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return cc, nil
	case nil:
		return cc, nil
	default:
		log.Printf("Unable to scan the row. %v", err)
	}

	log.Printf("Este es el problema?")
	cc.Id = strconv.Itoa(id)

	return cc, err
}

func updateCreditCard(CC models.PaymentSettings) int64 {

	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE creditcard SET number=$2, cvv=$3, date=$4 WHERE id=$1`

	res, err := db.Exec(sqlStatement, CC.Id, CC.Number, CC.Cvv, CC.Date)
	if err != nil {
		log.Printf("Unable to execute the query 13. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error while checking the affected rows. %v", err)
	}
	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

func getUserSettings(id int) (models.ProfileSetting, error) {

	var idAddress int

	db := createConnection()

	defer db.Close()

	var user models.ProfileSetting

	sqlStatement := `SELECT email, name, lastname, rncid, addressid FROM users WHERE id=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&user.Email, &user.Name, &user.LastName, &user.RNC, &idAddress)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Printf("Unable to scan the row. %v", err)
	}

	//Get Adrress Setting

	sqlStatement = `SELECT name, cityid FROM address WHERE id=$1`

	row = db.QueryRow(sqlStatement, id)

	err = row.Scan(&user.Address, &user.City)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Printf("Unable to scan the row. %v", err)
	}

	return user, err
}

func updateUserSettings(id int, user models.UserSettings) int64 {

	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE users SET email=$2, name=$3, lastname=$4, rncid=$5, status=$6 WHERE id=$1`

	res, err := db.Exec(sqlStatement, id, user.Email, user.Name, user.LastName, user.RNC, user.TypeRNC)
	if err != nil {
		log.Printf("Unable to execute the query 13. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error while checking the affected rows. %v", err)
	}
	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected

}

func getUserAddressId(id int) (string, error) {

	var idR string

	db := createConnection()

	defer db.Close()

	sqlStatement := `SELECT addressid FROM users WHERE id=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&idR)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return idR, nil
	case nil:
		return idR, nil
	default:
		log.Printf("Unable to scan the row. %v", err)
	}

	return idR, err
}

func updateContactSettings(id int, user models.ContactSettings) int64 {

	db := createConnection()

	defer db.Close()

	addresID, _ := getUserAddressId(id)

	sqlStatement := `UPDATE address SET name=$2, cityid=$3 WHERE id=$1`

	res, err := db.Exec(sqlStatement, addresID, user.Address, user.City)
	if err != nil {
		log.Printf("Unable to execute the query 14. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error while checking the affected rows. %v", err)
	}
	fmt.Printf("Total rows/record affected %v", rowsAffected)

	//Update PhoneNumber

	UpdatePhoneNumber(id, user.PhoneNumber)
	return rowsAffected
}

func UpdatePhoneNumber(id int, number string) {
	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE users SET phone=$2 WHERE id=$1`

	res, err := db.Exec(sqlStatement, id, number)
	if err != nil {
		log.Printf("Unable to execute the query 15. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

}

func insertRNC(rnc models.RNC) bool {
	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO rnc (rnc, type, expiration) VALUES ($1, $2, $3)`

	min, _ := strconv.Atoi(rnc.Min)
	max, _ := strconv.Atoi(rnc.Max)

	for i := min; i <= max; i++ {

		res, err := db.Exec(sqlStatement, i, rnc.Type, rnc.Expiration)

		if err != nil {
			log.Printf("Unable to execute the query 16. %v", err)
			return false
		}

		fmt.Printf("Total rows/record affected %v", res)

	}
	return true
}

func updateRNC(voucher string) {
	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE rnc SET status=false WHERE rnc=$1`

	res, err := db.Exec(sqlStatement, voucher)
	if err != nil {
		log.Printf("Unable to execute the query 15. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

}

func selectRnc() (models.SelectRnc, error) {

	var rnc models.SelectRnc

	db := createConnection()

	defer db.Close()

	sqlStatement := `SELECT rnc, type FROM rnc WHERE status=true`

	row := db.QueryRow(sqlStatement)

	err := row.Scan(&rnc.Rnc, &rnc.TypeR)

	//change rnc status
	updateRNC(rnc.Rnc)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return rnc, nil
	case nil:
		return rnc, nil
	default:
		log.Printf("Unable to scan the row. %v", err)
	}

	return rnc, err

}

package router

import (
	"proyecto/middleware"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	// Users
	router.HandleFunc("/api/user/{id}", middleware.GetUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/users", middleware.GetAllUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/register", middleware.CreateUser).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/user/{id}", middleware.UpdateUser).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/deleteuser/{id}", middleware.DeleteUser).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/login", middleware.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/payment", middleware.Payment).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/cities", middleware.GetCities).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/provinces", middleware.GetCountries).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/paymentsettings/{id}", middleware.GetPaymentSettings).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/paymentsettings/{id}", middleware.UpdatePaymentSettings).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/profile/{id}", middleware.GetProfileSettings).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/usersettings/{id}", middleware.UpdateUserSettings).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/contactsettings/{id}", middleware.UpdateContactSettings).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/insertrnc", middleware.InsertRNC).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/select", middleware.SelectRnc).Methods("GET", "OPTIONS")
	//router.HandleFunc("/api/uservouchers/{id}", middleware.GetVouchers).Methods("GET", "OPTIONS")
	return router
}

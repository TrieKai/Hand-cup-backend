package controllers

import "handCup-project-backend/api/middlewares"

func (s *Server) initializeRoutes() {
	// Home Route
	// s.Router.HandleFunc("/", middlewares.SetMiddlewareJSON(s.Home)).Methods("GET")

	// Sign up Route
	s.Router.HandleFunc("/signup", middlewares.SetMiddlewareJSON(s.CreateUser)).Methods("POST", "OPTIONS")

	// Login Route
	s.Router.HandleFunc("/login", middlewares.SetMiddlewareJSON(s.Login)).Methods("POST", "OPTIONS")

	// Users routes
	// s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.CreateUser)).Methods("POST")
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.GetUsers)).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(s.GetUser)).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateUser))).Methods("PUT", "OPTIONS")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareAuthentication(s.DeleteUser)).Methods("DELETE", "OPTIONS")

	// Reset password
	s.Router.HandleFunc("/reset", middlewares.SetMiddlewareJSON(s.ResetPassword)).Methods("POST", "OPTIONS")

	// TODO: Add authentication
	// Favorites routes
	s.Router.HandleFunc("/favorites/{user_id}", middlewares.SetMiddlewareJSON(s.GetFavorites)).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/favorites", middlewares.SetMiddlewareJSON(s.CreateFavorites)).Methods("POST", "OPTIONS")
	s.Router.HandleFunc("/favorites/{user_id}/{place_id}", middlewares.SetMiddlewareJSON(s.DeleteFavorites)).Methods("DELETE", "OPTIONS")

	// Visited routes
	s.Router.HandleFunc("/visited/{user_id}", middlewares.SetMiddlewareJSON(s.GetVisiteds)).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/visited", middlewares.SetMiddlewareJSON(s.CreateVisited)).Methods("POST", "OPTIONS")
	s.Router.HandleFunc("/visited/{user_id}/{place_id}", middlewares.SetMiddlewareJSON(s.DeleteVisited)).Methods("DELETE", "OPTIONS")

	// Posts routes
	// s.Router.HandleFunc("/posts", middlewares.SetMiddlewareJSON(s.CreatePost)).Methods("POST")
	// s.Router.HandleFunc("/posts", middlewares.SetMiddlewareJSON(s.GetPosts)).Methods("GET")
	// s.Router.HandleFunc("/posts/{id}", middlewares.SetMiddlewareJSON(s.GetPost)).Methods("GET")
	// s.Router.HandleFunc("/posts/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdatePost))).Methods("PUT")
	// s.Router.HandleFunc("/posts/{id}", middlewares.SetMiddlewareAuthentication(s.DeletePost)).Methods("DELETE")

	// Handcup routes
	s.Router.HandleFunc("/map", middlewares.SetMiddlewareJSON(s.GetHandcupList)).Methods("POST", "OPTIONS")
	s.Router.HandleFunc("/map/{placeId}/{language}", middlewares.SetMiddlewareJSON(s.GetPlaceDetail)).Methods("GET", "OPTIONS")
	s.Router.HandleFunc("/myMap/{userId}", middlewares.SetMiddlewareJSON(s.GetMyMapList)).Methods("GET", "OPTIONS")

	// Upload routes
	s.Router.HandleFunc("/upload", middlewares.SetMiddlewareJSON(s.UploadFile)).Methods("POST", "OPTIONS")
}

package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elllban/test-backend-elllban/internal/config"
	"github.com/elllban/test-backend-elllban/internal/handlers"
	"github.com/elllban/test-backend-elllban/internal/middleware"
	"github.com/elllban/test-backend-elllban/internal/repository"
	"github.com/elllban/test-backend-elllban/internal/service"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	_ "github.com/elllban/test-backend-elllban/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Room Booking Service API
// @version 1.0
// @description Сервис бронирования переговорок
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected")

	roomRepo := repository.NewRoomRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	slotRepo := repository.NewSlotRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo, roomRepo, slotRepo)
	slotService := service.NewSlotService(slotRepo, scheduleRepo)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, userRepo)

	authMiddleware := middleware.NewAuthMiddleware([]byte(cfg.JWTSecret))

	roomHandler := handlers.NewRoomHandler(roomService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)
	slotHandler := handlers.NewSlotHandler(slotService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	authHandler := handlers.NewAuthHandler([]byte(cfg.JWTSecret), userRepo)

	router := mux.NewRouter()
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// public
	router.HandleFunc("/_info", handlers.InfoHandler).Methods("GET")
	router.HandleFunc("/dummyLogin", authHandler.DummyLogin).Methods("POST")

	// protected
	api := router.PathPrefix("").Subrouter()
	api.Use(authMiddleware.Authenticate)

	// rooms
	api.HandleFunc("/rooms/list", roomHandler.ListRooms).Methods("GET")
	api.HandleFunc("/rooms/create", roomHandler.CreateRoom).Methods("POST")

	// schedules
	api.HandleFunc("/rooms/{roomId}/schedule/create", scheduleHandler.CreateSchedule).Methods("POST")

	// slots
	api.HandleFunc("/rooms/{roomId}/slots/list", slotHandler.ListSlots).Methods("GET")

	// bookings
	api.HandleFunc("/bookings/create", bookingHandler.CreateBooking).Methods("POST")
	api.HandleFunc("/bookings/list", bookingHandler.ListAllBookings).Methods("GET")
	api.HandleFunc("/bookings/my", bookingHandler.ListMyBookings).Methods("GET")
	api.HandleFunc("/bookings/{bookingId}/cancel", bookingHandler.CancelBooking).Methods("POST")

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

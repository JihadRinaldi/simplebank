package api

import (
	"net/http"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	"github.com/JihadRinaldi/simplebank/token"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	Store      db.Store
	TokenMaker token.Maker
	Router     *gin.Engine
	Config     util.Config
}

func NewServer(store db.Store, config util.Config) (*Server, error) {
	token, err := token.NewJWTMaker(config.SymmetricKey)
	if err != nil {
		return nil, err
	}

	server := &Server{
		Store:      store,
		TokenMaker: token,
		Config:     config,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.SetupRouter()

	return server, nil
}

func (server *Server) SetupRouter() {
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.POST("/users", server.CreateUser)
	router.POST("/users/login", server.LoginUser)

	authRoutes := router.Group("/").Use(AuthMiddleware(server.TokenMaker))
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)

	authRoutes.POST("/transfers", server.createTransfer)

	server.Router = router
}

func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.Router.ServeHTTP(w, r)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}

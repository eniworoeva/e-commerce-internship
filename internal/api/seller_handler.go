package api

import (
	"e-commerce/internal/middleware"
	"e-commerce/internal/models"
	"e-commerce/internal/util"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Create Seller
func (u *HTTPHandler) CreateSeller(c *gin.Context) {
	var seller *models.Seller
	if err := c.ShouldBind(&seller); err != nil {
		util.Response(c, "invalid request", 400, err.Error(), nil)
		return
	}

	_, err := u.Repository.FindSellerByEmail(seller.Email)
	if err == nil {
		util.Response(c, "User already exists", 400, "Bad request body", nil)
		return
	}

	// Hash the password
	hashedPassword, err := util.HashPassword(seller.Password)
	if err != nil {
		util.Response(c, "Internal server error", 500, err.Error(), nil)
		return
	}
	seller.Password = hashedPassword

	err = u.Repository.CreateSeller(seller)
	if err != nil {
		util.Response(c, "Seller not created", 500, err.Error(), nil)
		return
	}
	util.Response(c, "Seller created", 200, nil, nil)

}

// Login Seller
func (u *HTTPHandler) LoginSeller(c *gin.Context) {
	var loginRequest *models.LoginRequestSeller
	err := c.ShouldBind(&loginRequest)
	if err != nil {
		util.Response(c, "invalid request", 400, err.Error(), nil)
		return
	}

	loginRequest.Email = strings.TrimSpace(loginRequest.Email)
	loginRequest.Password = strings.TrimSpace(loginRequest.Password)

	if loginRequest.Email == "" {
		util.Response(c, "Email must not be empty", 400, nil, nil)
		return
	}
	if loginRequest.Password == "" {
		util.Response(c, "Password must not be empty", 400, nil, nil)
		return
	}

	Seller, err := u.Repository.FindSellerByEmail(loginRequest.Email)
	if err != nil {
		util.Response(c, "Email does not exist", 404, err.Error(), nil)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(Seller.Password), []byte(loginRequest.Password))
	if err != nil {
		util.Response(c, "Invalid password", 400, err.Error(), nil)
		return
	}

	accessClaims, refreshClaims := middleware.GenerateClaims(Seller.Email)

	secret := os.Getenv("JWT_SECRET")

	accessToken, err := middleware.GenerateToken(jwt.SigningMethodHS256, accessClaims, &secret)
	if err != nil {
		util.Response(c, "Error generating access token", 500, err.Error(), nil)
		return
	}

	refreshToken, err := middleware.GenerateToken(jwt.SigningMethodHS256, refreshClaims, &secret)
	if err != nil {
		util.Response(c, "Error generating refresh token", 500, err.Error(), nil)
		return
	}

	c.Header("access_token", *accessToken)
	c.Header("refresh_token", *refreshToken)

	util.Response(c, "Login successful", 200, gin.H{
		"Seller":        Seller,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil)
}

// create a new product
func (u *HTTPHandler) CreateProduct(c *gin.Context) {
	seller, err := u.GetSellerFromContext(c)
	if err != nil {
		util.Response(c, "Invalid token", 401, err.Error(), nil)
		return
	}

	var product *models.Product
	if err := c.ShouldBind(&product); err != nil {
		util.Response(c, "invalid request", 400, err.Error(), nil)
		return
	}

	product.SellerID = seller.ID

	err = u.Repository.CreateProduct(product)
	if err != nil {
		util.Response(c, "Product not created", 500, err.Error(), nil)
		return
	}
	util.Response(c, "Product created", 200, nil, nil)
}

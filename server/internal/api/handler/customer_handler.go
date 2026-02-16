package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
)

type CustomerHandler struct {
	customerRepo *repository.CustomerRepository
}

func NewCustomerHandler(customerRepo *repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{customerRepo: customerRepo}
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var body struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  body.Name,
		Email: body.Email,
		Active: true,
	}

	if err := h.customerRepo.Create(c.Request.Context(), customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func (h *CustomerHandler) List(c *gin.Context) {
	customers, err := h.customerRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"customers": customers})
}

func (h *CustomerHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	customer, err := h.customerRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	customer, err := h.customerRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	var body struct {
		Name   *string `json:"name"`
		Email  *string `json:"email"`
		Active *bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Name != nil {
		customer.Name = *body.Name
	}
	if body.Email != nil {
		customer.Email = *body.Email
	}
	if body.Active != nil {
		customer.Active = *body.Active
	}

	if err := h.customerRepo.Update(c.Request.Context(), customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

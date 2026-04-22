package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	redis_repo "github.com/assidik12/go-restfull-api/internal/repository/redis"
	"github.com/assidik12/go-restfull-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductRepo struct {
	mock.Mock
}

func (m *MockProductRepo) GetAll(ctx context.Context, page int, pageSize int) ([]domain.Product, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockProductRepo) FindById(ctx context.Context, id int) (domain.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepo) Save(ctx context.Context, product domain.Product) (domain.Product, error) {
	args := m.Called(ctx, product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepo) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	args := m.Called(ctx, product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepo) DecrementStock(ctx context.Context, tx *sql.Tx, productID int, qty int) error {
	args := m.Called(ctx, tx, productID, qty)
	return args.Error(0)
}

func setupProductServiceTesting() (*MockProductRepo, service.ProductService) {
	mockRepo := new(MockProductRepo)

	// Create a dummy redis client pointing to an invalid address so it immediately errors out (cache miss)
	rClient := redis.NewClient(&redis.Options{
		Addr:        "localhost:9999",
		DialTimeout: 10 * time.Millisecond,
	})
	wrapper := redis_repo.NewWrapper(rClient)
	validate := validator.New()

	productService := service.NewProductService(mockRepo, nil, wrapper, validate)
	return mockRepo, productService
}

func TestGetProductById_Success(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	expectedProduct := domain.Product{
		ID:          1,
		Name:        "Test Product",
		Price:       10000,
		Stock:       10,
		Description: "A test product",
		Img:         "test.png",
		CategoryId:  1,
	}

	mockRepo.On("FindById", mock.Anything, 1).Return(expectedProduct, nil)

	ctx := context.Background()
	product, err := productService.GetProductById(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedProduct, product)
	mockRepo.AssertExpectations(t)
}

func TestGetProductById_NotFound(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	mockRepo.On("FindById", mock.Anything, 1).Return(domain.Product{}, domain.ErrNotFound)

	ctx := context.Background()
	product, err := productService.GetProductById(ctx, 1)

	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, product)
	mockRepo.AssertExpectations(t)
}

func TestGetProductById_InvalidId(t *testing.T) {
	_, productService := setupProductServiceTesting()

	ctx := context.Background()
	product, err := productService.GetProductById(ctx, 0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Empty(t, product)
}

func TestGetAllProducts_Success(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	expectedProducts := []domain.Product{
		{ID: 1, Name: "Product A", Price: 10000, Stock: 5},
		{ID: 2, Name: "Product B", Price: 20000, Stock: 10},
	}

	mockRepo.On("GetAll", mock.Anything, 1, 10).Return(expectedProducts, nil)

	ctx := context.Background()
	products, err := productService.GetAllProducts(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, products, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetAllProducts_RepoError(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	mockRepo.On("GetAll", mock.Anything, 1, 10).Return([]domain.Product(nil), domain.ErrNotFound)

	ctx := context.Background()
	products, err := productService.GetAllProducts(ctx, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, products)
	mockRepo.AssertExpectations(t)
}

func TestCreateProduct_Success(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	req := dto.ProductRequest{
		Name:        "New Product",
		Price:       15000,
		Stock:       20,
		Img:         "img.png",
		Description: "A great product",
		CategoryId:  1,
	}

	savedProduct := domain.Product{
		ID:          3,
		Name:        req.Name,
		Price:       req.Price,
		Img:         req.Img,
		Description: req.Description,
		CategoryId:  req.CategoryId,
	}

	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(p domain.Product) bool {
		return p.Name == req.Name && p.Price == req.Price
	})).Return(savedProduct, nil)

	ctx := context.Background()
	product, err := productService.CreateProduct(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, savedProduct.ID, product.ID)
	assert.Equal(t, savedProduct.Name, product.Name)
	mockRepo.AssertExpectations(t)
}

func TestCreateProduct_RepoError(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	req := dto.ProductRequest{Name: "Bad Product", Price: 5000}

	mockRepo.On("Save", mock.Anything, mock.Anything).Return(domain.Product{}, domain.ErrInvalidInput)

	ctx := context.Background()
	product, err := productService.CreateProduct(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, product)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProduct_Success(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	req := dto.ProductRequest{
		Name:        "Updated Product",
		Price:       25000,
		Stock:       30,
		Img:         "updated.png",
		Description: "Updated description",
		CategoryId:  2,
	}

	updatedProduct := domain.Product{
		ID:          1,
		Name:        req.Name,
		Price:       req.Price,
		Img:         req.Img,
		Description: req.Description,
		CategoryId:  req.CategoryId,
	}

	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p domain.Product) bool {
		return p.ID == 1 && p.Name == req.Name
	})).Return(updatedProduct, nil)

	ctx := context.Background()
	product, err := productService.UpdateProduct(ctx, 1, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, product.ID)
	assert.Equal(t, req.Name, product.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProduct_NotFound(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	req := dto.ProductRequest{Name: "Ghost Product", Price: 9999}

	mockRepo.On("Update", mock.Anything, mock.Anything).Return(domain.Product{}, domain.ErrNotFound)

	ctx := context.Background()
	product, err := productService.UpdateProduct(ctx, 999, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, product)
	mockRepo.AssertExpectations(t)
}

func TestDeleteProduct_Success(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	mockRepo.On("Delete", mock.Anything, 1).Return(nil)

	ctx := context.Background()
	err := productService.DeleteProduct(ctx, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteProduct_NotFound(t *testing.T) {
	mockRepo, productService := setupProductServiceTesting()

	mockRepo.On("Delete", mock.Anything, 999).Return(domain.ErrNotFound)

	ctx := context.Background()
	err := productService.DeleteProduct(ctx, 999)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

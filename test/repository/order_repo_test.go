package order_test

import (
	"context"
	"log"
	"testing"

	"github.com/corradoisidoro/orders-api/internal/model"
	"github.com/corradoisidoro/orders-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.New(
			log.New(nil, "", 0),
			logger.Config{LogLevel: logger.Silent},
		),
	})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&model.Order{}, &model.LineItem{}))

	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })

	return db
}

//
// INSERT
//

func TestInsert_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)
	ctx := context.Background()

	o := &model.Order{CustomerID: 1}
	err := repo.Insert(ctx, o)

	require.NoError(t, err)
	assert.Greater(t, o.OrderID, int64(0))
}

func TestInsert_Fails_WhenNilOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.Insert(context.Background(), nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

func TestInsert_Fails_WhenCustomerIDZero(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.Insert(context.Background(), &model.Order{CustomerID: 0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

//
// FIND ALL
//

func TestFindAll_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		require.NoError(t, repo.Insert(ctx, &model.Order{CustomerID: int64(i)}))
	}

	result, err := repo.FindAll(ctx, repository.Page{Size: 2, Offset: 0})

	require.NoError(t, err)
	assert.Len(t, result.Orders, 2)
	assert.Equal(t, int64(2), result.Cursor)
}

func TestFindAll_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	result, err := repo.FindAll(context.Background(), repository.Page{Size: 10, Offset: 0})

	require.NoError(t, err)
	assert.Empty(t, result.Orders)
	assert.Equal(t, int64(0), result.Cursor)
}

func TestFindAll_OffsetBeyondRange(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)
	ctx := context.Background()

	require.NoError(t, repo.Insert(ctx, &model.Order{CustomerID: 1}))

	result, err := repo.FindAll(ctx, repository.Page{Size: 10, Offset: 50})

	require.NoError(t, err)
	assert.Empty(t, result.Orders)
	assert.Equal(t, int64(50), result.Cursor)
}

func TestFindAll_Fails_WhenNegativeOffset(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	result, err := repo.FindAll(context.Background(), repository.Page{Size: 10, Offset: -1})

	assert.Error(t, err)
	assert.Empty(t, result.Orders)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

func TestFindAll_Fails_WhenNegativeSize(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	result, err := repo.FindAll(context.Background(), repository.Page{Size: -5, Offset: 0})

	assert.Error(t, err)
	assert.Empty(t, result.Orders)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

//
// FIND BY ID
//

func TestFindByID_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)
	ctx := context.Background()

	o := &model.Order{CustomerID: 1}
	require.NoError(t, repo.Insert(ctx, o))

	found, err := repo.FindByID(ctx, o.OrderID)

	require.NoError(t, err)
	assert.Equal(t, o.OrderID, found.OrderID)
}

func TestFindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	_, err := repo.FindByID(context.Background(), 999)
	assert.ErrorIs(t, err, repository.ErrNotExist)
}

func TestFindByID_Fails_WhenIDZero(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	_, err := repo.FindByID(context.Background(), 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

//
// UPDATE
//

func TestUpdateByID_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)
	ctx := context.Background()

	o := &model.Order{CustomerID: 1}
	require.NoError(t, repo.Insert(ctx, o))

	o.CustomerID = 2
	require.NoError(t, repo.UpdateByID(ctx, o))

	updated, err := repo.FindByID(ctx, o.OrderID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), updated.CustomerID)
}

func TestUpdateByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.UpdateByID(context.Background(), &model.Order{OrderID: 999, CustomerID: 1})
	assert.ErrorIs(t, err, repository.ErrNotExist)
}

func TestUpdateByID_Fails_WhenNilOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.UpdateByID(context.Background(), nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

func TestUpdateByID_Fails_WhenOrderIDZero(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.UpdateByID(context.Background(), &model.Order{OrderID: 0, CustomerID: 1})
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

func TestUpdateByID_Fails_WhenCustomerIDZero(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.UpdateByID(context.Background(), &model.Order{OrderID: 1, CustomerID: 0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

//
// DELETE
//

func TestDeleteByID_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)
	ctx := context.Background()

	o := &model.Order{CustomerID: 1}
	require.NoError(t, repo.Insert(ctx, o))

	require.NoError(t, repo.DeleteByID(ctx, o.OrderID))

	_, err := repo.FindByID(ctx, o.OrderID)
	assert.ErrorIs(t, err, repository.ErrNotExist)
}

func TestDeleteByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.DeleteByID(context.Background(), 999)
	assert.ErrorIs(t, err, repository.ErrNotExist)
}

func TestDeleteByID_Fails_WhenIDZero(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewOrderRepo(db)

	err := repo.DeleteByID(context.Background(), 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrInvalidInput)
}

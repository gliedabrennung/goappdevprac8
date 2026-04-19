package service

import (
	"errors"
	"testing"

	"github.com/gliedabrennung/goappdevprac8/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Test", Email: "test@test.com"}

	t.Run("user already exists", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail("test@test.com").Return(user, nil)

		err := service.RegisterUser(user, "test@test.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("new user success", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail("new@test.com").Return(nil, nil)
		mockRepo.EXPECT().CreateUser(user).Return(nil)

		err := service.RegisterUser(user, "new@test.com")
		assert.NoError(t, err)
	})

	t.Run("repository error on get email", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail("error@test.com").Return(nil, errors.New("db error"))

		err := service.RegisterUser(user, "error@test.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error getting user")
	})
}

func TestUpdateUserName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)

	t.Run("empty name", func(t *testing.T) {
		err := service.UpdateUserName(1, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetUserByID(99).Return(nil, errors.New("not found"))

		err := service.UpdateUserName(99, "New Name")
		assert.Error(t, err)
	})

	t.Run("successful update", func(t *testing.T) {
		oldUser := &repository.User{ID: 1, Name: "Old Name"}

		mockRepo.EXPECT().GetUserByID(1).Return(oldUser, nil)
		mockRepo.EXPECT().UpdateUser(gomock.Any()).Do(func(u *repository.User) {
			assert.Equal(t, "New Name", u.Name, "Name should be updated before calling repo")
		}).Return(nil)

		err := service.UpdateUserName(1, "New Name")
		assert.NoError(t, err)
	})

	t.Run("update user fails", func(t *testing.T) {
		oldUser := &repository.User{ID: 1, Name: "Old Name"}
		mockRepo.EXPECT().GetUserByID(1).Return(oldUser, nil)
		mockRepo.EXPECT().UpdateUser(gomock.Any()).Return(errors.New("update failed"))

		err := service.UpdateUserName(1, "New Name")
		assert.Error(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)

	t.Run("attempt to delete admin", func(t *testing.T) {
		err := service.DeleteUser(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed to delete admin")
	})

	t.Run("successful delete", func(t *testing.T) {
		mockRepo.EXPECT().DeleteUser(2).Return(nil)

		err := service.DeleteUser(2)
		assert.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().DeleteUser(3).Return(errors.New("db connection lost"))

		err := service.DeleteUser(3)
		assert.Error(t, err)
	})
}

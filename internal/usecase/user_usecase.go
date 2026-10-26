package usecase

import (
	"errors"
	"fmt"

	"api/internal/domain"
)

var ErrInvalidCredentials = errors.New("email hoặc mật khẩu không chính xác")

type CredentialService interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
	GenerateToken(userID uint, email, role string) (string, error)
}

type UserUseCase struct {
	userRepo    domain.UserRepository
	credentials CredentialService
	unitOfWork  UnitOfWork
}

func NewUserUseCase(users domain.UserRepository, credentials CredentialService, unitOfWork UnitOfWork) *UserUseCase {
	return &UserUseCase{userRepo: users, credentials: credentials, unitOfWork: unitOfWork}
}

type RegisterDTO struct {
	Name     string
	Email    string
	Password string
}

type LoginDTO struct {
	Email    string
	Password string
}

type LoginResponse struct {
	Token string
	User  domain.User
}

func (uc *UserUseCase) Register(input RegisterDTO, ipAddress string) (uint, error) {
	hashedPassword, err := uc.credentials.HashPassword(input.Password)
	if err != nil {
		return 0, fmt.Errorf("không thể mã hóa mật khẩu: %w", err)
	}
	user, err := domain.NewUser(input.Name, input.Email, hashedPassword)
	if err != nil {
		return 0, err
	}
	if err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		created, err := repositories.Users.Create(user)
		if err != nil {
			return fmt.Errorf("email đã tồn tại hoặc không thể tạo tài khoản: %w", err)
		}
		user = created
		data := user.Data()
		return repositories.Audits.RecordLog(data.ID, "REGISTER", fmt.Sprintf("User #%d", data.ID), "Đăng ký tài khoản mới: "+string(data.Email), ipAddress)
	}); err != nil {
		return 0, fmt.Errorf("không thể hoàn tất đăng ký: %w", err)
	}
	return user.Data().ID, nil
}

func (uc *UserUseCase) Login(input LoginDTO, ipAddress string) (*LoginResponse, error) {
	user, err := uc.userRepo.GetByEmail(input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	userData := user.Data()
	if !uc.credentials.CheckPasswordHash(input.Password, userData.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	token, err := uc.credentials.GenerateToken(userData.ID, string(userData.Email), string(userData.Role))
	if err != nil {
		return nil, fmt.Errorf("không thể tạo token xác thực: %w", err)
	}
	if err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		return repositories.Audits.RecordLog(userData.ID, "LOGIN", fmt.Sprintf("User #%d", userData.ID), "Đăng nhập thành công vào hệ thống", ipAddress)
	}); err != nil {
		return nil, fmt.Errorf("đăng nhập thành công nhưng không thể ghi audit log: %w", err)
	}
	userData.PasswordHash = ""
	return &LoginResponse{Token: token, User: *domain.ReconstituteUser(userData)}, nil
}

func (uc *UserUseCase) GetAllUsers() ([]domain.User, error) {
	return uc.userRepo.GetAll()
}

func (uc *UserUseCase) UpdateUserRole(currentUserID, targetUserID uint, newRole, ipAddress string) error {
	if currentUserID == targetUserID {
		return domain.ErrCannotChangeOwnRole
	}
	if targetUserID == 1 {
		return domain.ErrCannotChangeRootAdmin
	}
	switch newRole {
	case "admin", "editor", "author":
	default:
		return domain.ErrInvalidRole
	}
	role := domain.UserRole(newRole)
	if role != domain.UserRoleAdmin && role != domain.UserRoleEditor && role != domain.UserRoleAuthor {
		return domain.ErrInvalidRole
	}
	if err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		target, err := repositories.Users.GetByID(targetUserID)
		if err != nil {
			return errors.New("không tìm thấy người dùng")
		}
		if target.Data().ID == 1 {
			return domain.ErrCannotChangeRootAdmin
		}
		if err := target.ChangeRole(role); err != nil {
			return err
		}
		if err := repositories.Users.UpdateRole(targetUserID, string(target.Data().Role)); err != nil {
			return err
		}
		return repositories.Audits.RecordLog(currentUserID, "UPDATE_USER_ROLE", fmt.Sprintf("User #%d", targetUserID), fmt.Sprintf("Đổi vai trò thành '%s'", target.Data().Role), ipAddress)
	}); err != nil {
		return err
	}
	return nil
}
